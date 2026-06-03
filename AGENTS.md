# AGENTS.md — Iracus Logistics

## Контекст проекта

Логистический сервис Китай → Россия. B2B платформа для управления грузами.
Учебный + коммерческий проект. Стек выбран осознанно, не менять без обсуждения.

## Стек

- Backend: Go + Gin + GORM + PostgreSQL
- Frontend: React + TypeScript + Vite + Telegram WebApp SDK
- Bot: go-telegram-bot-api (long polling на старте)
- Infra: Docker + Nginx

## Структура проекта

backend/
├── cmd/api/main.go
├── internal/
│ ├── config/
│ ├── domain/ # сущности: Manager, Client, Lead, Shipment, Message
│ ├── repository/ # GORM репозитории
│ ├── service/ # бизнес-логика
│ ├── handler/ # Gin хендлеры
│ ├── middleware/ # JWT, Telegram auth
│ └── bot/ # Telegram bot
├── migrations/ # SQL файлы миграций
frontend/ # React WebApp

## Схема БД

### Manager

- id UUID PK
- email VARCHAR UNIQUE NOT NULL
- password VARCHAR NOT NULL (bcrypt)
- name VARCHAR NOT NULL
- created_at TIMESTAMP

### Lead (заявка с сайта, до клиента)

- id UUID PK
- name VARCHAR NOT NULL
- phone VARCHAR NOT NULL
- from_city VARCHAR NOT NULL
- to_city VARCHAR NOT NULL
- weight DECIMAL
- volume DECIMAL
- cargo_type VARCHAR
- comment TEXT
- status ENUM(new, contacted, converted, rejected)
- created_at TIMESTAMP

### Client (Telegram auth)

- id UUID PK
- telegram_id BIGINT UNIQUE NOT NULL
- username VARCHAR
- name VARCHAR NOT NULL
- phone VARCHAR
- created_at TIMESTAMP

### Shipment

- id UUID PK
- client_id UUID FK → Client
- manager_id UUID FK → Manager
- tracking_key VARCHAR UNIQUE NOT NULL
- status ENUM(pending, picked_up, in_transit, customs_clear, in_warehouse, out_for_delivery, delivered, cancelled)
- status_comment TEXT
- weight DECIMAL
- volume DECIMAL
- from_city VARCHAR
- to_city VARCHAR
- price DECIMAL
- currency VARCHAR DEFAULT 'USD'
- estimated_at TIMESTAMP
- delivered_at TIMESTAMP
- created_at TIMESTAMP
- updated_at TIMESTAMP

### Message

- id UUID PK
- shipment_id UUID FK → Shipment (nullable)
- client_id UUID FK → Client
- manager_id UUID FK → Manager (nullable)
- text TEXT NOT NULL
- from_role ENUM(client, manager)
- created_at TIMESTAMP

## Пользовательский flow

1. Клиент заходит на сайт → калькулятор → видит диапазон цены
2. Оставляет заявку (Lead) — без регистрации
3. Менеджер получает уведомление в боте → считает точную цену → пишет клиенту
4. Клиент соглашается → менеджер отправляет ссылку на бота
5. Клиент сам проходит верификацию через Telegram → аккаунт создаётся автоматически
6. Для каждого груза генерируется tracking_key
7. Клиент отслеживает грузы через WebApp и общается с менеджером в боте
8. Быстрые команды (/status) работают через telegram_id — без ключа

## Решения и их причины

- Мультитенантности нет — продукт под одну компанию
- tracking_key на Shipment, не на Client — у клиента может быть несколько грузов
- Аккаунт создаётся только после подтверждения сделки — база остаётся чистой
- Long polling для бота на старте — webhook настроим при деплое
- Миграции SQL файлами, не EnsureSchema — контроль над схемой
- Язык интерфейса: русский

## Текущее состояние

Существующий код (backend/internal/shipment/) — базовый скелет от Codex.
Структуру слоёв сохраняем, доменную модель переписываем полностью.
Приоритет сейчас: домен → репозитории → сервисы → хендлеры → бот.

## Правила работы

- Не менять стек без явного указания
- Не добавлять зависимости без необходимости
- Комментарии на русском
- Коммиты атомарные, по одной логической единице
