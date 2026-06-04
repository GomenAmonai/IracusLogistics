import type { ReactNode } from 'react'

// Мелкие атомы WebApp: экран-обёртка, спиннер, состояние ошибки/пустоты. Держим в одном
// файле — каждый слишком мал для отдельного.

export function Screen({ children }: { children: ReactNode }) {
  return <div className="mx-auto flex min-h-screen w-full max-w-screen-sm flex-col px-4 pb-10 pt-5">{children}</div>
}

export function Spinner() {
  return (
    <span
      role="status"
      aria-label="Загрузка"
      className="inline-block size-6 animate-spin rounded-full border-2 border-line border-t-accent"
    />
  )
}

export function CenteredState({
  title,
  description,
  action,
}: {
  title: string
  description?: string
  action?: ReactNode
}) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
      <h2 className="font-display text-lg font-semibold text-ink">{title}</h2>
      {description && <p className="text-sm leading-relaxed text-ink-soft">{description}</p>}
      {action}
    </div>
  )
}

export function PrimaryButton({
  children,
  onClick,
  disabled,
  type = 'button',
}: {
  children: ReactNode
  onClick?: () => void
  disabled?: boolean
  type?: 'button' | 'submit'
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className="inline-flex min-h-11 items-center justify-center rounded-full bg-accent px-6 py-3 text-base font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep disabled:cursor-not-allowed disabled:opacity-50"
    >
      {children}
    </button>
  )
}
