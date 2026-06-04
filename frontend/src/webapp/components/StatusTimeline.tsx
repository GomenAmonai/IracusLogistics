import { formatDateTime } from '../lib/format'
import { STATUS_LABELS, type StatusEvent } from '../lib/types'

// История статусов сверху вниз в хронологическом порядке (как пришла с бэкенда). Последнее
// событие — текущее состояние, выделено акцентом; прошлые — приглушённые узлы.
export function StatusTimeline({ history }: { history: StatusEvent[] }) {
  if (history.length === 0) {
    return null
  }

  const lastIndex = history.length - 1

  return (
    <ol className="flex flex-col">
      {history.map((event, index) => {
        const isCurrent = index === lastIndex

        return (
          <li key={event.id} className="flex gap-3">
            <div className="flex flex-col items-center">
              <span
                aria-hidden
                className="mt-1 size-2.5 shrink-0 rounded-full"
                style={{ backgroundColor: isCurrent ? 'var(--color-accent)' : 'var(--color-line)' }}
              />
              {!isCurrent && <span aria-hidden className="w-px flex-1 bg-line" />}
            </div>
            <div className="pb-5">
              <p className={`text-sm font-semibold ${isCurrent ? 'text-ink' : 'text-ink-soft'}`}>
                {STATUS_LABELS[event.status]}
              </p>
              {event.comment && <p className="mt-0.5 text-sm leading-relaxed text-ink-soft">{event.comment}</p>}
              <p className="terminal mt-1 text-xs text-ink-soft">{formatDateTime(event.created_at)}</p>
            </div>
          </li>
        )
      })}
    </ol>
  )
}
