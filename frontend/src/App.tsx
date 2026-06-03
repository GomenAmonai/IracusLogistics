import { type FormEvent, type ReactNode, useEffect, useState } from 'react'
import './App.css'

const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

type ShipmentStatus =
  | 'new'
  | 'needs_clarification'
  | 'in_calculation'
  | 'priced'
  | 'offer_sent'
  | 'won'
  | 'lost'

type ShipmentRequest = {
  id: string
  status: ShipmentStatus
  customer_name: string
  contact: string
  company_name?: string | null
  cargo_name: string
  origin_city?: string | null
  destination_city: string
  weight_kg?: number | null
  volume_m3?: number | null
  boxes_count?: number | null
  cargo_value?: number | null
  cargo_currency?: string | null
  comment: string
  manager_comment?: string | null
  created_at: string
  updated_at: string
}

type ShipmentDraft = {
  status: ShipmentStatus
  managerComment: string
}

type ShipmentFormState = {
  customerName: string
  contact: string
  companyName: string
  cargoName: string
  originCity: string
  destinationCity: string
  weightKg: string
  volumeM3: string
  boxesCount: string
  cargoValue: string
  cargoCurrency: string
  comment: string
}

const statusOptions: Array<{ value: ShipmentStatus; label: string }> = [
  { value: 'new', label: 'Новая' },
  { value: 'needs_clarification', label: 'Уточнение' },
  { value: 'in_calculation', label: 'Расчёт' },
  { value: 'priced', label: 'Цена готова' },
  { value: 'offer_sent', label: 'Оффер отправлен' },
  { value: 'won', label: 'Сделка' },
  { value: 'lost', label: 'Потеряна' },
]

const statusLabels: Record<ShipmentStatus, string> = {
  new: 'Новая',
  needs_clarification: 'Уточнение',
  in_calculation: 'Расчёт',
  priced: 'Цена готова',
  offer_sent: 'Оффер отправлен',
  won: 'Сделка',
  lost: 'Потеряна',
}

const emptyForm: ShipmentFormState = {
  customerName: '',
  contact: '',
  companyName: '',
  cargoName: '',
  originCity: '',
  destinationCity: '',
  weightKg: '',
  volumeM3: '',
  boxesCount: '',
  cargoValue: '',
  cargoCurrency: '',
  comment: '',
}

