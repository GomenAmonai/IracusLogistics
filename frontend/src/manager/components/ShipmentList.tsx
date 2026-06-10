import { useEffect, useState } from 'react'

import { ApiError, listShipments } from '../lib/api'
import { formatDate, formatMoney } from '../lib/format'
import type { Shipment } from '../lib/types'
import { LaneBadge, ShipmentStatusBadge } from './badges'
import { CenteredState, ErrorNote, PrimaryButton, Spinner } from './ui'

export function ShipmentList({
  onAuthError,
  onOpen,
  onCreate,
}: {
  onAuthError: (err: unknown) => boolean
  onOpen: (id: string) => void
  onCreate: () => void
}) {
  const [shipments, setShipments] = useState<Shipment[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    listShipments()
      .then((loaded) => {
        if (active) setShipments(loaded)
      })
      .catch((err: unknown) => {
        if (active && !onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить грузы.')
        }
      })
    return () => {
      active = false
    }
  }, [onAuthError])

  if (error) {
    return <ErrorNote message={error} />
  }
  if (shipments === null) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-5">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="font-display text-2xl font-bold text-ink">Грузы</h1>
        <PrimaryButton onClick={onCreate}>Завести груз</PrimaryButton>
      </header>

      {shipments.length === 0 ? (
        <CenteredState
          title="Грузов нет"
          description="Заведите первый груз для зарегистрированного клиента."
          action={<PrimaryButton onClick={onCreate}>Завести груз</PrimaryButton>}
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {shipments.map((shipment) => (
            <li key={shipment.id}>
              <button
                type="button"
                onClick={() => onOpen(shipment.id)}
                className="flex w-full flex-col gap-2.5 rounded-2xl border border-line bg-surface p-5 text-left shadow-card transition-colors duration-200 hover:border-accent"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <span className="terminal text-sm font-medium text-ink">{shipment.tracking_key}</span>
                  <div className="flex items-center gap-2">
                    <LaneBadge lane={shipment.lane} />
                    <ShipmentStatusBadge status={shipment.status} />
                  </div>
                </div>
                <div className="flex flex-wrap items-center justify-between gap-3 text-sm">
                  <span className="text-ink-soft">
                    {[shipment.from_city, shipment.to_city].filter(Boolean).join(' → ') || 'Маршрут не указан'}
                  </span>
                  <span className="terminal text-ink">
                    {formatMoney(shipment.price, shipment.currency) ?? 'цена не указана'}
                  </span>
                </div>
                <span className="terminal text-xs text-ink-soft">
                  обновлён {formatDate(shipment.updated_at)}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
