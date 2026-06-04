# Фаза 2 — разбор для код-ревью

Это твой материал к утреннему разбору. Код Фазы 2 написан мной целиком (ты делегировал,
т.к. устал) — поэтому здесь не «как написать», а **«почему так написано»**, чтобы ты мог
объяснить каждый кусок и оспорить решения. Завтра идём по разделу 7 (чеклист) и гоняем
раздел 6 (руками + тесты).

> Учебная установка прежняя: ты должен уметь объяснить каждую строку. Если что-то не
> объясняется — это и есть тема для разбора, не пропускаем.

---

## 1. Что входит в Фазу 2

Из `AGENTS.md`:
- Auth менеджера (email+пароль → JWT) ✅
- CRUD лидов (защищённые ручки) ✅
- Уведомление менеджера в Telegram при новом лиде ✅ (бот написан раньше, теперь подключён)
- Конвертация Lead → Client ⚠️ **частично** — см. раздел 5, это главное решение на ратификацию

## 2. Карта файлов (по слоям)

**Новые:**
- `domain/errors.go` — общий sentinel `ErrNotFound`
- `repository/manager.go`
- `service/auth.go` — `AuthService` (Login, CreateManager, выдача JWT)
- `middleware/auth.go` — `RequireAuth` (проверка JWT) + `ManagerID`
- `http/handlers/auth.go` — `POST /api/auth/login`
- `cmd/createmanager/main.go` — CLI создания менеджера
- миграция `000006_add_client_lead_id` (up/down)
- тесты: `service/auth_test.go`, `service/lead_test.go`, `middleware/auth_test.go`, `repository/manager_test.go`

**Изменённые:**
- `config/config.go` — env `JWT_SECRET`, `JWT_TTL`, `TELEGRAM_BOT_TOKEN`, `MANAGER_CHAT_ID`
- `domain/lead.go` — `LeadStatus.IsValid()`
- `domain/client.go` — поле `LeadID *uuid.UUID`
- `repository/lead.go` — перешёл на `domain.ErrNotFound`, добавлен `UpdateStatus`
- `service/lead.go` — расширен `LeadStore`, добавлен `Notifier`, методы `GetByID`/`UpdateStatus`, фоновое уведомление
- `http/handlers/lead.go` — `GetByID`, `UpdateStatus`
- `http/router.go` — публичная и защищённая группы
- `cmd/api/main.go` — сборка зависимостей (бот, auth)

## 3. Три потока данных (читать с кодом в руках)

**Логин:**
`POST /api/auth/login` → `AuthHandler.Login` → `AuthService.Login` →
`ManagerStore.GetByEmail` (репозиторий) → `bcrypt.CompareHashAndPassword` →
`issueToken` (HS256, `Subject = managerID`) → `{ "token": "..." }`.

**Защищённый запрос:**
`GET /api/leads` → `middleware.RequireAuth` парсит `Authorization: Bearer <jwt>`,
проверяет подпись и алгоритм, кладёт `managerID` в `gin.Context` → хендлер → сервис.
Нет/битый токен → `401` и цепочка обрывается (`c.Abort...`).

**Новый лид + уведомление:**
`POST /api/leads` (публично) → `LeadService.Create` → `store.Create` → **горутина**
`notifier.NotifyNewLead` (фоном, ошибка только в лог) → ответ `201` клиенту не ждёт Telegram.

## 4. Ключевые решения и почему так

### 4.1 `domain.ErrNotFound` — перенёс из `repository`
Раньше `ErrNotFound` жил в `repository`. Я перенёс его в `domain`. **Почему:** «не найдено» —
слово из общего словаря; его *возвращает* repository и *понимает* service. Если оставить в
repository, то `AuthService`, чтобы отличить «нет такого email», был бы вынужден
импортировать `repository` — а мы специально держим service независимым от repository
(см. 4.2). `domain` импортируют оба слоя, поэтому sentinel там — и никакой циклической
зависимости. **Это изменило твой референс `repository/lead.go` и его тест** — на ревью покажу
диф, проверь, что согласен.

