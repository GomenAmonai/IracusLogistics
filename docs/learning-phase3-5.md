# Фазы 3–5 — разбор для код-ревью

Материал к разбору. Код Фаз 3–5 написан Claude целиком (ты делегировал всю сессию:
«реализуй все фазы MVP самостоятельно»). Поэтому здесь не «как написать», а **«почему так
написано»**, чтобы ты мог объяснить и оспорить каждое решение. Учебная установка прежняя:
должен уметь объяснить каждую строку.

> ⚠️ Это полная делегация, а не обычный режим (где ты пишешь GORM-структуры, а Claude — SQL).
> В частности, структуру `domain.ShipmentStatusEvent` и обе миграции (000007, 000008) написал
> Claude. На ревью пройди их как если бы писал сам — особенно SQL.

---

## 1. Что вошло

- **Фаза 3 — клиентский поток:** Telegram-авторизация клиента (HMAC `initData`), создание
  `Client` (`ClientRepository`) с привязкой `lead_id`, генерация `tracking_key`, команда
  `/status` в боте, регистрация через `/start`.
- **Фаза 4 — грузы и статусы:** CRUD `Shipment` (менеджер), история статусов
  (`shipment_status_events`), смена статуса + комментарий, уведомление клиента в Telegram.
- **Фаза 5 — WebApp:** Telegram Mini App (React) — список грузов, детали + таймлайн истории,
  чат с менеджером. Отдельная точка входа `webapp.html`, общая дизайн-система с лендингом.
- **Чат (Фаза 4/5):** `Message` end-to-end — клиент пишет из WebApp, менеджер отвечает ручкой,
  обе стороны получают Telegram-уведомления.

## 2. Карта новых файлов

**Бэкенд — новые пакеты:**
- `internal/token/` — единый контракт JWT (`Issue`/`Parse`, claims с `Role`).
- `internal/telegram/` — проверка подписи `initData` (HMAC) + тип `User`.

**Бэкенд — по слоям:**
- domain: `shipment.go` (+`ShipmentStatusEvent`, `ShipmentStatus.IsValid()`).
- repository: `client.go`, `shipment.go`, `message.go`.
- service: `client.go` (`ClientService`), `shipment.go` (`ShipmentService`), `message.go`
  (`MessageService`); `auth.go` переведён на `token.Issue`.
- middleware: `auth.go` — `RequireAuth` теперь проверяет `role=manager`; добавлен
  `RequireClientAuth` (`role=client`).
- handlers: `client_auth.go`, `client.go`, `shipment.go`, `app_shipment.go`; `response.go`
  (+`respondNotFoundOr500`).
- bot: `bot.go` (+`Run` long polling, методы-уведомители), `format.go` (форматтеры + подписи).
- migrations: `000007` (partial unique `clients.lead_id`), `000008` (`shipment_status_events`).
- cmd/api: сборка новых сервисов + горутина `bot.Run`.

**Фронтенд — WebApp:** `webapp.html`, `src/webapp/**` (App, `lib/{api,telegram,types,format}`,
`components/*`), `vite.config.ts` (две точки входа).

## 3. Три потока данных (читать с кодом)

**Авторизация WebApp:**
`POST /api/app/auth/telegram { init_data }` → `ClientService.AuthenticateWebApp` →
`telegram.ValidateInitData` (HMAC + свежесть) → `Register` (найти/создать клиента) →
`token.Issue(role=client)` → `{ token, client }`. Дальше WebApp шлёт `Authorization: Bearer`.

**Регистрация через бота:**
`/start <lead_id?>` → `Bot.handleStart` → `ClientRegistrar.Register(telegram_id, …, leadID)`.
Тот же `Register`, что и в WebApp — клиент один, ключ `telegram_id`.

**Смена статуса + уведомление:**
`PATCH /api/shipments/:id/status` → `ShipmentService.UpdateStatus` (валидирует enum) →
`store.UpdateStatus` (транзакция: апдейт груза + запись истории) → **горутина**
`NotifyShipmentStatus(client.telegram_id, …)`.

## 4. Ключевые решения и почему так

### 4.1 Роль в JWT и пакет `token` — закрыли privilege escalation
До Фазы 3 менеджерский токен нёс только `Subject = managerID`. Если бы клиент получил
валидно подписанный токен (а он получает — `role=client`), и `RequireAuth` НЕ проверял роль,
то клиентский токен прошёл бы менеджерскую проверку: `Subject` парсится как UUID, кладётся
как `managerID`, и `GET /leads` отдал бы клиенту все лиды. **Это и есть эскалация привилегий.**

Закрыли так: общий claim `Role` (`token.Claims`), `RequireAuth` требует `role=manager`,
`RequireClientAuth` — `role=client`. Логику выпуска/разбора вынесли в `internal/token`, чтобы
не дублировать security-критичный разбор (проверка `alg=HS256`, обязательный `exp`) в двух
местах. **.NET-аналогия:** это как `[Authorize(Roles="Manager")]` поверх общего хендлера JWT.

