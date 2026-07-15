import { Section } from './Section'

const SERVICES = [
  {
    title: 'Выкуп товара',
    text: 'По согласованной схеме можем организовать поиск, оплату поставщику и проверку партии перед отправкой.',
  },
  {
    title: 'Консолидация',
    text: 'При необходимости объединяем грузы от нескольких поставщиков в одну партию.',
  },
  {
    title: 'Таможенное оформление',
    text: 'Участники, комплект документов и порядок таможенного оформления согласовываются по конкретному грузу.',
  },
  {
    title: 'Страхование груза',
    text: 'Возможность, покрытие и стоимость страхования уточняются до отправки партии.',
  },
  {
    title: 'Доставка до двери',
    text: 'Конечная точка и формат выдачи фиксируются в согласованном маршруте.',
  },
]

const REASONS = [
  {
    title: 'Договор и ответственность за груз',
    text: 'Стороны, объём услуг, сроки и ответственность должны быть закреплены в договоре до начала перевозки.',
  },
  {
    title: 'Прозрачная растаможка',
    text: 'Менеджер письменно фиксирует предварительную структуру расходов и условия её изменения.',
  },
  {
    title: 'Один менеджер на сделку',
    text: 'По заявке назначается менеджер, который хранит договорённости и ведёт основной диалог.',
  },
  {
    title: 'Трекинг на каждом этапе',
    text: 'Кабинет хранит текущий статус, историю изменений, платежи и переписку по грузу.',
  },
]

export function Services() {
  return (
    <Section
      id="services"
      eyebrow="Услуги"
      title="Собираем подходящую схему поставки"
      intro="Набор услуг определяется после проверки груза: от консолидации и выкупа до оформления и доставки до согласованной точки."
    >
      {/* Услуги — редакционный нумерованный список с разделителями: номер-табло,
          заголовок и описание в три колонки. Контраст к карточным блокам страницы. */}
      <div className="border-y border-line">
        {SERVICES.map((service, index) => (
          <article
            key={service.title}
            className={`group grid gap-2 py-7 md:grid-cols-[3.5rem_minmax(0,1fr)_minmax(0,1.3fr)] md:items-baseline md:gap-8 ${
              index > 0 ? 'border-t border-line-soft' : ''
            }`}
          >
            <span className="terminal text-sm font-semibold text-accent">
              {String(index + 1).padStart(2, '0')}
            </span>
            <h3 className="font-display text-xl font-extrabold leading-snug tracking-[-0.02em] text-ink transition-transform duration-200 group-hover:translate-x-1 md:text-2xl">
              {service.title}
            </h3>
            {/* NOTE: не использовать класс text-base — токен --color-base перехватывает
                его как цвет (текст красится в цвет фона). Размер задаём явно. */}
            <p className="text-sm leading-relaxed text-ink-soft sm:text-[1rem]">
              {service.text}
            </p>
          </article>
        ))}
      </div>

      {/* Единственный кобальт-акцент блока: ключевой принцип процесса */}
      <p className="mt-5 flex items-baseline gap-4 rounded-2xl border-l-2 border-accent bg-accent-tint py-4 pl-5 pr-6 text-sm leading-relaxed text-ink">
        <span className="field-label shrink-0 text-accent">Принцип</span>
        Участников, документы, стоимость и ответственность фиксируем до начала перевозки.
      </p>

      {/* Причины выбрать — мягкие скруглённые surface-карточки */}
      <div className="mt-16 grid gap-5 sm:grid-cols-2">
        {REASONS.map((reason) => (
          <article
            key={reason.title}
            className="rounded-2xl border border-line bg-surface p-6 shadow-card transition-shadow duration-200 hover:shadow-soft"
          >
            <h3 className="font-display text-base font-bold tracking-[-0.01em] text-ink">
              {reason.title}
            </h3>
            <p className="mt-2.5 text-sm leading-relaxed text-ink-soft">{reason.text}</p>
          </article>
        ))}
      </div>
    </Section>
  )
}
