## Обучение

Этот проект — учебный в том числе. Правила:

- Перед написанием нетривиального кода объяснять решение кратко
- Если есть несколько подходов — назвать их и обосновать выбор
- Указывать на Go-идиомы и почему они предпочтительнее
- Не писать код который я не смогу объяснить

## Session Memory

В конце каждой рабочей сессии:

1. Обновить файл `PROGRESS.md` в корне репозитория
2. Записать: что сделано, текущий статус, следующий шаг, открытые вопросы

В начале каждой новой сессии:

1. Прочитать `PROGRESS.md`
2. Восстановить контекст перед началом работы
3. Уточнить у разработчика если что-то неясно

---

## Журнал

### 2026-06-03

**Сделано:**
- Подключён `golang-migrate`; runner `backend/cmd/migrate` (`go run ./cmd/migrate up|down|version`).
- Миграция `000001_create_managers` + структура `domain.Manager` (`uuid.UUID`).
- Миграция `000002_create_leads` + структура `domain.Lead` (`decimal.NullDecimal` для weight/volume, тип-enum `LeadStatus` + константы).
- Зависимости: `golang-migrate/v4`, `google/uuid`, `shopspring/decimal`.

**Принятые решения:**
- Схема через SQL-миграции, AutoMigrate/EnsureSchema не используем (AGENTS.md).
- `timestamptz`, а не `timestamp`. UUID генерит БД (`gen_random_uuid()`).
- enum в БД — `varchar + CHECK`, в Go — типизированные строковые константы.
- Деньги/вес/объём — `decimal`, не `float64`. Опциональные — `decimal.NullDecimal`.
- Workflow миграций: гибрид — Hiki пишет GORM-структуры, Claude пишет SQL.

- Миграции `000003_create_clients`, `000004_create_shipments`, `000005_create_messages` — написаны и проверены (up/down цепочка с FK обратима).
- FK: `shipments.client_id/manager_id` NOT NULL + ON DELETE RESTRICT; `messages.shipment_id` nullable + SET NULL, `client_id` NOT NULL, `manager_id` nullable + SET NULL. Индексы на FK-колонки добавлены вручную.

- Подключён GORM (`gorm.io/gorm` + `gorm.io/driver/postgres`); `internal/db.Connect` теперь возвращает `*gorm.DB`.
- Удалён старый `internal/shipment/` (chi+pgx CRUD) и его HTTP-хендлер; легаси-таблица `shipment_requests` дропнута из dev-БД.
- `health` и `main.go` переведены на GORM; chi-роутер оставлен временно только под `/api/health` (миграция на Gin — отдельный шаг).
- Написан `internal/repository.LeadRepository` (Create/List/GetByID) + интеграционные тесты (само-пропуск, если БД недоступна — CI без базы зелёный). Тесты проверены на живой БД: ID генерит база, `status` дефолтится в `new`, `GetByID` → `ErrNotFound`.

- Первый сквозной срез по `Lead`: `service.LeadService` (валидация/нормализация формы, интерфейс `LeadStore` на стороне сервиса) + хендлеры.
- **HTTP-слой переведён с chi на Gin** (chi удалён из зависимостей). Единый формат ошибок `{"error":{"code","message"}}` (`handlers/response.go`).
- Эндпоинты: `GET /api/health`, `POST /api/leads`, `GET /api/leads`. Проверены вживую: POST → 201 (БД генерит UUID, status=new, decimal через JSON), невалидный → 400, GET → 200.
- Логгер GORM приглушён (`IgnoreRecordNotFoundError`, уровень Warn) в `db.Connect`.

**Текущий статус:** работает сквозной поток `Gin → service → repository → GORM → Postgres` для лидов. `go build`/`go vet`/`gofmt`/`go test ./...` чисто. Стек теперь как в AGENTS.md (Gin + GORM). `manager_id` NOT NULL (подтверждено).

**Следующий шаг:** на выбор — (а) юнит-тесты `LeadService` (валидация, без БД, CI-friendly); (б) `ManagerRepository` + bcrypt + аутентификация (JWT middleware из AGENTS.md); (в) репозитории/сервисы/ручки для `Client`/`Shipment`/`Message`. Логично двигаться к auth, т.к. админка/менеджерские ручки потребуют защиты.

