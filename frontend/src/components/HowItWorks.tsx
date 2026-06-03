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
      <ol className="border-t border-rule md:grid md:grid-cols-5 md:border-t-0 md:border-l md:border-rule">
        {STEPS.map((step, index) => (
          <li
            key={step.title}
            // Реестр манифеста: строки на бумаге держатся линиями border-rule,
            // на десктопе — пять колонок с вертикальными делителями. Без карточек.
            className="group flex flex-col border-b border-rule bg-paper px-1 py-6 transition-colors duration-200 hover:bg-paper-raised focus-within:bg-paper-raised md:border-b-0 md:border-r md:px-5"
            onMouseEnter={() => onStepHover(index)}
            onMouseLeave={() => onStepHover(null)}
            onFocus={() => onStepHover(index)}
            onBlur={() => onStepHover(null)}
          >
            <div className="mb-4 flex items-center gap-3 border-b border-rule-soft pb-3">
              <span
                aria-hidden="true"
                className="tabular font-mono text-3xl font-extrabold leading-none tracking-[-0.02em] text-ink-soft transition-colors group-hover:text-stamp group-focus-within:text-stamp"
              >
                {String(index + 1).padStart(2, '0')}
              </span>
            </div>
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