function App() {
  const [form, setForm] = useState<ShipmentFormState>(emptyForm)
  const [requests, setRequests] = useState<ShipmentRequest[]>([])
  const [drafts, setDrafts] = useState<Record<string, ShipmentDraft>>({})
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [savingId, setSavingId] = useState<string | null>(null)
  const [message, setMessage] = useState('Готово к работе.')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void loadRequests()
  }, [])

  async function loadRequests() {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`${apiUrl}/api/shipment-requests`)
      if (!response.ok) {
        throw new Error(`API returned ${response.status}`)
      }

      const items = (await response.json()) as ShipmentRequest[]
      setRequests(items)
      setDrafts(buildDrafts(items))
      setMessage(`Загружено заявок: ${items.length}`)
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить заявки')
    } finally {
      setLoading(false)
    }
  }

  function buildDrafts(items: ShipmentRequest[]) {
    return items.reduce<Record<string, ShipmentDraft>>((acc, item) => {
      acc[item.id] = {
        status: item.status,
        managerComment: item.manager_comment ?? '',
      }
      return acc
    }, {})
  }

  function updateFormField<K extends keyof ShipmentFormState>(field: K, value: ShipmentFormState[K]) {
    setForm((current) => ({ ...current, [field]: value }))
  }

  function updateDraftField(id: string, field: keyof ShipmentDraft, value: string) {
    setDrafts((current) => ({
      ...current,
      [id]:
        field === 'status'
          ? {
              status: value as ShipmentStatus,
              managerComment: current[id]?.managerComment ?? '',
            }
          : {
              status: current[id]?.status ?? 'new',
              managerComment: value,
            },
    }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError(null)

    try {
      const response = await fetch(`${apiUrl}/api/shipment-requests`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(toCreatePayload(form)),
      })

      const payload = (await response.json()) as { error?: string; id?: string }
      if (!response.ok) {
        throw new Error(payload.error ?? `API returned ${response.status}`)
      }

      setForm(emptyForm)
      setMessage(`Заявка ${payload.id ?? 'создана'} сохранена.`)
      await loadRequests()
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Не удалось отправить заявку')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleSave(request: ShipmentRequest) {
    const draft = drafts[request.id]
    if (!draft) {
      return
    }

    setSavingId(request.id)
    setError(null)

    try {
      const response = await fetch(`${apiUrl}/api/shipment-requests/${request.id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: draft.status,
          manager_comment: draft.managerComment.trim() || null,
        }),
      })

      const payload = (await response.json()) as { error?: string }
      if (!response.ok) {
        throw new Error(payload.error ?? `API returned ${response.status}`)
      }

      setMessage(`Заявка ${request.id} обновлена.`)
      await loadRequests()
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : 'Не удалось обновить заявку')
    } finally {
      setSavingId(null)
    }
  }

  return (
    <main className="app-shell">
      <header className="hero">
        <div className="hero__copy">
          <p className="eyebrow">Iracus Logistic</p>
          <h1>Заявки на доставку из Китая</h1>
          <p className="lead">
            Публичная форма собирает вводные, backend сохраняет их в Postgres,
            менеджер ведет статус и комментарий в одном месте.
          </p>
        </div>

        <div className="hero__meta">
          <div className="meta-card">
            <span className="meta-card__label">API</span>
            <strong>{apiUrl}</strong>
          </div>
          <div className="meta-card">
            <span className="meta-card__label">Заявок</span>
            <strong>{requests.length}</strong>
          </div>
          <div className="meta-card meta-card--wide">
            <span className="meta-card__label">Статус</span>
            <strong>{message}</strong>
          </div>
        </div>
      </header>

      <section className="workspace">
        <article className="panel panel--form">
          <div className="panel__head">
            <div>
              <p className="panel__eyebrow">Публичная форма</p>
              <h2>Новая заявка</h2>
            </div>
            <span className="panel__hint">Нажимает клиент, заполняет менеджер</span>
          </div>

          <form className="form-grid" onSubmit={handleSubmit}>
            <Field label="Имя" required>
              <input
                value={form.customerName}
                onChange={(event) => updateFormField('customerName', event.target.value)}
                placeholder="Иван Петров"
              />
            </Field>

            <Field label="Телефон или мессенджер" required>
              <input
                value={form.contact}
                onChange={(event) => updateFormField('contact', event.target.value)}
                placeholder="+7 999 111-22-33"
              />
            </Field>

            <Field label="Компания">
              <input
                value={form.companyName}
                onChange={(event) => updateFormField('companyName', event.target.value)}
                placeholder="ООО Пример"
              />
            </Field>

            <Field label="Груз" required>
              <input
                value={form.cargoName}
                onChange={(event) => updateFormField('cargoName', event.target.value)}
                placeholder="Светильники"
              />
            </Field>

            <Field label="Город отправки">
              <input
                value={form.originCity}
                onChange={(event) => updateFormField('originCity', event.target.value)}
                placeholder="Guangzhou"
              />
            </Field>

            <Field label="Город доставки" required>
              <input
                value={form.destinationCity}
                onChange={(event) => updateFormField('destinationCity', event.target.value)}
                placeholder="Moscow"
              />
            </Field>

            <Field label="Вес, кг">
              <input
                inputMode="decimal"
                value={form.weightKg}
                onChange={(event) => updateFormField('weightKg', event.target.value)}
                placeholder="12.5"
              />
            </Field>

            <Field label="Объём, м3">
              <input
                inputMode="decimal"
                value={form.volumeM3}
                onChange={(event) => updateFormField('volumeM3', event.target.value)}
                placeholder="0.42"
              />
            </Field>

            <Field label="Коробок">
              <input
                inputMode="numeric"
                value={form.boxesCount}
                onChange={(event) => updateFormField('boxesCount', event.target.value)}
                placeholder="8"
              />
            </Field>

            <Field label="Стоимость груза">
              <div className="inline-fields">
                <input
                  inputMode="decimal"
                  value={form.cargoValue}
                  onChange={(event) => updateFormField('cargoValue', event.target.value)}
                  placeholder="1500"
                />
                <input
                  value={form.cargoCurrency}
                  onChange={(event) => updateFormField('cargoCurrency', event.target.value)}
                  placeholder="USD"
                />
              </div>
            </Field>

            <Field label="Комментарий" required className="field--full">
              <textarea
                value={form.comment}
                onChange={(event) => updateFormField('comment', event.target.value)}
                placeholder="Что важно учесть для расчёта?"
                rows={5}
              />
            </Field>

            <div className="form-actions field--full">
              <button type="submit" disabled={submitting}>
                {submitting ? 'Отправляем...' : 'Отправить заявку'}
              </button>
              <button type="button" className="secondary" onClick={() => void loadRequests()} disabled={loading}>
                {loading ? 'Обновляем...' : 'Обновить список'}
              </button>
            </div>
          </form>
        </article>

        <article className="panel panel--list">
          <div className="panel__head">
            <div>
              <p className="panel__eyebrow">Админка</p>
              <h2>Список заявок</h2>
            </div>
            <span className="panel__hint">
              {loading ? 'Загружаем...' : `${requests.length} записей`}
            </span>
          </div>

          {error ? <div className="notice notice--error">{error}</div> : null}

          <div className="requests">
            {requests.length === 0 ? (
              <div className="empty-state">
                Пока заявок нет. Первая отправка появится здесь после формы слева.
              </div>
            ) : (
              requests.map((request) => {
                const draft = drafts[request.id] ?? {
                  status: request.status,
                  managerComment: request.manager_comment ?? '',
                }

                return (
                  <section key={request.id} className="request-card">
                    <div className="request-card__top">
                      <div>
                        <p className="request-card__title">{request.customer_name}</p>
                        <p className="request-card__meta">
                          {request.cargo_name} · {request.destination_city}
                        </p>
                      </div>
                      <span className={`status status--${request.status}`}>
                        {statusLabels[request.status]}
                      </span>
                    </div>

                    <dl className="request-card__facts">
                      <div>
                        <dt>Контакт</dt>
                        <dd>{request.contact}</dd>
                      </div>
                      <div>
                        <dt>Маршрут</dt>
                        <dd>{request.origin_city ?? '—'} → {request.destination_city}</dd>
                      </div>
                      <div>
                        <dt>Габариты</dt>
                        <dd>
                          {formatQuantity(request.weight_kg, 'кг')}
                          {formatQuantity(request.volume_m3, 'м3', true)}
                          {request.boxes_count ? ` · ${request.boxes_count} кор.` : ''}
                        </dd>
                      </div>
                      <div>
                        <dt>Комментарий</dt>
                        <dd>{request.comment}</dd>
                      </div>
                    </dl>

                    <div className="request-card__editor">
                      <label>
                        <span>Статус</span>
                        <select
                          value={draft.status}
                          onChange={(event) =>
                            updateDraftField(request.id, 'status', event.target.value)
                          }
                        >
                          {statusOptions.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                      </label>

                      <label>
                        <span>Комментарий менеджера</span>
                        <textarea
                          rows={3}
                          value={draft.managerComment}
                          onChange={(event) =>
                            updateDraftField(request.id, 'managerComment', event.target.value)
                          }
                          placeholder="Что нужно уточнить клиенту?"
                        />
                      </label>

                      <button
                        type="button"
                        className="secondary"
                        onClick={() => void handleSave(request)}
                        disabled={savingId === request.id}
                      >
                        {savingId === request.id ? 'Сохраняем...' : 'Сохранить'}
                      </button>
                    </div>
                  </section>
                )
              })
            )}
          </div>
        </article>
      </section>
    </main>
  )
}

function Field({
  label,
  required = false,
  className,
  children,
}: {
  label: string
  required?: boolean
  className?: string
  children: ReactNode
}) {
  return (
    <label className={`field ${className ?? ''}`.trim()}>
      <span>
        {label}
        {required ? ' *' : ''}
      </span>
      {children}
    </label>
  )
}

function toCreatePayload(form: ShipmentFormState) {
  const payload: Record<string, unknown> = {
    customer_name: form.customerName.trim(),
    contact: form.contact.trim(),
    cargo_name: form.cargoName.trim(),
    destination_city: form.destinationCity.trim(),
    comment: form.comment.trim(),
  }

  if (form.companyName.trim()) {
    payload.company_name = form.companyName.trim()
  }
  if (form.originCity.trim()) {
    payload.origin_city = form.originCity.trim()
  }
  if (form.weightKg.trim()) {
    payload.weight_kg = Number(form.weightKg)
  }
  if (form.volumeM3.trim()) {
    payload.volume_m3 = Number(form.volumeM3)
  }
  if (form.boxesCount.trim()) {
    payload.boxes_count = Number(form.boxesCount)
  }
  if (form.cargoValue.trim()) {
    payload.cargo_value = Number(form.cargoValue)
  }
  if (form.cargoCurrency.trim()) {
    payload.cargo_currency = form.cargoCurrency.trim()
  }

  return payload
}

function formatQuantity(value: number | null | undefined, unit: string, prefix = false) {
  if (value === null || value === undefined) {
    return prefix ? '' : '—'
  }

  return prefix ? ` · ${value} ${unit}` : `${value} ${unit}`
}

export default App
