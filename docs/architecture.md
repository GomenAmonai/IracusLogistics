# IcarisLogistics Architecture

## Назначение Документа

Этот документ описывает архитектуру продукта для сервиса доставки из Китая в Россию.

Цель документа:

- зафиксировать понимание предметной области;
- показать связи между ролями, процессами и данными;
- отделить MVP от будущей платформы;
- служить учебной картой для проектирования backend, frontend и базы данных.

Документ должен обновляться по мере уточнения бизнес-процесса.

## Продуктовая Идея

Мы строим не просто калькулятор доставки, а платформу для обработки запросов на перевозку груза из Китая.

Публичный сайт и калькулятор являются входной точкой. Центральная сущность системы - заявка на перевозку.

```text
Клиент хочет:
- понять, можно ли доставить его товар;
- получить ориентир по цене и срокам;
- передать данные без долгой переписки;
- видеть понятный процесс.

Менеджер хочет:
- не терять заявки;
- быстро понимать вводные;
- запросить ставку у китайского партнера;
- подготовить предложение;
- вести клиента до завершения доставки.

Китайский партнер хочет:
- получить структурированный запрос;
- дать ставку;
- обновлять статусы;
- прикладывать документы, фото и комментарии.
```

## Роли

```mermaid
flowchart LR
    Visitor["Visitor\nпосетитель сайта"]
    Client["Client\nклиент"]
    Manager["Manager\nменеджер"]
    Partner["Partner\nкитайский партнер"]
    Admin["Admin\nадминистратор системы"]

    Visitor -->|"оставляет заявку"| Client
    Client -->|"получает расчет / статус"| Manager
    Manager -->|"запрашивает ставку"| Partner
    Partner -->|"возвращает цену / статус"| Manager
    Admin -->|"настраивает тарифы, услуги, пользователей"| Manager
```

В MVP реально реализуем роли:

- `Visitor`;
- `Manager`;
- `Admin` в упрощенном виде.

Роли `Client` и `Partner` проектируются заранее, но полноценные кабинеты для них не входят в первый MVP.

## Главный Поток MVP

```mermaid
sequenceDiagram
    participant V as Visitor
    participant Web as React App
    participant API as Go API
    participant DB as PostgreSQL
    participant TG as Telegram
    participant M as Manager

    V->>Web: Заполняет форму заявки
    Web->>API: POST /shipment-requests
    API->>DB: Сохраняет клиента, груз, маршрут, заявку
    API->>DB: Добавляет статус new
    API->>TG: Отправляет уведомление
    TG->>M: Новая заявка
    M->>API: Открывает админку
    API->>DB: Загружает список заявок
    M->>API: Меняет статус / добавляет комментарий / расчет
    API->>DB: Сохраняет изменения
```

## Границы MVP

MVP должен закрыть один сквозной процесс:

```text
Клиент оставил заявку
-> система сохранила структурированные данные
-> менеджер получил уведомление
-> менеджер увидел заявку в админке
-> менеджер сменил статус
-> менеджер добавил комментарий или предварительный расчет
```

Входит в MVP:

- публичная страница с описанием услуги;
- форма заявки;
- backend API на Go;
- PostgreSQL;
- сохранение заявок;
- простая админка;
- статусы заявки;
- внутренние комментарии;
- Telegram-уведомление;
- предварительный расчет вручную или полуавтоматически.

Не входит в MVP:

- полноценный личный кабинет клиента;
- кабинет китайского партнера;
- интеграция с 1С;
- интеграция с WeChat;
- точный тарифный движок;
- онлайн-оплата;
- автоматический трекинг;
- документооборот;
- сложная аналитика.

## Слои Системы

