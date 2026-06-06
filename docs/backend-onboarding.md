# Введение в бэкенд: Go для разработчика на .NET / EF Core

> Аудитория: ты — опытный разработчик на C# / ASP.NET Core / EF Core, новичок в Go, который владеет этим проектом и хочет *изучить* бэкенд, а не просто получить его пересказ. Всё, что описано ниже, привязано к уже знакомой тебе концепции из .NET, после чего выделяется отличие Go и его *причина*. Каждое утверждение ссылается на `file:line` в этом репозитории. Все пути к файлам — внутри `backend/`.

---

## 0. Что представляет собой этот код (в одну строку)

HTTP API на Go (**Gin** для маршрутизации, **GORM** поверх **Postgres** для работы с данными, бот **Telegram** для уведомлений), реализующий логистическую CRM, — те же слои, что ты знаешь по ASP.NET Core (`handler → service → repository → domain`), но соединённые руками, **без DI-контейнера**, с **ошибками в виде возвращаемых значений вместо исключений** и **интерфейсами, объявляемыми потребителем, а не заранее**.

Весь граф объектов строится единожды, сверху вниз, в `cmd/api/main.go:20-96`. Если читать только один файл первым — читай этот.

---

## 1. Архитектура и слои — composition root

### Картина

```
HTTP request
  → internal/http/router.go          (Gin routes + middleware groups)    ~ ASP.NET endpoint routing
  → internal/http/handlers/*.go      (bind JSON, map errors → status)     ~ Controller actions
  → internal/service/*.go            (business logic, validation)         ~ application/service layer
  → internal/repository/*.go         (GORM calls)                         ~ EF Core DbContext repo
  → internal/domain/*.go             (entities + enums + errors)          ~ EF entities / POCOs
Postgres (gorm.io/gorm)
```

### `cmd/api/main.go` — ЭТО твой `Program.cs` + `Startup.cs`, собранный руками

В .NET эта логика разнесена между `Program.cs` (хост, Kestrel, конфигурация) и `Startup.ConfigureServices` (регистрации через `AddScoped<…>`). Здесь это одна функция `main()`. Прочитай её сверху вниз — и у тебя есть вся история сборки целиком.

- **Загрузка конфигурации** — `cfg := config.Load()` (`main.go:24`). `config.Load()` читает переменные окружения с дефолтами и возвращает обычную структуру `Config` — как привязка `IConfiguration` к options-POCO, но это явный вызов функции, а не конвейер configuration-провайдеров.
- **Открытие БД** — `gdb, err := db.Connect(ctx, cfg.DatabaseURL)` (`main.go:29`). `gdb` имеет тип `*gorm.DB` — твой единственный общий хэндл к пулу соединений (подробнее в §2). Обрати внимание на **ручную проверку `err`** прямо на следующей строке (`main.go:30-33`): в Go нет исключений, поэтому каждый вызов, способный завершиться неудачей, возвращает `(value, error)`, и ты проверяешь `err` немедленно. `os.Exit(1)` = аварийное завершение при фатальной ошибке старта.
- **`defer` = твой `using` / `finally`** — `defer func(){ … sqlDB.Close() }()` (`main.go:34-38`) планирует очистку на момент возврата из `main`. `_ =` намеренно отбрасывает ошибку, которую Go иначе заставил бы обработать.
- **Создание репозиториев** (`main.go:40-44`), каждому передаётся один и тот же `gdb`. `repository.NewLeadRepository(gdb)` (`repository/lead.go:17-19`) возвращает `*LeadRepository`, хранящий хэндл к БД. `New…` — это **просто соглашение об именовании**: в Go нет конструкторов и нет ключевого слова `new` для этого.
- **Создание бота** (`main.go:46`) *перед* сервисами, потому что сервисы зависят от него как от нотификатора.
- **Создание сервисов** (`main.go:52-56`), репозитории (и бот) внедряются руками. Это твой `AddScoped<LeadService>()`, расписанный явными вызовами.

```go
leadService := service.NewLeadService(leadRepo, notifier)
shipmentService := service.NewShipmentService(shipmentRepo, clientRepo, notifier)
messageService := service.NewMessageService(messageRepo, shipmentRepo, clientRepo, notifier, notifier)
```

> Посмотри на последнюю строку: одно и то же значение `notifier` передаётся в **два параметра** (`main.go:56`). Это не баг копипаста — у этих двух параметров два *разных* интерфейсных типа, и один `*bot.Bot` удовлетворяет обоим. В этом суть §3.

- **Сборка роутера** через передачу всех сервисов в структуре `RouterDeps` (`main.go:61-69`), затем он отдаётся стандартному `http.Server` (`main.go:71-75`).

**Замечание о времени жизни (важный сдвиг относительно .NET):** здесь **нет scoped/transient/singleton-времён жизни**. Каждый репозиторий и сервис создаётся **единожды** и живёт весь процесс — фактически все они синглтоны. Состояние на запрос (id аутентифицированного пользователя, отмена) *не* живёт в этих объектах; оно течёт через аргументы `context.Context` в каждом вызове метода (см. §5). В .NET ты опирался бы на scoped DI + `DbContext` на запрос; здесь `*gorm.DB` общий, а scope запроса несётся в `ctx`.

**Конкурентность, которой ты не увидишь в ASP.NET (полностью раскрыта в §6):**
- `go notifier.Run(…)` (`main.go:59`) запускает Telegram-цикл long-poll в **горутине**.
- `go func(){ server.ListenAndServe() … }()` (`main.go:77-83`) запускает HTTP-сервер в другой горутине.
- `<-ctx.Done()` (`main.go:85`) **блокирует** до тех пор, пока SIGINT/SIGTERM не отменит корневой контекст (созданный в `main.go:21` через `signal.NotifyContext`), после чего `server.Shutdown(shutdownCtx)` (`main.go:90`) дренирует запросы в полёте с таймаутом 10 с. Это идиома graceful-shutdown в Go — аналог `IHostApplicationLifetime.ApplicationStopping` + `WaitForShutdownAsync`.

### Пакеты, правило `internal/` и контроль доступа

- **Модуль** = `icaris-logistic/backend` (`go.mod:1`). Грубо — твой solution + корневое пространство имён. Блок `require` — это твой список `<PackageReference>`.
- **Пакет = одна директория.** Каждый файл `.go` в `internal/service/` начинается с `package service`. Файлы одного пакета видят неэкспортируемые идентификаторы друг друга **без `import`** — здесь нет пофайлового пространства имён, как в C#. Ближе всего к `namespace` из C#, но одна директория = один пакет, это жёсткое правило.
- **Регистр букв — ЭТО модификатор доступа.** Нет ключевых слов `public`/`private`/`internal`. `LeadService` (с заглавной) экспортируется = `public`; `validateCreateLead` (`service/lead.go:122`, строчная) неэкспортируемый = package-private. Переименование `notifyNewLead` → `NotifyNewLead` молча сделало бы его частью публичного API — нет ключевого слова, которое можно было бы найти грепом.
- **`internal/` — это файрвол, навязываемый компилятором.** Всё, что лежит под `backend/internal/…`, импортируется только кодом с корнем в `backend/`. Поэтому весь реальный код живёт в `internal/`, и только `cmd/api`, `cmd/migrate`, `cmd/createmanager` находятся снаружи как запускаемые точки входа. Ближайшая аналогия из .NET: видимость `internal` + `[InternalsVisibleTo]`, ограниченные твоей собственной сборкой, — но навязанные раскладкой директорий.
- **Импорты — это пути, а не ссылки на сборки.** Алиас `apphttp "icaris-logistic/backend/internal/http"` (`main.go:15`) локально переименовывает пакет, потому что его настоящее имя `http` столкнулось бы со стандартным `net/http` (`main.go:6`). Это `using apphttp = …`.

