import { type FormEvent, useEffect, useState } from 'react'

import { ApiError, createLead, type CreateLeadInput } from '../lib/api'
import { Section } from './Section'

export type LeadPrefill = {
  toCity: string
  cargoType: string
  weight: string
  volume: string
}

type FormState = {
  name: string
  phone: string
  fromCity: string
  toCity: string
  cargoType: string
  weight: string
  volume: string
  comment: string
  consent: boolean
}

const EMPTY_FORM: FormState = {
  name: '',
  phone: '',
  fromCity: '',
  toCity: '',
  cargoType: '',
  weight: '',
  volume: '',
  comment: '',
  consent: false,
}

type FieldErrors = Partial<Record<'name' | 'phone' | 'fromCity' | 'toCity' | 'consent', string>>

type Status =
  | { kind: 'idle' }
  | { kind: 'submitting' }
  | { kind: 'success'; leadId: string }
  | { kind: 'error'; message: string }

const RESPONSE_HOURS = 2
const PRIVACY_POLICY_URL = (import.meta.env.VITE_PRIVACY_POLICY_URL ?? '').trim()
const PRIVACY_NOTICE_VERSION = (import.meta.env.VITE_PRIVACY_NOTICE_VERSION ?? '').trim()
const IS_LEAD_FORM_ENABLED = Boolean(PRIVACY_POLICY_URL && PRIVACY_NOTICE_VERSION)

function toNumber(raw: string): number | undefined {
  const trimmed = raw.trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed.replace(',', '.'))
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined
}

function validate(form: FormState): FieldErrors {
  const errors: FieldErrors = {}
  if (!form.name.trim()) errors.name = 'Укажите имя'
  if (!form.phone.trim()) errors.phone = 'Укажите телефон или Telegram'
  if (!form.fromCity.trim()) errors.fromCity = 'Укажите город отправки'
  if (!form.toCity.trim()) errors.toCity = 'Укажите город назначения'
  if (!form.consent) errors.consent = 'Нужно согласие на обработку данных'
  return errors
}

const FIELD_IDS: Record<keyof FieldErrors, string> = {
  name: 'lead-name',
  phone: 'lead-phone',
  fromCity: 'lead-from',
  toCity: 'lead-to',
  consent: 'lead-consent',
}

type LeadFormProps = {
  prefill: LeadPrefill | null
}

