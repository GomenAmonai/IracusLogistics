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
    <section id="top" className="relative overflow-hidden">
      {/* Мягкая прохладная аура — спокойный свет, без резких пятен */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(90%_60%_at_85%_-10%,var(--color-accent-tint),transparent_60%)]"
      />

      <div className="relative mx-auto grid w-full max-w-6xl gap-12 px-5 pb-20 pt-16 sm:px-8 md:grid-cols-[1.08fr_0.92fr] md:gap-12 md:pb-28 md:pt-24">
        <div className="flex flex-col justify-center">
          <p className="eyebrow mb-5 flex items-center gap-2.5">
            <span aria-hidden="true" className="h-2 w-2 rounded-full bg-accent" />
            Экспедирование Китай → Россия
          </p>
          <h1 className="font-display text-4xl font-extrabold leading-[1.05] tracking-[-0.025em] text-ink sm:text-5xl md:text-[3.5rem]">
            Грузы из Китая в Россию —{' '}
            <span className="text-accent">под контролем на каждом этапе</span>
          </h1>
          <p className="mt-6 max-w-xl text-base leading-relaxed text-ink-soft sm:text-lg">
            Выкуп, консолидация, таможенное оформление и доставка до двери.
            Ответственность за груз закреплена договором, вы видите статус
            контейнера на каждом этапе — от выкупа до выдачи с закрывающими
            документами.
          </p>

          <div className="mt-8 flex flex-col gap-3 xs:flex-row xs:items-center">
            <a
              href="#calc"
              className="inline-flex items-center justify-center rounded-full bg-accent px-7 py-3.5 text-base font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep"
            >
              Рассчитать стоимость
            </a>
            <a
              href="#how"
              className="inline-flex items-center justify-center rounded-full border border-line bg-surface px-7 py-3.5 text-base font-medium text-ink transition-colors duration-200 hover:border-accent hover:text-accent"
            >
              Как это работает
            </a>
          </div>

          {/* Показатели «терминалом» — моноширинные табличные числа, как на табло */}
          <dl className="mt-12 flex flex-wrap gap-x-8 gap-y-6">
            {METRICS.map((metric) => (
              <div key={metric.unit}>
                <dt className="sr-only">{metric.unit}</dt>
                <dd>
                  <span className="terminal text-xl font-semibold sm:text-2xl">
                    {metric.value}
                  </span>
                  <span className="mt-2 block text-xs leading-snug text-ink-soft">
                    {metric.unit}
                  </span>
                </dd>
              </div>
            ))}
          </dl>
        </div>

        {/* Коридор как мягкая карточка-виджет: тихий маршрут с узлами */}
        <div className="flex flex-col justify-center">
          <figure className="rounded-2xl border border-line bg-surface p-6 shadow-soft sm:p-7">
            <figcaption className="mb-6 flex items-center justify-between">
              <span className="eyebrow">Коридор доставки</span>
              <span className="terminal text-xs">CN → RU</span>
            </figcaption>

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
                stroke="var(--color-line)"
                strokeWidth={2}
              />
              {NODES.map((node, index) => {
                const isActive = activeNode === index
                return (
                  <g key={node.label}>
                    <circle
                      cx={node.x}
                      cy={40}
                      r={isActive ? 8 : 5}
                      fill={isActive ? 'var(--color-accent)' : 'var(--color-surface)'}
                      stroke="var(--color-accent)"
                      strokeWidth={2}
                      style={{
                        transition: 'r 220ms var(--ease-corridor), fill 220ms var(--ease-corridor)',
                      }}
                    />
                    <text
                      x={node.x}
                      y={70}
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

            <p className="mt-6 border-t border-line-soft pt-5 text-sm leading-relaxed text-ink-soft">
              Каждый груз получает отдельный трек-номер. Статус виден на каждом
              узле коридора — от выкупа в КНР до выдачи на складе в России.
            </p>
          </figure>
        </div>
      </div>
    </section>
  )
}
