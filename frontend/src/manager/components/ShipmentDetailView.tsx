import { useCallback, useEffect, useState } from 'react'

import { ApiError, getShipment, updateShipmentStatus } from '../lib/api'
import { formatDateTime, formatMoney, formatVolume, formatWeight } from '../lib/format'
import {
  SHIPMENT_STATUS_LABELS,
  type ShipmentDetail,
  type ShipmentStatus,
} from '../lib/types'
import { LaneBadge, ShipmentStatusBadge } from './badges'
import { Chat } from './Chat'
import { PaymentBlock } from './PaymentBlock'
import { ErrorNote, Field, PrimaryButton, Spinner } from './ui'

export function ShipmentDetailView({
  shipmentId,
  onAuthError,
  onBack,
}: {
  shipmentId: string
  onAuthError: (err: unknown) => boolean
  onBack: () => void
}) {
  const [detail, setDetail] = useState<ShipmentDetail | null>(null)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    getShipment(shipmentId)
      .then(setDetail)
      .catch((err: unknown) => {
        if (!onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить груз.')
        }
      })
  }, [shipmentId, onAuthError])

  useEffect(() => {
    load()
  }, [load])

  if (error) {
    return <ErrorNote message={error} />
  }
  if (detail === null) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    )
  }

  const { shipment, history } = detail
  const facts = [
    formatWeight(shipment.weight),
    formatVolume(shipment.volume),
    formatMoney(shipment.price, shipment.currency),
  ].filter(Boolean)

  return (
    <div className="flex flex-col gap-6">
      <button
        type="button"
        onClick={onBack}
        className="self-start text-sm font-medium text-accent hover:text-accent-deep"
      >
        ← К списку грузов
      </button>

      <header className="flex flex-col gap-2">
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="terminal text-2xl font-semibold text-ink">{shipment.tracking_key}</h1>
          <LaneBadge lane={shipment.lane} />
          <ShipmentStatusBadge status={shipment.status} />
        </div>
        <p className="text-sm text-ink-soft">
          {[shipment.from_city, shipment.to_city].filter(Boolean).join(' → ')}
          {facts.length > 0 && ` · ${facts.join(' · ')}`}
        </p>
      </header>

      <div className="grid gap-6 lg:grid-cols-[3fr_2fr]">
        <div className="flex flex-col gap-6">
          <StatusControl
            current={shipment.status}
            shipmentId={shipment.id}
            onAuthError={onAuthError}
            onUpdated={load}
          />
          <Timeline history={history} />
          <PaymentBlock shipmentId={shipment.id} shipmentCurrency={shipment.currency} onAuthError={onAuthError} />
        </div>
        <Chat shipmentId={shipment.id} onAuthError={onAuthError} />
      </div>
    </div>
  )
}

function StatusControl({
  current,
  shipmentId,
  onAuthError,
  onUpdated,
}: {
  current: ShipmentStatus
  shipmentId: string
  onAuthError: (err: unknown) => boolean
  onUpdated: () => void
}) {
  const [status, setStatus] = useState<ShipmentStatus>(current)
  const [comment, setComment] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setSubmitting] = useState(false)

  async function apply() {
    setSubmitting(true)
    setError(null)
    try {
      await updateShipmentStatus(shipmentId, status, comment.trim())
      setComment('')
      onUpdated()
    } catch (err) {
      if (!onAuthError(err)) {
        setError(err instanceof ApiError ? err.message : 'Не удалось сменить статус.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-5 shadow-card">
      <h2 className="font-display text-base font-bold text-ink">Смена статуса</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field id="next-status" label="Новый статус">
          <select
            id="next-status"
            value={status}
            onChange={(e) => setStatus(e.target.value as ShipmentStatus)}
            className="input-field"
          >
            {(Object.keys(SHIPMENT_STATUS_LABELS) as ShipmentStatus[]).map((value) => (
              <option key={value} value={value}>
                {SHIPMENT_STATUS_LABELS[value]}
              </option>
            ))}
          </select>
        </Field>
        <Field id="status-comment" label="Комментарий (виден клиенту)">
          <input
            id="status-comment"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="Прошли таможню в Хоргосе"
            className="input-field"
          />
        </Field>
      </div>
      {error && <ErrorNote message={error} />}
      <div>
        <PrimaryButton onClick={() => void apply()} disabled={isSubmitting || status === current}>
          {isSubmitting ? 'Применяем…' : 'Применить'}
        </PrimaryButton>
      </div>
      <p className="text-xs text-ink-soft">
        Клиент получит уведомление в Telegram о смене статуса.
      </p>
    </section>
  )
}

function Timeline({ history }: { history: ShipmentDetail['history'] }) {
  return (
    <section className="rounded-2xl border border-line bg-surface p-5 shadow-card">
      <h2 className="font-display text-base font-bold text-ink">История статусов</h2>
      <ol className="mt-4 flex flex-col gap-0">
        {history.map((event, index) => (
          <li key={event.id} className="relative flex gap-3 pb-4 last:pb-0">
            {index < history.length - 1 && (
              <span aria-hidden="true" className="absolute left-[4px] top-3 h-full w-px bg-line" />
            )}
            <span
              aria-hidden="true"
              className="relative z-10 mt-1.5 h-2 w-2 shrink-0 rounded-full bg-accent"
            />
            <div className="flex flex-col gap-0.5">
              <span className="text-sm font-semibold text-ink">
                {SHIPMENT_STATUS_LABELS[event.status]}
              </span>
              {event.comment && <span className="text-sm text-ink-soft">{event.comment}</span>}
              <span className="terminal text-xs text-ink-soft">{formatDateTime(event.created_at)}</span>
            </div>
          </li>
        ))}
      </ol>
    </section>
  )
}
