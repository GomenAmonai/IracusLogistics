import { Section } from './Section'

const ROUTES = [
  {
    from: 'Гуанчжоу',
    to: 'Москва',
    mode: 'Авто',
    eta: '18–28 дней',
    cargo: 'Электроника, сборный груз',
  },
  {
    from: 'Иу',
    to: 'Екатеринбург',
    mode: 'ЖД',
    eta: '25–40 дней',
    cargo: 'Товары для дома, текстиль',
  },
  {
    from: 'Шэньчжэнь',
    to: 'Новосибирск',
    mode: 'Авто',
    eta: '20–30 дней',
    cargo: 'Оборудование, комплектующие',
  },
]

export function Routes() {
  return (
    <Section
      surface
      eyebrow="Направления"
      title="Маршруты, которые мы возим регулярно"
      intro="Это не отзывы со стоков, а действующие коридоры с предсказуемыми сроками и понятным типом груза."
    >
      <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {ROUTES.map((route) => (
          <article
            key={`${route.from}-${route.to}`}
            className="flex flex-col rounded-2xl border border-line bg-surface p-6 shadow-card"
          >
            <div className="flex flex-wrap items-center gap-x-2.5 gap-y-1 font-display text-xl font-extrabold tracking-[-0.02em] text-ink">
              <span>{route.from}</span>
              <span aria-hidden="true" className="text-accent">
                →
              </span>
              <span>{route.to}</span>
            </div>

            <div className="mt-5 flex items-center justify-between gap-4">
              <div className="flex flex-col gap-1">
                <span className="field-label">Способ</span>
                <span className="font-mono uppercase tracking-[0.04em] text-ink">{route.mode}</span>
              </div>
              <div className="flex flex-col items-end gap-1.5">
                <span className="field-label">Срок</span>
                <span className="terminal tabular text-sm font-semibold">{route.eta}</span>
              </div>
            </div>

            <div className="mt-5 border-t border-line-soft pt-4">
              <span className="field-label">Груз</span>
              <p className="mt-1 text-sm leading-relaxed text-ink-soft">{route.cargo}</p>
            </div>
          </article>
        ))}
      </div>
    </Section>
  )
}
