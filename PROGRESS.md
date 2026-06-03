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

**Текущий статус:** приложение поднимается на GORM, `/api/health` → 200. `go build`/`go vet`/`go test ./...` чисто. `manager_id` оставлен NOT NULL (подтверждено).

**Следующий шаг:** репозитории для остальных сущностей (`Manager` — понадобится для auth, `Client`, `Shipment`, `Message`). Затем сервисный слой и переход HTTP с chi на **Gin**.

**Открытые вопросы / мелочи:**
- Приглушить шумный дефолтный логгер GORM (логирует `record not found` как error) — настроить `logger.Config{IgnoreRecordNotFoundError: true}` при `gorm.Open`.
- chi → Gin: переписать HTTP-слой на Gin как отдельный шаг (сейчас на chi только health).