### 4.2 Проверка `initData` Telegram (HMAC) — `internal/telegram`
Telegram подписывает данные WebApp. Алгоритм (точно по спецификации):
`secret_key = HMAC_SHA256(key="WebAppData", msg=botToken)`, затем
`hash == HMAC_SHA256(key=secret_key, msg=data_check_string)`, где `data_check_string` — все
поля кроме `hash`, отсортированные по ключу и склеенные `key=value` через `\n`. Сравнение —
`hmac.Equal` (постоянное время, анти-тайминг). Дополнительно:
- **Пустой botToken → отказ** (fail closed): без секрета подделку не отличить.
- **Свежесть `auth_date`** (окно 24ч): защита от воспроизведения перехваченной строки.
Тест `initdata_test.go` собирает подпись независимой реализацией и проверяет round-trip:
валидная проходит, подделанная/чужим токеном/протухшая/без токена — отклоняются.

### 4.3 `Register` — идемпотентный get-or-create, общий для бота и WebApp
Клиент создаётся и через `/start`, и через WebApp; оба сходятся на `telegram_id` (UNIQUE).
`Register` сначала ищет по `telegram_id`, при отсутствии создаёт. **Гонка:** два параллельных
`Register` одним `telegram_id` — второй упрётся в unique-индекс; ловим ошибку `Create` и
перечитываем (это нормальный исход, а не сбой). Параметры примитивные (`telegram_id, username,
name`), чтобы `service` не зависел от пакета `telegram`.

### 4.4 `tracking_key` — генерация
Формат `ICR-XXXXXXXXXX`, алфавит Crockford base32 без похожих символов (нет I, L, O, U) —
ключ можно надиктовать. **Без смещения по модулю:** `len(alphabet)==32`, `256 % 32 == 0`,
поэтому `byte % 32` распределён ровно. 10 символов ≈ 50 бит. Уникальность — unique-индекс;
сервис делает пред-проверку в цикле, чтобы не отдавать 500 на (практически невозможную)
коллизию. `crypto/rand`, не `math/rand`.

### 4.5 Транзакции для груза и истории
`ShipmentRepository.Create` и `UpdateStatus` пишут **груз и запись истории одной
транзакцией** (`db.Transaction`): груз без начального события оставил бы таймлайн неполным,
а смена статуса без записи истории — соврала бы клиенту. `UpdateStatus` при первом переходе
в `delivered` проставляет `delivered_at` и **перечитывает** груз (т.к. `Updates` с map не
пишет `now()` обратно в структуру). Транзакция живёт в репозитории (это забота слоя данных),
сервис её не видит — иначе пришлось бы протаскивать `*gorm.DB` через интерфейсы.

### 4.6 Принадлежность груза → 404, не 403
`DetailForClient`/`ListForClient`/`SendFromClient` проверяют `shipment.ClientID == clientID`
и при чужом грузе возвращают `domain.ErrNotFound` (→ 404), а **не** 403. Так мы не
подтверждаем существование чужого груза (не раскрываем, что id валиден). `clientID` берётся
из токена, не из запроса — клиент не может попросить чужой `client_id`.

### 4.7 Уведомители — интерфейсы на стороне потребителя; бот их реализует
`ShipmentService` объявляет `ClientNotifier` (статус), `MessageService` — `ManagerNotifier`
и `ClientMessageNotifier`. `*bot.Bot` структурно удовлетворяет всем — `main` передаёт один и
тот же бот в каждый сервис. Бот при этом НЕ импортирует `service`; для команд (`/start`,
`/status`) он объявляет **свои** интерфейсы (`ClientRegistrar`, `ShipmentLister`), которые
реализуют сервисы. Никаких циклов: `bot → domain`, `service → domain`, `main` склеивает.

### 4.8 Бот: long polling в горутине
`bot.Run(ctx, deps)` крутит `GetUpdatesChan` и в `select` слушает `ctx.Done()` → при отмене
`StopReceivingUpdates()` и выход. Запускается из `main` как `go notifier.Run(...)` с тем же
`ctx`, что и graceful shutdown сервера. В no-op режиме (нет токена) `Run` сразу возвращается —
dev поднимается без Telegram. Ошибки отправки не пробрасываем с первопричиной (строка из
`Send` содержит URL с токеном).

### 4.9 `shipment_status_events` — новая таблица под «историю статусов»
AGENTS (Фаза 5) требует «история статусов», а в схеме был только текущий `status`. Завёл
таблицу-историю: `shipment_id` (FK, `ON DELETE CASCADE` — события принадлежат грузу),
`changed_by` (FK на менеджера, `SET NULL` — историю храним, даже если менеджера удалили),
`status` (CHECK тот же enum). Индекс на FK в одной миграции с таблицей — для **новой**
таблицы это ок (откатывается вместе с ней), в отличие от индекса на существующую таблицу
(`000007` вынесен отдельно). **Тебе на ревью:** прошёл бы ты `up`/`down` сам? Оба обратимы
(проверено `up→down→up`).