### 4.2 Интерфейсы на стороне потребителя
`service/auth.go` объявляет `ManagerStore`, `service/lead.go` — `LeadStore` и `Notifier`.
Это тот же идиом, что мы обсуждали: интерфейс там, где используется, перечисляет только
нужные методы; конкретные `*ManagerRepository`/`*Bot` удовлетворяют им **структурно**, без
`implements`. Профит: сервис не зависит от GORM и от пакета bot → в тестах подменяется
фейком (см. `auth_test.go` `stubManagerStore`, `lead_test.go` `fakeLeadStore`/`noopNotifier`).

### 4.3 JWT
- Алгоритм **HS256** (симметричный, один секрет). В `RequireAuth` keyfunc **проверяет тип
  метода** (`*jwt.SigningMethodHMAC`) — это защита от alg-confusion: иначе атакующий мог бы
  подсунуть токен с `alg: none` или RS256 и обойти подпись. Запомни этот приём.
- В токене стандартные claims: `Subject = managerID`, `IssuedAt`, `ExpiresAt`. Парсер v5
  сам проверяет `exp`.
- Секрет приходит из `config`, наружу/в логи не попадает.

### 4.4 bcrypt и отсутствие user-enumeration
- Хеш создаётся только в `AuthService.CreateManager` (одно место) — поэтому CLI идёт через
  него, а не дублирует bcrypt.
- `CompareHashAndPassword` сравнивает за **постоянное время** (анти-тайминг).
- На «нет такого email» и «неверный пароль» возвращается **один и тот же**
  `ErrInvalidCredentials` → `401`. Так мы не подсказываем, какие email существуют.

### 4.5 Middleware и группы маршрутов
`router.go`: публичная группа (`/health`, `POST /leads`, `/auth/login`) и защищённая
(`protected.Use(RequireAuth)` → `GET /leads`, `GET /leads/:id`, `PATCH /leads/:id`).
`POST /leads` остаётся **публичным** — это форма с сайта. `ManagerID(c)` достаёт id из
контекста (пригодится, когда будем писать «кто изменил статус»).

### 4.6 Смена статуса лида
- `repository.UpdateStatus`: GORM на `Update` без совпадений ошибку не даёт, поэтому
  `RowsAffected == 0` → `domain.ErrNotFound` (иначе PATCH несуществующего лида вернул бы 200).
- `service.UpdateStatus`: проверяет `status.IsValid()` (enum) → иначе `ErrValidation` → 400.
  Переходы **свободные** (можно вернуть из rejected) — осознанное MVP, без стейт-машины.

### 4.7 Бот: внедрение зависимости и fire-and-forget
- `LeadService` получает `Notifier` через конструктор (DI). `main.go` всегда передаёт
  `bot.New(...)`; пустой токен → бот no-op, приложение поднимается без Telegram.
- Уведомление шлётся в **горутине** с `context.Background()`: латентность Telegram не тормозит
  ответ клиенту, а `ctx` запроса к этому моменту уже отменён. Ошибка — только в лог: заявка
  не должна падать из-за бота. Минусы (нет ретраев/ожидания при остановке) записаны в техдолг.

### 4.8 Что ужесточили после авто-ревью (хороший чеклист безопасности)
После написания я прогнал адверсариальное ревью (несколько агентов независимо ищут дефекты,
каждый вердикт перепроверяется ещё одним). Подтверждённые правки — отличные темы для разбора:
- **JWT обязан иметь `exp`.** Парсер `golang-jwt/v5` по умолчанию считает `exp` *необязательным*:
  корректно подписанный токен без `exp` жил бы вечно. Добавил `jwt.WithExpirationRequired()`
  (и `WithValidMethods(["HS256"])` — дублирующая проверка алгоритма). Есть тест на это.
- **Анти-enumeration по таймингу.** Раньше при «нет такого email» ответ возвращался быстро (без
  bcrypt), а при существующем — медленно (bcrypt). По времени можно было угадывать валидные
  email. Теперь в ветке «нет менеджера» прогоняется bcrypt против фиктивного хеша — время выровнено.
- **bcrypt ≤ 72 байт.** bcrypt молча усекает длинные пароли; теперь `CreateManager` явно
  отклоняет `> 72` байт.
- **Защита от nil.** `Login` не разыменует nil-менеджера (контракт `ManagerStore`), а
  `notifyNewLead` не паникует при nil-notifier — паника в фоновой горутине минует `gin.Recovery`
  и уронила бы процесс.

