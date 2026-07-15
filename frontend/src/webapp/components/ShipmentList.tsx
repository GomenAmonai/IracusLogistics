import type { Client, Shipment } from '../lib/types'
import { ShipmentCard } from './ShipmentCard'

type ShipmentListProps = {
  client: Client | null
  shipments: Shipment[]
  onOpen: (id: string) => void
}

// Что произойдёт после заявки — пустой список превращаем в онбординг, а не тупик.
const NEXT_STEPS = [
  'Менеджер согласует с вами условия и стоимость.',
  'Заведёт партию — здесь появится груз с трек-номером.',
  'Дальше статусы и чат по грузу обновляются в этом приложении.',
]

export function ShipmentList({ client, shipments, onOpen }: ShipmentListProps) {
  return (
    <div className="flex flex-1 flex-col">
      <header className="mb-5">
        <p className="eyebrow">Icaris · отслеживание</p>
        <h1 className="mt-1 font-display text-2xl font-bold text-ink">
          {client?.name ? `Грузы · ${client.name}` : 'Мои грузы'}
        </h1>
      </header>

      {shipments.length === 0 ? <EmptyState /> : (
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

function EmptyState() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-7 py-8">
      {/* Коридор CN→RU с пунктиром: груз ещё не в пути. Тот же мотив, что на лендинге. */}
      <div className="w-full max-w-xs" aria-hidden="true">
        <div className="flex items-center gap-2">
          <span className="h-2.5 w-2.5 shrink-0 rounded-full bg-accent" />
          <span className="flex-1 border-t border-dashed border-line" />
          <span className="h-2.5 w-2.5 shrink-0 rounded-full border-2 border-accent bg-surface" />
        </div>
        <div className="mt-2 flex justify-between font-mono text-[0.7rem] uppercase tracking-[0.06em] text-ink-soft">
          <span>Гуанчжоу</span>
          <span>Москва</span>
        </div>
      </div>

      <div className="px-2 text-center">
        <h2 className="font-display text-lg font-bold text-ink">Пока нет грузов</h2>
        <p className="mt-2 text-sm leading-relaxed text-ink-soft">
          Когда первая партия будет оформлена, она появится здесь — со статусами и чатом
          с менеджером.
        </p>
      </div>

      <ol className="w-full max-w-xs rounded-2xl border border-line bg-surface p-5 shadow-card">
        {NEXT_STEPS.map((step, index) => (
          <li
            key={step}
            className={`flex items-baseline gap-3 ${index > 0 ? 'mt-4 border-t border-line-soft pt-4' : ''}`}
          >
            <span className="terminal text-xs font-semibold text-accent">
              {String(index + 1).padStart(2, '0')}
            </span>
            <span className="text-sm leading-relaxed text-ink-soft">{step}</span>
          </li>
        ))}
      </ol>
    </div>
  )
}
