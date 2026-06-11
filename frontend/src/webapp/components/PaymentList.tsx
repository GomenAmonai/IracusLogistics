import { useEffect, useState } from 'react'

import { listPayments } from '../lib/api'
import { formatDate, formatMoney } from '../lib/format'
import { PAYMENT_CHANNEL_LABELS, PAYMENT_STATUS_LABELS, type Payment, type PaymentStatus } from '../lib/types'

// Тона статусов платежа — те же токены, что у StatusBadge груза: получен — зелёный,
// возвращён — alert, ожидает — акцент.
const STATUS_TONES: Record<PaymentStatus, { color: string; bg: string }> = {
  pending: { color: 'var(--color-accent)', bg: 'var(--color-accent-tint)' },
  confirmed: { color: 'var(--color-cargo)', bg: 'color-mix(in srgb, var(--color-cargo) 12%, transparent)' },
  refunded: { color: 'var(--color-alert)', bg: 'color-mix(in srgb, var(--color-alert) 12%, transparent)' },
}

// Блок «Платежи» в деталях груза. Без платежей (или пока бэкенд без этой ручки — 404)
// секция не рендерится вовсе: клиенту незачем видеть пустой блок.
export function PaymentList({ shipmentId }: { shipmentId: string }) {
  const [payments, setPayments] = useState<Payment[]>([])

  useEffect(() => {
    let active = true
    listPayments(shipmentId)
      .then((loaded) => {
        if (active) {
          setPayments(loaded ?? [])
        }
      })
      .catch(() => {
        // Платежи — дополнение к деталям: любая ошибка загрузки скрывает блок, не экран.
      })
    return () => {
      active = false
    }
  }, [shipmentId])

  if (payments.length === 0) {
    return null
  }

  return (
    <section className="flex flex-col gap-3">
      <h2 className="font-display text-base font-semibold text-ink">Платежи</h2>
      <ul className="flex flex-col divide-y divide-line rounded-2xl border border-line bg-surface shadow-card">
        {payments.map((payment) => {
          const tone = STATUS_TONES[payment.status]
          return (
            <li key={payment.id} className="flex flex-col gap-1 p-4">
              <div className="flex items-center justify-between gap-3">
                <span className="terminal text-sm font-semibold text-ink">
                  {formatMoney(payment.amount, payment.currency)}
                </span>
                <span
                  className="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold"
                  style={{ color: tone.color, backgroundColor: tone.bg }}
                >
                  {PAYMENT_STATUS_LABELS[payment.status]}
                </span>
              </div>
              <div className="flex items-center justify-between gap-3 text-xs text-ink-soft">
                <span>{PAYMENT_CHANNEL_LABELS[payment.channel] ?? payment.channel}</span>
                <span>{formatDate(payment.created_at)}</span>
              </div>
              {payment.comment && <p className="text-xs leading-relaxed text-ink-soft">{payment.comment}</p>}
            </li>
          )
        })}
      </ul>
    </section>
  )
}
