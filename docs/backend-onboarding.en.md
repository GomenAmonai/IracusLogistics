# Backend Onboarding: Go for a .NET / EF Core Developer

> Audience: you — an experienced C# / ASP.NET Core / EF Core developer, new to Go, who owns this project and wants to *learn* the backend, not just have it narrated. Everything below is mapped to a .NET concept you already know, then the Go difference and the *why* is called out. Every claim cites `file:line` in this repo. All file paths are under `backend/`.

---

## 0. What this codebase is (one line)

A Go HTTP API (**Gin** for routing, **GORM** over **Postgres** for data, **Telegram** bot for notifications) implementing a logistics CRM — the same layering you know from ASP.NET Core (`handler → service → repository → domain`), but wired by hand with **no DI container**, with **errors as return values instead of exceptions**, and **interfaces declared by the consumer instead of up front**.

The whole object graph is constructed once, top-to-bottom, in `cmd/api/main.go:20-96`. If you read one file first, read that one.

---

## 1. Architecture & layering — the composition root

### The picture

```
HTTP request
  → internal/http/router.go          (Gin routes + middleware groups)    ~ ASP.NET endpoint routing
  → internal/http/handlers/*.go      (bind JSON, map errors → status)     ~ Controller actions
  → internal/service/*.go            (business logic, validation)         ~ application/service layer
  → internal/repository/*.go         (GORM calls)                         ~ EF Core DbContext repo
  → internal/domain/*.go             (entities + enums + errors)          ~ EF entities / POCOs
Postgres (gorm.io/gorm)
```

### `cmd/api/main.go` IS your `Program.cs` + `Startup.cs` — hand-wired

In .NET this logic is split between `Program.cs` (host, Kestrel, config) and `Startup.ConfigureServices` (the `AddScoped<…>` registrations). Here it is one `main()` function. Read it top to bottom and you have the entire wiring story.

- **Config load** — `cfg := config.Load()` (`main.go:24`). `config.Load()` reads env vars with fallbacks and returns a plain `Config` struct — like binding `IConfiguration` to an options POCO, but it's an explicit function call, not the configuration-provider pipeline.
- **Open the DB** — `gdb, err := db.Connect(ctx, cfg.DatabaseURL)` (`main.go:29`). `gdb` is a `*gorm.DB` — your single shared connection-pool handle (more in §2). Note the **manual `err` check** on the very next line (`main.go:30-33`): Go has no exceptions, so every fallible call returns `(value, error)` and you check `err` immediately. `os.Exit(1)` = crash on a fatal startup error.
- **`defer` = your `using` / `finally`** — `defer func(){ … sqlDB.Close() }()` (`main.go:34-38`) schedules cleanup to run when `main` returns. The `_ =` deliberately discards an error Go would otherwise force you to handle.
- **Construct repositories** (`main.go:40-44`), each handed the same `gdb`. `repository.NewLeadRepository(gdb)` (`repository/lead.go:17-19`) returns a `*LeadRepository` holding the db handle. `New…` is **just a naming convention** — Go has no constructors and no `new` keyword for this.
- **Construct the bot** (`main.go:46`) *before* the services, because services depend on it as a notifier.
- **Construct services** (`main.go:52-56`), injecting repos (and the bot) by hand. This is your `AddScoped<LeadService>()` written out as explicit calls.

```go
leadService := service.NewLeadService(leadRepo, notifier)
shipmentService := service.NewShipmentService(shipmentRepo, clientRepo, notifier)
messageService := service.NewMessageService(messageRepo, shipmentRepo, clientRepo, notifier, notifier)
```

> Look at the last line: the same `notifier` value is passed into **two parameters** (`main.go:56`). Not a copy-paste bug — the two parameters have two *different* interface types, and one `*bot.Bot` satisfies both. This is the heart of §3.

- **Build the router** by passing all services in a `RouterDeps` struct (`main.go:61-69`), then hand it to a standard `http.Server` (`main.go:71-75`).

**Lifetime note (important shift from .NET):** there are **no scoped/transient/singleton lifetimes**. Every repo and service is constructed **once** and lives for the whole process — effectively all singletons. Per-request state (the authenticated user id, cancellation) does *not* live in these objects; it flows through `context.Context` arguments on each method call (see §5). In .NET you'd lean on scoped DI + a `DbContext` per request; here the `*gorm.DB` is shared and request scope is carried by `ctx`.

**Concurrency you won't see in ASP.NET (covered fully in §6):**
- `go notifier.Run(…)` (`main.go:59`) launches the Telegram long-poll loop on a **goroutine**.
- `go func(){ server.ListenAndServe() … }()` (`main.go:77-83`) runs the HTTP server on another goroutine.
- `<-ctx.Done()` (`main.go:85`) **blocks** until SIGINT/SIGTERM cancels the root context (built at `main.go:21` via `signal.NotifyContext`), then `server.Shutdown(shutdownCtx)` (`main.go:90`) drains in-flight requests with a 10s timeout. This is Go's graceful-shutdown idiom — analogous to `IHostApplicationLifetime.ApplicationStopping` + `WaitForShutdownAsync`.

### Packages, the `internal/` rule, and access control

- **Module** = `icaris-logistic/backend` (`go.mod:1`). Roughly your solution + root namespace. The `require` block is your `<PackageReference>` list.
- **Package = one directory.** Every `.go` file in `internal/service/` starts with `package service`. Files in the same package see each other's unexported identifiers with **no `import`** needed — there is no per-file namespace like C#. Closest to a C# `namespace`, but one directory = one package, hard rule.
- **Capitalization IS the access modifier.** No `public`/`private`/`internal` keyword. `LeadService` (capital) is exported = `public`; `validateCreateLead` (`service/lead.go:122`, lowercase) is unexported = package-private. Renaming `notifyNewLead` → `NotifyNewLead` would silently make it part of the public API — there's no keyword to grep for.
- **`internal/` is a compiler-enforced firewall.** Anything under `backend/internal/…` is importable only by code rooted at `backend/`. That's why all real code lives in `internal/` and only `cmd/api`, `cmd/migrate`, `cmd/createmanager` sit outside as runnable entry points. Closest .NET analogy: `internal` visibility + `[InternalsVisibleTo]` scoped to your own assembly — but enforced by directory layout.
- **Imports are paths, not assembly refs.** The alias `apphttp "icaris-logistic/backend/internal/http"` (`main.go:15`) renames the package locally because its real name `http` would collide with the stdlib `net/http` (`main.go:6`). That's a `using apphttp = …`.