## 5. ⚠️ Решение на ратификацию: конверсия Lead → Client

`Client.telegram_id` — `NOT NULL` (схема AGENTS.md), а у `Lead` его нет. Telegram-id
появляется только при Telegram-авторизации клиента — это **Фаза 3**. Поэтому полноценно
«создать Client из Lead» в Фазе 2 нельзя без одного из:
- (а) сделать `telegram_id` nullable — но это против AGENTS.md («аккаунт после Telegram-auth»);
- (б) вписать фейковый telegram_id — это мусор в данных.

**Что я сделал вместо этого:** статус `converted` доступен через `PATCH`, и готова **схема**
под связь — миграция `client.lead_id` (nullable FK, `ON DELETE SET NULL`) и поле
`domain.Client.LeadID`. Саму запись `Client` и `ClientRepository` создаст Фаза 3 при
Telegram-авторизации, связав с исходным лидом через `lead_id`. (Репозиторий без вызывающего
сейчас — это мёртвый код, поэтому заранее его не вношу.)

**Кардинальность для Фазы 3:** один лид → один клиент (тогда `lead_id` нужен partial UNIQUE
`where lead_id is not null`) или несколько? Сейчас не зафиксировано — реши к Фазе 3.

**Твоё решение завтра:** согласен с такой трактовкой, или хочешь иначе (например, явный
статус «ожидает Telegram»)? От этого зависит начало Фазы 3.

## 6. Как погонять руками (завтра)

```bash
# БД и схема
docker compose up -d postgres
cd backend && go run ./cmd/migrate up

# менеджер
go run ./cmd/createmanager -email=admin@icaris.io -name="Админ" -password=secret

# API (в отдельном терминале)
go run ./cmd/api

# логин → токен
TOKEN=$(curl -s -X POST localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@icaris.io","password":"secret"}' | jq -r .token)

# без токена — должно быть 401
curl -i localhost:8080/api/leads

# с токеном — список
curl -s localhost:8080/api/leads -H "Authorization: Bearer $TOKEN" | jq

# создать лид (публично, без токена) — в логе API увидишь попытку уведомления
curl -s -X POST localhost:8080/api/leads -H 'Content-Type: application/json' \
  -d '{"name":"Иван","phone":"+79990001122","from_city":"Guangzhou","to_city":"Moscow"}' | jq

# сменить статус
curl -s -X PATCH localhost:8080/api/leads/<LEAD_ID> -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"status":"contacted"}' | jq
```

Тесты:
```bash
cd backend && go test ./...
# repository-тесты сами пропустятся, если БД недоступна (CI без базы зелёный)
```

## 7. Чеклист на ревью (пройдём вместе)

- [ ] `domain.ErrNotFound`: согласен с переносом из repository? (4.1)
- [ ] Понимаешь, почему `service` не импортирует `repository`? (4.2)
- [ ] Проверка алгоритма в `RequireAuth` — зачем именно она? (4.3)
- [ ] Почему единый `ErrInvalidCredentials` на оба случая? (4.4)
- [ ] `RowsAffected == 0 → ErrNotFound` в `UpdateStatus` — зачем? (4.6)
- [ ] Горутина+`context.Background()` в уведомлении — объясни tradeoff. (4.7)
- [ ] Понимаешь, зачем `WithExpirationRequired()` и выравнивание тайминга? (4.8)
- [ ] Решение по конверсии Lead→Client — ратифицируем или меняем? (5)
- [ ] Кардинальность `client.lead_id`: 1:1 (UNIQUE) или 1:N? (5 / к Фазе 3)
- [ ] Прогнали раздел 6, 401/201/200 ведут себя как ожидается?

## 8. Чего я НЕ делал (осознанно)

- Полное создание `Client` (Фаза 3, см. 5).
- Refresh-токены, httpOnly-cookie, рейт-лимит логина — MVP, см. `docs/tech-debt.md`.
- Стейт-машину переходов статуса — MVP.
- Ручку «менеджер создаёт менеджера» — пока только CLI.

Связанные документы: [`AGENTS.md`](../AGENTS.md), [`docs/tech-debt.md`](./tech-debt.md).
