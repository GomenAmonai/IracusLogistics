import { resolve } from 'node:path'

import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

// Две точки входа: index.html — лендинг, webapp.html — Telegram Mini App. Это «отдельные
// приложения» (свой HTML, свой React-root, без общего роутинга), но дизайн-система общая —
// обе тянут src/index.css. Прокси /api → Go API на :8080: фронт ходит на относительный
// /api, vite проксирует, без CORS-плясок локально.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        webapp: resolve(__dirname, 'webapp.html'),
      },
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