### The `cmd/` layout

`backend/cmd/` holds three `package main` programs, each compiling to one binary: `cmd/api` (this HTTP server), `cmd/migrate` (runs SQL migrations), `cmd/createmanager` (seeds a manager account). Go convention: **`cmd/<name>/main.go` per executable**, everything reusable under `internal/`. .NET analogy: one solution with multiple console projects all referencing a shared class library.

---

## 2. The data layer — GORM and golang-migrate for an EF veteran

**One line:** GORM is "EF Core for Go" over Postgres, but the schema is **not** generated from the structs — migrations are **hand-written SQL** run by `golang-migrate`, and the structs are kept in sync by hand.

### Models: struct tags, not Fluent API or DataAnnotations

GORM has only one mapping mechanism: **struct tags** (string metadata parsed via reflection). There is **no `OnModelCreating`** anywhere in this project. `domain/lead.go:12-24`:

```go
type Lead struct {
    ID        uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name      string              `gorm:"type:varchar(255);not null" json:"name"`
    Weight    decimal.NullDecimal `gorm:"type:numeric" json:"weight"`
    Status    LeadStatus          `gorm:"type:varchar(20);not null;default:new" json:"status"`
    CreatedAt time.Time           `gorm:"not null;default:now()" json:"created_at"`
}
```

The backtick string carries two namespaces: `gorm:"…"` (column mapping — your `[Column]`/`HasColumnType`/`IsRequired`/`HasDefaultValueSql`) and `json:"…"` (wire serialization — your `[JsonPropertyName]`). One field, both ORM mapping and HTTP contract.

| GORM tag | EF Core equivalent |
|---|---|
| `primaryKey` | `[Key]` / `HasKey` |
| `type:uuid` | `.HasColumnType("uuid")` |
| `not null` | `.IsRequired()` |
| `default:gen_random_uuid()` | `.HasDefaultValueSql(...)` |
| `uniqueIndex` | `.HasIndex(...).IsUnique()` |
| `json:"-"` | `[JsonIgnore]` |

Two things to clock immediately:
1. **`json:"-"` is security-critical.** `Manager.Password` (`domain/manager.go:14`) is tagged `json:"-"`, so the bcrypt hash is never serialized even though structs are returned directly from handlers. There is no separate DTO/view-model layer guarding it — that one tag is the guard.
2. **Field → column is convention-based snake_case.** `FromCity` → `from_city`, `TelegramID` → `telegram_id`, `ID` → `id`. You won't see `column:` tags because the convention already produces the right name.

### No `DbSet<T>`, no model registry, no change tracker

- There is **no `DbSet`** and no central entity registry. A struct becomes a "model" the moment you hand it to a query: `r.db.Create(lead)` (`repository/lead.go:24`) tells GORM the table for `*domain.Lead` is `leads` (pluralized type name) right there.
- The `*gorm.DB` (built once in `db/postgres.go:14-42`, `gorm.Open(postgres.Open(url), …)` at line 25) is a **single, long-lived, thread-safe handle wrapping a connection pool**, shared by every repo. Contrast EF's `DbContext`, which is per-request, scoped, and not thread-safe. Per-request scoping is done with `.WithContext(ctx)` on each call, not by constructing a new context object.
- **There is no change tracker and no `SaveChanges()`.** Every write is explicit and immediate.

### Queries: method chains, not LINQ

No `IQueryable`, no expression trees. Queries are method chains on `*gorm.DB` that execute on a *finisher* (`Find`, `First`, `Create`, `Count`). The WHERE clause is a **parameterized SQL string fragment**, not a lambda. EF vs GORM:

```csharp
db.Shipments.Where(s => s.ClientId == clientId).OrderByDescending(s => s.CreatedAt).ToListAsync();
```
```go
// repository/shipment.go:49-54
r.db.WithContext(ctx).Where("client_id = ?", clientID).Order("created_at desc").Find(&shipments).Error
```

Three things that bite a LINQ user:
1. **The result is an out-parameter.** `Find(&shipments)` takes a pointer to a slice and fills it; the method returns `*gorm.DB` and the error lives on `.Error`. That's why every chain ends `.Find(...).Error`. The list is the variable you passed in (`var shipments []domain.Shipment`).
2. **`Where("client_id = ?", clientID)` is a raw, parameterized string.** `?` is a real bound parameter (injection-safe). But the column name is hand-typed — a typo is a **runtime** error, not a build error. You lose `s => s.ClientId` refactor-safety.
3. **`First` vs `Find`.** `First` (`repository/shipment.go:64`) **errors with `gorm.ErrRecordNotFound`** if nothing matches. `Find` returns an empty slice and *no* error. That asymmetry drives the two different "not found" mechanisms below.

### "Not found is an error" → translated to `domain.ErrNotFound`

In EF, `FirstOrDefaultAsync` returns `null`. In GORM, `First` returns a typed error. Every read repo does this translation (`repository/lead.go:36-47`, `repository/shipment.go:62-73`):

```go
err := r.db.WithContext(ctx).First(&lead, "id = ?", id).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, domain.ErrNotFound   // re-map so the service never imports GORM
}
if err != nil { return nil, err }
return &lead, nil
```

The repo deliberately re-maps the GORM error to the project's own `domain.ErrNotFound` (`domain/errors.go:8`) — a clean ports-and-adapters boundary, exactly like returning a domain `Result`/`null` instead of leaking `DbException`. The *why* is documented at `repository/manager.go:24-25`.

