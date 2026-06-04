import { useEffect, useRef, useState } from 'react'

import { ApiError, listMessages, sendMessage } from '../lib/api'
import { formatDateTime } from '../lib/format'
import { haptic } from '../lib/telegram'
import type { Message } from '../lib/types'

// Чат с менеджером по конкретному грузу. Загружает ленту при монтировании, дозагружает
// отправленное. NOTE: MVP — без авто-обновления/поллинга; новые ответы видны при повторном
// открытии; см. docs/tech-debt.md
export function Chat({ shipmentId }: { shipmentId: string }) {
  const [messages, setMessages] = useState<Message[]>([])
  const [text, setText] = useState('')
  const [isLoading, setLoading] = useState(true)
  const [isSending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let active = true
    listMessages(shipmentId)
      .then((loaded) => {
        if (active) {
          setMessages(loaded)
        }
      })
      .catch((err: unknown) => {
        if (active) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить переписку.')
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [shipmentId])

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: 'nearest' })
  }, [messages])

  async function handleSend(event: React.FormEvent) {
    event.preventDefault()
    const trimmed = text.trim()
    if (!trimmed || isSending) {
      return
    }

    setSending(true)
    setError(null)
    try {
      const created = await sendMessage(shipmentId, trimmed)
      setMessages((prev) => [...prev, created])
      setText('')
      haptic('success')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось отправить сообщение.')
      haptic('error')
    } finally {
      setSending(false)
    }
  }

  return (
    <section aria-label="Чат с менеджером" className="flex flex-col gap-3">
      <h2 className="font-display text-base font-semibold text-ink">Чат с менеджером</h2>

      <div className="flex flex-col gap-2">
        {isLoading ? (
          <p role="status" className="text-sm text-ink-soft">
            Загрузка переписки…
          </p>
        ) : messages.length === 0 ? (
          <p className="text-sm text-ink-soft">Сообщений пока нет. Напишите менеджеру первым.</p>
        ) : (
          messages.map((message) => <Bubble key={message.id} message={message} />)
        )}
        <div ref={endRef} />
      </div>

      {/* role=alert → ассистивные технологии озвучат ошибку; цвет не единственный сигнал. */}
      {error && (
        <p role="alert" className="text-sm text-alert">
          {error}
        </p>
      )}

      <form onSubmit={handleSend} className="flex items-end gap-2">
        <label htmlFor="chat-input" className="sr-only">
          Сообщение менеджеру
        </label>
        <input
          id="chat-input"
          className="input-field"
          value={text}
          onChange={(event) => setText(event.target.value)}
          placeholder="Сообщение менеджеру"
          autoComplete="off"
        />
        <button
          type="submit"
          disabled={!text.trim() || isSending}
          className="inline-flex min-h-11 shrink-0 items-center justify-center rounded-full bg-accent px-5 py-3 text-sm font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSending ? '…' : 'Отправить'}
        </button>
      </form>
    </section>
  )
}

function Bubble({ message }: { message: Message }) {
  const isClient = message.from_role === 'client'
  return (
    <div className={`flex flex-col ${isClient ? 'items-end' : 'items-start'}`}>
      <div
        className={`max-w-[85%] rounded-2xl px-3.5 py-2 text-sm leading-relaxed ${
          isClient ? 'bg-accent text-surface' : 'border border-line bg-surface text-ink'
        }`}
      >
        {message.text}
      </div>
      <span className="terminal mt-0.5 px-1 text-[0.6875rem] text-ink-soft">
        {formatDateTime(message.created_at)}
      </span>
    </div>
  )
}
