import { formatDate } from '../lib/format'
import type { Shipment } from '../lib/types'
import { StatusBadge } from './StatusBadge'

type ShipmentCardProps = {
  shipment: Shipment
  onOpen: (id: string) => void
}

export function ShipmentCard({ shipment, onOpen }: ShipmentCardProps) {
  const route = [shipment.from_city, shipment.to_city].filter(Boolean).join(' → ')

  return (
    <button
      type="button"
      onClick={() => onOpen(shipment.id)}
      className="flex w-full flex-col gap-3 rounded-2xl border border-line bg-surface p-4 text-left shadow-card transition-colors duration-200 hover:border-accent"
    >
      <div className="flex items-center justify-between gap-3">
        <span className="terminal text-sm font-medium text-ink">{shipment.tracking_key}</span>
        <StatusBadge status={shipment.status} />
      </div>
      {route && <p className="text-sm text-ink-soft">{route}</p>}
      <p className="terminal text-xs text-ink-soft">обновлён {formatDate(shipment.updated_at)}</p>
    </button>
  )
}
