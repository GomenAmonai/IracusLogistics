// Типы клиентского WebApp. Поля 1:1 с доменными моделями бэкенда (Shipment, Message,
// ShipmentStatusEvent). decimal-поля приходят строками — бэкенд сериализует numeric как
// строку, чтобы не терять точность; в UI их форматируем, но не считаем.

export type ShipmentStatus =
  | 'pending'
  | 'picked_up'
  | 'in_transit'
  | 'customs_clear'
  | 'in_warehouse'
  | 'out_for_delivery'
  | 'delivered'
  | 'cancelled'

export type Lane = 'cargo' | 'white' | 'buyout'

export type Shipment = {
  id: string
  tracking_key: string
  lane: Lane
  status: ShipmentStatus
  status_comment: string
  from_city: string
  to_city: string
  weight: string | null
  volume: string | null
  price: string | null
  currency: string
  estimated_at: string | null
  delivered_at: string | null
  created_at: string
  updated_at: string
}

export type StatusEvent = {
  id: string
  status: ShipmentStatus
  comment: string
  created_at: string
}

export type ShipmentDetail = {
  shipment: Shipment
  history: StatusEvent[]
}

export type PaymentChannel = 'bank_transfer' | 'card_sbp' | 'cash' | 'crypto'

export type PaymentStatus = 'pending' | 'confirmed' | 'refunded'

export type Payment = {
  id: string
  amount: string
  currency: string
  channel: PaymentChannel
  status: PaymentStatus
  comment: string
  created_at: string
}

export type Message = {
  id: string
  text: string
  from_role: 'client' | 'manager'
  created_at: string
}

export type Client = {
  id: string
  name: string
  username: string
}

export const LANE_LABELS: Record<Lane, string> = {
  cargo: 'Карго',
  white: 'Белый импорт',
  buyout: 'Выкуп',
}

export const PAYMENT_CHANNEL_LABELS: Record<PaymentChannel, string> = {
  bank_transfer: 'Безнал по счёту',
  card_sbp: 'Карта / СБП',
  cash: 'Наличные',
  crypto: 'Криптовалюта',
}

export const PAYMENT_STATUS_LABELS: Record<PaymentStatus, string> = {
  pending: 'Ожидает оплаты',
  confirmed: 'Получен',
  refunded: 'Возвращён',
}

export const STATUS_LABELS: Record<ShipmentStatus, string> = {
  pending: 'Создан',
  picked_up: 'Забран',
  in_transit: 'В пути',
  customs_clear: 'Таможня',
  in_warehouse: 'На складе',
  out_for_delivery: 'Доставляется',
  delivered: 'Доставлен',
  cancelled: 'Отменён',
}
