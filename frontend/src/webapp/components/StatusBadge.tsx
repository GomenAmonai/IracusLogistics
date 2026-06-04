import { STATUS_LABELS, type ShipmentStatus } from '../lib/types'

// Тон бейджа задаётся токенами дизайн-системы напрямую (color-mix для мягкой заливки):
// доставлен — «груз» зелёный, отменён — alert, в работе — акцент, создан — нейтральный.
type Tone = { color: string; bg: string }

const DONE: Tone = { color: 'var(--color-cargo)', bg: 'color-mix(in srgb, var(--color-cargo) 12%, transparent)' }
const CANCELLED: Tone = { color: 'var(--color-alert)', bg: 'color-mix(in srgb, var(--color-alert) 12%, transparent)' }
const ACTIVE: Tone = { color: 'var(--color-accent)', bg: 'var(--color-accent-tint)' }
const IDLE: Tone = { color: 'var(--color-ink-soft)', bg: 'var(--color-surface-soft)' }

function toneFor(status: ShipmentStatus): Tone {
  if (status === 'delivered') {
    return DONE
  }
  if (status === 'cancelled') {
    return CANCELLED
  }
  if (status === 'pending') {
    return IDLE
  }
  return ACTIVE
}

export function StatusBadge({ status }: { status: ShipmentStatus }) {
  const tone = toneFor(status)
  return (
    <span
      className="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold"
      style={{ color: tone.color, backgroundColor: tone.bg }}
    >
      {STATUS_LABELS[status]}
    </span>
  )
}
