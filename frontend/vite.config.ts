import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

// Прокси /api → Go API на :8080, чтобы публичная форма и калькулятор работали
// локально без CORS-плясок: фронт ходит на относительный /api, vite проксирует.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
