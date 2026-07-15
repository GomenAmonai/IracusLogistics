import { useState } from 'react'

import { MODE_LABELS, type ShippingMode } from '../lib/calc'
import { Section } from './Section'

export type CalcPrefill = {
  weight: string
  volume: string
  cargoType: string
  toCity: string
  mode: ShippingMode
}

const CARGO_TYPES = ['Электроника', 'Одежда и текстиль', 'Мебель', 'Оборудование', 'Сборный груз']
const MODES: ShippingMode[] = ['auto', 'rail', 'air']

type CalculatorProps = {
  onSendAsLead: (prefill: CalcPrefill) => void
}

export function Calculator({ onSendAsLead }: CalculatorProps) {
  const [weight, setWeight] = useState('800')
  const [volume, setVolume] = useState('4')
  const [mode, setMode] = useState<ShippingMode>('auto')
  const [cargoType, setCargoType] = useState(CARGO_TYPES[0])
  const [toCity, setToCity] = useState('Москва')

  const weightNum = Number(weight.replace(',', '.')) || 0
  const volumeNum = Number(volume.replace(',', '.')) || 0

  const hasInput = weightNum > 0 || volumeNum > 0

  return (
    <Section
      id="calc"
      surface
      eyebrow="Расчёт"
      title="Подготовьте данные для расчёта"
      intro="Укажите основные параметры груза и перенесите их в заявку. Стоимость, маршрут и срок менеджер рассчитает после проверки данных."
    >
      <div className="grid gap-6 rounded-2xl border border-line bg-surface p-6 shadow-card sm:p-8 lg:grid-cols-[1fr_1fr] lg:gap-8">
        {/* Панель-инпуты */}
        <div>
          <div className="grid gap-5 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <span className="field-label mb-2 block">Способ доставки</span>
              {/* Мягкий сегмент-контрол: активный — кобальтовая pill-заливка */}
              <div
                role="group"
                aria-label="Способ доставки"
                className="grid grid-cols-3 gap-1 rounded-full border border-line bg-surface-soft p-1"
              >
                {MODES.map((value) => {
                  const isActive = mode === value
                  return (
                    <button
                      key={value}
                      type="button"
                      aria-pressed={isActive}
                      onClick={() => setMode(value)}
                      className={`min-h-11 rounded-full px-3 py-2.5 font-mono text-sm uppercase tracking-[0.06em] transition-colors duration-200 ${
                        isActive
                          ? 'bg-accent text-surface'
                          : 'text-ink-soft hover:text-ink'
                      }`}
                    >
                      {MODE_LABELS[value]}
                    </button>
                  )
                })}
              </div>
            </div>

            <div>
              <label htmlFor="calc-weight" className="field-label mb-2 block">
                Вес, кг
              </label>
              <input
                id="calc-weight"
                className="input-field tabular font-mono"
                inputMode="decimal"
                value={weight}
                onChange={(event) => setWeight(event.target.value)}
                placeholder="800"
              />
            </div>

            <div>
              <label htmlFor="calc-volume" className="field-label mb-2 block">
                Объём, м³
              </label>
              <input
                id="calc-volume"
                className="input-field tabular font-mono"
                inputMode="decimal"
                value={volume}
                onChange={(event) => setVolume(event.target.value)}
                placeholder="4"
              />
            </div>

            <div>
              <label htmlFor="calc-cargo" className="field-label mb-2 block">
                Тип груза
              </label>
              <select
                id="calc-cargo"
                className="input-field"
                value={cargoType}
                onChange={(event) => setCargoType(event.target.value)}
              >
                {CARGO_TYPES.map((type) => (
                  <option key={type} value={type}>
                    {type}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="calc-city" className="field-label mb-2 block">
                Город назначения
              </label>
              <input
                id="calc-city"
                className="input-field"
                value={toCity}
                onChange={(event) => setToCity(event.target.value)}
                placeholder="Москва"
              />
            </div>
          </div>

          <p className="mt-5 text-sm leading-relaxed text-ink-soft">
            Эти параметры сохранятся в форме ниже. Никакие ставки или сроки на этом шаге
            не подставляются автоматически.
          </p>
        </div>

        {/* Блок результата — мягкая surface-карточка с терминал-показаниями */}
        <div className="flex flex-col justify-between rounded-2xl bg-surface-soft p-6 shadow-soft sm:p-7">
          <div>
            <p className="eyebrow mb-4">Вводные для менеджера</p>
            <div aria-live="polite">
              <p className="field-label">Предпочтительный способ</p>
              <p className="terminal mt-3 text-3xl font-bold tracking-[-0.02em] text-ink">
                {MODE_LABELS[mode]}
              </p>
              <div className="mt-6 flex items-center justify-between border-t border-line-soft pt-4">
                <span className="field-label">Параметры</span>
                <span className="terminal text-base">{weight || '—'} кг · {volume || '—'} м³</span>
              </div>
            </div>

            <p className="mt-5 text-sm leading-relaxed text-ink-soft">
              Публичный расчёт будет включён только после появления подтверждённых тарифов.
              Пока менеджер письменно фиксирует индивидуальные условия после проверки груза.
            </p>
          </div>

          <button
            type="button"
            disabled={!hasInput}
            onClick={() =>
              onSendAsLead({ weight, volume, cargoType, toCity, mode })
            }
            className="mt-6 inline-flex items-center justify-center rounded-full bg-accent px-6 py-3.5 text-base font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep disabled:cursor-not-allowed disabled:opacity-50"
          >
            Перенести в заявку
          </button>
        </div>
      </div>
    </Section>
  )
}
