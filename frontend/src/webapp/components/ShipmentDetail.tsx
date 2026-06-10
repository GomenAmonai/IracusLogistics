import { useEffect, useState } from 'react'

import { ApiError, getShipment } from '../lib/api'
import { formatDate, formatMoney, formatVolume, formatWeight } from '../lib/format'
import { LANE_LABELS, type ShipmentDetail as ShipmentDetailData } from '../lib/types'
import { Chat } from './Chat'
import { StatusBadge } from './StatusBadge'
import { StatusTimeline } from './StatusTimeline'
import { CenteredState, Spinner } from './ui'

type Fact = { label: string; value: string }

function facts(detail: ShipmentDetailData): Fact[] {
  const { shipment } = detail
  const route = [shipment.from_city, shipment.to_city].filter(Boolean).join(' → ')
  const entries: Array<[string, string | null]> = [
    ['Маршрут', route || null],
    ['Полоса', LANE_LABELS[shipment.lane] ?? shipment.lane],
    ['Вес', formatWeight(shipment.weight)],
    ['Объём', formatVolume(shipment.volume)],
    ['Стоимость', formatMoney(shipment.price, shipment.currency)],
    ['Доставлен', shipment.delivered_at ? formatDate(shipment.delivered_at) : null],
  ]
  return entries.filter((entry): entry is [string, string] => entry[1] !== null).map(([label, value]) => ({ label, value }))
}

export function ShipmentDetail({ shipmentId }: { shipmentId: string }) {
  const [detail, setDetail] = useState<ShipmentDetailData | null>(null)
  const [isLoading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    setLoading(true)
    getShipment(shipmentId)
      .then((loaded) => {
        if (active) {
          setDetail(loaded)
        }
      })
      .catch((err: unknown) => {
        if (active) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить груз.')
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

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    )
  }

  if (error || !detail) {
    return <CenteredState title="Груз недоступен" description={error ?? 'Попробуйте позже.'} />
  }

  const { shipment } = detail

  return (
    <div className="flex flex-1 flex-col gap-6">
      <header className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-3">
          <span className="terminal text-lg font-semibold text-ink">{shipment.tracking_key}</span>
          <StatusBadge status={shipment.status} />
        </div>
        {shipment.status_comment && <p className="text-sm leading-relaxed text-ink-soft">{shipment.status_comment}</p>}
      </header>

      <dl className="grid grid-cols-2 gap-x-4 gap-y-3 rounded-2xl border border-line bg-surface p-4 shadow-card">
        {facts(detail).map((fact) => (
          <div key={fact.label} className="flex flex-col gap-0.5">
            <dt className="field-label">{fact.label}</dt>
            <dd className="text-sm text-ink">{fact.value}</dd>
          </div>
        ))}
      </dl>

      <section className="flex flex-col gap-3">
        <h2 className="font-display text-base font-semibold text-ink">История статусов</h2>
        <StatusTimeline history={detail.history} />
      </section>

      <Chat shipmentId={shipmentId} />
    </div>
  )
}
