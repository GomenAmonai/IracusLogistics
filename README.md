# IcarisLogistics

B2B-сервис экспедирования грузов **Китай → Россия**: публичный лендинг с калькулятором и
заявками, менеджерская обработка, Telegram-бот и Mini App для клиента (отслеживание грузов +
чат). Учебный и коммерческий проект.

Источник истины по направлению и фазам — [`AGENTS.md`](./AGENTS.md). Осознанные MVP-упрощения
и техдолг — [`docs/tech-debt.md`](./docs/tech-debt.md). Разборы к ревью — [`docs/`](./docs).

## Что умеет (MVP)

- **Лендинг** — описание услуги, калькулятор диапазона цены, форма заявки → `Lead` в БД.
- **Менеджер** — вход по email+паролю (JWT), CRUD лидов и грузов, смена статуса с историей,
  чат с клиентом, платежи по грузу. Панель — `/manager.html`; уведомление о новом лиде в Telegram.
- **Клиент** — авторизация через Telegram (бот `/start` или Mini App по `initData`),
  быстрый `/status` в боте, уведомления о смене статуса и платежах.
- **WebApp (Telegram Mini App)** — список грузов, детали + таймлайн истории статусов, платежи,
  чат с менеджером.

## Стек

- **Backend:** Go + Gin + GORM + PostgreSQL. Схема — SQL-миграции (не AutoMigrate).
- **Frontend:** React + TypeScript + Vite + Tailwind v4. Три точки входа (лендинг + Mini App +
  панель менеджера), общая дизайн-система. Telegram WebApp SDK.
- **Bot:** go-telegram-bot-api, long polling или webhook (`TELEGRAM_WEBHOOK_*`); исходящие
  уведомления через outbox с ретраями.
- **Локально:** Docker Compose (Postgres).

## Архитектура

Слои бэкенда: `http (Gin)` → `service` (бизнес-логика) → `repository` (GORM) → `db`.
Интерфейсы объявлены на стороне потребителя; бот и сервисы связаны контрактами, без циклов.

```
backend/
  cmd/{api,migrate,createmanager}   # точки входа и CLI
  internal/
    domain/        # сущности: Manager, Lead, Client, Shipment, Message, ShipmentStatusEvent, Payment, Notification
    repository/    # GORM-репозитории
    service/       # бизнес-логика (Lead, Auth, Client, Shipment, Message, Payment)
    http/          # Gin-роутер и хендлеры
    middleware/    # JWT-проверка (RequireAuth / RequireClientAuth)
    token/         # выпуск/разбор JWT с ролью (manager | client)
    telegram/      # проверка подписи Telegram initData (HMAC)
    bot/           # Telegram-бот (команды + уведомления)
  migrations/      # SQL-миграции (схема v12)
frontend/
  index.html  + src/                # лендинг
  webapp.html + src/webapp/         # Telegram Mini App
  manager.html + src/manager/       # панель менеджера
  src/index.css                     # общие дизайн-токены
```

## Пользовательский поток

1. Посетитель на лендинге считает калькулятор и оставляет заявку (`Lead`) — без регистрации.
2. Менеджер получает уведомление в Telegram, связывается, согласует цену.
3. Клиент проходит верификацию в боте (`/start`) — создаётся аккаунт `Client` (по `telegram_id`).
4. Менеджер заводит груз (`Shipment`) с уникальным `tracking_key` и ведёт статусы.
5. Клиент следит за грузами в Mini App (детали + история) и общается с менеджером в чате;
   быстрый `/status` работает прямо в боте.

## Локальный запуск

```bash
docker compose up -d postgres        # Postgres на :5433

cd backend
go run ./cmd/migrate up              # применить миграции (до v12)
go run ./cmd/createmanager -email=manager@example.com -name="Имя" -password='<свой-пароль>'
go run ./cmd/api                     # http://localhost:8080 (без TELEGRAM_BOT_TOKEN бот в no-op)

cd ../frontend
npm install
npm run dev                          # :5173 — лендинг, /webapp.html — Mini App, /manager.html — панель; /api → :8080
```

Без `TELEGRAM_BOT_TOKEN` бот и авторизация WebApp по `initData` выключены; для отладки UI
WebApp без Telegram в DEV принимается `/webapp.html?token=<client-jwt>`.

## Конфигурация (env)

В development `backend/.env` автозагружается (godotenv); в проде переменные приходят из
окружения платформы. Шаблон — `backend/.env.example`. Вне development `config.Validate()`
останавливает старт при dev-дефолтах: пустой/дефолтный `JWT_SECRET` и `DATABASE_URL`,
`sslmode=disable`, отсутствие токена/чата бота.

| Переменная | Назначение | Дефолт |
|---|---|---|
| `DATABASE_URL` | строка подключения Postgres | dev-Postgres на `:5433` |
| `HTTP_ADDR` | адрес API (приоритетнее `PORT`) | `:8080` |
| `PORT` | порт от облачной платформы (Railway/Render) | — |
| `APP_ENV` | окружение | `development` |
| `JWT_SECRET` | секрет подписи JWT | dev-дефолт (**в проде обязателен**) |
| `JWT_TTL` | время жизни токена | `4h` |
| `TELEGRAM_BOT_TOKEN` | токен бота; пусто → бот выключен | `""` |
| `MANAGER_CHAT_ID` | чат для уведомлений менеджеру | `""` |
| `ALLOWED_ORIGINS` | CORS-белый список origin'ов, через запятую | пусто (вне dev кросс-домен запрещён) |
| `TRUSTED_PROXIES` | IP/CIDR прокси, чьим X-Forwarded-For верим | приватные диапазоны + CGNAT |
| `TELEGRAM_WEBHOOK_URL` | https-URL ручки webhook; задан → бот без polling | пусто (long polling) |
| `TELEGRAM_WEBHOOK_SECRET` | секрет webhook (обязателен вместе с URL) | `""` |

