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
      id="routes"
      surface
      eyebrow="Направления"
      title="Маршруты, которые мы возим регулярно"
      intro="Это не отзывы со стоков, а действующие коридоры с предсказуемыми сроками и понятным типом груза."
    >
      <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {ROUTES.map((route) => (
          <article
            key={`${route.from}-${route.to}`}
            className="group flex flex-col rounded-2xl border border-line bg-surface p-6 shadow-card transition-shadow duration-200 hover:shadow-soft"
          >
            <div className="flex items-baseline justify-between gap-3 font-display text-xl font-extrabold tracking-[-0.02em] text-ink">
              <span>{route.from}</span>
              <span>{route.to}</span>
            </div>
            {/* Мини-коридор маршрута: линия с узлами, способ перевозки — «на линии».
                Тот же мотив, что маршрутная полоса hero. */}
            <div aria-hidden="true" className="mt-3 flex items-center gap-2">
              <span className="h-2 w-2 shrink-0 rounded-full bg-accent" />
              <span className="h-px flex-1 bg-line" />
              <span className="font-mono text-[0.7rem] uppercase tracking-[0.08em] text-ink-soft">
                {route.mode}
              </span>
              <span className="h-px flex-1 bg-line" />
              <span className="h-2 w-2 shrink-0 rounded-full border-2 border-accent bg-surface transition-colors duration-200 group-hover:bg-accent" />
            </div>
            <span className="sr-only">Способ: {route.mode}</span>

            <div className="mt-6 flex items-baseline justify-between gap-4">
              <span className="field-label">Срок</span>
              <span className="terminal tabular text-xl font-semibold text-ink">{route.eta}</span>
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