### Раскладка `cmd/`

`backend/cmd/` содержит три программы `package main`, каждая компилируется в один бинарник: `cmd/api` (этот HTTP-сервер), `cmd/migrate` (запускает SQL-миграции), `cmd/createmanager` (создаёт учётку менеджера). Соглашение Go: **`cmd/<name>/main.go` на каждый исполняемый файл**, всё переиспользуемое — под `internal/`. Аналогия из .NET: один solution с несколькими консольными проектами, ссылающимися на общую библиотеку классов.

---

## 2. Слой данных — GORM и golang-migrate для ветерана EF

**В одну строку:** GORM — это «EF Core для Go» поверх Postgres, но схема **не** генерируется из структур: миграции — это **написанный руками SQL**, запускаемый `golang-migrate`, а структуры синхронизируются со схемой вручную.

### Модели: struct-теги, а не Fluent API или DataAnnotations

У GORM есть лишь один механизм маппинга — **struct-теги** (строковые метаданные, разбираемые через рефлексию). Нигде в этом проекте **нет `OnModelCreating`**. `domain/lead.go:12-24`:

```go
type Lead struct {
    ID        uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name      string              `gorm:"type:varchar(255);not null" json:"name"`
    Weight    decimal.NullDecimal `gorm:"type:numeric" json:"weight"`
    Status    LeadStatus          `gorm:"type:varchar(20);not null;default:new" json:"status"`
    CreatedAt time.Time           `gorm:"not null;default:now()" json:"created_at"`
}
```

Строка в обратных кавычках несёт два пространства имён: `gorm:"…"` (маппинг колонок — твои `[Column]`/`HasColumnType`/`IsRequired`/`HasDefaultValueSql`) и `json:"…"` (сериализация по проводу — твой `[JsonPropertyName]`). Одно поле — и ORM-маппинг, и HTTP-контракт.

| GORM-тег | Эквивалент в EF Core |
|---|---|
| `primaryKey` | `[Key]` / `HasKey` |
| `type:uuid` | `.HasColumnType("uuid")` |
| `not null` | `.IsRequired()` |
| `default:gen_random_uuid()` | `.HasDefaultValueSql(...)` |
| `uniqueIndex` | `.HasIndex(...).IsUnique()` |
| `json:"-"` | `[JsonIgnore]` |

Две вещи, которые нужно сразу зафиксировать:
1. **`json:"-"` критичен для безопасности.** `Manager.Password` (`domain/manager.go:14`) помечен `json:"-"`, поэтому bcrypt-хэш никогда не сериализуется, хотя структуры возвращаются из хэндлеров напрямую. Здесь нет отдельного слоя DTO/view-model, который бы его охранял, — этот один тег и есть охрана.
2. **Поле → колонка по соглашению — snake_case.** `FromCity` → `from_city`, `TelegramID` → `telegram_id`, `ID` → `id`. Тегов `column:` ты не увидишь, потому что соглашение уже даёт нужное имя.

### Нет `DbSet<T>`, нет реестра моделей, нет change tracker

- **Нет `DbSet`** и нет центрального реестра сущностей. Структура становится «моделью» в тот момент, когда ты передаёшь её в запрос: `r.db.Create(lead)` (`repository/lead.go:24`) прямо там сообщает GORM, что таблица для `*domain.Lead` — это `leads` (имя типа во множественном числе).
- `*gorm.DB` (создаётся единожды в `db/postgres.go:14-42`, `gorm.Open(postgres.Open(url), …)` на строке 25) — это **единственный, долгоживущий, потокобезопасный хэндл, оборачивающий пул соединений**, общий для каждого репозитория. Сравни с `DbContext` из EF, который создаётся на запрос, scoped и не потокобезопасен. Привязка к запросу делается через `.WithContext(ctx)` на каждом вызове, а не созданием нового объекта контекста.
- **Нет change tracker и нет `SaveChanges()`.** Каждая запись явная и немедленная.

### Запросы: цепочки методов, а не LINQ

Нет `IQueryable`, нет деревьев выражений. Запросы — это цепочки методов на `*gorm.DB`, которые выполняются на *финишере* (`Find`, `First`, `Create`, `Count`). Условие WHERE — это **параметризованный фрагмент SQL-строки**, а не лямбда. EF против GORM:

```csharp
db.Shipments.Where(s => s.ClientId == clientId).OrderByDescending(s => s.CreatedAt).ToListAsync();
```
```go
// repository/shipment.go:49-54
r.db.WithContext(ctx).Where("client_id = ?", clientID).Order("created_at desc").Find(&shipments).Error
```

Три вещи, которые кусают LINQ-пользователя:
1. **Результат — это out-параметр.** `Find(&shipments)` принимает указатель на слайс и заполняет его; метод возвращает `*gorm.DB`, а ошибка живёт на `.Error`. Поэтому каждая цепочка заканчивается на `.Find(...).Error`. Список — это та переменная, которую ты передал (`var shipments []domain.Shipment`).
2. **`Where("client_id = ?", clientID)` — это сырая параметризованная строка.** `?` — это настоящий привязанный параметр (защита от инъекций). Но имя колонки набрано руками — опечатка будет ошибкой **во время выполнения**, а не при сборке. Ты теряешь рефакторинг-безопасность `s => s.ClientId`.
3. **`First` против `Find`.** `First` (`repository/shipment.go:64`) **возвращает ошибку `gorm.ErrRecordNotFound`**, если ничего не нашлось. `Find` возвращает пустой слайс и *никакой* ошибки. Эта асимметрия порождает два разных механизма «не найдено» ниже.

### «Не найдено — это ошибка» → транслируется в `domain.ErrNotFound`

В EF `FirstOrDefaultAsync` возвращает `null`. В GORM `First` возвращает типизированную ошибку. Каждый репозиторий на чтение делает эту трансляцию (`repository/lead.go:36-47`, `repository/shipment.go:62-73`):

```go
err := r.db.WithContext(ctx).First(&lead, "id = ?", id).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, domain.ErrNotFound   // re-map so the service never imports GORM
}
if err != nil { return nil, err }
return &lead, nil
```

Репозиторий намеренно перемаппит ошибку GORM в собственную ошибку проекта `domain.ErrNotFound` (`domain/errors.go:8`) — чистая граница ports-and-adapters, ровно как возвращать доменный `Result`/`null` вместо протечки `DbException`. *Почему* — задокументировано в `repository/manager.go:24-25`.

