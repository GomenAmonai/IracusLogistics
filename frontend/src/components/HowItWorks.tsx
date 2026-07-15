import { Section } from './Section'

const STEPS = [
  {
    title: 'Заявка',
    text: 'Оставляете вводные без регистрации: что везём, откуда и куда, примерные вес и объём.',
  },
  {
    title: 'Предварительная оценка',
    text: 'Блок расчёта помогает собрать параметры груза и перенести их в заявку без вымышленных ставок.',
  },
  {
    title: 'Менеджер рассчитывает условия',
    text: 'Менеджер проверяет документы и характеристики, затем письменно фиксирует доступную схему и условия.',
  },
  {
    title: 'Подготовка и отправка',
    text: 'Подключаем только согласованные услуги и передаём груз участникам выбранного маршрута.',
  },
  {
    title: 'Транзит и выдача',
    text: 'Обновляем статусы до согласованной точки выдачи и передаём предусмотренные схемой документы.',
  },
]

export function HowItWorks() {
  return (
    <Section
      id="how"
      eyebrow="Как работаем"
      title="Пять шагов от заявки до груза на вашем складе"
      intro="Заявка не требует регистрации. Все существенные условия появляются только после проверки груза и письменного подтверждения менеджером."
    >
      {/* Открытый степпер на линии-коридоре — рифма с маршрутной полосой hero:
          без карточек, узлы на общей линии, контент «висит» под своим узлом. */}
      <ol className="relative grid gap-9 sm:grid-cols-2 md:grid-cols-5 md:gap-7">
        <span
          aria-hidden="true"
          className="absolute left-1 right-16 top-[5px] hidden h-px bg-line md:block"
        />
        {STEPS.map((step, index) => (
          <li key={step.title} className="group relative md:pr-2">
            {/* Вертикальная нить шага на мобиле, узел на общей линии на десктопе */}
            <span
              aria-hidden="true"
              className="relative z-10 mb-5 flex h-2.5 w-2.5 rounded-full border-2 border-accent bg-base transition-colors duration-200 group-hover:bg-accent"
            />
            <span className="terminal text-sm font-semibold text-accent">
              {String(index + 1).padStart(2, '0')}
            </span>
            <h3 className="mt-2 font-display text-lg font-extrabold leading-tight tracking-[-0.02em] text-ink">
              {step.title}
            </h3>
            <p className="mt-3 text-sm leading-relaxed text-ink-soft">{step.text}</p>
          </li>
        ))}
      </ol>
    </Section>
  )
}
