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
      {/* Реестр услуг: нумерованные ruled-строки, как позиции манифеста */}
      <div className="border-y border-rule">
        {SERVICES.map((service, index) => (
          <article
            key={service.title}
            className="grid grid-cols-[auto_1fr] items-baseline gap-x-5 gap-y-1.5 border-b border-rule-soft py-5 last:border-b-0 sm:grid-cols-[auto_minmax(0,16rem)_1fr] sm:gap-x-8"
          >
            <span className="tabular font-mono text-sm text-ink-soft">
              {String(index + 1).padStart(2, '0')}
            </span>
            <h3 className="font-display text-lg font-bold tracking-[-0.01em] text-ink">
              {service.title}
            </h3>
            <p className="col-start-2 text-sm leading-relaxed text-ink-soft sm:col-start-3">
              {service.text}
            </p>
          </article>
        ))}
      </div>

      {/* Единственный штамп-акцент блока: ключевой принцип процесса */}
      <p className="mt-5 flex items-baseline gap-4 border-l-2 border-stamp bg-paper-raised py-4 pl-5 pr-6 text-sm leading-relaxed text-ink">
        <span className="field-label shrink-0 text-stamp">Принцип</span>
        Каждый этап закрыт документами — груз не «зависает» между подрядчиками.
      </p>

      {/* Причины выбрать: bordered ruled-ячейки, острые углы */}
      <div className="mt-16 grid border-l border-t border-rule sm:grid-cols-2">
        {REASONS.map((reason) => (
          <article
            key={reason.title}
            className="border-b border-r border-rule p-6"
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
