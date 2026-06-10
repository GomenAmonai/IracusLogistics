import {
  LANE_LABELS,
  LEAD_STATUS_LABELS,
  PAYMENT_STATUS_LABELS,
  SHIPMENT_STATUS_LABELS,
  type Lane,
  type LeadStatus,
  type PaymentStatus,
  type ShipmentStatus,
} from '../lib/types'

// Тоновые бейджи (как StatusBadge в WebApp): токены дизайн-системы напрямую, мягкая
// заливка через color-mix. done — зелёный «груз», alert — красный, active — бронза.
type Tone = { color: string; bg: string }

const DONE: Tone = { color: 'var(--color-cargo)', bg: 'color-mix(in srgb, var(--color-cargo) 12%, transparent)' }
const ALERT: Tone = { color: 'var(--color-alert)', bg: 'color-mix(in srgb, var(--color-alert) 12%, transparent)' }
const ACTIVE: Tone = { color: 'var(--color-accent)', bg: 'var(--color-accent-tint)' }
const IDLE: Tone = { color: 'var(--color-ink-soft)', bg: 'var(--color-surface-soft)' }

function Badge({ tone, label }: { tone: Tone; label: string }) {
  return (
    <span
      className="inline-flex items-center whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-semibold"
      style={{ color: tone.color, backgroundColor: tone.bg }}
    >
      {label}
    </span>
  )
}

export function LeadStatusBadge({ status }: { status: LeadStatus }) {
  const tone =
    status === 'converted' ? DONE : status === 'rejected' ? ALERT : status === 'new' ? ACTIVE : IDLE
  return <Badge tone={tone} label={LEAD_STATUS_LABELS[status]} />
}

export function ShipmentStatusBadge({ status }: { status: ShipmentStatus }) {
  const tone =
    status === 'delivered' ? DONE : status === 'cancelled' ? ALERT : status === 'pending' ? IDLE : ACTIVE
  return <Badge tone={tone} label={SHIPMENT_STATUS_LABELS[status]} />
}

export function LaneBadge({ lane }: { lane: Lane }) {
  // Полоса — атрибут, не состояние: всегда нейтральный тон, отличается только текстом.
  return <Badge tone={IDLE} label={LANE_LABELS[lane]} />
}

export function PaymentStatusBadge({ status }: { status: PaymentStatus }) {
  const tone = status === 'confirmed' ? DONE : status === 'refunded' ? ALERT : ACTIVE
  return <Badge tone={tone} label={PAYMENT_STATUS_LABELS[status]} />
}
