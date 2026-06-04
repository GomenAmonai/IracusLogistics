/// <reference types="vite/client" />

interface ImportMetaEnv {
  // Базовый URL backend для WebApp на задеплоенном фронте. Пусто локально (прокси Vite).
  readonly VITE_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
