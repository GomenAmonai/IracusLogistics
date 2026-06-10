import { useEffect, useState } from 'react'

import { ApiError, listClients } from '../lib/api'
import { formatDate } from '../lib/format'
import type { Client } from '../lib/types'
import { CenteredState, ErrorNote, GhostButton, Spinner } from './ui'

export function ClientList({
  onAuthError,
  onCreateShipment,
}: {
  onAuthError: (err: unknown) => boolean
  onCreateShipment: (clientId: string) => void
}) {
  const [clients, setClients] = useState<Client[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    listClients()
      .then((loaded) => {
        if (active) setClients(loaded)
      })
      .catch((err: unknown) => {
        if (active && !onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить клиентов.')
        }
      })
    return () => {
      active = false
    }
  }, [onAuthError])

  if (error) {
    return <ErrorNote message={error} />
  }
  if (clients === null) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-5">
      <h1 className="font-display text-2xl font-bold text-ink">Клиенты</h1>

      {clients.length === 0 ? (
        <CenteredState
          title="Клиентов нет"
          description="Клиент появляется после подтверждения в боте (/start)."
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {clients.map((client) => (
            <li
              key={client.id}
              className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-line bg-surface p-5 shadow-card"
            >
              <div className="flex flex-col gap-0.5">
                <span className="font-display text-base font-bold text-ink">
                  {client.name || 'Без имени'}
                </span>
                <span className="terminal text-xs text-ink-soft">
                  {client.username ? `@${client.username} · ` : ''}
                  с {formatDate(client.created_at)}
                </span>
              </div>
              <GhostButton onClick={() => onCreateShipment(client.id)}>Завести груз</GhostButton>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
