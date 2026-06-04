# IcarisLogistics Frontend

React + TypeScript + Vite. Две точки входа одного проекта, общая дизайн-система (`src/index.css`):

- `index.html` — публичный лендинг (`src/main.tsx`, `src/components/*`)
- `webapp.html` — Telegram Mini App (`src/webapp/*`)

## Команды

```bash
npm install
npm run dev      # http://localhost:5173 (лендинг), /webapp.html (Mini App); /api → :8080
npm run build    # tsc -b && vite build → dist/ (обе точки входа)
npm run lint     # tsc -b --pretty false
```

## Конфигурация

`VITE_API_BASE` — базовый URL бэкенда (без пути; код сам добавляет `/api`).

- Локально: **не задавать** — запросы идут на относительный `/api`, который Vite проксирует на `:8080`.
- На задеплоенном фронте (Vercel): задать `VITE_API_BASE` на публичный backend (бэкенд отдаёт CORS `*`, авторизация по Bearer — куки не используются).

Полная картина (стек, запуск бэкенда, деплой, API) — в [корневом README](../README.md).
