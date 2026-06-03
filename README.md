# Iracus Logistic

Учебный и рабочий проект сервиса заявок на доставку товаров из Китая.

## Stack

- Backend: Go
- Frontend: React + TypeScript + Vite
- Database: PostgreSQL
- Local environment: Docker Compose

## Local Run

Start PostgreSQL:

```bash
docker compose up -d postgres
```

Run backend:

```bash
cd backend
cp .env.example .env
go run ./cmd/api
```

Run frontend:

```bash
cd frontend
npm install
npm run dev
```

## Current MVP

- public form for shipment requests
- Postgres persistence
- admin list with status and manager comment

## API

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/shipment-requests
```

```http
POST /api/shipment-requests
GET /api/shipment-requests
GET /api/shipment-requests/{id}
PATCH /api/shipment-requests/{id}
```

Frontend reads the backend URL from `VITE_API_URL` and defaults to `http://localhost:8080`.
