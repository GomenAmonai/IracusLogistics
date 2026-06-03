// NOTE: MVP — формула цены захардкожена; см. docs/tech-debt.md
// Результат калькулятора — оценка, в БД НЕ сохраняется. Заявка пишет только сырые
// входы (weight/volume/cargo_type/to_city). Реальную цену фиксирует менеджер.

export type ShippingMode = 'auto' | 'rail' | 'air'

type Rate = {
  perKg: number
  perM3: number
  minUsd: number
  etaDays: [number, number]
}

// Базовые ставки Китай→Россия (USD), правдоподобные для MVP.
const RATES: Record<ShippingMode, Rate> = {
  auto: { perKg: 2.6, perM3: 190, minUsd: 120, etaDays: [18, 28] },
  rail: { perKg: 2.1, perM3: 150, minUsd: 150, etaDays: [25, 40] },
  air: { perKg: 6.5, perM3: 0, minUsd: 90, etaDays: [6, 12] },
}

// Фиксированная неопределённость MVP: курс, таможенная стоимость, упаковка.
const SPREAD = 0.18

// Округление вверх до $10 для «инструментального» вида показаний.
function roundTo10(value: number): number {
  return Math.ceil(value / 10) * 10
}

export type CalcResult = {
  low: number
  high: number
  etaDays: [number, number]
  chargeableKg: number
}

export function calcPriceRange(
  weightKg: number,
  volumeM3: number,
  mode: ShippingMode,
): CalcResult {
  const rate = RATES[mode]

  // Объёмный вес (отраслевой коэффициент): авиа 167 кг/м³, авто/жд 250 (карго плотнее).
  const volWeightKg = volumeM3 * (mode === 'air' ? 167 : 250)
  // Расчётный вес — больше из фактического и объёмного, как считают реальные карго.
  const chargeableKg = Math.max(weightKg, volWeightKg)

  const byWeight = chargeableKg * rate.perKg
  // Авиа считает только по весу.
  const byVolume = mode === 'air' ? 0 : volumeM3 * rate.perM3
  // Защита от микрозаявок минимальной ставкой.
  const base = Math.max(byWeight, byVolume, rate.minUsd)

  return {
    low: roundTo10(base * (1 - SPREAD)),
    high: roundTo10(base * (1 + SPREAD)),
    etaDays: rate.etaDays,
    chargeableKg: Math.round(chargeableKg),
  }
}

export const MODE_LABELS: Record<ShippingMode, string> = {
  auto: 'Авто',
  rail: 'ЖД',
  air: 'Авиа',
}
