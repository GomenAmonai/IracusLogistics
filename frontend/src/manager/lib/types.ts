// Типы панели менеджера. Поля 1:1 с доменными моделями бэкенда; decimal-поля приходят
// строками (numeric сериализуется строкой, чтобы не терять точность).

export type LeadStatus = 'new' | 'contacted' | 'converted' | 'rejected'

export type Lead = {
  id: string
  name: string
  phone: string
  from_city: string
  to_city: string
  weight: string | null
  volume: string | null
  cargo_type: string
  comment: string
  status: LeadStatus
  created_at: string
}

export type Client = {
  id: string
  telegram_id: number
  username: string
  name: string
  phone: string
  lead_id: string | null
  created_at: string
}

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
  client_id: string
  manager_id: string
  tracking_key: string
  lane: Lane
  status: ShipmentStatus
  status_comment: string
  weight: string | null
  volume: string | null
  from_city: string
  to_city: string
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

export type Message = {
  id: string
  text: string
  from_role: 'client' | 'manager'
  created_at: string
}

export type PaymentChannel = 'bank_transfer' | 'card_sbp' | 'cash' | 'crypto'

export type PaymentStatus = 'pending' | 'confirmed' | 'refunded'

export type Payment = {
  id: string
  shipment_id: string
  amount: string
  currency: string
  channel: PaymentChannel
  status: PaymentStatus
  comment: string
  created_at: string
  updated_at: string
}

export const LEAD_STATUS_LABELS: Record<LeadStatus, string> = {
  new: 'Новый',
  contacted: 'На связи',
  converted: 'Конвертирован',
  rejected: 'Отклонён',
}

export const SHIPMENT_STATUS_LABELS: Record<ShipmentStatus, string> = {
  pending: 'Создан',
  picked_up: 'Забран',
  in_transit: 'В пути',
  customs_clear: 'Таможня',
  in_warehouse: 'На складе',
  out_for_delivery: 'Доставляется',
  delivered: 'Доставлен',
  cancelled: 'Отменён',
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
  crypto: 'USDT / крипта',
}

export const PAYMENT_STATUS_LABELS: Record<PaymentStatus, string> = {
  pending: 'Ожидается',
  confirmed: 'Получен',
  refunded: 'Возврат',
}
