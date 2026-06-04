// Тонкая обёртка над Telegram WebApp SDK (telegram-web-app.js, подключён в webapp.html).
// Вне Telegram (обычный браузер при разработке) window.Telegram отсутствует — функции
// деградируют в no-op, а initData пустой, поэтому авторизация уходит в dev-ветку.

type BackButton = {
  show: () => void
  hide: () => void
  onClick: (cb: () => void) => void
  offClick: (cb: () => void) => void
}

type HapticFeedback = {
  impactOccurred: (style: 'light' | 'medium' | 'heavy') => void
  notificationOccurred: (type: 'error' | 'success' | 'warning') => void
}

type TelegramWebApp = {
  initData: string
  ready: () => void
  expand: () => void
  colorScheme: 'light' | 'dark'
  BackButton: BackButton
  HapticFeedback?: HapticFeedback
}

declare global {
  interface Window {
    Telegram?: { WebApp?: TelegramWebApp }
  }
}

export function getWebApp(): TelegramWebApp | undefined {
  return window.Telegram?.WebApp
}

export function getInitData(): string {
  return getWebApp()?.initData ?? ''
}

// ready/expand сообщают Telegram, что интерфейс готов, и разворачивают мини-апп на весь
// экран. Безопасны вне Telegram (no-op).
export function initWebApp(): void {
  const app = getWebApp()
  app?.ready()
  app?.expand()
}

// useBackButton показывает нативную кнопку «Назад» Telegram, пока активен экран деталей,
// и вешает на неё обработчик. Возвращает функцию-отписку.
export function bindBackButton(onClick: () => void): () => void {
  const button = getWebApp()?.BackButton
  if (!button) {
    return () => {}
  }
  button.onClick(onClick)
  button.show()
  return () => {
    button.offClick(onClick)
    button.hide()
  }
}

export function haptic(type: 'success' | 'error'): void {
  getWebApp()?.HapticFeedback?.notificationOccurred(type)
}
