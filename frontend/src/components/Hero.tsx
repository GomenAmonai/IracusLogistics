import type { CSSProperties } from 'react'

const METRICS = [
  { value: '18–28', unit: 'дней транзит авто' },
  { value: '100%', unit: 'грузов с трекингом' },
  { value: '2 ч', unit: 'среднее время ответа' },
]

// Три равноправных способа везти — клиент выбирает по своей ситуации.
const LANES = [
  {
    title: 'Карго',
    note: 'Быстрая доставка через Казахстан и Азербайджан — для физлиц и небольших партий.',
  },
  {
    title: 'Белый импорт',
    note: 'С документами, НДС и Честным знаком — для маркетплейсов и юрлиц.',
  },
  {
    title: 'Выкуп',
    note: '找货, оплата поставщику и проверка товара на складе в Гуанчжоу.',
  },
]

// Узлы коридора CN→RU — маршрут груза «виден на каждом этапе». Подсветка идёт по
// очереди сама (CSS, см. .corridor-node), без зависимости от других секций.
const ROUTE = ['Гуанчжоу', 'Консолидация', 'Граница', 'Транзит', 'Москва']

export function Hero() {
  return (
    <section id="top" className="relative isolate overflow-hidden bg-night text-white">
      {/* Фоновый кадр коридора. Декоративный — смысл несёт заголовок. */}
      <img
        src="/hero/corridor-1920.jpg"
        alt=""
        aria-hidden="true"
        fetchPriority="high"
        width={1920}
        height={1080}
        className="absolute inset-0 -z-10 h-full w-full object-cover object-[70%_center]"
      />
      {/* Затемнение слева-направо под текст + увод низа в почти-чёрный к светлому блоку */}
      <div
        aria-hidden="true"
        className="absolute inset-0 -z-10 bg-[linear-gradient(100deg,#0b0e13_8%,rgba(11,14,19,0.82)_38%,rgba(11,14,19,0.35)_70%,rgba(11,14,19,0.55)_100%)]"
      />
      <div
        aria-hidden="true"
        className="absolute inset-x-0 bottom-0 -z-10 h-40 bg-[linear-gradient(to_top,#0b0e13,transparent)]"
      />

      <div className="relative mx-auto w-full max-w-6xl px-5 pb-12 pt-28 sm:px-8 md:pb-16 md:pt-36">
        <div className="max-w-2xl">
          <p className="eyebrow mb-5 flex items-center gap-2.5 text-amber">
            <span aria-hidden="true" className="h-2 w-2 rounded-full bg-amber" />
            Экспедирование Китай → Россия
          </p>
          <h1 className="font-display text-[2.6rem] font-extrabold leading-[1.04] tracking-[-0.03em] sm:text-6xl md:text-[4rem]">
            Грузы из Китая <span className="text-amber">под контролем</span> на каждом этапе
          </h1>
          <p className="mt-6 max-w-xl text-base leading-relaxed text-white/75 sm:text-lg">
            Выкуп, консолидация, таможня и доставка до двери. Карго или белый
            импорт — на выбор. За операциями в Китае стоит действующая
            логистическая компания в Гуанчжоу, статус груза виден на каждом шаге.
          </p>

          <div className="mt-9 flex flex-col gap-3 xs:flex-row xs:items-center">
            <a
              href="#calc"
              className="inline-flex items-center justify-center rounded-full bg-amber px-7 py-3.5 text-base font-semibold text-night shadow-soft transition-transform duration-200 hover:-translate-y-0.5"
            >
              Рассчитать стоимость
            </a>
            <a
              href="#how"
              className="inline-flex items-center justify-center rounded-full border border-white/25 bg-white/5 px-7 py-3.5 text-base font-medium text-white backdrop-blur-sm transition-colors duration-200 hover:border-white/50 hover:bg-white/10"
            >
              Как это работает
            </a>
          </div>

          <dl className="mt-12 flex flex-wrap gap-x-9 gap-y-6">
            {METRICS.map((metric) => (
              <div key={metric.unit}>
                <dt className="sr-only">{metric.unit}</dt>
                <dd>
                  <span className="terminal text-2xl font-semibold sm:text-3xl">{metric.value}</span>
                  <span className="mt-1.5 block text-xs leading-snug text-white/60">{metric.unit}</span>
                </dd>
              </div>
            ))}
          </dl>
        </div>

        {/* Коридор доставки: узлы загораются по очереди — «трекинг на каждом этапе».
            Десктоп-акцент; на узких экранах прячем, чтобы не дробить 5 подписей. */}
        <div className="mt-12 hidden max-w-2xl sm:block" aria-hidden="true">
          <div className="relative flex items-start justify-between">
            <span className="absolute inset-x-[5px] top-[5px] -z-0 h-px bg-white/15" />
            {ROUTE.map((label, index) => (
              <span
                key={label}
                className="relative flex flex-col items-center gap-2"
                style={{ '--i': index } as CSSProperties}
              >
                <span className="corridor-node h-2.5 w-2.5 rounded-full" />
                <span className="font-mono text-[0.7rem] tracking-wide text-white/55">{label}</span>
              </span>
            ))}
          </div>
        </div>

        {/* Три способа доставки — стеклянные карточки, мост в светлый функциональный низ */}
        <ul className="mt-12 grid gap-3 sm:grid-cols-3 sm:gap-4">
          {LANES.map((lane) => (
            <li
              key={lane.title}
              className="rounded-2xl border border-white/10 bg-white/[0.06] p-5 backdrop-blur-md"
            >
              <h2 className="font-display text-lg font-bold text-white">{lane.title}</h2>
              <p className="mt-2 text-sm leading-relaxed text-white/65">{lane.note}</p>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}
