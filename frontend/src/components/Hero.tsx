// Узлы коридора CN→RU. Тот же набор подсвечивается из таймлайна «как работаем»:
// при hover на шаге i загорается узел i — невидимый процесс становится видимым.
const NODES = [
  { x: 60, label: 'Гуанчжоу' },
  { x: 175, label: 'Консолидация' },
  { x: 290, label: 'Граница' },
  { x: 405, label: 'Транзит' },
  { x: 520, label: 'Москва' },
]

const METRICS = [
  { value: '18–28', unit: 'дней транзит авто' },
  { value: '100%', unit: 'грузов с трекингом' },
  { value: '2 ч', unit: 'среднее время ответа' },
]

type HeroProps = {
  activeNode: number | null
}

export function Hero({ activeNode }: HeroProps) {
  return (
    <section id="top">
      {/* Строка регистрации документа — как шапка манифеста */}
      <div className="border-b border-rule">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-2.5 font-mono text-[0.7rem] uppercase tracking-[0.12em] text-ink-soft sm:px-8">
          <span>Манифест · CN → RU</span>
          <span className="hidden truncate sm:inline">
            Выкуп · консолидация · таможня · доставка до двери
          </span>
          <span>Ред. 2026</span>
        </div>
      </div>

      <div className="mx-auto grid max-w-6xl gap-12 px-5 py-14 sm:px-8 md:grid-cols-[1.12fr_0.88fr] md:gap-10 md:py-20">
        <div className="flex flex-col justify-center">
          <h1 className="font-display text-[clamp(2.75rem,7.5vw,5.25rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.03em] text-ink">
            Грузы из Китая
            <br />в Россию
          </h1>
          <p className="mt-7 max-w-xl text-lg leading-relaxed text-ink-soft">
            Выкуп, консолидация, таможенное оформление и доставка до двери.
            Ответственность за груз закреплена договором — статус контейнера виден
            на каждом этапе коридора, от выкупа до выдачи с закрывающими документами.
          </p>

          <div className="mt-9 flex flex-col gap-3 xs:flex-row xs:items-center">
            <a
              href="#calc"
              className="inline-flex items-center justify-center bg-stamp px-7 py-3.5 font-mono text-sm font-semibold uppercase tracking-[0.08em] text-paper transition-colors duration-200 hover:bg-stamp-deep"
            >
              Рассчитать стоимость
            </a>
            <a
              href="#how"
              className="inline-flex items-center justify-center border border-ink px-7 py-3.5 font-mono text-sm font-medium uppercase tracking-[0.08em] text-ink transition-colors duration-200 hover:bg-ink hover:text-paper"
            >
              Как это работает
            </a>
          </div>

          <dl className="mt-12 grid max-w-xl grid-cols-3 border-y border-rule">
            {METRICS.map((metric, index) => (
              <div
                key={metric.unit}
                className={`py-5 ${index > 0 ? 'border-l border-rule pl-5' : ''}`}
              >
                <dt className="sr-only">{metric.unit}</dt>
                <dd>
                  <span className="tabular block font-display text-2xl font-extrabold tracking-[-0.02em] text-ink sm:text-[1.75rem]">
                    {metric.value}
                  </span>
                  <span className="mt-1.5 block font-mono text-[0.7rem] uppercase leading-snug tracking-[0.04em] text-ink-soft">
                    {metric.unit}
                  </span>
                </dd>
              </div>
            ))}
          </dl>
        </div>

        {/* Коридор как мотив документа: оттиск-штамп + линия маршрута с узлами */}
        <div className="relative flex flex-col justify-center">
          <div
            aria-hidden="true"
            className="stamp absolute -top-2 right-1 rotate-[-8deg] text-xs sm:right-4"
          >
            Под контролем
          </div>

          <figure className="border border-rule bg-paper-raised">
            <figcaption className="flex items-center justify-between border-b border-rule px-5 py-3 font-mono text-[0.7rem] uppercase tracking-[0.1em] text-ink-soft">
              <span>Коридор доставки</span>
              <span className="text-marine">CN → RU</span>
            </figcaption>

            <div className="px-5 py-7">
              <svg
                viewBox="0 0 580 96"
                className="h-auto w-full"
                role="img"
                aria-label="Коридор доставки от Гуанчжоу до Москвы через консолидацию, границу и транзит"
              >
                <line
                  x1={NODES[0].x}
                  y1={40}
                  x2={NODES[NODES.length - 1].x}
                  y2={40}
                  stroke="var(--color-marine)"
                  strokeWidth={1.5}
                />
                {NODES.map((node, index) => {
                  const isActive = activeNode === index
                  return (
                    <g key={node.label}>
                      <rect
                        x={node.x - (isActive ? 7 : 4)}
                        y={40 - (isActive ? 7 : 4)}
                        width={isActive ? 14 : 8}
                        height={isActive ? 14 : 8}
                        fill={isActive ? 'var(--color-stamp)' : 'var(--color-marine)'}
                        stroke="var(--color-paper-raised)"
                        strokeWidth={3}
                        style={{
                          transition:
                            'all 220ms var(--ease-corridor)',
                        }}
                      />
                      <text
                        x={node.x}
                        y={72}
                        textAnchor="middle"
                        className="font-mono"
                        fontSize={10.5}
                        fill={isActive ? 'var(--color-ink)' : 'var(--color-ink-soft)'}
                      >
                        {node.label}
                      </text>
                    </g>
                  )
                })}
              </svg>
            </div>

            <p className="border-t border-rule px-5 py-4 text-sm leading-relaxed text-ink-soft">
              Каждый груз получает отдельный трек-номер. Статус виден на каждом
              узле коридора — от выкупа в КНР до выдачи на складе в России.
            </p>
          </figure>
        </div>
      </div>
    </section>
  )
}
