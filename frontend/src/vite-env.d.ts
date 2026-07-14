/// <reference types="vite/client" />

interface ImportMetaEnv {
  // Базовый URL backend для WebApp на задеплоенном фронте. Пусто локально (прокси Vite).
  readonly VITE_API_BASE?: string
  // Публичная политика и её стабильная версия обязательны для включения формы заявок.
  readonly VITE_PRIVACY_POLICY_URL?: string
  readonly VITE_PRIVACY_NOTICE_VERSION?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