**Два разных способа определить «не найдено» — нужно знать, какой применять:**
- **Чтения:** `First` возвращает ошибку → проверяй `errors.Is(err, gorm.ErrRecordNotFound)`.
- **Записи:** `Update` **не** возвращает ошибку при нуле совпадений, поэтому репозиторий проверяет `result.RowsAffected == 0` и синтезирует `domain.ErrNotFound` (`repository/lead.go:51-63`). Неправильная проверка молча пропускает этот случай.

### Связи: намеренно НЕТ навигационных свойств / нет `Include`

GORM поддерживает `Preload("Orders")` (твой `Include`), но этот код **его не использует**. `Shipment.ClientID` / `ManagerID` (`domain/shipment.go:14-15`) — это голые поля `uuid.UUID` с тегами `index`; **нет навигационного свойства `Client Client`**. Связи по внешним ключам целиком живут в SQL миграций (`client_id uuid not null references clients (id)` в `migrations/000004_create_shipments.up.sql`), и никогда не выводятся из структур. Когда сервису нужен связанный клиент, он делает **второй явный вызов репозитория по id**. Поведение `ON DELETE` (`restrict`/`set null`/`cascade`) живёт только в SQL — из Go-модели его не видно.

Ментальная модель: **структуры описывают колонки; миграции описывают связи и ограничения. Два отдельных источника истины, синхронизируемых человеком.**

### Транзакции: API на замыканиях (чище, чем `BeginTransaction` в EF)

Ты передаёшь функцию; вернуть `error` → откат, вернуть `nil` → коммит. Никакого `using`, никакого ручного `CommitAsync`. Канонический пример атомарно записывает отправление + его начальное событие статуса (`repository/shipment.go:23-38`):

```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(shipment).Error; err != nil { return err }
    event := &domain.ShipmentStatusEvent{ ShipmentID: shipment.ID, Status: shipment.Status, ChangedBy: &shipment.ManagerID }
    return tx.Create(event).Error
})
```

Три несущих детали: (1) **используй `tx`, а не `r.db`, для каждой записи внутри** — использование `r.db` вышло бы за пределы транзакции; (2) `shipment.ID` заполняется *первым* `Create` (GORM считывает сгенерированный БД UUID обратно через `RETURNING`), поэтому строка 30 может ссылаться на него для FK события; (3) `ChangedBy: &shipment.ManagerID` использует `&`, потому что поле — это nullable-указатель (§4).

Подводный камень GORM, который стоит отметить: `UpdateStatus` (`repository/shipment.go:78-131`) обновляет через `map[string]any` + `gorm.Expr("now()")`, и **обновление через map НЕ записывает значения обратно в твою структуру**, поэтому приходится повторно делать `First` строки (строка 124), чтобы вернуть свежие данные. Change tracker из EF обновил бы сущность за тебя.

### Перечисления: VARCHAR + CHECK + типизированные строковые константы Go (тройка, координируемая руками)

Здесь **нет настоящего enum** ни в Go, *ни* в Postgres. Паттерн (на примере `LeadStatus`, `domain/lead.go:26-44`; то же для `ShipmentStatus`, `domain/shipment.go:31-55`):

1. **Go:** `type LeadStatus string` + блок `const (…)` с типизированными строковыми значениями. Отдельный тип, чтобы случайно нельзя было передать голую строку, но под капотом значение — это строка, хранимая в `varchar(20)`.
2. **БД:** `status varchar(20) … check (status in ('new','contacted',…))` — ограничение CHECK навязывает набор значений.
3. **Мост:** метод `IsValid()` на типе (`func (s LeadStatus) IsValid() bool`), вызываемый сервисом перед записью, чтобы плохое значение возвращало **400** вместо **500** от нарушения CHECK.

**Критичный факт сопровождения:** эти три списка правятся **вместе, руками**. Добавить статус = отредактировать константы Go, switch в `IsValid()` И написать новую миграцию, изменяющую CHECK. Ничто не генерирует одно из другого.

### Миграции: написанный руками парный SQL — намеренная противоположность EF code-first

EF code-first: меняешь модель, `dotnet ef migrations add`, EF делает diff и генерирует `Up()`/`Down()`. **В этом проекте — написанный руками сырой SQL в парных файлах** (`NNNNNN_name.up.sql` + `.down.sql` в `migrations/`). Здесь **нет diff’а, нет генерации и нет `AutoMigrate`** (поищи грепом по бэкенду — он не встречается нигде). Структуры **не** создают и не изменяют таблицы; схема существует только потому, что кто-то написал SQL. Обе половины ты пишешь сам.

Раннер — `cmd/migrate/main.go`: `migrate up` → `m.Up()` (применить все ожидающие); `migrate down` → `m.Steps(-1)` (откатить **ровно одну** — безопаснее, чем `database update 0` в EF); применённые версии отслеживаются в `schema_migrations`, с **флагом `dirty`** для частичных сбоев, которые нужно разруливать вручную. Реальная дисциплина эволюции схемы видна в миграциях 6–8 (например, `000007` — это **частичный уникальный индекс** `… where lead_id is not null`, кодирующий «один lead → не более одного клиента»).

**Где живёт каждая концепция:** модели/теги → `internal/domain/*.go`; «DbContext» → `internal/db/postgres.go`; запросы/транзакции/трансляция ошибок → `internal/repository/*.go`; схема/FK/CHECK → `migrations/*.sql`; раннер → `cmd/migrate/main.go`.

---

## 3. Идиома, которая ощущается «наизнанку»: интерфейсы объявляются ПОТРЕБИТЕЛЕМ

Это важнейшая идиома Go, которую нужно усвоить, потому что она **противоположна** привычке из C#. Она встречается повсюду, поэтому ей отведён отдельный раздел.

**Привычка из C#:** определить `ILeadRepository` рядом с репозиторием, написать `LeadRepository : ILeadRepository`, зарегистрировать маппинг в контейнере. Интерфейс объявляется заранее, и реализация *явно именует* его.

**Что делает Go здесь — две инверсии одновременно:**
1. **Интерфейсы удовлетворяются структурно (duck typing).** Тип реализует интерфейс, *имея подходящие методы*, — синтаксиса `: LeadStore` нигде нет.
2. **Интерфейс объявляется там, где он *потребляется*.** **Сервис** объявляет тот маленький интерфейс, который ему нужен.

`service/lead.go:20-25`:
```go
type LeadStore interface {
    Create(ctx context.Context, lead *domain.Lead) error
    List(ctx context.Context) ([]domain.Lead, error)
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) error
}
```

`*repository.LeadRepository` (`repository/lead.go:13`) **никогда не упоминает `LeadStore`** — но поскольку его методы совпадают, он *им является*. Компилятор проверяет это в точке присваивания в `main.go:52`, где `leadRepo` передаётся туда, где ожидается `LeadStore`. Чтобы найти, кто реализует интерфейс, ты идёшь по сигнатурам методов, а не по объявлению (команда «Implementations» в твоей IDE — твой друг).

