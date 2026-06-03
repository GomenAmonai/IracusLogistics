import { useMemo, useState } from 'react'

import { calcPriceRange, MODE_LABELS, type ShippingMode } from '../lib/calc'
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

  const result = useMemo(
    () => calcPriceRange(weightNum, volumeNum, mode),
    [weightNum, volumeNum, mode],
  )

  const hasInput = weightNum > 0 || volumeNum > 0

  return (
    <Section
      id="calc"
      surface
      eyebrow="Тариф"
      title="Калькулятор диапазона цены"
      intro="Введите параметры груза — увидите предварительную вилку и срок доставки. Это оценка по нашим базовым ставкам, а не оффер."
    >
      <div className="grid border border-rule bg-paper lg:grid-cols-[1fr_1fr]">
        {/* Панель-инпуты */}
        <div className="border-b border-rule p-6 sm:p-8 lg:border-b-0 lg:border-r">
          <div className="grid gap-5 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <span className="field-label mb-2 block">Способ доставки</span>
              {/* Сегменты-тоггл: активный — чернильная заливка, не цвет; разделители-линии */}
              <div
                role="group"
                aria-label="Способ доставки"
                className="grid grid-cols-3 border border-ink"
              >
                {MODES.map((value, index) => {
                  const isActive = mode === value
                  return (
                    <button
                      key={value}
                      type="button"
                      aria-pressed={isActive}
                      onClick={() => setMode(value)}
                      className={`px-3 py-3 font-mono text-sm uppercase tracking-[0.06em] transition-colors duration-200 ${
                        index > 0 ? 'border-l border-ink' : ''
                      } ${
                        isActive
                          ? 'bg-ink text-paper'
                          : 'bg-paper-raised text-ink-soft hover:text-ink'
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
            Расчётный вес — больше из фактического и объёмного:{' '}
            <span className="tabular font-mono text-ink">{result.chargeableKg} кг</span>.
          </p>
        </div>

        {/* Вывод вилки — панель-выписка манифеста */}
        <div className="flex flex-col justify-between p-6 sm:p-8">
          <div>
            <p className="eyebrow mb-4">Предварительная оценка</p>
            <div
              className="border border-rule bg-paper-raised"
              aria-live="polite"
            >
              <p className="border-b border-rule px-5 py-3 font-mono text-[0.7rem] uppercase tracking-[0.1em] text-ink-soft">
                Диапазон стоимости
              </p>
              {/* Цена — единственный штамп-акцент блока */}
              <p className="tabular flex flex-wrap items-baseline gap-x-2 gap-y-1 px-5 py-5 font-mono text-2xl font-bold leading-tight tracking-[-0.02em] text-ink sm:text-3xl md:text-4xl">
                <span className="text-ink-soft">от</span>
                <span className="whitespace-nowrap text-stamp">
                  ${result.low.toLocaleString('ru-RU')}
                </span>
                <span className="text-ink-soft">до</span>
                <span className="whitespace-nowrap text-stamp">
                  ${result.high.toLocaleString('ru-RU')}
                </span>
              </p>
              <div className="flex items-center justify-between border-t border-rule px-5 py-4">
                <span className="font-mono text-[0.7rem] uppercase tracking-[0.1em] text-ink-soft">
                  Срок доставки
                </span>
                <span className="tabular font-mono text-base text-ink">
                  {result.etaDays[0]}–{result.etaDays[1]} дней
                </span>
              </div>
            </div>

            <p className="mt-5 text-sm leading-relaxed text-ink-soft">
              Предварительная оценка — точную сумму фиксирует менеджер после проверки
              документов и характеристик груза.
            </p>
          </div>

          <button
            type="button"
            disabled={!hasInput}
            onClick={() =>
              onSendAsLead({ weight, volume, cargoType, toCity, mode })
            }
            className="mt-6 inline-flex items-center justify-center bg-stamp px-6 py-3.5 font-mono text-sm font-semibold uppercase tracking-[0.08em] text-paper transition-colors duration-200 hover:bg-stamp-deep disabled:cursor-not-allowed disabled:opacity-50"
          >
            Отправить как заявку
          </button>
        </div>
      </div>
    </Section>
  )
}
