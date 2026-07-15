import { useReveal } from '../lib/hooks'

type Stat = {
  value: string
  label: string
}

const STATS: Stat[] = [
  { value: 'Заявка', label: 'вводные по грузу' },
  { value: 'Менеджер', label: 'расчёт и условия' },
  { value: 'Трекинг', label: 'статусы, платежи и чат' },
]

const MODES = ['Авто', 'ЖД', 'Авиа', 'Море']

function StatCell({ stat }: { stat: Stat }) {
  return (
    <div className="flex flex-col items-center px-4 py-8 text-center sm:px-5">
      <span className="terminal text-3xl font-semibold sm:text-4xl">
        {stat.value}
      </span>
      <span className="mt-3 block font-mono text-[0.7rem] uppercase leading-snug tracking-[0.04em] text-ink-soft">
        {stat.label}
      </span>
    </div>
  )
}

export function Kpi() {
  const { ref, isVisible } = useReveal<HTMLElement>()

  return (
    <section
      ref={ref}
      className={`reveal ${isVisible ? 'reveal-in' : ''} border-t border-line bg-base-tint`}
      aria-label="Контур сервиса"
    >
      <div className="mx-auto w-full max-w-6xl px-5 py-16 sm:px-8">
        {/* Воздушный ряд показателей: мягкая surface-карточка, делители-линии вместо острых рамок */}
        <dl className="grid rounded-2xl border border-line bg-surface shadow-card sm:grid-cols-3 sm:divide-x sm:divide-line-soft max-sm:divide-y max-sm:divide-line-soft">
          {STATS.map((stat) => (
            <div key={stat.label}>
              <dt className="sr-only">{stat.label}</dt>
              <dd>
                <StatCell stat={stat} />
              </dd>
            </div>
          ))}
        </dl>

        <div className="mt-8 flex flex-wrap items-center gap-x-3 gap-y-4">
          <span className="eyebrow flex items-center gap-2.5">
            <span aria-hidden="true" className="h-2 w-2 rounded-full bg-accent" />
            Способы перевозки
          </span>
          <div className="flex flex-wrap gap-2">
            {MODES.map((mode) => (
              <span
                key={mode}
                className="rounded-full border border-line bg-surface px-3.5 py-1.5 font-mono text-sm uppercase tracking-[0.06em] text-ink-soft"
              >
                {mode}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