**Почему это мощно — конкретно в этом репозитории:** одно физическое значение `*bot.Bot` потребляется через *четыре разных узких* интерфейса, каждый определён своим сервисом:
- `service.Notifier` — `NotifyNewLead` (`service/notifier.go:11`)
- `service.ClientNotifier` — `NotifyShipmentStatus` (`service/shipment.go:45`)
- `service.ManagerNotifier` — `NotifyClientMessage` (`service/message.go:26`)
- `service.ClientMessageNotifier` — `NotifyManagerReply` (`service/message.go:31`)

`*bot.Bot` реализует их все как обычные методы (`bot/bot.go:210,215,220,225`) и **никогда не импортирует `package service`** — это подтверждается комментарием в `bot/bot.go:18-20`: связь существует «только по контракту методов». Поэтому `main.go:56` пишет `…, notifier, notifier`: один бот удовлетворяет и `ManagerNotifier`, и `ClientMessageNotifier`. В C# у тебя был бы `Bot : INotifier, IClientNotifier, …` с регистрацией каждого маппинга; здесь бот в блаженном неведении о существовании этих интерфейсов.

**Направление зависимостей:** стрелки указывают *внутрь*. `service` определяет контракты; `repository` и `bot` — внешние реализации, которые их удовлетворяют. Компилятор следит за тем, чтобы `service` импортировал только `domain` (не `repository` и не `bot`). Ты получаешь выгоду Dependency Inversion от `ILeadRepository` **без шага регистрации**, и интерфейсы остаются крошечными, потому что каждый потребитель вырезает ровно те методы, которые вызывает (например, `ShipmentReader` в `service/message.go:21` — это «срез» из одного метода). Широкий `IRepository<T>` идиоматичен для C#; маленькие интерфейсы на стороне потребителя идиоматичны для Go.

**Подводный камень:** переименуй метод на `LeadRepository`, и он молча перестанет удовлетворять `LeadStore` — сбой всплывёт на строке связывания `main.go:52`, а не в репозитории.

---

## 4. Ошибки, null, нулевые значения, ресиверы, встраивание — ежедневные идиомы

### Ошибки — это значения, а не исключения (самый большой сдвиг)

Каждая функция, способная завершиться неудачей, возвращает `(value, error)`; ты проверяешь `err` перед использованием значения. Никаких `throw`/`catch`. `service/auth.go:46-49` типичен:

```go
manager, err := s.managers.GetByEmail(ctx, email)
if err != nil && !errors.Is(err, domain.ErrNotFound) {
    return "", err
}
```

`""` — это заглушка «нет полезного значения», которую ты возвращаешь рядом с ошибкой. **Забыть проверить `err` компилируется нормально** и молча продолжает работу с нулевым значением — нет брошенного исключения, которое можно поймать.

- **Оборачивание через `%w`** — это твой `new XException(msg, inner)`. `fmt.Errorf("%w: name is required", ErrValidation)` (`service/lead.go:124`) оборачивает sentinel, чтобы вызывающие всё ещё могли его распознать. `%v` сплющил бы его в текст и *разорвал* цепочку — `service/client.go:107` намеренно использует `%w` для sentinel и `%v` для нижележащей причины (`fmt.Errorf("%w: %v", ErrTelegramAuth, err)`), чтобы telegram-ошибку нельзя было развернуть и слить.
- **Sentinel-ошибки + `errors.Is`** — это твой `catch (XException)`. `domain.ErrNotFound` (`domain/errors.go:8`) и `service.ErrValidation` (`service/lead.go:16`) — это синглтоны уровня пакета; `errors.Is` идёт по цепочке `%w`, чтобы их сопоставить.
- **Хэндлер — это твой `ExceptionFilter`, но явный.** `handlers/lead.go:90-99` маппит sentinel’ы на коды статуса через `switch`:

```go
switch {
case errors.Is(err, service.ErrValidation):
    respondError(c, http.StatusBadRequest, "validation", err.Error())
case errors.Is(err, domain.ErrNotFound):
    respondError(c, http.StatusNotFound, "not_found", "lead not found")
default:
    respondError(c, http.StatusInternalServerError, "internal", "internal server error")
}
```

Форма ошибки `{ error: { code, message } }` (`handlers/response.go`) и правило «никогда не сливать внутренности на 500» соответствуют твоему `error-handling.md`. Выгода от ошибки-как-значения: пути сбоя видимы и локальны, а не невидимая раскрутка стека на три слоя выше. Цена: многословность.

### Nullability через указатели (и `NullDecimal`)

В Go нет `Nullable<T>`. Для значимого типа, который может отсутствовать, используется **указатель**; `nil` означает «отсутствует»:
- `LeadID *uuid.UUID` (`domain/client.go:17`) ≈ `Guid?`
- `EstimatedAt *time.Time` / `DeliveredAt *time.Time` (`domain/shipment.go:25-26`) ≈ `DateTime?`

Создаёшь через `&` (например, `ChangedBy: &shipment.ManagerID`, `repository/shipment.go:33`) и **обязан проверять перед разыменованием** (`if current.DeliveredAt == nil`, `repository/shipment.go:104`). Разыменование nil-указателя — это паника во время выполнения — твой `NullReferenceException` — и **компилятор Go НЕ предупреждает тебя** так, как анализ nullable-ссылок в C#.

**Два представления null, не взаимозаменяемые:** указатели для FK/таймстампов; `decimal.NullDecimal` (структура `{ Valid bool; Decimal }`, `domain/shipment.go:19-23`) для денег/измерений, читается через `.Valid`, затем `.Decimal` (`bot/format.go:40`). `NullDecimal` — это форма `Nullable<decimal>`. Деньги никогда не используют `float64` — `shopspring/decimal` — это аналог `System.Decimal`, выбранный намеренно, чтобы избежать денежных багов двоичной плавающей точки.

### Нулевые значения против null

У каждого типа есть **нулевое значение** «бесплатно» — нет состояния «не инициализировано». `string`→`""`, числа→`0`, `bool`→`false`, указатели/интерфейсы/слайсы/мапы→`nil`, структуры→все поля занулены. `var leads []domain.Lead` сразу же — пригодный (nil) слайс.

Это **используется как логика**: `if input.ClientID == uuid.Nil` (`service/shipment.go:80`) трактует UUID из одних нулей как «отсутствует»; `telegramID == 0` и `currency == ""` означают «отсутствует/по умолчанию» в других местах. Острый край: **отсутствующее JSON-поле молча десериализуется в нулевое значение** — нет автоматической ошибки «поле было обязательным», поэтому валидация обязательных полей — ручная. Чтобы восстановить различие «присутствует/отсутствует», код возвращает `(uuid.UUID, bool)` — например, `middleware.ManagerID` (`auth.go:60-63`), «защита от тихого uuid.Nil». Этот булев `, ok` и есть идиома.

### Ресиверы по значению против ресиверов-указателей

