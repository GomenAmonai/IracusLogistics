import type { ReactNode } from 'react'

// Мелкие атомы панели: спиннер, состояния, кнопки, поле формы. Один файл — каждый
// слишком мал для отдельного (как в WebApp).

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
    <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center">
      <h2 className="font-display text-lg font-semibold text-ink">{title}</h2>
      {description && <p className="max-w-md text-sm leading-relaxed text-ink-soft">{description}</p>}
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
      className="inline-flex min-h-10 items-center justify-center rounded-full bg-accent px-5 py-2 text-sm font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep disabled:cursor-not-allowed disabled:opacity-50"
    >
      {children}
    </button>
  )
}

export function GhostButton({
  children,
  onClick,
  disabled,
}: {
  children: ReactNode
  onClick?: () => void
  disabled?: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="inline-flex min-h-10 items-center justify-center rounded-full border border-line bg-surface px-5 py-2 text-sm font-medium text-ink transition-colors duration-200 hover:border-accent disabled:cursor-not-allowed disabled:opacity-50"
    >
      {children}
    </button>
  )
}

// Поле формы: подпись + контрол. htmlFor/id связывает label с input для a11y.
export function Field({
  id,
  label,
  children,
}: {
  id: string
  label: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="field-label">
        {label}
      </label>
      {children}
    </div>
  )
}

export function ErrorNote({ message }: { message: string }) {
  return (
    <p role="alert" className="rounded-xl border border-alert/30 bg-alert/5 px-4 py-3 text-sm text-alert">
      {message}
    </p>
  )
}