**Открытые вопросы / мелочи:**
- У `LeadService` пока нет юнит-тестов (валидацию стоит покрыть — чистая логика, без БД).
- Gin стартует в debug-режиме (лог-варнинг) — для прода выставить `gin.SetMode(gin.ReleaseMode)` по `AppEnv`.

### 2026-06-04

**Сделано:**
- **Бот** (`internal/bot`) — исходящие Telegram-уведомления о новом лиде + интерфейс `service.Notifier` (на стороне потребителя). Авто-ревью нашло утечку токена в ошибке `Send` (`*url.Error` содержит URL с токеном) — исправлено.
- **Лендинг** (`frontend/`) — Vite + React + TS + Tailwind, дизайн «Iron Corridor», калькулятор (MVP-формула) + форма → `POST /api/leads`. Сборка зелёная, a11y-минорки подчищены.
- **Фаза 2 (бэкенд) целиком:** auth менеджера (bcrypt + JWT HS256), JWT-middleware, `POST /api/auth/login`, CLI `cmd/createmanager`; защищённый CRUD лидов (`GET /leads`, `GET /leads/:id`, `PATCH /leads/:id`), публичный `POST /leads`; рефактор роутера (public/protected); бот подключён к `LeadService` (fire-and-forget в горутине).
- Миграция `000006` (`client.lead_id` nullable FK, `ON DELETE SET NULL`) + поле `domain.Client.LeadID`. Схема **v6**.
- Общий sentinel `domain.ErrNotFound` (перенесён из `repository`, чтобы `service` не зависел от `repository`).
- Тесты: `AuthService` (login/bcrypt/JWT/normalize), `LeadService` (валидация/статусы), `middleware` (401/200/без-exp), `ManagerRepository` (интеграционные, само-пропуск без БД). `go build/vet/test` зелёные.
- Адверсариальное ревью Фазы 2 (workflow, 5 агентов): 0 подтверждённых blocker/major. Исправлены security-минорки: требование `exp` в JWT (+`WithValidMethods`), анти-enumeration по таймингу (холостой bcrypt), лимит bcrypt 72 байта, nil-guards в `Login`/`notifyNewLead`. Удалён мёртвый `ClientRepository`.
- Документация: README переписан под актуальный API; AGENTS — Фаза 2 отмечена; `docs/tech-debt.md` (14 пунктов); `docs/learning-phase2.md` — разбор к утреннему ревью.

**Принятые решения:**
- Auth: JWT (HS256) в заголовке `Authorization`, без refresh; секрет/TTL из config — MVP.
- Создание менеджера — через CLI; bcrypt-хеш создаётся только в `AuthService.CreateManager`.
- Интерфейсы на стороне потребителя (`ManagerStore`, `LeadStore`, `Notifier`).
- Смена статуса лида — свободная, валидируется принадлежность enum (без стейт-машины).
- **Конверсия Lead→Client отложена в Фазу 3** (`Client.telegram_id` NOT NULL появляется только при Telegram-auth); в Фазе 2 — статус `converted` + схема `client.lead_id`. ⚠️ Требует ратификации, см. `docs/learning-phase2.md §5`.

**Текущий статус:** Фазы 1–2 готовы и проверены. `go build/vet/test` чисто, схема v6. Бот и лендинг сданы. Стек по AGENTS.md.

**Следующий шаг (после утреннего ревью):** Фаза 3 — клиентский поток: Telegram-авторизация (HMAC `initData`), создание `Client` (+`ClientRepository`) с привязкой `lead_id`, генерация `tracking_key`, команда `/status`. Сперва ратифицировать решение по конверсии и кардинальность `client.lead_id`.