```mermaid
flowchart TB
    subgraph Frontend["Frontend: React + TypeScript"]
        PublicSite["Публичный сайт"]
        RequestForm["Форма заявки"]
        AdminUI["Админка"]
    end

    subgraph Backend["Backend: Go API"]
        HTTP["HTTP handlers"]
        Services["Application services"]
        Domain["Domain model"]
        Repos["Repositories"]
        Notifications["Notifications"]
        Integrations["Integrations"]
    end

    subgraph Data["Data layer"]
        DB[("PostgreSQL")]
        Files[("File storage\nfuture")]
    end

    subgraph External["External services"]
        Telegram["Telegram Bot"]
        Email["Email\nfuture"]
        CRM["CRM / 1C / WeChat\nfuture"]
    end

    PublicSite --> HTTP
    RequestForm --> HTTP
    AdminUI --> HTTP

    HTTP --> Services
    Services --> Domain
    Services --> Repos
    Services --> Notifications
    Services --> Integrations
    Repos --> DB
    Repos --> Files
    Notifications --> Telegram
    Notifications --> Email
    Integrations --> CRM
```

## Backend Слои

Backend не должен превращаться в набор случайных обработчиков.

Базовый поток:

```text
HTTP handler
-> application service
-> domain logic
-> repository
-> database
```

Предлагаемая структура:

```text
cmd/api
internal/config
internal/http
internal/domain
internal/service
internal/repository
internal/db
internal/notification
internal/integration
```

Ответственность слоев:

- `http`: принимает HTTP-запросы, парсит JSON, возвращает ответы;
- `service`: выполняет бизнес-сценарии;
- `domain`: хранит основные типы и правила предметной области;
- `repository`: работает с базой данных;
- `notification`: отправляет Telegram/email;
- `integration`: будущие интеграции с CRM, 1С, WeChat.

## Центральная Domain-Модель

Главная сущность:

```text
ShipmentRequest
```

Это не просто лид и не просто результат формы. Это запрос на перевозку, который может пройти путь от первой заявки до доставки.

Основные сущности:

```text
Customer
ShipmentRequest
Cargo
Route
Estimate
EstimateItem
ServiceOption
StatusHistory
InternalComment
FileAttachment
```

## Связи Сущностей

```mermaid
erDiagram
    CUSTOMER ||--o{ SHIPMENT_REQUEST : creates
    SHIPMENT_REQUEST ||--|| CARGO : contains
    SHIPMENT_REQUEST ||--|| ROUTE : has
    SHIPMENT_REQUEST ||--o{ ESTIMATE : has
    ESTIMATE ||--o{ ESTIMATE_ITEM : contains
    SHIPMENT_REQUEST ||--o{ STATUS_HISTORY : tracks
    SHIPMENT_REQUEST ||--o{ INTERNAL_COMMENT : has
    SHIPMENT_REQUEST ||--o{ FILE_ATTACHMENT : includes
    SHIPMENT_REQUEST }o--o{ SERVICE_OPTION : requests

    CUSTOMER {
        uuid id
        text name
        text phone
        text email
        text telegram
        text company_name
        timestamp created_at
    }

    SHIPMENT_REQUEST {
        uuid id
        uuid customer_id
        text status
        text source
        text comment
        uuid manager_id
        timestamp created_at
        timestamp updated_at
    }

    CARGO {
        uuid id
        uuid shipment_request_id
        text name
        text description
        text category
        numeric declared_value
        text currency
        numeric weight_kg
        numeric volume_m3
        int boxes_count
        bool has_batteries
        bool has_liquids
        bool is_branded
    }

    ROUTE {
        uuid id
        uuid shipment_request_id
        text origin_country
        text origin_city
        text destination_country
        text destination_city
    }

    ESTIMATE {
        uuid id
        uuid shipment_request_id
        numeric price_from
        numeric price_to
        text currency
        text confidence
        bool requires_manager_review
        timestamp created_at
    }

    ESTIMATE_ITEM {
        uuid id
        uuid estimate_id
        text title
        numeric amount
        text currency
        text comment
    }

    STATUS_HISTORY {
        uuid id
        uuid shipment_request_id
        text from_status
        text to_status
        uuid changed_by
        timestamp created_at
    }

    INTERNAL_COMMENT {
        uuid id
        uuid shipment_request_id
        uuid author_id
        text body
        timestamp created_at
    }

    FILE_ATTACHMENT {
        uuid id
        uuid shipment_request_id
        text file_name
        text file_url
        text file_type
        timestamp created_at
    }

    SERVICE_OPTION {
        uuid id
        text code
        text title
        text description
        bool is_active
    }
```