### 4.10 WebApp как вторая точка входа Vite, светлая бренд-тема
WebApp — `webapp.html` + `src/webapp/` в том же Vite-проекте: отдельный HTML, отдельный
React-root, **без общего роутинга** с лендингом, но **дизайн-система общая** (оба тянут
`src/index.css`). Это прагматичная трактовка «отдельное приложение, общая дизайн-система».
**Решение:** Mini App использует нашу светлую бренд-тему («Тихая гавань»), а не подстраивается
под тему Telegram — ради единообразия с лендингом. Числа (трек-ключ, даты, вес, цена) —
моноширинные (`.terminal`), как договаривались.

## 5. ⚠️ Решения на ратификацию

1. **Кардинальность `client.lead_id` = 1:1** (миграция `000007`, partial unique). Логика:
   один лид — это один запрос одного человека, он конвертируется максимум в одного клиента.
   1:1 легко ослабить до 1:N позже (удалить индекс); обратно — больно (дедуп). Согласен?
2. **Дизайн-система уехала в светлую тему.** Память фиксировала «строгий тёмный Linear»,
   но актуальный `index.css` (твои незакоммиченные правки) — светлая «Тихая гавань» с тиловым
   акцентом. WebApp сделан под **актуальную** светлую тему. Если тёмная — переключим.
3. **Менеджерский reply — только ручкой `POST /shipments/:id/messages`** (без менеджерского
   UI). Достаточно для MVP? (см. tech-debt #16).

## 6. Как погонять руками

```bash
docker compose up -d postgres
cd backend && go run ./cmd/migrate up        # до версии 8
go run ./cmd/createmanager -email=admin@icaris.io -name="Админ" -password=secret
go run ./cmd/api                              # bot disabled без TELEGRAM_BOT_TOKEN

# менеджер логинится
TOKEN=$(curl -s -X POST localhost:8080/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@icaris.io","password":"secret"}' | jq -r .token)

# клиент появляется только после Telegram-auth. Для ручной проверки без бота —
# создать клиента в БД и завести ему груз:
#   psql ... "insert into clients (telegram_id, name) values (111, 'Тест') returning id;"
#   curl -X POST localhost:8080/api/shipments -H "Authorization: Bearer $TOKEN" \
#     -d '{"client_id":"<id>","from_city":"Гуанчжоу","to_city":"Москва","weight":150}'
#   curl -X PATCH localhost:8080/api/shipments/<sid>/status -H "Authorization: Bearer $TOKEN" \
#     -d '{"status":"in_transit","comment":"в пути"}'

# WebApp локально: vite dev (5173) + открыть /webapp.html?token=<client-jwt> (только DEV).
cd ../frontend && npm run dev

cd ../backend && go test ./...   # repo-тесты сами пропустятся без БД
```

## 7. Чеклист на ревью

- [ ] Роль в JWT: понимаешь, какую именно эскалацию закрывает `role`-claim? (4.1)
- [ ] Алгоритм `initData`: почему два HMAC и почему `hmac.Equal`, а не `==`? (4.2)
- [ ] `Register`: зачем перечитывать после ошибки `Create`? (4.3)
- [ ] `tracking_key`: почему `256 % 32 == 0` важно и почему `crypto/rand`? (4.4)
- [ ] Транзакция в `UpdateStatus`: зачем перечитывать груз после `Updates`? (4.5)
- [ ] Чужой груз → 404, а не 403 — обоснуй. (4.6)
- [ ] Где объявлены интерфейсы уведомителей и почему бот не импортирует `service`? (4.7)
- [ ] Бот выходит из `Run` по `ctx.Done()` — проследи путь. (4.8)
- [ ] `shipment_status_events`: cascade vs set null на FK — почему так? (4.9)
- [ ] Прошёл бы миграции `000007`/`000008` (up/down) сам?
- [ ] Ратифицируем 1:1, светлую тему, reply-через-ручку? (раздел 5)

## 8. Чего НЕ делал (осознанно)

- Менеджерский UI и маршрутизацию ответа из Telegram-группы (tech-debt #16).
- Реалтайм-чат (поллинг/WS) — лента грузится при открытии (#17).
- Webhook вместо long polling, outbox для уведомлений — MVP-долг (#7, #11, #15).
- Persist client-JWT между перезагрузками вне Telegram (#18).
- Стейт-машину переходов статуса груза — свободные переходы, как у лида.

Связанные документы: [`AGENTS.md`](../AGENTS.md), [`docs/tech-debt.md`](./tech-debt.md),
[`docs/learning-phase2.md`](./learning-phase2.md).
