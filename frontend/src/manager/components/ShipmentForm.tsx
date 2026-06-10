import { useEffect, useState, type FormEvent } from 'react'

import { ApiError, createShipment, listClients, type CreateShipmentInput } from '../lib/api'
import { LANE_LABELS, type Client, type Lane } from '../lib/types'
import { ErrorNote, Field, GhostButton, PrimaryButton, Spinner } from './ui'

export function ShipmentForm({
  onAuthError,
  presetClientId,
  onCancel,
  onCreated,
}: {
  onAuthError: (err: unknown) => boolean
  presetClientId?: string
  onCancel: () => void
  onCreated: (id: string) => void
}) {
  const [clients, setClients] = useState<Client[] | null>(null)
  const [clientId, setClientId] = useState(presetClientId ?? '')
  const [lane, setLane] = useState<Lane>('cargo')
  const [fromCity, setFromCity] = useState('Гуанчжоу')
  const [toCity, setToCity] = useState('Москва')
  const [weight, setWeight] = useState('')
  const [volume, setVolume] = useState('')
  const [price, setPrice] = useState('')
  const [currency, setCurrency] = useState('USD')
  const [statusNote, setStatusNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setSubmitting] = useState(false)

  useEffect(() => {
    let active = true
    listClients()
      .then((loaded) => {
        if (active) setClients(loaded)
      })
      .catch((err: unknown) => {
        if (active && !onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить клиентов.')
        }
      })
    return () => {
      active = false
    }
  }, [onAuthError])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)

    // Пустые опциональные поля не отправляем: бэкенд трактует отсутствие как «не задано»
    // (NullDecimal), а пустая строка в decimal-поле была бы ошибкой парсинга.
    const input: CreateShipmentInput = { client_id: clientId, lane }
    if (fromCity.trim()) input.from_city = fromCity.trim()
    if (toCity.trim()) input.to_city = toCity.trim()
    if (weight.trim()) input.weight = weight.trim()
    if (volume.trim()) input.volume = volume.trim()
    if (price.trim()) input.price = price.trim()
    if (currency.trim()) input.currency = currency.trim()
    if (statusNote.trim()) input.status_note = statusNote.trim()

    try {
      const shipment = await createShipment(input)
      onCreated(shipment.id)
    } catch (err) {
      if (!onAuthError(err)) {
        setError(err instanceof ApiError ? err.message : 'Не удалось создать груз.')
      }
      setSubmitting(false)
    }
  }

  if (clients === null && !error) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    )
  }

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="flex max-w-2xl flex-col gap-5">
      <header>
        <h1 className="font-display text-2xl font-bold text-ink">Новый груз</h1>
        <p className="mt-1 text-sm text-ink-soft">
          Трек-ключ сгенерируется автоматически; клиент получит доступ к грузу в WebApp.
        </p>
      </header>

      <div className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6 shadow-card">
        <Field id="client" label="Клиент *">
          <select
            id="client"
            required
            value={clientId}
            onChange={(e) => setClientId(e.target.value)}
            className="input-field"
          >
            <option value="" disabled>
              Выберите клиента
            </option>
            {(clients ?? []).map((client) => (
              <option key={client.id} value={client.id}>
                {client.name || 'Без имени'}
                {client.username ? ` (@${client.username})` : ''}
              </option>
            ))}
          </select>
        </Field>

        <Field id="lane" label="Полоса">
          <select
            id="lane"
            value={lane}
            onChange={(e) => setLane(e.target.value as Lane)}
            className="input-field"
          >
            {(Object.keys(LANE_LABELS) as Lane[]).map((value) => (
              <option key={value} value={value}>
                {LANE_LABELS[value]}
              </option>
            ))}
          </select>
        </Field>

        <div className="grid gap-4 sm:grid-cols-2">
          <Field id="from-city" label="Откуда">
            <input
              id="from-city"
              value={fromCity}
              onChange={(e) => setFromCity(e.target.value)}
              className="input-field"
            />
          </Field>
          <Field id="to-city" label="Куда">
            <input
              id="to-city"
              value={toCity}
              onChange={(e) => setToCity(e.target.value)}
              className="input-field"
            />
          </Field>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <Field id="weight" label="Вес, кг">
            <input
              id="weight"
              inputMode="decimal"
              placeholder="800"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
              className="input-field"
            />
          </Field>
          <Field id="volume" label="Объём, м³">
            <input
              id="volume"
              inputMode="decimal"
              placeholder="4"
              value={volume}
              onChange={(e) => setVolume(e.target.value)}
              className="input-field"
            />
          </Field>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <Field id="price" label="Цена">
            <input
              id="price"
              inputMode="decimal"
              placeholder="2500"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              className="input-field"
            />
          </Field>
          <Field id="currency" label="Валюта">
            <select
              id="currency"
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              className="input-field"
            >
              <option value="USD">USD</option>
              <option value="RUB">RUB</option>
              <option value="CNY">CNY</option>
            </select>
          </Field>
        </div>

        <Field id="status-note" label="Комментарий к начальному статусу">
          <input
            id="status-note"
            placeholder="Партия согласована, ждём выкуп"
            value={statusNote}
            onChange={(e) => setStatusNote(e.target.value)}
            className="input-field"
          />
        </Field>
      </div>

      {error && <ErrorNote message={error} />}

      <div className="flex items-center gap-3">
        <PrimaryButton type="submit" disabled={isSubmitting || !clientId}>
          {isSubmitting ? 'Создаём…' : 'Создать груз'}
        </PrimaryButton>
        <GhostButton onClick={onCancel}>Отмена</GhostButton>
      </div>
    </form>
  )
}
