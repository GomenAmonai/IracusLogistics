import { useEffect, useState, type FormEvent } from 'react'

import {
  ApiError,
  createPayment,
  listPayments,
  updatePaymentStatus,
  type CreatePaymentInput,
} from '../lib/api'
import { formatDateTime, formatMoney } from '../lib/format'
import {
  PAYMENT_CHANNEL_LABELS,
  PAYMENT_STATUS_LABELS,
  type Payment,
  type PaymentChannel,
  type PaymentStatus,
} from '../lib/types'
import { PaymentStatusBadge } from './badges'
import { ErrorNote, Field, GhostButton, PrimaryButton, Spinner } from './ui'

export function PaymentBlock({
  shipmentId,
  shipmentCurrency,
  onAuthError,
}: {
  shipmentId: string
  shipmentCurrency: string
  onAuthError: (err: unknown) => boolean
}) {
  const [payments, setPayments] = useState<Payment[] | null>(null)
  const [isAdding, setAdding] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    listPayments(shipmentId)
      .then((loaded) => {
        if (active) setPayments(loaded)
      })
      .catch((err: unknown) => {
        if (active && !onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить платежи.')
        }
      })
    return () => {
      active = false
    }
  }, [shipmentId, onAuthError])

  async function handleStatusChange(payment: Payment, status: PaymentStatus) {
    try {
      const updated = await updatePaymentStatus(shipmentId, payment.id, status)
      setPayments((current) => current?.map((p) => (p.id === updated.id ? updated : p)) ?? null)
    } catch (err) {
      if (!onAuthError(err)) {
        setError(err instanceof ApiError ? err.message : 'Не удалось сменить статус платежа.')
      }
    }
  }

  return (
    <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <h2 className="font-display text-base font-bold text-ink">Платежи</h2>
        {!isAdding && <GhostButton onClick={() => setAdding(true)}>Записать платёж</GhostButton>}
      </div>

      {error && <ErrorNote message={error} />}

      {payments === null ? (
        <Spinner />
      ) : payments.length === 0 && !isAdding ? (
        <p className="text-sm text-ink-soft">Платежей пока нет.</p>
      ) : (
        <ul className="flex flex-col divide-y divide-line-soft">
          {payments.map((payment) => (
            <li key={payment.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
              <div className="flex flex-col gap-0.5">
                <span className="terminal text-sm font-semibold text-ink">
                  {formatMoney(payment.amount, payment.currency)}
                </span>
                <span className="text-xs text-ink-soft">
                  {PAYMENT_CHANNEL_LABELS[payment.channel]} · {formatDateTime(payment.created_at)}
                  {payment.comment && ` · ${payment.comment}`}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <PaymentStatusBadge status={payment.status} />
                <label htmlFor={`payment-status-${payment.id}`} className="sr-only">
                  Статус платежа
                </label>
                <select
                  id={`payment-status-${payment.id}`}
                  value={payment.status}
                  onChange={(e) => void handleStatusChange(payment, e.target.value as PaymentStatus)}
                  className="input-field max-w-36 py-1 text-xs"
                >
                  {(Object.keys(PAYMENT_STATUS_LABELS) as PaymentStatus[]).map((status) => (
                    <option key={status} value={status}>
                      {PAYMENT_STATUS_LABELS[status]}
                    </option>
                  ))}
                </select>
              </div>
            </li>
          ))}
        </ul>
      )}

      {isAdding && (
        <PaymentForm
          shipmentId={shipmentId}
          defaultCurrency={shipmentCurrency}
          onAuthError={onAuthError}
          onCancel={() => setAdding(false)}
          onCreated={(payment) => {
            setPayments((current) => [...(current ?? []), payment])
            setAdding(false)
          }}
        />
      )}
    </section>
  )
}

function PaymentForm({
  shipmentId,
  defaultCurrency,
  onAuthError,
  onCancel,
  onCreated,
}: {
  shipmentId: string
  defaultCurrency: string
  onAuthError: (err: unknown) => boolean
  onCancel: () => void
  onCreated: (payment: Payment) => void
}) {
  const [amount, setAmount] = useState('')
  const [currency, setCurrency] = useState(defaultCurrency)
  const [channel, setChannel] = useState<PaymentChannel>('bank_transfer')
  const [status, setStatus] = useState<PaymentStatus>('pending')
  const [comment, setComment] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)

    const input: CreatePaymentInput = { amount: amount.trim(), channel, status }
    if (currency.trim()) input.currency = currency.trim()
    if (comment.trim()) input.comment = comment.trim()

    try {
      onCreated(await createPayment(shipmentId, input))
    } catch (err) {
      if (!onAuthError(err)) {
        setError(err instanceof ApiError ? err.message : 'Не удалось записать платёж.')
      }
      setSubmitting(false)
    }
  }

  return (
    <form
      onSubmit={(e) => void handleSubmit(e)}
      className="flex flex-col gap-4 rounded-xl border border-line-soft bg-surface-soft p-4"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <Field id="payment-amount" label="Сумма *">
          <input
            id="payment-amount"
            required
            inputMode="decimal"
            placeholder="1250.50"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="input-field"
          />
        </Field>
        <Field id="payment-currency" label="Валюта">
          <select
            id="payment-currency"
            value={currency}
            onChange={(e) => setCurrency(e.target.value)}
            className="input-field"
          >
            {[defaultCurrency, 'RUB', 'USD', 'CNY']
              .filter((value, index, list) => list.indexOf(value) === index)
              .map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
          </select>
        </Field>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <Field id="payment-channel" label="Канал">
          <select
            id="payment-channel"
            value={channel}
            onChange={(e) => setChannel(e.target.value as PaymentChannel)}
            className="input-field"
          >
            {(Object.keys(PAYMENT_CHANNEL_LABELS) as PaymentChannel[]).map((value) => (
              <option key={value} value={value}>
                {PAYMENT_CHANNEL_LABELS[value]}
              </option>
            ))}
          </select>
        </Field>
        <Field id="payment-status" label="Статус">
          <select
            id="payment-status"
            value={status}
            onChange={(e) => setStatus(e.target.value as PaymentStatus)}
            className="input-field"
          >
            {(Object.keys(PAYMENT_STATUS_LABELS) as PaymentStatus[]).map((value) => (
              <option key={value} value={value}>
                {PAYMENT_STATUS_LABELS[value]}
              </option>
            ))}
          </select>
        </Field>
      </div>

      <Field id="payment-comment" label="Комментарий">
        <input
          id="payment-comment"
          placeholder="Аванс 50%"
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          className="input-field"
        />
      </Field>

      {error && <ErrorNote message={error} />}

      <div className="flex items-center gap-3">
        <PrimaryButton type="submit" disabled={isSubmitting || !amount.trim()}>
          {isSubmitting ? 'Записываем…' : 'Записать'}
        </PrimaryButton>
        <GhostButton onClick={onCancel}>Отмена</GhostButton>
      </div>
    </form>
  )
}
