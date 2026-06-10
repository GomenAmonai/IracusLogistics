import { useState, type FormEvent } from 'react'

import { ApiError, login } from '../lib/api'
import { ErrorNote, Field, PrimaryButton } from './ui'

export function Login({ onSuccess }: { onSuccess: () => void }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await login(email.trim(), password)
      onSuccess()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось войти.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="flex w-full max-w-sm flex-col gap-5 rounded-2xl border border-line bg-surface p-7 shadow-card"
      >
        <header>
          <p className="eyebrow">Icaris · панель менеджера</p>
          <h1 className="mt-1 font-display text-2xl font-bold text-ink">Вход</h1>
        </header>

        <Field id="email" label="Email">
          <input
            id="email"
            type="email"
            autoComplete="username"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="input-field"
          />
        </Field>
        <Field id="password" label="Пароль">
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="input-field"
          />
        </Field>

        {error && <ErrorNote message={error} />}

        <PrimaryButton type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Входим…' : 'Войти'}
        </PrimaryButton>
      </form>
    </div>
  )
}
