import { useEffect, useState } from 'react'

import { ApiError, listLeads, updateLeadStatus } from '../lib/api'
import { formatDateTime, formatVolume, formatWeight } from '../lib/format'
import { LEAD_STATUS_LABELS, type Lead, type LeadStatus } from '../lib/types'
import { LeadStatusBadge } from './badges'
import { CenteredState, ErrorNote, Spinner } from './ui'

const FILTERS: (LeadStatus | 'all')[] = ['all', 'new', 'contacted', 'converted', 'rejected']

export function LeadList({ onAuthError }: { onAuthError: (err: unknown) => boolean }) {
  const [leads, setLeads] = useState<Lead[] | null>(null)
  const [filter, setFilter] = useState<LeadStatus | 'all'>('all')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    listLeads()
      .then((loaded) => {
        if (active) setLeads(loaded)
      })
      .catch((err: unknown) => {
        if (active && !onAuthError(err)) {
          setError(err instanceof ApiError ? err.message : 'Не удалось загрузить лиды.')
        }
      })
    return () => {
      active = false
    }
  }, [onAuthError])

  async function handleStatusChange(lead: Lead, status: LeadStatus) {
    try {
      const updated = await updateLeadStatus(lead.id, status)
      setLeads((current) => current?.map((l) => (l.id === updated.id ? updated : l)) ?? null)
    } catch (err) {
      if (!onAuthError(err)) {
        setError(err instanceof ApiError ? err.message : 'Не удалось сменить статус.')
      }
    }
  }

  if (error) {
    return <ErrorNote message={error} />
  }
  if (leads === null) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    )
  }

  const visible = filter === 'all' ? leads : leads.filter((l) => l.status === filter)

  return (
    <div className="flex flex-col gap-5">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="font-display text-2xl font-bold text-ink">Лиды</h1>
        <div className="flex flex-wrap gap-1.5" role="group" aria-label="Фильтр по статусу">
          {FILTERS.map((value) => (
            <button
              key={value}
              type="button"
              onClick={() => setFilter(value)}
              aria-pressed={filter === value}
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors duration-200 ${
                filter === value
                  ? 'border-accent bg-accent-tint text-accent'
                  : 'border-line bg-surface text-ink-soft hover:border-accent'
              }`}
            >
              {value === 'all' ? 'Все' : LEAD_STATUS_LABELS[value]}
            </button>
          ))}
        </div>
      </header>

      {visible.length === 0 ? (
        <CenteredState
          title="Лидов нет"
          description={filter === 'all' ? 'Новые заявки с сайта появятся здесь.' : 'В этом статусе пусто.'}
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {visible.map((lead) => (
            <li key={lead.id}>
              <LeadCard lead={lead} onStatusChange={(status) => void handleStatusChange(lead, status)} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function LeadCard({
  lead,
  onStatusChange,
}: {
  lead: Lead
  onStatusChange: (status: LeadStatus) => void
}) {
  const route = [lead.from_city, lead.to_city].filter(Boolean).join(' → ')
  const cargo = [lead.cargo_type, formatWeight(lead.weight), formatVolume(lead.volume)]
    .filter(Boolean)
    .join(' · ')

  return (
    <article className="flex flex-col gap-3 rounded-2xl border border-line bg-surface p-5 shadow-card">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <h2 className="font-display text-base font-bold text-ink">{lead.name}</h2>
          <LeadStatusBadge status={lead.status} />
        </div>
        <span className="terminal text-xs text-ink-soft">{formatDateTime(lead.created_at)}</span>
      </div>

      <dl className="grid gap-x-8 gap-y-1 text-sm sm:grid-cols-2">
        <div className="flex gap-2">
          <dt className="field-label shrink-0">Контакт</dt>
          <dd className="terminal text-ink">{lead.phone}</dd>
        </div>
        {route && (
          <div className="flex gap-2">
            <dt className="field-label shrink-0">Маршрут</dt>
            <dd className="text-ink">{route}</dd>
          </div>
        )}
        {cargo && (
          <div className="flex gap-2">
            <dt className="field-label shrink-0">Груз</dt>
            <dd className="text-ink-soft">{cargo}</dd>
          </div>
        )}
      </dl>
      {lead.comment && <p className="text-sm leading-relaxed text-ink-soft">{lead.comment}</p>}

      <div className="flex items-center gap-2 border-t border-line-soft pt-3">
        <label htmlFor={`lead-status-${lead.id}`} className="field-label">
          Статус
        </label>
        <select
          id={`lead-status-${lead.id}`}
          value={lead.status}
          onChange={(e) => onStatusChange(e.target.value as LeadStatus)}
          className="input-field max-w-48 py-1.5 text-sm"
        >
          {(Object.keys(LEAD_STATUS_LABELS) as LeadStatus[]).map((status) => (
            <option key={status} value={status}>
              {LEAD_STATUS_LABELS[status]}
            </option>
          ))}
        </select>
      </div>
    </article>
  )
}
