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

If commands need to use the local VPN proxy from Happ Tunnel, run them through
the project wrapper:

```bash
scripts/with-proxy.sh npm --prefix frontend install
scripts/with-proxy.sh npm --prefix frontend run build
scripts/with-proxy.sh go -C backend mod download
scripts/with-proxy.sh curl -I https://example.com
```

By default the wrapper uses `127.0.0.1:10808`, which is exposed by Happ Tunnel
as HTTP CONNECT and SOCKS5. Override it when needed:

```bash
PROXY_PORT=10820 scripts/with-proxy.sh curl -I https://example.com
```

Health check:

```bash
curl http://localhost:8080/api/health
```
# IracusLogistics