export function LeadForm({ prefill }: LeadFormProps) {
  const [form, setForm] = useState<FormState>(EMPTY_FORM)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [status, setStatus] = useState<Status>({ kind: 'idle' })

  // Калькулятор передаёт введённые данные — предзаполняем форму при их изменении.
  useEffect(() => {
    if (!prefill) return
    setForm((current) => ({
      ...current,
      toCity: prefill.toCity || current.toCity,
      cargoType: prefill.cargoType || current.cargoType,
      weight: prefill.weight || current.weight,
      volume: prefill.volume || current.volume,
    }))
  }, [prefill])

  function update<K extends keyof FormState>(field: K, value: FormState[K]) {
    setForm((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!IS_LEAD_FORM_ENABLED) {
      setStatus({
        kind: 'error',
        message: 'Онлайн-заявка временно недоступна. Напишите менеджеру в Telegram.',
      })
      return
    }

    const errors = validate(form)
    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors)
      // Перенос фокуса на первое невалидное поле: иначе SR-пользователь не узнаёт, что сабмит не прошёл (WCAG 3.3.1).
      const firstInvalid = Object.keys(errors)[0] as keyof FieldErrors
      document.getElementById(FIELD_IDS[firstInvalid])?.focus()
      return
    }

    // Тело строго по контракту: обязательные всегда, пустые опциональные не включаем.
    const payload: CreateLeadInput = {
      name: form.name.trim(),
      phone: form.phone.trim(),
      from_city: form.fromCity.trim(),
      to_city: form.toCity.trim(),
      privacy_notice_version: PRIVACY_NOTICE_VERSION,
    }
    const weight = toNumber(form.weight)
    const volume = toNumber(form.volume)
    if (weight !== undefined) payload.weight = weight
    if (volume !== undefined) payload.volume = volume
    if (form.cargoType.trim()) payload.cargo_type = form.cargoType.trim()
    if (form.comment.trim()) payload.comment = form.comment.trim()

    setStatus({ kind: 'submitting' })
    try {
      const lead = await createLead(payload)
      setStatus({ kind: 'success', leadId: lead.id })
      setForm(EMPTY_FORM)
    } catch (error) {
      const message =
        error instanceof ApiError ? error.message : 'Не удалось отправить заявку. Попробуйте позже.'
      setStatus({ kind: 'error', message })
    }
  }

  if (status.kind === 'success') {
    return (
      <Section id="lead" surface eyebrow="Заявка">
        <div
          role="status"
          className="mx-auto max-w-xl rounded-2xl border border-line bg-surface p-8 text-center shadow-soft"
        >
          {/* Мягкий cargo-бейдж «принято» — спокойный статус успеха */}
          <span
            aria-hidden="true"
            className="mx-auto mb-6 inline-flex items-center gap-2 rounded-full bg-cargo/10 px-4 py-1.5 text-sm font-semibold text-cargo"
          >
            ✓ Принято
          </span>
          <h2 className="font-display text-2xl font-extrabold tracking-[-0.02em] text-ink">
            Заявка принята
          </h2>
          <p className="mt-3 text-base leading-relaxed text-ink-soft">
            Менеджер свяжется с вами в течение {RESPONSE_HOURS} часов и зафиксирует точную
            стоимость после проверки документов.
          </p>
          <p className="mt-6 flex flex-wrap items-center justify-center gap-x-2 gap-y-2 border-t border-line-soft pt-5 text-sm text-ink-soft">
            Номер заявки:{' '}
            <span className="terminal break-all text-sm font-semibold">{status.leadId}</span>
          </p>
          <button
            type="button"
            onClick={() => setStatus({ kind: 'idle' })}
            className="mt-6 inline-flex items-center justify-center rounded-full border border-line bg-surface px-6 py-2.5 text-sm font-medium text-ink transition-colors duration-200 hover:border-accent hover:text-accent"
          >
            Отправить ещё одну
          </button>
        </div>
      </Section>
    )
  }

  const isSubmitting = status.kind === 'submitting'

  return (
    <Section
      id="lead"
      surface
      eyebrow="Заявка"
      title="Оставьте заявку — посчитаем точную стоимость"
      intro={`Ответим в течение ${RESPONSE_HOURS} часов в рабочее время. Регистрация не нужна.`}
    >
      <form className="grid gap-6 lg:grid-cols-[1.4fr_0.6fr]" onSubmit={handleSubmit} noValidate>
        <div className="grid gap-5 rounded-2xl border border-line bg-surface p-6 shadow-soft sm:grid-cols-2 sm:p-8">
          <div>
            <label htmlFor="lead-name" className="field-label mb-2 block">
              Имя <span className="text-accent">*</span>
            </label>
            <input
              id="lead-name"
              className="input-field"
              value={form.name}
              onChange={(event) => update('name', event.target.value)}
              aria-invalid={Boolean(fieldErrors.name)}
              aria-describedby={fieldErrors.name ? 'err-name' : undefined}
              placeholder="Иван Петров"
              autoComplete="name"
            />
            {fieldErrors.name && (
              <p id="err-name" className="mt-1.5 text-sm text-alert">
                {fieldErrors.name}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="lead-phone" className="field-label mb-2 block">
              Телефон или Telegram <span className="text-accent">*</span>
            </label>
            <input
              id="lead-phone"
              className="input-field"
              value={form.phone}
              onChange={(event) => update('phone', event.target.value)}
              aria-invalid={Boolean(fieldErrors.phone)}
              aria-describedby={fieldErrors.phone ? 'err-phone' : undefined}
              placeholder="+7 999 111-22-33 / @username"
              autoComplete="tel"
            />
            {fieldErrors.phone && (
              <p id="err-phone" className="mt-1.5 text-sm text-alert">
                {fieldErrors.phone}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="lead-from" className="field-label mb-2 block">
              Город отправки <span className="text-accent">*</span>
            </label>
            <input
              id="lead-from"
              className="input-field"
              value={form.fromCity}
              onChange={(event) => update('fromCity', event.target.value)}
              aria-invalid={Boolean(fieldErrors.fromCity)}
              aria-describedby={fieldErrors.fromCity ? 'err-from' : undefined}
              placeholder="Гуанчжоу"
            />
            {fieldErrors.fromCity && (
              <p id="err-from" className="mt-1.5 text-sm text-alert">
                {fieldErrors.fromCity}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="lead-to" className="field-label mb-2 block">
              Город назначения <span className="text-accent">*</span>
            </label>
            <input
              id="lead-to"
              className="input-field"
              value={form.toCity}
              onChange={(event) => update('toCity', event.target.value)}
              aria-invalid={Boolean(fieldErrors.toCity)}
              aria-describedby={fieldErrors.toCity ? 'err-to' : undefined}
              placeholder="Москва"
            />
            {fieldErrors.toCity && (
              <p id="err-to" className="mt-1.5 text-sm text-alert">
                {fieldErrors.toCity}
              </p>
            )}
          </div>

          <div>
            <label htmlFor="lead-cargo" className="field-label mb-2 block">
              Тип груза
            </label>
            <input
              id="lead-cargo"
              className="input-field"
              value={form.cargoType}
              onChange={(event) => update('cargoType', event.target.value)}
              placeholder="Электроника"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="lead-weight" className="field-label mb-2 block">
                Вес, кг
              </label>
              <input
                id="lead-weight"
                className="input-field tabular font-mono"
                inputMode="decimal"
                value={form.weight}
                onChange={(event) => update('weight', event.target.value)}
                placeholder="800"
              />
            </div>
            <div>
              <label htmlFor="lead-volume" className="field-label mb-2 block">
                Объём, м³
              </label>
              <input
                id="lead-volume"
                className="input-field tabular font-mono"
                inputMode="decimal"
                value={form.volume}
                onChange={(event) => update('volume', event.target.value)}
                placeholder="4"
              />
            </div>
          </div>

          <div className="sm:col-span-2">
            <label htmlFor="lead-comment" className="field-label mb-2 block">
              Комментарий
            </label>
            <textarea
              id="lead-comment"
              className="input-field resize-y"
              rows={3}
              value={form.comment}
              onChange={(event) => update('comment', event.target.value)}
              placeholder="Что важно учесть: сроки, поставщик, особые требования"
            />
          </div>

          <div className="sm:col-span-2">
            <label className="flex cursor-pointer items-start gap-3 text-sm leading-relaxed text-ink-soft">
              <input
                id="lead-consent"
                type="checkbox"
                className="mt-1 h-4 w-4 shrink-0 accent-accent"
                checked={form.consent}
                disabled={!IS_LEAD_FORM_ENABLED}
                onChange={(event) => update('consent', event.target.checked)}
                aria-invalid={Boolean(fieldErrors.consent)}
                aria-describedby={fieldErrors.consent ? 'err-consent' : undefined}
              />
              <span>
                Согласен на обработку указанных данных для ответа на заявку и связь со мной
                по телефону или в Telegram в соответствии с{' '}
                {PRIVACY_POLICY_URL ? (
                  <a
                    href={PRIVACY_POLICY_URL}
                    target="_blank"
                    rel="noreferrer"
                    className="text-ink underline underline-offset-2 hover:text-accent"
                  >
                    политикой обработки персональных данных
                  </a>
                ) : (
                  <span className="text-alert">политикой, которая пока не опубликована</span>
                )}
                .
                {fieldErrors.consent && (
                  <span id="err-consent" className="mt-1 block text-alert">
                    {fieldErrors.consent}
                  </span>
                )}
              </span>
            </label>
          </div>
        </div>

        {/* Сайдбар действия */}
        <aside className="flex flex-col gap-5 rounded-2xl border border-line bg-surface p-6 shadow-soft sm:p-8">
          <div>
            <p className="eyebrow mb-4">Что дальше</p>
            {/* Мягкие шаги: кобальтовый порядковый чип ведёт взгляд по процессу */}
            <ul className="flex flex-col gap-4">
              <li className="flex gap-3 text-sm leading-relaxed text-ink-soft">
                <span
                  aria-hidden="true"
                  className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-accent-tint font-mono text-xs font-semibold text-accent"
                >
                  01
                </span>
                Менеджер проверит вводные и документы по грузу.
              </li>
              <li className="flex gap-3 text-sm leading-relaxed text-ink-soft">
                <span
                  aria-hidden="true"
                  className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-accent-tint font-mono text-xs font-semibold text-accent"
                >
                  02
                </span>
                Зафиксирует точную стоимость и срок.
              </li>
              <li className="flex gap-3 text-sm leading-relaxed text-ink-soft">
                <span
                  aria-hidden="true"
                  className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-accent-tint font-mono text-xs font-semibold text-accent"
                >
                  03
                </span>
                Свяжется с вами выбранным способом.
              </li>
            </ul>
          </div>

          {status.kind === 'error' && (
            <p role="alert" className="rounded-xl border border-alert/40 bg-alert/5 px-4 py-3 text-sm text-alert">
              {status.message}
            </p>
          )}

          {!IS_LEAD_FORM_ENABLED && status.kind !== 'error' && (
            <p role="status" className="rounded-xl border border-alert/40 bg-alert/5 px-4 py-3 text-sm text-alert">
              Онлайн-заявка временно закрыта до публикации политики обработки данных. Связаться
              можно с менеджером{' '}
              <a className="underline underline-offset-2" href="https://t.me/hikill8">
                @hikill8
              </a>
              .
            </p>
          )}

          <button
            type="submit"
            disabled={isSubmitting || !IS_LEAD_FORM_ENABLED}
            className="inline-flex items-center justify-center rounded-full bg-accent px-6 py-3.5 text-base font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep disabled:cursor-not-allowed disabled:opacity-60"
            aria-busy={isSubmitting}
          >
            {isSubmitting
              ? 'Отправляем…'
              : IS_LEAD_FORM_ENABLED
                ? 'Отправить заявку'
                : 'Приём заявок временно закрыт'}
          </button>

          <p className="text-xs leading-relaxed text-ink-soft">
            После включения формы условия обработки данных доступны по ссылке рядом с
            согласием. Не отправляйте сведения о грузе через форму, пока политика недоступна.
          </p>
        </aside>
      </form>
    </Section>
  )
}