Фронтенд: `VITE_API_BASE` — базовый URL бэкенда (без пути, код добавляет `/api`). Запекается
в бандл **на этапе сборки**. Локально не задаётся (относительный `/api` через прокси Vite).

## Деплой (staging)

**Backend + Postgres — Railway, frontend — Vercel.** Бот живёт в том же процессе API
(long polling), поэтому хост не должен засыпать — Railway с `sleepApplication: false`
(Render free отпал: сервисы спят 15 минут и убивают polling).

- **Railway:** сервис `icaris-api` + Postgres `icaris-db`. `DATABASE_URL` — reference-переменная
  `${{icaris-db.DATABASE_URL}}`; плюс `APP_ENV=production`, `JWT_SECRET`, `TELEGRAM_BOT_TOKEN`,
  `MANAGER_CHAT_ID`, `ALLOWED_ORIGINS=<vercel-домен>`. Healthcheck — `/api/health`.
  Билдер RAILPACK собирает только `cmd/api`, поэтому **миграции гоняются с локальной машины**
  по публичному URL БД: `DATABASE_URL='<public url>' go run ./cmd/migrate up`.
- **Vercel:** проект с Root Directory = `frontend`, все три точки входа собираются. В переменных
  проекта (Production+Preview) задан `VITE_API_BASE` = URL Railway-API — он запекается в бандл
  при сборке; после смены значения нужен redeploy без кэша.
- **Mini App в боте:** кнопка-меню → `https://<project>.vercel.app/webapp.html`
  (@BotFather → *Bot Settings → Menu Button*). Именно страница WebApp, не корень и не API.
  Telegram кэширует Mini App — после деплоя бандла иногда нужно чистить кэш Telegram.

> Не запускать второй экземпляр API с тем же токеном бота (локально + облако): Telegram
> отдаёт 409 на конкурирующий long polling. Локально — без `TELEGRAM_BOT_TOKEN`.
> Webhook-режим и outbox реализованы; на Railway webhook включается переменными
> `TELEGRAM_WEBHOOK_*` (см. `docs/tech-debt.md` #22).
> `render.yaml` оставлен как запасной блюпринт (на free-плане Render не годится).

## API

Публичные:

```http
GET  /api/health
POST /api/leads               # форма с сайта (без авторизации)
POST /api/auth/login          # менеджер: { "email", "password" } → { "token" }
POST /api/app/auth/telegram   # клиент: { "init_data" } (Telegram WebApp) → { "token", "client" }
```

Менеджерские (`Authorization: Bearer <token>`, role=manager):

```http
GET   /api/leads
GET   /api/leads/{id}
PATCH /api/leads/{id}                 # { "status": "new|contacted|converted|rejected" }

GET   /api/clients                    # зарегистрированные клиенты

POST  /api/shipments                  # { "client_id", "lane", "from_city", "to_city", "weight", "volume", "price", "currency", "status_note" }
GET   /api/shipments
GET   /api/shipments/{id}             # → { "shipment", "history": [...] }
PATCH /api/shipments/{id}/status      # { "status", "comment" } — статус из enum груза
GET   /api/shipments/{id}/messages
POST  /api/shipments/{id}/messages    # { "text" } — ответ менеджера

POST  /api/shipments/{id}/payments              # { "amount", "currency"?, "channel", "status"?, "comment"? }
GET   /api/shipments/{id}/payments
PATCH /api/shipments/{id}/payments/{paymentID}  # { "status": "pending|confirmed|refunded" }
```

Клиентские WebApp (`Authorization: Bearer <token>`, role=client):

```http
GET   /api/app/shipments              # только свои грузы
GET   /api/app/shipments/{id}         # → { "shipment", "history": [...] }
GET   /api/app/shipments/{id}/messages
POST  /api/app/shipments/{id}/messages  # { "text" }
GET   /api/app/shipments/{id}/payments  # платежи своего груза
```

Списочные ручки принимают `?limit=&offset=` (дефолт 100, максимум 200), сортировка —
новые сверху. Тела запросов ограничены 64 КиБ.

Статусы груза: `pending, picked_up, in_transit, customs_clear, in_warehouse,
out_for_delivery, delivered, cancelled`. Полосы: `cargo, white, buyout` (метка на грузе).
Каналы платежа: `bank_transfer, card_sbp, cash, crypto`. Чужой груз для клиента → 404
(существование не раскрываем). Ошибки в едином формате:
`{"error": {"code": "...", "message": "..."}}`.

Уведомления Telegram идут через outbox (таблица `notifications` + диспетчер с ретраями).
Бот принимает апдейты long polling'ом либо webhook'ом (`POST /api/telegram/webhook`,
включается переменными `TELEGRAM_WEBHOOK_*`).

## Тесты

```bash
cd backend && go test ./...   # репозиторий-тесты сами пропускаются без БД (CI без базы зелёный)
cd frontend && npm run build  # tsc + сборка всех трёх точек входа
```

## Документы

- [`AGENTS.md`](./AGENTS.md) — направление и фазы (источник истины).
- [`docs/architecture.md`](./docs/architecture.md) — доменная модель и границы MVP.
- [`docs/tech-debt.md`](./docs/tech-debt.md) — осознанные MVP-упрощения и техдолг.
- [`docs/learning-phase2.md`](./docs/learning-phase2.md), [`docs/learning-phase3-5.md`](./docs/learning-phase3-5.md) — разборы к ревью.
