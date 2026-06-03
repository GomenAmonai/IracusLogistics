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

Health check:

```bash
curl http://localhost:8080/api/health
```