Каждый метод объявляет, получает ли он *копию* (ресивер по значению) или *указатель*:
- **Ресивер-указатель** — `func (s *LeadService) Create(...)` (`service/lead.go:49`), `func (r *LeadRepository) Create(...)` (`repository/lead.go:23`). Общий экземпляр, может мутировать поля. Долгоживущие зависимости всегда имеют ресиверы-указатели.
- **Ресивер по значению** — `func (i CreateLeadInput) normalized()` (`service/lead.go:112`), `func (s LeadStatus) IsValid()` (`domain/lead.go:37`), `func (h LeadHandler) Create(...)` (`handlers/lead.go:22`). Получает **копию**.

Поучительный случай: `normalized()` мутирует свои поля, но, будучи ресивером по значению, мутирует только **копию** — поэтому он возвращает копию, и вызывающий **обязан переприсвоить**: `input = input.normalized()` (`service/lead.go:50`). Забудь переприсваивание — и изменение молча отбрасывается, без ошибки компиляции. Структуры Go — значимые типы (как `struct` в C#, а не `class`); единственное, что похоже на ссылки, — это указатели/слайсы/мапы.

### Нет наследования — композиция / встраивание

В Go **нет базовых классов и нет наследования**. Переиспользование — через **встраивание**: помещение одной структуры внутрь другой без имени поля, что продвигает её поля/методы. `token/token.go:16-19`:

```go
type Claims struct {
    jwt.RegisteredClaims        // embedded — no field name
    Role domain.Role `json:"role"`
}
```

`claims.Subject` / `claims.ExpiresAt` работают как унаследованные — но `Claims` **не** является «is-a» `RegisteredClaims` (нет подтипизации, нет virtual/override). `*Claims` нельзя передать туда, где ожидается `*RegisteredClaims`. Полиморфизм идёт **только** от интерфейсов (§3), никогда от иерархии классов. Везде в остальном это обычная композиция через именованные поля (`ShipmentService` держит `store`, `clients`, `notifier` как поля — никакого `BaseService`).

### Сторонние типы для UUID и decimal

В стандартной библиотеке Go **нет ни UUID, ни decimal-типа**. Оба приходят из модулей, явно импортируемых в каждом файле, который их касается (`domain/lead.go:6-7`): `github.com/google/uuid`, `github.com/shopspring/decimal`. Следствие: это библиотечные типы с **методами, а не операторами** — нет `+` для decimal (вызывай `.Add(...)`), сравнивай UUID через `== uuid.Nil` (а не `Guid.Empty`). Плюс: корректность. Цена: каждый потребляющий файл импортирует пакет, и ты теряешь синтаксический сахар операторов.

---

## 5. Прослеженный путь: `POST /api/leads` (форма с лендинга → уведомление менеджеру в Telegram)

Это самый удачный запрос для изучения бэкенда — он касается каждого слоя ровно один раз, плюс асинхронный Telegram-сайд-канал.

**Прыжок 1 — маршрут.** `router.go:41`: `api.POST("/leads", lead.Create)`. `gin.New()` (`router.go:25`) ≈ `WebApplication...Build()`; `router.Use(gin.Recovery())` (`router.go:26`) — это ловец паник (ближайшее в Go к middleware для необработанных исключений) — но он защищает только код в горутине HTTP-запроса, не в горутине бота. `router.Group("/api")` ≈ `MapGroup`. Маршрут lead регистрируется **до** группы аутентификации `manager` (`router.go:46-47`), поэтому он **публичный** — без JWT. `lead.Create` — это *значение метода*, привязанное к хэндлеру (как передача `RequestDelegate`).

**Прыжок 2 — хэндлер (`handlers/lead.go:22-40`).**
- `var input service.CreateLeadInput` имеет **нулевое значение, а не null** — самый большой ментальный сдвиг.
- `c.ShouldBindJSON(&input)` (`lead.go:24`) — это model binding `[FromBody]`, управляемый тегами `json:"…"` (`service/lead.go:38-47`). **`&` несущий**: без него Gin заполняет копию, а твой `input` остаётся нулевым — без ошибки компиляции, молча пустые данные.
- Привязка возвращает `error` (а не throw). Воронка ошибок ниже (`lead.go:30-37`) — это типизированная обработка исключений, сделанная через `errors.Is`: `ErrValidation` → 400, всё остальное → общий 500 (внутренности не сливаются).
- `c.JSON(http.StatusCreated, lead)` (`lead.go:39`) ≈ `Results.Created(...)`, сериализует через теги `json` структуры.

**Прыжок 3 — сервис (`service/lead.go:49-73`).** `input = input.normalized()` (копия ресивера по значению, §4), затем `validateCreateLead` (guard-условия, возвращающие `fmt.Errorf("%w: …", ErrValidation)`). Строит `lead := &domain.Lead{…}` как **композитный литерал, у которого берётся адрес** (`new Lead { … }` + `&`, чтобы репозиторий мог его мутировать). Он намеренно **не** задаёт `ID`/`Status`/`CreatedAt` — оставляет нулевыми, чтобы Postgres заполнил дефолты колонок. Поле `store` — это интерфейс `LeadStore` (§3).

**Прыжок 4 — репозиторий (`repository/lead.go:23-25`).** `r.db.WithContext(ctx).Create(lead).Error`. GORM ≈ `Add` + `SaveChanges`, слитые воедино, привязанные к отмене запроса через `WithContext(ctx)`. **Нет change tracker, нет `SaveChanges`** — INSERT выполняется немедленно и синхронно. GORM мутирует `lead` на месте, дозаполняя `ID`/`status`/`created_at` (вот почему мы передали указатель; комментарий в `repository/lead.go:21-22`), поэтому ответ 201 несёт новый UUID. Дефолтами владеет БД (`migrations/000002_create_leads.up.sql`).

**Прыжок 5 — уведомление по принципу fire-and-forget (`service/lead.go:101-110`).** После вставки `s.notifyNewLead(lead)` порождает `go func(){ … }()` — **горутину** (§6). Три тонкости: она использует `context.Background()` (НЕ контекст запроса), потому что контекст запроса отменяется в момент записи 201 (комментарий в `lead.go:96-98`); ошибки идут **только в лог** через `slog.Error` (упавшее уведомление не должно завалить запрос — lead уже сохранён); и проверка `if s.notifier == nil` важна, потому что интерфейсы могут быть nil.

**Прыжок 6 — 201 возвращается ДО того, как предпринимается попытка Telegram.** `handlers/lead.go:39` выполняется, пока горутина ещё в полёте. 201 означает «строка вставлена», НЕ «менеджер уведомлён».

**Прыжок 7 — асинхронный хвост (`bot/bot.go:210` → `bot/format.go:32`).** `NotifyNewLead` вызывает `b.send(b.chatID, formatLeadMessage(lead))`. `formatLeadMessage` — чистая функция, использующая `strings.Builder` (твой `StringBuilder`), читающая nullable-decimal через `if lead.Weight.Valid { … }` (`format.go:40`). `b.send` (`bot/bot.go:231`) возвращается рано в режиме без токена, а при сбое возвращает общий `fmt.Errorf("bot: send message failed")` **без** оборачивания ошибки библиотеки — потому что эта строка может содержать токен бота в URL (намеренная редакция секрета, соответствует `security.md`).

**Весь путь целиком:** `POST /api/leads` (`router.go:41`) → bind+воронка (`handlers/lead.go:22`) → нормализация/валидация/сборка (`service/lead.go:49`) → синхронный INSERT, БД заполняет дефолты (`repository/lead.go:23`) → `go func()` с `context.Background()` (`service/lead.go:101`) → **201 возвращён без ожидания** → параллельно `Bot.NotifyNewLead → formatLeadMessage → b.send → Telegram`.

---

## 6. Прослеженный путь: вход через Telegram Mini App (HMAC initData → JWT → RequireClientAuth)

**В одно предложение:** WebApp делает POST подписанного Telegram’ом `init_data` на `POST /api/app/auth/telegram`; бэкенд заново вычисляет HMAC и **сравнивает за константное время**, делает upsert `Client` по `telegram_id`, выпускает HS256 JWT с `role=client`, и каждый последующий запрос `/api/app/*` пропускается через `RequireClientAuth`, который парсит JWT и фиксирует `clientID` вызывающего в контексте запроса.

**Шаг 0 — маршруты (`router.go:43, 64-71`).** Вход **публичный**; клиентская поверхность — это `api.Group("/app")` + `app.Use(middleware.RequireClientAuth(deps.JWTSecret))`. Это `MapGroup("/app").RequireAuthorization("ClientOnly")` — но скрытого конвейера нет: `app.Use(...)` буквально дописывает функцию в слайс, который Gin вызывает по порядку. Заметь: у `gin.New()` **нет** default-middleware; Recovery и CORS добавляются руками (`router.go:25-27`).

**Шаг 1 — хэндлер (`handlers/client_auth.go:25-43`).** Привязывает JSON (`init_data` через `json:"init_data"`), делегирует `AuthenticateWebApp` (который возвращает **три** значения `(string, *domain.Client, error)` — совершенно нормально в Go) и маппит `service.ErrTelegramAuth` → **401**, всё остальное → 500.

**Шаг 2 — сервис (`service/client.go:103-121`).** Три обязанности по порядку: проверить подпись (`telegram.ValidateInitData`), сделать upsert клиента (`Register`), подписать токен (`token.Issue`). Оборачивает сбои через `%w`, чтобы `errors.Is` хэндлера сопоставлял через границу пакета. Зависимости внедрены конструктором единожды в `main.go:54`; репозиторий держит один долгоживущий `*gorm.DB`, привязанный к запросу на каждом вызове через `.WithContext(ctx)`.

**Шаг 3 — валидация HMAC (`telegram/initdata.go:95-121`), ядро безопасности.** Ни один фреймворк .NET не делает это за тебя — это задокументированная схема Telegram, реализованная руками:
1. Построить `data_check_string`: все ключи, кроме `hash`, **отсортированные**, склеенные как строки `key=value` через `\n`.
2. Вывести секрет: `secret_key = HMAC_SHA256(key="WebAppData", msg=botToken)` — константа является ключом, токен является сообщением (контринтуитивно, но по спецификации).
3. Подписать check-строку этим выведенным ключом, закодировать в hex.
4. **Сравнение за константное время:** `hmac.Equal(computed, hash)` (`initdata.go:120`). **Не заменяй на `==`** — обычное сравнение прерывается на первом отличающемся байте и сливает тайминг, что позволяет побайтовую подделку. Эквивалент в .NET — `CryptographicOperations.FixedTimeEquals`.

Плюс отклонение пустого токена (fail closed, `initdata.go:50-52`) и 24-часовое окно защиты от повтора через `auth_date` (`initDataMaxAge`, `client.go:23`). JSON пользователя парсится **только после** того, как подпись прошла, поэтому непроверенным полям никогда не доверяют.

**Шаг 4 — `Register`: идемпотентный upsert (`service/client.go:48-99`).** Поиск по `telegram_id`; если найден — вернуть его; если `domain.ErrNotFound` — создать. **Конкурентность, которую ты редко писал бы руками в EF:** если два первых входа гонятся, проигравший упирается в уникальный индекс на `telegram_id` (`gorm:"uniqueIndex"`, `domain/client.go:12`), и код **перечитывает и возвращает строку победителя** вместо 500. Внутрипроцессная проверка not-found сама по себе имеет TOCTOU-разрыв — именно уникальный индекс БД делает её безопасной. `LeadID *uuid.UUID` равен nil на пути WebApp, поэтому ветка повторной привязки lead — мёртвый код на *этом* маршруте (она для пути deep-link бота `/start`, который делит `Register`).

**Шаг 5 — выпуск JWT (`token/token.go:23-35`).** HS256 (**симметричный** — один и тот же `cfg.JWTSecret` подписывает и проверяет). `Claims` встраивает `jwt.RegisteredClaims` (встраивание из §4) и добавляет `Role`. `Subject = client.ID` ≈ `sub` / `ClaimTypes.NameIdentifier`. **Критично:** *тот же* секрет выпускает токены менеджера (`main.go:53`) и токены клиента (`main.go:54`), поэтому claim `Role` — *единственное*, что разделяет эти две аудитории — комментарий в `token.go:12-15` прямо описывает угрозу.

**Шаг 6 — `RequireClientAuth` (`middleware/auth.go:29-58`).** **Функция высшего порядка**: `RequireClientAuth(secret)` возвращает замыкание `gin.HandlerFunc` (вычисляется единожды при старте, `router.go:65`). На каждый запрос: извлечь bearer через `strings.CutPrefix`; `token.Parse` проверяет подпись + `exp`; затем `claims.Role != role` отклоняет токен менеджера несмотря на идеальную подпись; затем `c.Set("clientID", subject)` прячет идентичность. `token.Parse` (`token/token.go:40-56`) фиксирует алгоритм **и** в keyfunc (утверждая `*jwt.SigningMethodHMAC`), **и** в `WithValidMethods(["HS256"])` — вместе они убивают атаку alg-confusion (подмена `none`/`RS256`). `WithExpirationRequired()` отклоняет токены без `exp`.

**В сравнении с аутентификацией ASP.NET:** две стадии (`AddJwtBearer` строит `HttpContext.User`, затем `[Authorize(Roles=…)]`) **слиты в один `gin.HandlerFunc`**. Последствия, которые кусают:
- **Нет ambient `User`/`ClaimsPrincipal`.** Нижележащий код вызывает `middleware.ClientID(c)` (`auth.go:67`), который делает `c.Get("clientID")` и приводит тип к `uuid.UUID`, возвращая `(uuid.UUID, bool)`.
- **Булев `ok` — реальный предохранитель** (`auth.go:60-61`): хэндлер, смонтированный вне защищённой группы, получает `ok=false` → 401, вместо того чтобы молча работать как `uuid.Nil`.
- **Значения в контексте нетипизированы (`any`)**, отсюда приведение типа + bool.

**Чистый эффект — изоляция арендаторов:** id клиента берётся **только** из проверенного JWT, никогда из URL или тела. Поэтому `GET /api/app/shipments/:id` не может прочитать чужое отправление — `DetailForClient(ctx, id, clientID)` (`app_shipment.go:59`) фильтрует по доверенному `clientID`. Именно middleware делает этот id доверенным.

---

## 7. Крупнейшие сдвиги ментальной модели (быстрый справочник)

Они подробно каталогизированы выше; это твоя шпаргалка «разучиться писать как в C#» — см. `mental_model_shifts` для сжатого списка.

---

## Приложение: индекс идиом → `file:line`

- Собранный руками DI / composition root: `cmd/api/main.go:40-69`
- Структурный интерфейс, определённый потребителем: `service/lead.go:20-25`, удовлетворён `repository/lead.go:23,27,36,51`, утверждён в `main.go:52`
- Одно значение, четыре интерфейса: `bot/bot.go:18-20,210-225`; `main.go:56`
- Ошибка-как-значение + воронка: `handlers/lead.go:30-37,90-99`; `%w` в `service/lead.go:124`; sentinel `domain/errors.go:8`
- Трансляция ошибок репозитория: `repository/lead.go:39-41` (чтение), `repository/lead.go:59-61` (запись)
- Теги модели GORM: `domain/lead.go:12-24`; охрана `json:"-"` `domain/manager.go:14`
- Транзакция-замыкание: `repository/shipment.go:23-38`; подводный камень обновления через map `repository/shipment.go:96-124`
- Тройка enum: `domain/lead.go:26-44`; CHECK в `migrations/000004_create_shipments.up.sql`
- Раннер миграций: `cmd/migrate/main.go`
- Горутина fire-and-forget: `service/lead.go:101-110`; корневой ctx `main.go:21`; shutdown `main.go:85-90`
- Nullability через указатель / `NullDecimal`: `domain/shipment.go:19,25-26`
- Ресивер по значению против ресивера-указателя: `service/lead.go:49` (ptr) против `service/lead.go:112` (value)
- Встраивание: `token/token.go:16-19`
- HMAC + сравнение за константное время: `telegram/initdata.go:95-121`
- Фиксация alg в JWT: `token/token.go:40-56`
- Middleware как функция высшего порядка: `middleware/auth.go:29-58`

---

## Сдвиги мышления — шпаргалка «разучиться писать как в C#»

1. Ошибки — это обычные возвращаемые значения, а не исключения. Каждый вызов, способный завершиться неудачей, возвращает (value, error), и ты ОБЯЗАН проверить err на следующей строке — нет try/catch и нет раскрутки стека. Sentinel-значения (domain.ErrNotFound), сопоставляемые через errors.Is, — это «типы исключений»; %w — это связь InnerException, а switch хэндлера по errors.Is — ЭТО твой ExceptionFilter, написанный руками (service/lead.go:124, handlers/lead.go:90-99). Забыть проверить err компилируется нормально и молча продолжает с нулевым значением.

2. Интерфейсы неявные и объявляются ПОТРЕБИТЕЛЕМ, а не реализатором. Нигде нет ': ILeadStore' — тип удовлетворяет интерфейсу просто потому, что имеет подходящие методы, что проверяется в точке присваивания (main.go:52). Контракт живёт в сервисе, которому он нужен (service/lead.go:20), остаётся крошечным, а репозиторий/бот никогда его не импортируют и не именуют. Это Dependency Inversion без DI-контейнера; один *bot.Bot удовлетворяет четырём интерфейсам нотификаторов, о которых никогда не слышал (bot/bot.go:18-20).

3. Для значимых типов нет null — есть нулевые значения, и отсутствующее JSON-поле молча становится одним из них. Свежеобъявленная структура имеет поля '' / 0 / false / nil; код ИСПОЛЬЗУЕТ uuid.Nil / '' / 0 как «отсутствует» (service/shipment.go:80). Только указатели/слайсы/мапы/интерфейсы могут быть nil. Nullability для значимых типов выражается переходом на указатель (*uuid.UUID ≈ Guid?) или decimal.NullDecimal (≈ decimal?) — а разыменование nil-указателя приводит к панике БЕЗ предупреждения компилятора, в отличие от анализа nullable-ссылок в C#.

4. Нет DI-контейнера и нет времени жизни на запрос. Весь граф объектов собирается руками единожды в main.go (40-69), в порядке зависимостей, который навязывает компилятор, и каждый репозиторий/сервис фактически является синглтоном на весь процесс. *gorm.DB — это один общий, потокобезопасный, пулящий хэндл БЕЗ change tracker и БЕЗ SaveChanges — каждая запись явная и немедленная (repository/lead.go:24). Scope/отмена на запрос несётся в context.Context, передаваемом руками в каждый метод, а не новыми объектами.

5. GORM — это не EF code-first, и схема НЕ генерируется из твоих структур. Struct-теги — это документация, которая ЗЕРКАЛИТ написанные руками SQL-миграции (golang-migrate); нет AutoMigrate, нет diff’а моделей, нет OnModelCreating. Связи, FK, поведение ON DELETE, ограничения CHECK и перечисления живут в SQL миграций, и ты сам пишешь и up.sql, и down.sql. Структура + миграция — это два источника истины, синхронизируемых человеком; рассинхрон валится во время выполнения, а не на компиляции.

6. Конкурентность — это горутины + context.Context + каналы, а не async/await/Task. 'go f()' — это fire-and-forget БЕЗ awaitable-хэндла (service/lead.go:105); нет окрашивания функций (function coloring). context.Context — аналог CancellationToken (плюс дедлайны и значения, привязанные к запросу), передаётся явно первым аргументом везде — нет ambient HttpContext. Намеренная подмена контекста запроса на context.Background() в фоновой горутине здесь корректна, потому что контекст запроса отменяется в момент записи HTTP-ответа.

7. Регистр букв — ЭТО модификатор доступа, а директория — ЭТО пакет. Идентификатор с заглавной = public (экспортируемый), со строчной = package-private — нет ключевых слов public/private/internal, которые можно искать грепом. Все файлы одной директории делят плоскую область видимости и видят неэкспортируемые имена друг друга без import. Директория internal/ — это файрвол видимости, навязываемый компилятором по пути директории, а не атрибут.

8. Нет наследования. Переиспользование — это композиция: встраивание (token/token.go:16 продвигает поля RegisteredClaims, но это НЕ подтипизация — нет virtual/override, нет is-a) и интерфейсы для всего полиморфизма. Методы явно выбирают ресивер по значению или по указателю, что меняет поведение: метод с ресивером по значению мутирует только КОПИЮ, поэтому normalized() обязан вернуть её, а вызывающий обязан переприсвоить (service/lead.go:50,112) — забытое переприсваивание молча отбрасывает изменение.

---

## Маршрут чтения (по порядку)

1. backend/cmd/api/main.go — вся история сборки в одном файле; это ЭТО твои Program.cs + Startup, читай сверху вниз первым.
2. backend/internal/http/router.go — таблица маршрутов + группы middleware; показывает каждый эндпоинт и какая позиция аутентификации (публичный / менеджер / клиент) его охраняет.
3. backend/internal/domain/lead.go — самая чистая сущность: struct-теги (gorm + json), паттерн типизированного-строкового-enum (LeadStatus + IsValid). Твоя EF-сущность без мистики.
4. backend/internal/domain/errors.go — sentinel domain.ErrNotFound; крошечный, но это хребет всей обработки ошибок.
5. backend/internal/http/handlers/lead.go — действия контроллера: model binding ShouldBindJSON и воронка errors.Is → код статуса (твой ExceptionFilter, сделанный явным).
6. backend/internal/service/lead.go — определённый потребителем интерфейс LeadStore, оборачивание ошибок через %w и горутина уведомления по принципу fire-and-forget. Самый богатый на идиомы Go файл.
7. backend/internal/repository/lead.go — тончайший GORM-репозиторий: Create/First/Update, gorm.ErrRecordNotFound → domain.ErrNotFound и RowsAffected==0 для записей.
8. backend/internal/db/postgres.go — «фабрика DbContext»: один долгоживущий, общий, потокобезопасный *gorm.DB без change tracker.
9. backend/internal/service/notifier.go + backend/internal/bot/bot.go (фокус на bot.go:18-20, 210-225) — как один *bot.Bot структурно удовлетворяет четырём интерфейсам нотификаторов со стороны сервиса, ничего не импортируя из service.
10. backend/internal/domain/shipment.go — связи как голые FK-UUID (без навигационных свойств), nullable-указатели (*time.Time), деньги NullDecimal и вторая тройка enum.
11. backend/internal/repository/shipment.go — транзакция на замыкании (Create отправления + начальное событие статуса атомарно) и подводный камень «обновление через map не обновляет структуру».
12. backend/internal/middleware/auth.go — middleware как функция высшего порядка, возвращающая gin.HandlerFunc; навязывание роли и предохранитель идентичности (uuid.UUID, bool) (нет ambient HttpContext.User).
13. backend/internal/token/token.go — выпуск/парсинг HS256, встраивание структуры RegisteredClaims и фиксация alg, блокирующая атаку alg-confusion для JWT.
14. backend/internal/telegram/initdata.go — реализованная руками валидация HMAC Telegram и сравнение за константное время hmac.Equal (читай это, когда освоишься; это ядро безопасности).
15. backend/cmd/migrate/main.go + backend/migrations/000002_create_leads.up.sql — как применяются написанные руками парные SQL-миграции (нет AutoMigrate, нет diff’а моделей); убедись, что struct-теги зеркалят SQL вручную.

---

## Первое практическое задание

ЦЕЛЬ: добавить новый эндпоинт с аутентификацией менеджера `GET /api/clients/:id`, который возвращает одного клиента по id, зеркаля существующий поток `GET /api/leads/:id`. Это максимально безопасная первая реальная задача на Go: она касается каждого слоя бэкенда (маршрут → хэндлер → сервис → интерфейс-потребитель → репозиторий → domain.ErrNotFound → 404), но НЕ требует миграции и НЕ требует изменения схемы, потому что метод репозитория и таблица БД уже существуют. Оценка времени для новичка в Go: одна сосредоточенная сессия.

ПОЧЕМУ ЭТО РАБОТАЕТ «ИЗ КОРОБКИ»: `ClientRepository.GetByID(ctx, id)` уже существует в backend/internal/repository/client.go:25 (он используется внутренне в Register), и таблица `clients` уже на месте. Не хватает лишь обвязки над ним: записи в интерфейсе со стороны потребителя, метода сервиса, хэндлера и одной строки маршрута. Ты буквально увидишь, как работает идиома структурных интерфейсов — добавь один метод в интерфейс, и репозиторий удовлетворит его бесплатно.

ЧИТАТЬ СНАЧАЛА (в этом порядке, ~20 мин): (1) backend/internal/http/handlers/lead.go:52-70 — `LeadHandler.GetByID`: uuid.Parse(c.Param('id')), воронка errors.Is(domain.ErrNotFound) → 404, c.JSON(200). Это твой точный шаблон. (2) backend/internal/service/lead.go:79-81 — `LeadService.GetByID`, делегирующий store. (3) backend/internal/service/client.go:25-31 — интерфейс `ClientStore` (сейчас GetByTelegramID, Create, List — ОБРАТИ ВНИМАНИЕ, что он пока НЕ объявляет GetByID). (4) backend/internal/http/handlers/client.go — `ClientHandler` (есть List, нет GetByID). (5) backend/internal/http/router.go:49-53 — как `manager.GET('/leads/:id', lead.GetByID)` регистрируется под группой RequireAuth и где сидит `manager.GET('/clients', client.List)`.

ШАГИ:
1. В backend/internal/service/client.go добавь `GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)` в интерфейс `ClientStore` (около строк 27-31). Репозиторий уже это реализует, поэтому код компилируется сразу — это и есть урок структурных интерфейсов.
2. В том же файле добавь `func (s *ClientService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)`, который просто возвращает `s.clients.GetByID(ctx, id)` — зеркаль `LeadService.GetByID` (service/lead.go:79-81). (uuid уже импортирован в client.go.)
3. В backend/internal/http/handlers/client.go добавь `func (h ClientHandler) GetByID(c *gin.Context)`, скопированный из `LeadHandler.GetByID` (handlers/lead.go:52-70), заменив вызов сервиса на `h.service.GetByID(...)` и сообщение not-found на 'client not found'. Тебе нужно добавить импорты `errors`, `github.com/google/uuid` и `icaris-logistic/backend/internal/domain` (скопируй блок импортов из handlers/lead.go).
4. В backend/internal/http/router.go добавь `manager.GET('/clients/:id', client.GetByID)` рядом с существующей строкой `manager.GET('/clients', client.List)` (router.go:53), чтобы он унаследовал аутентификацию по manager-JWT.

ПРОВЕРКА:
- Сборка: из директории backend выполни `go build ./...` — должно скомпилироваться без единой ошибки. (Если запись интерфейса из шага 1 не совпала точно с сигнатурой репозитория, сбой всплывёт здесь, на связывании, а не в репозитории — ровно тот подводный камень, что описан в гайде.)
- Vet: `go vet ./...` должен быть чистым.
- Запусти: подними Postgres + API (`go run ./cmd/api`), выпусти токен менеджера через существующий `POST /api/auth/login` (при необходимости создай менеджера через `go run ./cmd/createmanager`), затем `curl -H 'Authorization: Bearer <token>' http://localhost:<port>/api/clients`, чтобы взять реальный id клиента, и `curl -H 'Authorization: Bearer <token>' http://localhost:<port>/api/clients/<that-id>` — ожидай 200 с JSON клиента.
- Негативные пути, чтобы подтвердить корректность воронки: случайный UUID валидного формата → 404 `{ error: { code: 'not_found', ... } }`; некорректный id вроде `abc` → 400 `invalid_id`; тот же запрос БЕЗ заголовка Authorization → 401 (доказывает, что он под группой менеджера).
- Дополнительно (только после того, как всё выше зелёное): напиши table-driven юнит-тест для `ClientHandler.GetByID` или `ClientService.GetByID` рядом с существующими файлами `*_test.go`, проверяющий, что случай not-found маппится в domain.ErrNotFound — по правилам тестирования проекта (проверяй поведение/вывод, одно утверждение на тест, без проверок количества вызовов мока).