## Жизненный Цикл Заявки

Статус заявки - это не декоративное поле, а описание бизнес-процесса.

MVP-статусы:

```text
new
needs_clarification
in_calculation
priced
offer_sent
won
lost
```

```mermaid
stateDiagram-v2
    [*] --> new
    new --> needs_clarification
    new --> in_calculation
    needs_clarification --> in_calculation
    in_calculation --> priced
    priced --> offer_sent
    offer_sent --> won
    offer_sent --> lost
    needs_clarification --> lost
    in_calculation --> lost
    won --> [*]
    lost --> [*]
```

Будущие статусы после MVP:

```text
accepted
payment_pending
paid
cargo_expected_at_warehouse
cargo_received_in_china
inspection
consolidation
in_transit
customs_clearance
arrived_in_russia
last_mile_delivery
delivered
completed
```

## Estimate И Quote

Важно разделять предварительный расчет и финальное предложение.

```text
Estimate
```

Ориентировочный расчет. Может быть автоматическим или полуавтоматическим. Не является обещанием финальной цены.

```text
Quote
```

Финальное коммерческое предложение, подтвержденное менеджером. В MVP можно не делать отдельную сущность `Quote`, но архитектурно держим это различие.

```mermaid
flowchart LR
    Request["ShipmentRequest"]
    Estimate["Estimate\nпредварительный расчет"]
    Quote["Quote\nподтвержденное предложение"]
    Deal["Deal / Shipment\nбудущая операционная сущность"]

    Request --> Estimate
    Estimate --> Quote
    Quote --> Deal
```

## Данные Формы Заявки

Форма на сайте может быть короче, чем структура в базе.

Обязательные поля для первого MVP:

- имя;
- телефон или мессенджер;
- название груза;
- город отправки в Китае, если известен;
- город доставки в России;
- вес или объем;
- комментарий.

Желательные поля:

- email;
- название компании;
- стоимость груза;
- валюта;
- количество коробок;
- нужна ли помощь с выкупом;
- нужна ли таможня;
- есть ли поставщик;
- фото, инвойс или packing list.

Поля, которые не стоит требовать сразу:

- точный ТН ВЭД;
- полные габариты каждой коробки;
- юридические реквизиты;
- детальная структура документов.

Принцип:

```text
Форма должна быть достаточно простой для клиента,
но backend должен хранить данные структурно.
```

## Будущая Платформа

MVP должен быть первым слоем будущей системы, а не одноразовой формой.

```mermaid
flowchart TB
    MVP["MVP\nзаявки, админка, уведомления"]
    V2["Version 2\nстатусы, расчеты, история"]
    V3["Version 3\nтарифы, клиентский кабинет"]
    V4["Version 4\nпартнерский кабинет, трекинг"]
    V5["Version 5\nCRM, 1C, WeChat, аналитика"]

    MVP --> V2
    V2 --> V3
    V3 --> V4
    V4 --> V5
```

## Открытые Вопросы

Эти вопросы нужно уточнять постепенно, не до начала разработки целиком:

- Какие услуги точно оказываем в первой версии?
- Какие маршруты наиболее важны?
- Как китайский партнер сейчас считает цену?
- Какие данные партнеру нужны для ставки?
- Кто будет менеджером в системе?
- Нужно ли хранить документы в MVP?
- Какой канал уведомлений первый: Telegram, email или оба?
- Нужно ли сразу делать авторизацию для админки?
- Будет ли сайт на одном языке или нужен китайский/английский позже?

## Ближайший Следующий Шаг

Следующий архитектурный шаг - подробно разобрать `ShipmentRequest`:

- какие поля обязательные;
- какие поля опциональные;
- какие поля заполняет клиент;
- какие поля заполняет менеджер;
- какие поля нужны только будущим версиям;
- какие таблицы появятся в первой миграции.

