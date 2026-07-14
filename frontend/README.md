# IcarisLogistics Frontend

React + TypeScript + Vite. Три точки входа одного проекта, общая дизайн-система (`src/index.css`):

- `index.html` — публичный лендинг (`src/main.tsx`, `src/components/*`)
- `webapp.html` — Telegram Mini App (`src/webapp/*`)
- `manager.html` — панель менеджера (`src/manager/*`)

## Команды

```bash
npm install
npm run dev      # http://localhost:5173 (лендинг), /webapp.html (Mini App), /manager.html (панель); /api → :8080
npm run build    # tsc -b && vite build → dist/ (все три точки входа)
npm run lint     # tsc -b --pretty false
```

## Конфигурация

`VITE_API_BASE` — базовый URL бэкенда (без пути; код сам добавляет `/api`).

- Локально: **не задавать** — запросы идут на относительный `/api`, который Vite проксирует на `:8080`.
- При раздельном деплое: задать `VITE_API_BASE` на публичный backend (домен фронта должен быть в `ALLOWED_ORIGINS` бэкенда; авторизация по Bearer — куки не используются).

Публичная форма заявки также требует:

- `VITE_PRIVACY_POLICY_URL` — URL опубликованной политики обработки данных;
- `VITE_PRIVACY_NOTICE_VERSION` — версия документа, сохраняемая с заявкой.

Без обеих переменных форма остаётся выключенной.

Полная картина (стек, запуск бэкенда, деплой, API) — в [корневом README](../README.md).
