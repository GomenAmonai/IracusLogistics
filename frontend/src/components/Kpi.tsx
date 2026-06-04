import { useReveal } from '../lib/hooks'

type Stat = {
  target: number
  suffix: string
  decimals: number
  label: string
}

const STATS: Stat[] = [
  { target: 9, suffix: '', decimals: 0, label: 'лет на рынке КНР → РФ' },
  { target: 640, suffix: '', decimals: 0, label: 'тонн груза в месяц' },
  { target: 22, suffix: '', decimals: 0, label: 'дней средний транзит' },
  { target: 100, suffix: '%', decimals: 0, label: 'грузов с трекингом' },
]

const MODES = ['Авто', 'ЖД', 'Авиа', 'Море']

function StatCell({ stat }: { stat: Stat }) {
  return (
    <div className="flex flex-col items-center px-4 py-7 text-center sm:px-5">
      <span className="terminal text-2xl font-semibold sm:text-3xl">
        {stat.target.toFixed(stat.decimals)}
        {stat.suffix}
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
      aria-label="Показатели работы"
    >
      <div className="mx-auto w-full max-w-6xl px-5 py-16 sm:px-8">
        {/* Воздушный ряд показателей: мягкая surface-карточка, делители-линии вместо острых рамок */}
        <dl className="grid grid-cols-2 rounded-2xl border border-line bg-surface shadow-card md:grid-cols-4 md:divide-x md:divide-line-soft [&>:nth-child(-n+2)]:border-b [&>:nth-child(-n+2)]:border-line-soft md:[&>:nth-child(-n+2)]:border-b-0 [&>:nth-child(2n)]:border-l [&>:nth-child(2n)]:border-line-soft md:[&>:nth-child(2n)]:border-l-0">
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