**Two different "not found" detections — know which applies:**
- **Reads:** `First` errors → check `errors.Is(err, gorm.ErrRecordNotFound)`.
- **Writes:** `Update` does **not** error on zero matches, so the repo checks `result.RowsAffected == 0` and synthesizes `domain.ErrNotFound` (`repository/lead.go:51-63`). Using the wrong check silently misses the case.

### Relations: deliberately NO navigation properties / no `Include`

GORM supports `Preload("Orders")` (your `Include`), but this codebase **does not use it**. `Shipment.ClientID` / `ManagerID` (`domain/shipment.go:14-15`) are bare `uuid.UUID` fields with `index` tags — there is **no `Client Client` nav property**. FK relationships live entirely in the migration SQL (`client_id uuid not null references clients (id)` in `migrations/000004_create_shipments.up.sql`), never inferred from structs. When a service needs the related client, it makes a **second explicit repo call by id**. `ON DELETE` behavior (`restrict`/`set null`/`cascade`) lives only in the SQL — invisible from the Go model.

Mental model: **structs describe columns; migrations describe relationships and constraints. Two separate sources of truth, kept in sync by a human.**

### Transactions: closure API (cleaner than EF's BeginTransaction)

You pass a function; return `error` → rollback, return `nil` → commit. No `using`, no manual `CommitAsync`. The canonical example writes a shipment + its initial status event atomically (`repository/shipment.go:23-38`):

```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(shipment).Error; err != nil { return err }
    event := &domain.ShipmentStatusEvent{ ShipmentID: shipment.ID, Status: shipment.Status, ChangedBy: &shipment.ManagerID }
    return tx.Create(event).Error
})
```

Three load-bearing details: (1) **use `tx`, not `r.db`, for every write inside** — using `r.db` would escape the transaction; (2) `shipment.ID` is populated by the *first* `Create` (GORM reads the DB-generated UUID back via `RETURNING`), which is why line 30 can reference it for the event FK; (3) `ChangedBy: &shipment.ManagerID` uses `&` because the field is a nullable pointer (§4).

A GORM gotcha worth flagging: `UpdateStatus` (`repository/shipment.go:78-131`) updates via a `map[string]any` + `gorm.Expr("now()")`, and **a map update does NOT write values back into your struct**, so it must re-`First` the row (line 124) to return fresh data. EF's change tracker would have refreshed the entity for you.

### Enums: VARCHAR + CHECK + Go typed-string constants (a hand-coordinated triple)

