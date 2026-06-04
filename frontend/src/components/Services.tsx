import { Section } from './Section'

const SERVICES = [
  {
    title: 'Выкуп товара',
    text: 'Находим и выкупаем товар у поставщика в КНР, проверяем партию перед отправкой.',
  },
  {
    title: 'Консолидация',
    text: 'Собираем груз от нескольких поставщиков в одну партию на складе в Китае.',
  },
  {
    title: 'Таможенное оформление',
    text: 'Готовим документы, рассчитываем платежи, проходим таможню без сюрпризов по стоимости.',
  },
  {
    title: 'Страхование груза',
    text: 'Оформляем страховку на стоимость партии — на случай повреждения или утраты.',
  },
  {
    title: 'Доставка до двери',
    text: 'Довозим до вашего склада или адреса в России, а не только до терминала в Москве.',
  },
]

const REASONS = [
  {
    title: 'Договор и ответственность за груз',
    text: 'Отношения зафиксированы договором: за сохранность и сроки отвечаем мы, а не «как получится».',
  },
  {
    title: 'Прозрачная растаможка',
    text: 'Заранее показываем структуру таможенных платежей — без скрытых доплат в конце.',
  },
  {
    title: 'Один менеджер на сделку',
    text: 'С вами работает один человек от заявки до выдачи — не нужно объяснять всё заново.',
  },
  {
    title: 'Трекинг на каждом этапе',
    text: 'Видите статус груза по трек-номеру — где партия и что с ней происходит сейчас.',
  },
]

export function Services() {
  return (
    <Section
      id="services"
      eyebrow="Услуги"
      title="Полный цикл от поставщика до вашего склада"
      intro="Берём на себя то, на чём импортёр обычно теряет деньги и время: выкуп, сборку партии, таможню и доставку до двери."
    >
      {/* Услуги — мягкие скруглённые карточки: номер позиции как терминал-табло */}
      <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {SERVICES.map((service, index) => (
          <article
            key={service.title}
            className="rounded-2xl border border-line bg-surface p-6 shadow-card transition-shadow duration-200 hover:shadow-soft"
          >
            <span className="terminal text-sm font-semibold">
              {String(index + 1).padStart(2, '0')}
            </span>
            <h3 className="mt-4 font-display text-lg font-bold tracking-[-0.01em] text-ink">
              {service.title}
            </h3>
            <p className="mt-2 text-sm leading-relaxed text-ink-soft">
              {service.text}
            </p>
          </article>
        ))}
      </div>

      {/* Единственный кобальт-акцент блока: ключевой принцип процесса */}
      <p className="mt-5 flex items-baseline gap-4 rounded-2xl border-l-2 border-accent bg-accent-tint py-4 pl-5 pr-6 text-sm leading-relaxed text-ink">
        <span className="field-label shrink-0 text-accent">Принцип</span>
        Каждый этап закрыт документами — груз не «зависает» между подрядчиками.
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