**Открытые вопросы:**
- Кардинальность `client.lead_id`: 1:1 (partial UNIQUE) или 1:N — решить к Фазе 3.
- Прод-готовность: `JWT_SECRET` обязательным вне dev, CORS-белый список, `gin.ReleaseMode`, таймаут на отправку бота (tech-debt #5, #9, #12).
- WebApp (Фаза 5) — отдельное приложение, общая дизайн-система с лендингом.

### 2026-06-04 (вечер) — Фазы 3–6 одной делегированной сессией

Hiki делегировал реализацию всех оставшихся фаз MVP целиком (build-then-teach). Разбор —
`docs/learning-phase3-5.md`.

**Сделано:**
- **Фаза 3 (клиентский поток):** пакет `internal/telegram` (HMAC-проверка `initData`,
  свежесть `auth_date`, fail-closed на пустой токен); `ClientRepository`;
  `ClientService` (`Register` — идемпотентный get-or-create с восстановлением на
  unique-конфликт; `AuthenticateWebApp`); генерация `tracking_key` (Crockford base32, 50 бит,
  без modulo-bias); бот `/start` (регистрация + привязка `lead_id`) и `/status`.
- **Фаза 4 (грузы):** `ShipmentRepository` (Create и UpdateStatus транзакционно пишут запись
  истории; `delivered_at` ставится/снимается со статусом); `ShipmentService`; менеджерские
  ручки CRUD + смена статуса; миграция `000008 shipment_status_events` + `domain.ShipmentStatusEvent`;
  уведомление клиента в Telegram при смене статуса.
- **Чат:** `MessageRepository`/`MessageService`; клиент пишет из WebApp, менеджер отвечает
  ручкой; уведомления в обе стороны.
- **Фаза 5 (WebApp):** Telegram Mini App — `webapp.html` + `src/webapp/` (вторая точка входа
  Vite, общий `index.css`); авторизация по `initData`, список грузов, детали + таймлайн
  истории, чат. Светлая тема «Тихая гавань» (как актуальный лендинг).
- **Безопасность ролей:** пакет `internal/token` (claims с `Role`); `RequireAuth`
  (role=manager) и `RequireClientAuth` (role=client) — закрыта эскалация (клиентский токен
  больше не проходит менеджерскую проверку).
- **Миграция `000007`:** partial unique `clients.lead_id` (кардинальность 1:1).
- **Тесты:** `telegram` (round-trip подписи), `token`/middleware (разделение ролей),
  `ClientService`/`ShipmentService` (фейки), `ShipmentRepository` (интеграционные,
  само-пропуск без БД), регрессия на nil-`From` в боте. `go build/vet/test` и
  `npm run build` зелёные. Потоки менеджера и клиента проверены end-to-end (curl + Playwright).
- **Адверсариальное ревью (workflow, 25 агентов, 5 измерений):** 8 подтверждённых из 10.
  Исправлено: nil-`From` в боте (паника валила процесс — guard + recover); `delivered_at`
  не снимался при откате из delivered; нестабильный порядок таймлайна (вторичный ключ `id`);
  потеря копеек в `formatMoney` (форматируем строку, не через `Number`); `bot.New` глотал
  категорию ошибки; чат без `aria-live`. Два «находки» отклонены как недостижимые.

**Принятые решения (на ратификацию, см. learning-phase3-5 §5):**
- `client.lead_id` 1:1 (partial unique) — легко ослабить до 1:N позже.
- WebApp использует светлую бренд-тему, а не тему Telegram.
- Менеджерский reply — только ручкой (менеджерского UI в MVP нет).

**Текущий статус:** все фазы MVP (1–6) готовы, схема **v8**. Новый техдолг — `docs/tech-debt.md`
#15–19.

**Следующий шаг:** утренний разбор по `docs/learning-phase3-5.md` (чеклист §7), ратификация
решений §5; затем прод-готовность (webhook, outbox, CORS-whitelist, обязательный JWT_SECRET,
ReleaseMode).

**Открытые вопросы:**
- Тёмная vs светлая тема (память `frontend-design-system` обновлена под актуальную светлую).
- Реальный `TELEGRAM_BOT_TOKEN` для живой проверки бота и WebApp внутри Telegram (локально бот
  в no-op, WebApp гонялся через dev `?token`).

---

### 2026-06-08 — Дизайн: закрыта ветка `redesign/dark-hero`

**Сделано:**
- Ратифицирована тема: **тёмный кинематографичный hero → светлый функциональный низ** (Hiki
  выбрал из вариантов dark-hero / полностью светлая / полностью тёмная).
- Финиш hero (pass 2, `eaaed3d`): добавлена само-достаточная корридор-полоса (Гуанчжоу→Москва)
  с поочерёдной пульсацией узлов (CSS, reduced-motion-safe). Удалён мёртвый кросс-секционный
  sync таймлайн→hero: hero и «как работаем» разнесены на ~1000px по вертикали, подсветка
  `activeNode` была не видна никогда → убран `activeNode` из App/Hero и хендлеры hover/focus
  из HowItWorks.
- `npm run build` зелёный (tsc + vite). `redesign/dark-hero` смержена в `main` (fast-forward,
  `eaaed3d`); **локально, не запушено**.
- Память `frontend-design-system` исправлена под факт: акцент — бронза `#a4661c` (была ошибочно
  записана как тил), hero — тёмный `--color-night`/`--color-amber`.

**Закрытый вопрос:** тёмная vs светлая — решено (dark hero + light body).

**Следующий шаг:** прод-готовность (webhook, outbox, CORS-whitelist, обязательный `JWT_SECRET`,
`gin.ReleaseMode`) — частью как backend-задачи для Hiki; либо ратификация решений §5 из
`learning-phase3-5.md`.

---

### 2026-06-09 — Аудит проекта (бэкенд/бизнес/данные) + прод-гейт A0–A3

**Аудит (3 агента):** бэкенд-код здоров; главный разрыв — доменная модель моделирует
однокомпанийного курьера, а бизнес двусторонний (РФ + партнёр Гуанчжоу), трёхполосный
(карго/белый/выкуп), мультиканальный по платежам. Развилка решена Hiki: **сначала прод-гейт,
потом домен**; платежи — единственный доменный кусок этой итерации; полосы/партнёр — заморожены
до ратификации.

**Модель работы уточнена (value-based split):** Hiki параллельно учит .NET руками, поэтому
прод-гейт/инфру пишет Claude (Hiki ревьюит дифы), а высокоценный доменный Go (платежи) Hiki
пишет сам с менторством. Не возврат к build-then-teach. (память `working-model` обновлена.)

**Сделано (прод-гейт, всё под `go build/vet/test` зелёное; миграции и репо-тесты прогнаны на живой БД):**
- **A0 (БЛОКЕР):** `config.Validate()` (fail-fast вне dev: пустой/дефолтный `JWT_SECRET`,
  `DATABASE_URL`, `sslmode=disable`, отсутствие токена/чата бота → `os.Exit(1)`, `errors.Join`).
  Вызов из `main`. Юнит-тесты на Validate.
- **A1a (баг):** `ClientService.Register` больше не маскирует любую ошибку `Create` под гонку —
  репо транслирует `23505` в `domain.ErrClientExists` / `ErrLeadAlreadyClaimed` (по
  `ConstraintName`, через `pgconn.PgError`), сервис разбирает их `switch`'ем, прочие ошибки
  пробрасывает. Регресс-тест на не-duplicate ошибку.
- **A1b:** `service.Background` (WaitGroup-дренаж фоновых задач). Все fire-and-forget уведомления
  + цикл бота идут через него; `main` ждёт `bg.Wait` после `server.Shutdown`. Тесты на
  блокировку/таймаут.
- **A1c:** `lead.UpdateStatus` → одно `UPDATE … RETURNING` (атомарно, без read-after-write
  гонки); `shipment.UpdateStatus` читает строку с `FOR UPDATE` (гонка `delivered_at`).
- **A2:** CORS-whitelist (`ALLOWED_ORIGINS`), `gin.ReleaseMode` вне dev, пул БД
  (`SetMaxOpenConns` и пр.), HTTP-таймауты (`Read/Write/Idle`), rate-limit middleware на
  публичных POST (token-bucket по IP, без deps), multi-stage Dockerfile (distroless nonroot,
  бинари `api`+`migrate`), app+migrate сервисы в compose, CI бэкенда (Postgres-сервис,
  `build`/`vet`/`migrate up`/`test`). Docker-образ собран и проверен.
- **A3:** миграция `000009` — индексы на `messages.manager_id`, `shipment_status_events.changed_by`.
  Up/down прогнаны (8→9→8→9).
- `docs/tech-debt.md`: #5/#9/#10 → Закрыто; #11/#15 частично закрыты (дренаж есть, outbox нет);
  новый #20 (rate-limit доверяет X-Forwarded-For).

**Следующий шаг:** **Фаза B — модель платежей (`Payment`), пишет Hiki** с менторством
(домен → миграция `000010` (SQL за Claude) → репо → сервис → хендлеры `POST/GET
/api/shipments/:id/payments` → тесты). Референсы-паттерны из этой сессии: sentinel-ошибки
(A1a), транзакции/`RETURNING` (A1c). Затем — ратификация замороженных вопросов (полосы,
партнёр, поток статусов, кардинальность lead_id, явная конверсия Lead→Client) и webhook при
выборе хоста.

**Открытые вопросы (на ратификацию, до кода):** полосы карго/белый/выкуп (3 потока vs метка);
партнёр Гуанчжоу (актор vs вне системы); каналы платежей (конкретный список значений).
