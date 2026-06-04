import type { Client, Shipment } from '../lib/types'
import { ShipmentCard } from './ShipmentCard'
import { CenteredState } from './ui'

type ShipmentListProps = {
  client: Client | null
  shipments: Shipment[]
  onOpen: (id: string) => void
}

export function ShipmentList({ client, shipments, onOpen }: ShipmentListProps) {
  return (
    <div className="flex flex-1 flex-col">
      <header className="mb-5">
        <p className="eyebrow">Icaris · отслеживание</p>
        <h1 className="mt-1 font-display text-2xl font-bold text-ink">
          {client?.name ? `Грузы · ${client.name}` : 'Мои грузы'}
        </h1>
      </header>

      {shipments.length === 0 ? (
        <CenteredState
          title="Пока нет грузов"
          description="Когда менеджер заведёт груз, он появится здесь. Статусы обновляются автоматически."
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {shipments.map((shipment) => (
            <li key={shipment.id}>
              <ShipmentCard shipment={shipment} onOpen={onOpen} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
