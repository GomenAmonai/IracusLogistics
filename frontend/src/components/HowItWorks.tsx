import { Section } from './Section'

const STEPS = [
  {
    title: 'Заявка',
    text: 'Оставляете вводные без регистрации: что везём, откуда и куда, примерные вес и объём.',
  },
  {
    title: 'Калькулятор показывает вилку',
    text: 'Сразу видите предварительный диапазон цены и срок — чтобы понять порядок затрат.',
  },
  {
    title: 'Менеджер считает точную цену',
    text: 'Проверяем документы и характеристики груза, фиксируем точную сумму и связываемся с вами.',
  },
  {
    title: 'Выкуп, консолидация, таможня',
    text: 'Выкупаем товар, собираем партию на складе в КНР, проходим таможенное оформление.',
  },
  {
    title: 'Транзит и выдача',
    text: 'Доставляем по выбранному маршруту и выдаём груз с закрывающими документами.',
  },
]

type HowItWorksProps = {
  onStepHover: (index: number | null) => void
}

export function HowItWorks({ onStepHover }: HowItWorksProps) {
  return (
    <Section
      id="how"
      eyebrow="Как работаем"
      title="Пять шагов от заявки до груза на вашем складе"
      intro="Схема прозрачна с первого касания: сайт даёт вилку цены, заявка не требует регистрации, точную сумму фиксирует менеджер после проверки документов."
    >
      <ol className="grid gap-4 sm:grid-cols-2 md:grid-cols-5">
        {STEPS.map((step, index) => (
          <li
            key={step.title}
            // Мягкая карточка-степпер: при hover/фокусе граница уходит в кобальт —
            // активный шаг подсвечивает соответствующий узел коридора в Hero.
            className="group flex flex-col rounded-2xl border border-line bg-surface p-5 shadow-card transition-colors duration-200 hover:border-accent focus-within:border-accent"
            onMouseEnter={() => onStepHover(index)}
            onMouseLeave={() => onStepHover(null)}
            onFocus={() => onStepHover(index)}
            onBlur={() => onStepHover(null)}
          >
            {/* Номер шага — числовое табло терминала: контраст в мягком UI */}
            <span className="terminal mb-4 self-start text-sm font-semibold">
              {String(index + 1).padStart(2, '0')}
            </span>
            <h3 className="font-display text-lg font-extrabold leading-tight tracking-[-0.02em] text-ink">
              {step.title}
            </h3>
            <p className="mt-3 text-sm leading-relaxed text-ink-soft">{step.text}</p>
          </li>
        ))}
      </ol>
    </Section>
  )
}
