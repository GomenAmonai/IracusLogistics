import { useCallback, useState } from 'react'

import { clearStoredToken, getStoredToken, isAuthError } from './lib/api'
import { ClientList } from './components/ClientList'
import { LeadList } from './components/LeadList'
import { Login } from './components/Login'
import { ShipmentDetailView } from './components/ShipmentDetailView'
import { ShipmentForm } from './components/ShipmentForm'
import { ShipmentList } from './components/ShipmentList'

type View =
  | { name: 'leads' }
  | { name: 'clients' }
  | { name: 'shipments' }
  | { name: 'newShipment'; clientId?: string }
  | { name: 'shipment'; id: string }

const NAV: { view: View['name']; label: string }[] = [
  { view: 'leads', label: 'Лиды' },
  { view: 'clients', label: 'Клиенты' },
  { view: 'shipments', label: 'Грузы' },
]

export function App() {
  const [isAuthed, setAuthed] = useState(() => getStoredToken() !== '')
  const [view, setView] = useState<View>({ name: 'leads' })

  // Просроченный/отозванный JWT: любой экран при 401 разлогинивает обратно на вход.
  const handleAuthError = useCallback((err: unknown): boolean => {
    if (isAuthError(err)) {
      clearStoredToken()
      setAuthed(false)
      return true
    }
    return false
  }, [])

  function logout() {
    clearStoredToken()
    setAuthed(false)
  }

  if (!isAuthed) {
    return <Login onSuccess={() => setAuthed(true)} />
  }

  // Подсветка активной вкладки: детали/форма груза остаются «внутри» вкладки Грузы.
  const activeTab =
    view.name === 'shipment' || view.name === 'newShipment' ? 'shipments' : view.name

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-10 border-b border-line bg-base/85 backdrop-blur-md">
        <div className="mx-auto flex h-14 w-full max-w-5xl items-center justify-between gap-4 px-5">
          <div className="flex items-center gap-2.5">
            <span
              aria-hidden="true"
              className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent font-display text-sm font-bold text-surface"
            >
              I
            </span>
            <span className="font-display text-base font-extrabold tracking-[-0.01em] text-ink">
              Icaris
            </span>
            <span className="hidden font-mono text-xs tracking-wide text-ink-soft sm:inline">
              панель
            </span>
          </div>

          <nav aria-label="Разделы" className="flex items-center gap-1">
            {NAV.map((item) => (
              <button
                key={item.view}
                type="button"
                onClick={() => setView({ name: item.view } as View)}
                aria-current={activeTab === item.view ? 'page' : undefined}
                className={`rounded-full px-3.5 py-1.5 text-sm transition-colors duration-200 ${
                  activeTab === item.view
                    ? 'bg-accent-tint font-semibold text-accent'
                    : 'text-ink-soft hover:text-ink'
                }`}
              >
                {item.label}
              </button>
            ))}
          </nav>

          <button
            type="button"
            onClick={logout}
            className="text-sm text-ink-soft transition-colors hover:text-ink"
          >
            Выйти
          </button>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col px-5 py-8">
        {view.name === 'leads' && <LeadList onAuthError={handleAuthError} />}
        {view.name === 'clients' && (
          <ClientList
            onAuthError={handleAuthError}
            onCreateShipment={(clientId) => setView({ name: 'newShipment', clientId })}
          />
        )}
        {view.name === 'shipments' && (
          <ShipmentList
            onAuthError={handleAuthError}
            onOpen={(id) => setView({ name: 'shipment', id })}
            onCreate={() => setView({ name: 'newShipment' })}
          />
        )}
        {view.name === 'newShipment' && (
          <ShipmentForm
            onAuthError={handleAuthError}
            presetClientId={view.clientId}
            onCancel={() => setView({ name: 'shipments' })}
            onCreated={(id) => setView({ name: 'shipment', id })}
          />
        )}
        {view.name === 'shipment' && (
          <ShipmentDetailView
            shipmentId={view.id}
            onAuthError={handleAuthError}
            onBack={() => setView({ name: 'shipments' })}
          />
        )}
      </main>
    </div>
  )
}
