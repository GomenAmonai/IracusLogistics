# IcarisLogistics

B2B-сервис экспедирования грузов **Китай → Россия**: публичный лендинг с калькулятором и
заявками, менеджерская обработка, Telegram-бот и Mini App для клиента (отслеживание грузов +
чат). Учебный и коммерческий проект.

Источник истины по направлению и фазам — [`AGENTS.md`](./AGENTS.md). Осознанные MVP-упрощения
и техдолг — [`docs/tech-debt.md`](./docs/tech-debt.md). Разборы к ревью — [`docs/`](./docs).

## Что умеет (MVP)

- **Лендинг** — описание услуги, калькулятор диапазона цены, форма заявки → `Lead` в БД.
- **Менеджер** — вход по email+паролю (JWT), CRUD лидов и грузов, смена статуса с историей,
  чат с клиентом. Уведомление о новом лиде в Telegram.
- **Клиент** — авторизация через Telegram (бот `/start` или Mini App по `initData`),
  быстрый `/status` в боте, уведомления о смене статуса.
- **WebApp (Telegram Mini App)** — список грузов, детали + таймлайн истории статусов, чат с
  менеджером.

## Стек

- **Backend:** Go + Gin + GORM + PostgreSQL. Схема — SQL-миграции (не AutoMigrate).
- **Frontend:** React + TypeScript + Vite + Tailwind v4. Две точки входа (лендинг + Mini App),
  общая дизайн-система. Telegram WebApp SDK.
- **Bot:** go-telegram-bot-api, long polling (`/start`, `/status` + исходящие уведомления).
- **Локально:** Docker Compose (Postgres).

## Архитектура

Слои бэкенда: `http (Gin)` → `service` (бизнес-логика) → `repository` (GORM) → `db`.
Интерфейсы объявлены на стороне потребителя; бот и сервисы связаны контрактами, без циклов.

```
backend/
  cmd/{api,migrate,createmanager}   # точки входа и CLI
  internal/
    domain/        # сущности: Manager, Lead, Client, Shipment, Message, ShipmentStatusEvent
    repository/    # GORM-репозитории
    service/       # бизнес-логика (Lead, Auth, Client, Shipment, Message)
    http/          # Gin-роутер и хендлеры
    middleware/    # JWT-проверка (RequireAuth / RequireClientAuth)
    token/         # выпуск/разбор JWT с ролью (manager | client)
    telegram/      # проверка подписи Telegram initData (HMAC)
    bot/           # Telegram-бот (команды + уведомления)
  migrations/      # SQL-миграции (схема v8)
frontend/
  index.html  + src/                # лендинг
  webapp.html + src/webapp/         # Telegram Mini App
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
go run ./cmd/migrate up              # применить миграции (до v8)
go run ./cmd/createmanager -email=admin@icaris.io -name="Админ" -password=secret
go run ./cmd/api                     # http://localhost:8080 (без TELEGRAM_BOT_TOKEN бот в no-op)

cd ../frontend
npm install
npm run dev                          # :5173 — лендинг, /webapp.html — Mini App; /api → :8080
```

Без `TELEGRAM_BOT_TOKEN` бот и авторизация WebApp по `initData` выключены; для отладки UI
WebApp без Telegram в DEV принимается `/webapp.html?token=<client-jwt>`.

## Конфигурация (env)

Переменные читаются из окружения (не из `.env` — автозагрузки нет). Шаблон — `backend/.env.example`.

| Переменная | Назначение | Дефолт |
|---|---|---|
| `DATABASE_URL` | строка подключения Postgres | `postgres://icaris:icaris@localhost:5433/icaris_logistic?sslmode=disable` |
| `HTTP_ADDR` | адрес API | `:8080` |
| `APP_ENV` | окружение | `development` |
| `JWT_SECRET` | секрет подписи JWT | dev-дефолт (**в проде обязателен**) |
| `JWT_TTL` | время жизни токена | `24h` |
| `TELEGRAM_BOT_TOKEN` | токен бота; пусто → бот выключен | `""` |
| `MANAGER_CHAT_ID` | чат для уведомлений менеджеру | `""` |

Фронтенд: `VITE_API_BASE` — базовый URL бэкенда (без пути, код добавляет `/api`). Локально не
задаётся (относительный `/api` через прокси Vite).

## Деплой

MVP-схема: **фронтенд на Vercel, backend локально через туннель** (ngrok), фронт ходит на
backend по `VITE_API_BASE`.

1. **Backend наружу.** Поднять API с токеном бота и пробросить порт:
   ```bash
   TELEGRAM_BOT_TOKEN=<token> MANAGER_CHAT_ID=<chat_id> go run ./cmd/api
   ngrok http 8080                  # → публичный https://<...>.ngrok-free.app
   ```
2. **Frontend на Vercel.** Root Directory = `frontend` (Vite определится сам, output `dist`,
   обе точки входа собираются). В переменных проекта задать `VITE_API_BASE` = ngrok-URL backend,
   затем redeploy.
3. **Mini App в боте.** Кнопку-меню бота навести на `https://<project>.vercel.app/webapp.html`
   (через @BotFather → *Bot Settings → Menu Button*, либо Bot API `setChatMenuButton`).

> Бесплатный ngrok-URL меняется при перезапуске — тогда обновить `VITE_API_BASE` и сделать
> redeploy. Постоянный хостинг backend (+ managed Postgres, webhook вместо long polling) —
> следующий шаг, см. `docs/tech-debt.md`.

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

POST  /api/shipments                  # { "client_id", "from_city", "to_city", "weight", "volume", "price", "currency", "status_note" }
GET   /api/shipments
GET   /api/shipments/{id}             # → { "shipment", "history": [...] }
PATCH /api/shipments/{id}/status      # { "status", "comment" } — статус из enum груза
GET   /api/shipments/{id}/messages
POST  /api/shipments/{id}/messages    # { "text" } — ответ менеджера
```

Клиентские WebApp (`Authorization: Bearer <token>`, role=client):

```http
GET   /api/app/shipments              # только свои грузы
GET   /api/app/shipments/{id}         # → { "shipment", "history": [...] }
GET   /api/app/shipments/{id}/messages
POST  /api/app/shipments/{id}/messages  # { "text" }
```

Статусы груза: `pending, picked_up, in_transit, customs_clear, in_warehouse,
out_for_delivery, delivered, cancelled`. Чужой груз для клиента → 404 (существование не
раскрываем). Ошибки в едином формате: `{"error": {"code": "...", "message": "..."}}`.

## Тесты

```bash
cd backend && go test ./...   # репозиторий-тесты сами пропускаются без БД (CI без базы зелёный)
cd frontend && npm run build  # tsc + сборка обеих точек входа
```

## Документы

- [`AGENTS.md`](./AGENTS.md) — направление и фазы (источник истины).
- [`docs/architecture.md`](./docs/architecture.md) — доменная модель и границы MVP.
- [`docs/tech-debt.md`](./docs/tech-debt.md) — осознанные MVP-упрощения и техдолг.
- [`docs/learning-phase2.md`](./docs/learning-phase2.md), [`docs/learning-phase3-5.md`](./docs/learning-phase3-5.md) — разборы к ревью.
