# Iracus Logistic

Учебный и коммерческий проект: B2B-сервис экспедирования грузов Китай → Россия.
Источник истины по направлению и фазам — [`AGENTS.md`](./AGENTS.md). Технический долг и
осознанные MVP-упрощения — [`docs/tech-debt.md`](./docs/tech-debt.md).

## Стек

- Backend: Go + Gin + GORM + PostgreSQL
- Frontend: React + TypeScript + Vite + Tailwind (Telegram WebApp SDK — в Фазе 5)
- Bot: go-telegram-bot-api (исходящие уведомления; long polling — позже)
- Локально: Docker Compose

## Схема и миграции

Схема задаётся SQL-миграциями (не AutoMigrate). Раннер — `cmd/migrate`.

```bash
docker compose up -d postgres        # поднять БД
cd backend
go run ./cmd/migrate up              # применить миграции
go run ./cmd/migrate version         # текущая версия схемы
```

## Запуск

```bash
# 1) первый менеджер (публичной регистрации нет)
cd backend
go run ./cmd/createmanager -email=admin@iracus.io -name="Админ" -password=secret

# 2) API
go run ./cmd/api                     # http://localhost:8080

# 3) лендинг
cd ../frontend
npm install
npm run dev                          # http://localhost:5173, /api проксируется на :8080
```

## Конфигурация (env)

| Переменная | Назначение | Дефолт |
|---|---|---|
| `DATABASE_URL` | строка подключения Postgres | dev-локальная |
| `HTTP_ADDR` | адрес API | `:8080` |
| `JWT_SECRET` | секрет подписи JWT | dev-дефолт (в проде обязателен) |
| `JWT_TTL` | время жизни токена | `24h` |
| `TELEGRAM_BOT_TOKEN` | токен бота; пусто → уведомления выключены | `""` |
| `MANAGER_CHAT_ID` | чат для уведомлений о лидах | `""` |

## API

Публичные:

```http
GET  /api/health
POST /api/leads            # форма с сайта (без авторизации)
POST /api/auth/login       # { "email", "password" } → { "token" }
```

Защищённые (заголовок `Authorization: Bearer <token>`):

```http
GET   /api/leads           # список лидов
GET   /api/leads/{id}      # лид по id
PATCH /api/leads/{id}      # { "status": "new|contacted|converted|rejected" }
```

Ошибки в едином формате: `{"error": {"code": "...", "message": "..."}}`.
