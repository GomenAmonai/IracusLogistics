export type ShippingMode = 'auto' | 'rail' | 'air'

export const MODE_LABELS: Record<ShippingMode, string> = {
  auto: 'Авто',
  rail: 'ЖД',
  air: 'Авиа',
}
