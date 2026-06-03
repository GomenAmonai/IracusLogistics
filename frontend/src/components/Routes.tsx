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

// Сетка-манифест: на md+ четыре колонки с шапкой-реестром, на мобиле — строки
// с подписями-полями. Без боксов-карточек — таблица декларации.
const COLS = 'md:grid md:grid-cols-[1.6fr_0.7fr_0.9fr_1.4fr] md:items-baseline md:gap-6 md:px-1'

export function Routes() {
  return (
    <Section
      surface
      eyebrow="Направления"
      title="Маршруты, которые мы возим регулярно"
      intro="Это не отзывы со стоков, а действующие коридоры с предсказуемыми сроками и понятным типом груза."
    >
      <div className="border-y border-rule">
        <div className={`hidden border-b border-rule py-3 ${COLS}`}>
          <span className="field-label">Маршрут</span>
          <span className="field-label">Способ</span>
          <span className="field-label">Срок</span>
          <span className="field-label">Груз</span>
        </div>

        {ROUTES.map((route) => (
          <article
            key={`${route.from}-${route.to}`}
            className={`border-b border-rule-soft py-5 last:border-b-0 ${COLS}`}
          >
            <div className="flex flex-wrap items-center gap-2 font-display text-lg font-extrabold tracking-[-0.02em] text-ink sm:text-xl">
              <span>{route.from}</span>
              <span aria-hidden="true" className="text-ink-soft">
                →
              </span>
              <span>{route.to}</span>
            </div>

            <div className="mt-3 flex items-baseline justify-between md:mt-0 md:block">
              <span className="field-label md:hidden">Способ</span>
              <span className="font-mono uppercase tracking-[0.04em] text-ink">{route.mode}</span>
            </div>

            <div className="mt-2 flex items-baseline justify-between md:mt-0 md:block">
              <span className="field-label md:hidden">Срок</span>
              <span className="tabular font-mono text-ink">{route.eta}</span>
            </div>

            <div className="mt-2 flex items-baseline justify-between gap-4 md:mt-0 md:block">
              <span className="field-label shrink-0 md:hidden">Груз</span>
              <span className="text-right text-ink-soft md:text-left">{route.cargo}</span>
            </div>
          </article>
        ))}
      </div>
    </Section>
  )
}
