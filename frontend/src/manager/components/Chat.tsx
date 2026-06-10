import { useEffect, useRef, useState, type FormEvent } from 'react'

import { ApiError, listMessages, sendMessage } from '../lib/api'
import { formatDateTime } from '../lib/format'
import type { Message } from '../lib/types'
import { ErrorNote, PrimaryButton, Spinner } from './ui'

export function Chat({
  shipmentId,
  onAuthError,
}: {
  shipmentId: string
  onAuthError: (err: unknown) => boolean
}) {
  const [messages, setMessages] = useState<Message[] | null>(null)
  const [draft, setDraft] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSending, setSending] = useState(false)
  const listRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let active = true
    listMessages(shipmentId)
      .then((loaded) => {
        if (active) setMessages(loaded)
      })
      .catch((err: unknown) => {
        if (active && !onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить переписку.')
        }
      })
    return () => {
      active = false
    }
  }, [shipmentId, onAuthError])

  // Прокрутка к последнему сообщению после загрузки и после отправки.
  useEffect(() => {
    listRef.current?.scrollTo({ top: listRef.current.scrollHeight })
  }, [messages])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const text = draft.trim()
    if (!text) {
      return
    }
    setSending(true)
    setError(null)
    try {
      const sent = await sendMessage(shipmentId, text)
      setMessages((current) => [...(current ?? []), sent])
      setDraft('')
    } catch (err) {
      if (!onAuthError(err)) {
        setError(err instanceof ApiError ? err.message : 'Не удалось отправить сообщение.')
      }
    } finally {
      setSending(false)
    }
  }

  return (
    <section className="flex h-fit flex-col gap-4 rounded-2xl border border-line bg-surface p-5 shadow-card lg:sticky lg:top-20">
      <h2 className="font-display text-base font-bold text-ink">Чат с клиентом</h2>

      {messages === null ? (
        <Spinner />
      ) : (
        <div
          ref={listRef}
          aria-live="polite"
          className="flex max-h-96 flex-col gap-2.5 overflow-y-auto pr-1"
        >
          {messages.length === 0 && (
            <p className="text-sm text-ink-soft">Сообщений пока нет.</p>
          )}
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex max-w-[85%] flex-col gap-0.5 rounded-2xl px-3.5 py-2.5 ${
                message.from_role === 'manager'
                  ? 'self-end rounded-br-md bg-accent-tint'
                  : 'self-start rounded-bl-md bg-surface-soft'
              }`}
            >
              <p className="whitespace-pre-wrap text-sm leading-relaxed text-ink">{message.text}</p>
              <span className="terminal self-end text-[0.65rem] text-ink-soft">
                {formatDateTime(message.created_at)}
              </span>
            </div>
          ))}
        </div>
      )}

      {error && <ErrorNote message={error} />}

      <form onSubmit={(e) => void handleSubmit(e)} className="flex flex-col gap-2.5">
        <label htmlFor="chat-draft" className="sr-only">
          Сообщение клиенту
        </label>
        <textarea
          id="chat-draft"
          rows={3}
          placeholder="Ответ клиенту — он получит уведомление в Telegram"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          className="input-field resize-none"
        />
        <div>
          <PrimaryButton type="submit" disabled={isSending || !draft.trim()}>
            {isSending ? 'Отправляем…' : 'Отправить'}
          </PrimaryButton>
        </div>
      </form>
    </section>
  )
}