There is **no real enum** in Go *or* Postgres. The pattern (using `LeadStatus`, `domain/lead.go:26-44`; same for `ShipmentStatus`, `domain/shipment.go:31-55):

1. **Go:** `type LeadStatus string` + a `const (…)` block of typed string values. A distinct type so you can't pass a bare string by accident, but the underlying value is the string stored in `varchar(20)`.
2. **DB:** `status varchar(20) … check (status in ('new','contacted',…))` — the CHECK constraint enforces the set.
3. **Bridge:** an `IsValid()` method on the type (`func (s LeadStatus) IsValid() bool`), called by the service before writing so a bad value returns **400** instead of a **500** from the CHECK violation.

**Critical maintenance fact:** these three lists are edited **together, by hand**. Adding a status = edit the Go consts, the `IsValid()` switch, AND write a new migration to alter the CHECK. Nothing generates one from another.

### Migrations: hand-written paired SQL — the deliberate inverse of EF code-first

EF code-first: change the model, `dotnet ef migrations add`, EF diffs and generates `Up()`/`Down()`. **This project: hand-written raw SQL in paired files** (`NNNNNN_name.up.sql` + `.down.sql` in `migrations/`). There is **no diffing, no generation, and no `AutoMigrate`** (grep the backend — it appears nowhere). The structs do **not** create or alter tables; the schema exists only because someone wrote the SQL. You author both halves yourself.

Runner is `cmd/migrate/main.go`: `migrate up` → `m.Up()` (apply all pending); `migrate down` → `m.Steps(-1)` (roll back **exactly one** — safer than EF's `database update 0`); applied versions tracked in `schema_migrations`, with a **`dirty` flag** for partial failures you must resolve manually. Real schema-evolution discipline is visible in migrations 6-8 (e.g. `000007` is a **partial unique index** `… where lead_id is not null` encoding "one lead → at most one client").

**Where each concept lives:** models/tags → `internal/domain/*.go`; "DbContext" → `internal/db/postgres.go`; queries/transactions/error-mapping → `internal/repository/*.go`; schema/FKs/CHECKs → `migrations/*.sql`; runner → `cmd/migrate/main.go`.

---

## 3. The one idiom that feels backwards: interfaces declared by the CONSUMER

This is the most important Go idiom to internalize, because it is the **opposite** of the C# habit. It shows up everywhere, so it gets its own section.

**The C# habit:** define `ILeadRepository` near the repository, write `LeadRepository : ILeadRepository`, register the mapping in the container. The interface is declared up front and the implementation *explicitly names* it.

**What Go does here — two inversions at once:**
1. **Interfaces are satisfied structurally (duck typing).** A type implements an interface by *having matching methods* — there is no `: LeadStore` syntax anywhere.
2. **The interface is declared where it's *consumed*.** The **service** declares the small interface it needs.

`service/lead.go:20-25`:
```go
type LeadStore interface {
    Create(ctx context.Context, lead *domain.Lead) error
    List(ctx context.Context) ([]domain.Lead, error)
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.LeadStatus) error
}
```

`*repository.LeadRepository` (`repository/lead.go:13`) **never mentions `LeadStore`** — yet because its methods match, it *is* one. The compiler verifies this at the assignment in `main.go:52`, where `leadRepo` is passed where a `LeadStore` is expected. To find who implements an interface you follow method signatures, not a declaration (your IDE's "Implementations" command is your friend).

**Why this is powerful — concretely in this repo:** one physical `*bot.Bot` value is consumed through *four different narrow* interfaces, each defined by a different service:
- `service.Notifier` — `NotifyNewLead` (`service/notifier.go:11`)
- `service.ClientNotifier` — `NotifyShipmentStatus` (`service/shipment.go:45`)
- `service.ManagerNotifier` — `NotifyClientMessage` (`service/message.go:26`)
- `service.ClientMessageNotifier` — `NotifyManagerReply` (`service/message.go:31`)

`*bot.Bot` implements all of them as plain methods (`bot/bot.go:210,215,220,225`) and **never imports `package service`** — confirmed by the comment at `bot/bot.go:18-20`: the link is "by method contract only." That's why `main.go:56` writes `…, notifier, notifier`: the one bot satisfies both `ManagerNotifier` and `ClientMessageNotifier`. In C# you'd have `Bot : INotifier, IClientNotifier, …` and register each mapping; here the bot is blissfully unaware those interfaces exist.

**Dependency direction:** arrows point *inward*. `service` defines the contracts; `repository` and `bot` are outer implementations that satisfy them. The compiler enforces that `service` imports only `domain` (not `repository` or `bot`). You get the Dependency-Inversion benefit of `ILeadRepository` **without a registration step**, and interfaces stay tiny because each consumer carves out exactly the methods it calls (e.g. `ShipmentReader` at `service/message.go:21` is a one-method slice). Broad `IRepository<T>` is idiomatic C#; small consumer-side interfaces are idiomatic Go.

**The gotcha:** rename a method on `LeadRepository` and it silently stops satisfying `LeadStore` — the failure surfaces at the `main.go:52` wiring line, not at the repository.

---

## 4. Errors, nulls, zero values, receivers, embedding — the daily idioms

### Errors are values, not exceptions (the single biggest shift)

Every fallible function returns `(value, error)`; you check `err` before using the value. No `throw`/`catch`. `service/auth.go:46-49` is typical:

```go
manager, err := s.managers.GetByEmail(ctx, email)
if err != nil && !errors.Is(err, domain.ErrNotFound) {
    return "", err
}
```

The `""` is the "no useful value" placeholder you return alongside an error. **Forgetting to check `err` compiles fine** and silently proceeds with a zero value — there's no thrown exception to catch.

- **Wrapping with `%w`** is your `new XException(msg, inner)`. `fmt.Errorf("%w: name is required", ErrValidation)` (`service/lead.go:124`) wraps the sentinel so callers can still recognize it. `%v` would flatten to text and *sever* the chain — `service/client.go:107` deliberately uses `%w` for the sentinel and `%v` for the underlying cause (`fmt.Errorf("%w: %v", ErrTelegramAuth, err)`) so the telegram error can't be unwrapped and leaked.
- **Sentinel errors + `errors.Is`** is your `catch (XException)`. `domain.ErrNotFound` (`domain/errors.go:8`) and `service.ErrValidation` (`service/lead.go:16`) are package-level singleton values; `errors.Is` walks the `%w` chain to match them.
- **The handler is your `ExceptionFilter`, but explicit.** `handlers/lead.go:90-99` maps sentinels to status codes with a `switch`:

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

The error shape `{ error: { code, message } }` (`handlers/response.go`) and the "never leak internals on 500" rule match your `error-handling.md`. Payoff of error-as-value: failure paths are visible and local, never an invisible unwind three layers up. Cost: verbosity.

### Nullability via pointers (and `NullDecimal`)

Go has no `Nullable<T>`. For a value type that may be absent, use a **pointer**; `nil` means "absent":
- `LeadID *uuid.UUID` (`domain/client.go:17`) ≈ `Guid?`
- `EstimatedAt *time.Time` / `DeliveredAt *time.Time` (`domain/shipment.go:25-26`) ≈ `DateTime?`

You construct with `&` (e.g. `ChangedBy: &shipment.ManagerID`, `repository/shipment.go:33`) and **must guard before dereferencing** (`if current.DeliveredAt == nil`, `repository/shipment.go:104`). A `nil`-pointer dereference is a runtime panic — your `NullReferenceException` — and **the Go compiler does NOT warn you** the way C#'s nullable-reference analysis does.

**Two null representations, not interchangeable:** pointers for FKs/timestamps; `decimal.NullDecimal` (a `{ Valid bool; Decimal }` struct, `domain/shipment.go:19-23`) for money/measures, read via `.Valid` then `.Decimal` (`bot/format.go:40`). `NullDecimal` is the shape of `Nullable<decimal>`. Money never uses `float64` — `shopspring/decimal` is the `System.Decimal` analog, chosen deliberately to avoid binary-float money bugs.

### Zero values vs null

Every type has a **zero value** for free — there's no "uninitialized." `string`→`""`, numbers→`0`, `bool`→`false`, pointers/interfaces/slices/maps→`nil`, structs→all fields zeroed. `var leads []domain.Lead` is immediately a usable (nil) slice.

This is **used as logic**: `if input.ClientID == uuid.Nil` (`service/shipment.go:80`) treats the all-zeroes UUID as "missing"; `telegramID == 0` and `currency == ""` mean "missing/default" elsewhere. The sharp edge: a **missing JSON field deserializes to the zero value silently** — there's no automatic "field was required" error, so required-field validation is manual. To recover the present/absent distinction, the codebase returns `(uuid.UUID, bool)` — e.g. `middleware.ManagerID` (`auth.go:60-63`), "защита от тихого uuid.Nil." That `, ok` bool is the idiom.

### Value vs pointer receivers

Each method declares whether it gets a *copy* (value receiver) or a *pointer*:
- **Pointer receiver** — `func (s *LeadService) Create(...)` (`service/lead.go:49`), `func (r *LeadRepository) Create(...)` (`repository/lead.go:23`). Shared instance, can mutate fields. Long-lived dependencies are always pointer receivers.
- **Value receiver** — `func (i CreateLeadInput) normalized()` (`service/lead.go:112`), `func (s LeadStatus) IsValid()` (`domain/lead.go:37`), `func (h LeadHandler) Create(...)` (`handlers/lead.go:22`). Gets a **copy**.

The instructive one: `normalized()` mutates its fields but, being a value receiver, mutates only the **copy** — so it returns the copy and the caller **must reassign**: `input = input.normalized()` (`service/lead.go:50`). Forget the reassignment and the change is silently discarded — no compiler error. Go structs are value types (like C# `struct`, not `class`); the only reference-like things are pointers/slices/maps.

### No inheritance — composition / embedding

Go has **no base classes and no inheritance**. Reuse is by **embedding** — putting one struct inside another with no field name, promoting its fields/methods. `token/token.go:16-19`:

```go
type Claims struct {
    jwt.RegisteredClaims        // embedded — no field name
    Role domain.Role `json:"role"`
}
```

`claims.Subject` / `claims.ExpiresAt` work as if inherited — but `Claims` is **not** an "is-a" `RegisteredClaims` (no subtyping, no virtual/override). A `*Claims` can't be passed where a `*RegisteredClaims` is expected. Polymorphism comes **only** from interfaces (§3), never a class hierarchy. Everywhere else it's plain composition by named fields (`ShipmentService` holds `store`, `clients`, `notifier` as fields — no `BaseService`).

### Third-party types for UUID and decimal

Go's stdlib has **no UUID and no decimal type**. Both come from modules, imported explicitly in every file that touches them (`domain/lead.go:6-7`): `github.com/google/uuid`, `github.com/shopspring/decimal`. Consequence: these are library types with **methods, not operators** — no `+` on a decimal (call `.Add(...)`), compare UUIDs with `== uuid.Nil` (not `Guid.Empty`). Upside: correctness. Cost: every consuming file imports the package and you lose operator sugar.

---

## 5. Traced path: `POST /api/leads` (landing form → manager Telegram notification)

This is the single best request to learn the backend — it touches every layer exactly once, plus the async Telegram side-channel.

**Hop 1 — route.** `router.go:41`: `api.POST("/leads", lead.Create)`. `gin.New()` (`router.go:25`) ≈ `WebApplication...Build()`; `router.Use(gin.Recovery())` (`router.go:26`) is the panic-catcher (Go's nearest thing to unhandled-exception middleware) — but it only protects code in the HTTP request goroutine, not the bot goroutine. `router.Group("/api")` ≈ `MapGroup`. The lead route is registered **before** the `manager` auth group (`router.go:46-47`), so it is **public** — no JWT. `lead.Create` is a *method value* bound to the handler (like passing a `RequestDelegate`).

**Hop 2 — handler (`handlers/lead.go:22-40`).**
- `var input service.CreateLeadInput` is **zero-valued, not null** — biggest mental shift.
- `c.ShouldBindJSON(&input)` (`lead.go:24`) is `[FromBody]` model binding driven by the `json:"…"` tags (`service/lead.go:38-47`). The **`&` is load-bearing**: without it Gin fills a copy and your `input` stays zero — no compile error, silently empty data.
- Binding returns an `error` (not a throw). The error funnel below (`lead.go:30-37`) is typed exception handling done with `errors.Is`: `ErrValidation` → 400, everything else → generic 500 (no internals leaked).
- `c.JSON(http.StatusCreated, lead)` (`lead.go:39`) ≈ `Results.Created(...)`, serializing via the struct's `json` tags.

**Hop 3 — service (`service/lead.go:49-73`).** `input = input.normalized()` (value-receiver copy, §4), then `validateCreateLead` (guard clauses returning `fmt.Errorf("%w: …", ErrValidation)`). Builds `lead := &domain.Lead{…}` as a **composite literal taking its address** (`new Lead { … }` + `&` so the repo can mutate it). It deliberately sets **no** `ID`/`Status`/`CreatedAt` — left at zero so Postgres fills its column defaults. The `store` field is the `LeadStore` interface (§3).

**Hop 4 — repository (`repository/lead.go:23-25`).** `r.db.WithContext(ctx).Create(lead).Error`. GORM ≈ `Add` + `SaveChanges` fused, scoped to the request's cancellation via `WithContext(ctx)`. **No change tracker, no `SaveChanges`** — the INSERT runs immediately and synchronously. GORM mutates `lead` in place to backfill `ID`/`status`/`created_at` (that's why we passed a pointer; comment at `repository/lead.go:21-22`), so the 201 response carries the new UUID. The DB owns the defaults (`migrations/000002_create_leads.up.sql`).

**Hop 5 — fire-and-forget notification (`service/lead.go:101-110`).** After the insert, `s.notifyNewLead(lead)` spawns `go func(){ … }()` — a **goroutine** (§6). Three subtleties: it uses `context.Background()` (NOT the request ctx) because the request ctx cancels the instant the 201 is written (comment at `lead.go:96-98`); errors are **log-only** via `slog.Error` (a failed notification must not fail the request — the lead is already persisted); and the `if s.notifier == nil` guard matters because interfaces can be nil.

**Hop 6 — the 201 returns BEFORE Telegram is attempted.** `handlers/lead.go:39` runs while the goroutine is still in flight. A 201 means "row inserted," NOT "manager notified."

**Hop 7 — async tail (`bot/bot.go:210` → `bot/format.go:32`).** `NotifyNewLead` calls `b.send(b.chatID, formatLeadMessage(lead))`. `formatLeadMessage` is a pure function using `strings.Builder` (your `StringBuilder`), reading nullable decimals via `if lead.Weight.Valid { … }` (`format.go:40`). `b.send` (`bot/bot.go:231`) returns early in no-token mode, and on failure returns a generic `fmt.Errorf("bot: send message failed")` **without** wrapping the library error — because that string can contain the bot token in a URL (deliberate secret redaction, matching `security.md`).

**The whole path:** `POST /api/leads` (`router.go:41`) → bind+funnel (`handlers/lead.go:22`) → normalize/validate/build (`service/lead.go:49`) → synchronous INSERT, DB fills defaults (`repository/lead.go:23`) → `go func()` with `context.Background()` (`service/lead.go:101`) → **201 returned without waiting** → concurrently `Bot.NotifyNewLead → formatLeadMessage → b.send → Telegram`.

---

## 6. Traced path: Telegram Mini App login (initData HMAC → JWT → RequireClientAuth)

**One sentence:** the WebApp POSTs Telegram-signed `init_data` to `POST /api/app/auth/telegram`; the backend re-derives the HMAC and **constant-time compares** it, upserts a `Client` by `telegram_id`, mints an HS256 JWT with `role=client`, and every later `/api/app/*` request is gated by `RequireClientAuth`, which parses the JWT and pins the caller's `clientID` into the request context.

**Step 0 — routes (`router.go:43, 64-71`).** Login is **public**; the client surface is `api.Group("/app")` + `app.Use(middleware.RequireClientAuth(deps.JWTSecret))`. This is `MapGroup("/app").RequireAuthorization("ClientOnly")` — but there's no hidden pipeline: `app.Use(...)` literally appends a function to a slice Gin calls in order. Note `gin.New()` has **no** default middleware; Recovery and CORS are added by hand (`router.go:25-27`).

**Step 1 — handler (`handlers/client_auth.go:25-43`).** Binds JSON (`init_data` via `json:"init_data"`), delegates to `AuthenticateWebApp` (which returns **three** values `(string, *domain.Client, error)` — completely normal in Go), and maps `service.ErrTelegramAuth` → **401**, everything else → 500.

**Step 2 — service (`service/client.go:103-121`).** Three responsibilities in order: validate signature (`telegram.ValidateInitData`), upsert client (`Register`), sign token (`token.Issue`). Wraps failures with `%w` so the handler's `errors.Is` matches across the package boundary. Dependencies are constructor-injected once at `main.go:54`; the repo holds one long-lived `*gorm.DB`, request-scoped per call via `.WithContext(ctx)`.

**Step 3 — HMAC validation (`telegram/initdata.go:95-121`), the security core.** No .NET framework does this for you — it's Telegram's documented scheme by hand:
1. Build the `data_check_string`: all keys except `hash`, **sorted**, joined as `key=value` lines with `\n`.
2. Derive the secret: `secret_key = HMAC_SHA256(key="WebAppData", msg=botToken)` — the constant is the key, the token is the message (counter-intuitive but per spec).
3. Sign the check string with that derived key, hex-encode.
4. **Constant-time compare:** `hmac.Equal(computed, hash)` (`initdata.go:120`). **Do not replace with `==`** — a normal compare short-circuits on the first differing byte and leaks timing, enabling byte-by-byte forgery. The .NET equivalent is `CryptographicOperations.FixedTimeEquals`.

Plus an empty-token reject (fail closed, `initdata.go:50-52`) and a 24h replay window via `auth_date` (`initDataMaxAge`, `client.go:23`). User JSON is parsed **only after** the signature passes, so unverified fields are never trusted.

**Step 4 — `Register`: idempotent upsert (`service/client.go:48-99`).** Look up by `telegram_id`; if found, return it; if `domain.ErrNotFound`, create. **Concurrency you'd rarely hand-write in EF:** if two first-logins race, the loser hits the unique index on `telegram_id` (`gorm:"uniqueIndex"`, `domain/client.go:12`) and the code **re-reads and returns the winner's row** instead of 500-ing. The in-app not-found check alone has a TOCTOU gap — the DB unique index is what makes it safe. `LeadID *uuid.UUID` is nil on the WebApp path, so the lead-binding retry branch is dead code on *this* route (it's for the bot `/start` deep-link path that shares `Register`).

**Step 5 — JWT issuance (`token/token.go:23-35`).** HS256 (**symmetric** — same `cfg.JWTSecret` signs and verifies). `Claims` embeds `jwt.RegisteredClaims` (§4 embedding) and adds `Role`. `Subject = client.ID` ≈ the `sub` / `ClaimTypes.NameIdentifier`. **Critical:** the *same* secret mints manager tokens (`main.go:53`) and client tokens (`main.go:54`), so the `Role` claim is the *only* thing separating the two audiences — the comment at `token.go:12-15` spells out the threat.

**Step 6 — `RequireClientAuth` (`middleware/auth.go:29-58`).** A **higher-order function**: `RequireClientAuth(secret)` returns a `gin.HandlerFunc` closure (evaluated once at boot, `router.go:65`). Per request: extract bearer via `strings.CutPrefix`; `token.Parse` verifies signature + `exp`; then `claims.Role != role` rejects a manager token despite a perfect signature; then `c.Set("clientID", subject)` stashes the identity. `token.Parse` (`token/token.go:40-56`) pins the algorithm in **both** the keyfunc (asserting `*jwt.SigningMethodHMAC`) and `WithValidMethods(["HS256"])` — together these kill the alg-confusion attack (`none`/`RS256` swap). `WithExpirationRequired()` rejects tokens with no `exp`.

**Compared to ASP.NET auth:** two stages (`AddJwtBearer` builds `HttpContext.User`, then `[Authorize(Roles=…)]`) are **fused into one `gin.HandlerFunc`**. Consequences that bite:
- **No ambient `User`/`ClaimsPrincipal`.** Downstream code calls `middleware.ClientID(c)` (`auth.go:67`), which does `c.Get("clientID")` and type-asserts to `uuid.UUID`, returning `(uuid.UUID, bool)`.
- **The `ok` bool is a real guardrail** (`auth.go:60-61`): a handler mounted outside the protected group gets `ok=false` → 401, instead of silently operating as `uuid.Nil`.
- **Context values are untyped (`any`)**, hence the type assertion + bool.

**Net effect — tenant isolation:** the client id comes **only** from the verified JWT, never from URL or body. So `GET /api/app/shipments/:id` can't read another client's shipment — `DetailForClient(ctx, id, clientID)` (`app_shipment.go:59`) filters by the trusted `clientID`. The middleware is what makes that id trustworthy.

---

## 7. The biggest mental-model shifts (quick reference)

These are catalogued in detail above; this is your "unlearn the C# instinct" cheat sheet — see `mental_model_shifts` for the condensed list.

---

## Appendix: idiom → `file:line` index

- Hand-wired DI / composition root: `cmd/api/main.go:40-69`
- Consumer-defined structural interface: `service/lead.go:20-25`, satisfied by `repository/lead.go:23,27,36,51`, asserted at `main.go:52`
- One value, four interfaces: `bot/bot.go:18-20,210-225`; `main.go:56`
- Error-as-value + funnel: `handlers/lead.go:30-37,90-99`; `%w` at `service/lead.go:124`; sentinel `domain/errors.go:8`
- Repo error translation: `repository/lead.go:39-41` (read), `repository/lead.go:59-61` (write)
- GORM model tags: `domain/lead.go:12-24`; `json:"-"` guard `domain/manager.go:14`
- Transaction closure: `repository/shipment.go:23-38`; map-update gotcha `repository/shipment.go:96-124`
- Enum triple: `domain/lead.go:26-44`; CHECK in `migrations/000004_create_shipments.up.sql`
- Migrations runner: `cmd/migrate/main.go`
- Goroutine fire-and-forget: `service/lead.go:101-110`; root ctx `main.go:21`; shutdown `main.go:85-90`
- Nullability via pointer / `NullDecimal`: `domain/shipment.go:19,25-26`
- Value vs pointer receiver: `service/lead.go:49` (ptr) vs `service/lead.go:112` (value)
- Embedding: `token/token.go:16-19`
- HMAC + constant-time compare: `telegram/initdata.go:95-121`
- JWT alg-pinning: `token/token.go:40-56`
- Middleware as higher-order function: `middleware/auth.go:29-58`

---

## Сдвиги мышления — шпаргалка «разучиться писать как в C#»

1. Errors are ordinary return values, not exceptions. Every fallible call returns (value, error) and you MUST check err on the next line — there is no try/catch and no stack unwinding. Sentinel values (domain.ErrNotFound) matched by errors.Is are the 'exception types'; %w is the InnerException link, and the handler's switch over errors.Is IS your ExceptionFilter, written by hand (service/lead.go:124, handlers/lead.go:90-99). Forgetting to check err compiles fine and silently proceeds with a zero value.

2. Interfaces are implicit and declared by the CONSUMER, not the implementer. There is no ': ILeadStore' anywhere — a type satisfies an interface just by having matching methods, verified at the assignment site (main.go:52). The contract lives in the service that needs it (service/lead.go:20), kept tiny, and the repo/bot never import or name it. This is Dependency Inversion without a DI container; one *bot.Bot satisfies four notifier interfaces it has never heard of (bot/bot.go:18-20).

3. There is no null for value types — there are zero values, and a missing JSON field silently becomes one. A freshly declared struct has '' / 0 / false / nil fields; the code USES uuid.Nil / '' / 0 as 'absent' (service/shipment.go:80). Only pointers/slices/maps/interfaces can be nil. Nullability for value types is expressed by switching to a pointer (*uuid.UUID ≈ Guid?) or decimal.NullDecimal (≈ decimal?) — and dereferencing a nil pointer panics with NO compiler warning, unlike C# nullable-reference analysis.

4. There is no DI container and no per-request lifetime. The entire object graph is hand-wired once in main.go (40-69), in dependency order the compiler enforces, and every repo/service is effectively a process-wide singleton. The *gorm.DB is one shared, thread-safe, pooled handle with NO change tracker and NO SaveChanges — every write is explicit and immediate (repository/lead.go:24). Per-request scope/cancellation is carried by context.Context passed by hand into every method, not by new objects.

5. GORM is not EF code-first and the schema is NOT generated from your structs. Struct tags are documentation that MIRRORS the hand-written SQL migrations (golang-migrate); there is no AutoMigrate, no model diffing, no OnModelCreating. Relationships, FKs, ON DELETE behavior, CHECK constraints, and enums live in the migration SQL, and you author both the up.sql and down.sql yourself. Struct + migration are two sources of truth synced by a human; drift fails at runtime, not compile time.

6. Concurrency is goroutines + context.Context + channels, not async/await/Task. 'go f()' is fire-and-forget with NO awaitable handle (service/lead.go:105); there is no function coloring. context.Context is the CancellationToken analogue (plus deadlines and request-scoped values), passed explicitly as the first arg everywhere — there is no ambient HttpContext. Deliberately swapping the request ctx for context.Background() in a background goroutine is correct here, because the request ctx cancels the instant the HTTP response is written.

7. Capitalization IS the access modifier and a directory IS the package. Uppercase identifier = public (exported), lowercase = package-private — no public/private/internal keywords to grep for. All files in one directory share a flat scope and see each other's unexported names with no import. The internal/ directory is a compiler-enforced visibility firewall keyed off the directory path, not an attribute.

8. There is no inheritance. Reuse is composition: embedding (token/token.go:16 promotes RegisteredClaims' fields but is NOT subtyping — no virtual/override, no is-a) and interfaces for all polymorphism. Methods choose value vs pointer receiver explicitly, which changes behavior: a value-receiver method mutates only a COPY, so normalized() must return it and the caller must reassign (service/lead.go:50,112) — forgetting the reassignment silently discards the change.

---

## Маршрут чтения (по порядку)

1. backend/cmd/api/main.go — the whole wiring story in one file; this IS your Program.cs + Startup, read top to bottom first.
2. backend/internal/http/router.go — the route table + middleware groups; shows every endpoint and which auth posture (public / manager / client) gates it.
3. backend/internal/domain/lead.go — the cleanest entity: struct tags (gorm + json), the typed-string-enum pattern (LeadStatus + IsValid). Your EF entity, demystified.
4. backend/internal/domain/errors.go — the domain.ErrNotFound sentinel; tiny, but it's the spine of all error handling.
5. backend/internal/http/handlers/lead.go — controller actions: ShouldBindJSON model binding and the errors.Is → status-code funnel (your ExceptionFilter, made explicit).
6. backend/internal/service/lead.go — the consumer-defined LeadStore interface, %w error wrapping, and the fire-and-forget notification goroutine. The single richest file for Go idioms.
7. backend/internal/repository/lead.go — the thinnest GORM repo: Create/First/Update, gorm.ErrRecordNotFound → domain.ErrNotFound, and RowsAffected==0 for writes.
8. backend/internal/db/postgres.go — the 'DbContext factory': one long-lived, shared, thread-safe *gorm.DB with no change tracker.
9. backend/internal/service/notifier.go + backend/internal/bot/bot.go (focus bot.go:18-20, 210-225) — how one *bot.Bot satisfies four service-side notifier interfaces structurally, importing nothing from service.
10. backend/internal/domain/shipment.go — relations as bare FK UUIDs (no nav properties), nullable pointers (*time.Time), NullDecimal money, and a second enum triple.
11. backend/internal/repository/shipment.go — the closure-based transaction (Create shipment + initial status event atomically) and the map-update-doesn't-refresh-the-struct gotcha.
12. backend/internal/middleware/auth.go — middleware as a higher-order function returning a gin.HandlerFunc; role enforcement and the (uuid.UUID, bool) identity guardrail (there is no ambient HttpContext.User).
13. backend/internal/token/token.go — HS256 issue/parse, struct embedding of RegisteredClaims, and alg-pinning that blocks the JWT alg-confusion attack.
14. backend/internal/telegram/initdata.go — the hand-rolled Telegram HMAC validation and the constant-time hmac.Equal compare (read this once you're comfortable; it's the security core).
15. backend/cmd/migrate/main.go + backend/migrations/000002_create_leads.up.sql — how hand-written paired SQL migrations are applied (no AutoMigrate, no model diffing); confirm the struct tags mirror the SQL by hand.

---

## Первое практическое задание

GOAL: add a new manager-authenticated endpoint `GET /api/clients/:id` that returns a single client by id, mirroring the existing `GET /api/leads/:id` flow. This is the safest possible first real Go task: it touches every backend layer (route → handler → service → consumer-interface → repository → domain.ErrNotFound → 404) but needs NO migration and NO schema change, because the repository method and the DB table already exist. Estimated time for a Go beginner: one focused session.

WHY IT WORKS OUT OF THE BOX: `ClientRepository.GetByID(ctx, id)` already exists at backend/internal/repository/client.go:25 (it's used internally by Register), and the `clients` table is already there. What's missing is only the plumbing above it: the consumer-side interface entry, the service method, the handler, and one route line. You will literally watch the structural-interface idiom work — add one method to the interface, and the repo satisfies it for free.

READ FIRST (in this order, ~20 min): (1) backend/internal/http/handlers/lead.go:52-70 — `LeadHandler.GetByID`: uuid.Parse(c.Param('id')), the errors.Is(domain.ErrNotFound) → 404 funnel, c.JSON(200). This is your exact template. (2) backend/internal/service/lead.go:79-81 — `LeadService.GetByID` delegating to the store. (3) backend/internal/service/client.go:25-31 — the `ClientStore` interface (currently GetByTelegramID, Create, List — NOTE it does NOT yet declare GetByID). (4) backend/internal/http/handlers/client.go — `ClientHandler` (has List, no GetByID). (5) backend/internal/http/router.go:49-53 — how `manager.GET('/leads/:id', lead.GetByID)` is registered under the RequireAuth group, and where `manager.GET('/clients', client.List)` sits.

STEPS:
1. In backend/internal/service/client.go, add `GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)` to the `ClientStore` interface (around line 27-31). The repo already implements it, so this compiles immediately — that's the structural-interface lesson.
2. In the same file, add a `func (s *ClientService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error)` that just returns `s.clients.GetByID(ctx, id)` — mirror `LeadService.GetByID` (service/lead.go:79-81). (uuid is already imported in client.go.)
3. In backend/internal/http/handlers/client.go, add `func (h ClientHandler) GetByID(c *gin.Context)` copied from `LeadHandler.GetByID` (handlers/lead.go:52-70), swapping the service call to `h.service.GetByID(...)` and the not-found message to 'client not found'. You'll need to add the `errors`, `github.com/google/uuid`, and `icaris-logistic/backend/internal/domain` imports (copy the import block from handlers/lead.go).
4. In backend/internal/http/router.go, add `manager.GET('/clients/:id', client.GetByID)` next to the existing `manager.GET('/clients', client.List)` line (router.go:53), so it inherits manager-JWT auth.

VERIFY:
- Build: from the backend directory run `go build ./...` — it must compile with zero errors. (If step 1's interface entry didn't match the repo's signature exactly, the failure shows up here, at the wiring, not at the repo — exactly the gotcha described in the guide.)
- Vet: `go vet ./...` should be clean.
- Run it: start Postgres + the API (`go run ./cmd/api`), mint a manager token via the existing `POST /api/auth/login` (seed a manager with `go run ./cmd/createmanager` if needed), then `curl -H 'Authorization: Bearer <token>' http://localhost:<port>/api/clients` to grab a real client id, and `curl -H 'Authorization: Bearer <token>' http://localhost:<port>/api/clients/<that-id>` — expect 200 with the client JSON.
- Negative paths to confirm you wired the funnel correctly: a random valid-format UUID → 404 `{ error: { code: 'not_found', ... } }`; a malformed id like `abc` → 400 `invalid_id`; the same request with NO Authorization header → 401 (proving it's under the manager group).
- Optional stretch (only after the above is green): write a table-driven unit test for `ClientHandler.GetByID` or `ClientService.GetByID` next to the existing `*_test.go` files, asserting the not-found case maps to domain.ErrNotFound — per the project's testing rules (verify behavior/output, one assertion per test, no mock-call-count assertions).
