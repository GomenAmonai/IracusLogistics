import { useCallback, useEffect, useState } from 'react'

import { ShipmentDetail } from './components/ShipmentDetail'
import { ShipmentList } from './components/ShipmentList'
import { CenteredState, PrimaryButton, Screen, Spinner } from './components/ui'
import { ApiError, authTelegram, listShipments, setAuthToken } from './lib/api'
import { bindBackButton, getInitData, initWebApp } from './lib/telegram'
import type { Client, Shipment } from './lib/types'

type AuthState = 'loading' | 'ready' | 'needTelegram' | 'error'
type View = { name: 'list' } | { name: 'detail'; id: string }

export function App() {
  const [authState, setAuthState] = useState<AuthState>('loading')
  const [authError, setAuthError] = useState<string | null>(null)
  const [client, setClient] = useState<Client | null>(null)
  const [view, setView] = useState<View>({ name: 'list' })
  const [shipments, setShipments] = useState<Shipment[] | null>(null)
  const [listError, setListError] = useState<string | null>(null)
  const [reloadCount, setReloadCount] = useState(0)

  const bootstrap = useCallback(async () => {
    setAuthState('loading')
    setAuthError(null)

    // Dev-обход: вне Telegram (обычный браузер) initData пустой и подписи нет. Чтобы можно
    // было гонять UI локально, в DEV принимаем заранее выпущенный client-токен из ?token=.
    if (import.meta.env.DEV) {
      const devToken = new URLSearchParams(window.location.search).get('token')
      if (devToken) {
        setAuthToken(devToken)
        setAuthState('ready')
        return
      }
    }

    const initData = getInitData()
    if (!initData) {
      setAuthState('needTelegram')
      return
    }

    try {
      const result = await authTelegram(initData)
      setAuthToken(result.token)
      setClient(result.client)
      setAuthState('ready')
    } catch (err) {
      setAuthError(err instanceof ApiError ? err.message : 'Не удалось авторизоваться.')
      setAuthState('error')
    }
  }, [])

  useEffect(() => {
    initWebApp()
    void bootstrap()
  }, [bootstrap])

  // Список грузов грузим при готовности и каждый раз при возврате на экран списка —
  // так статусы освежаются после просмотра деталей.
  useEffect(() => {
    if (authState !== 'ready' || view.name !== 'list') {
      return
    }
    let active = true
    setListError(null)
    listShipments()
      .then((loaded) => {
        if (active) {
          setShipments(loaded)
        }
      })
      .catch((err: unknown) => {
        if (active) {
          setListError(err instanceof ApiError ? err.message : 'Не удалось загрузить грузы.')
        }
      })
    return () => {
      active = false
    }
  }, [authState, view.name, reloadCount])

  // Нативная кнопка «Назад» Telegram на экране деталей.
  useEffect(() => {
    if (view.name !== 'detail') {
      return
    }
    return bindBackButton(() => setView({ name: 'list' }))
  }, [view.name])

  if (authState === 'loading') {
    return (
      <Screen>
        <div className="flex flex-1 items-center justify-center">
          <Spinner />
        </div>
      </Screen>
    )
  }

  if (authState === 'needTelegram') {
    return (
      <Screen>
        <CenteredState
          title="Откройте в Telegram"
          description="Это приложение работает внутри Telegram. Откройте его через бота Icaris."
        />
      </Screen>
    )
  }

  if (authState === 'error') {
    return (
      <Screen>
        <CenteredState
          title="Не удалось войти"
          description={authError ?? undefined}
          action={<PrimaryButton onClick={() => void bootstrap()}>Повторить</PrimaryButton>}
        />
      </Screen>
    )
  }

  return (
    <Screen>
      {view.name === 'detail' ? (
        <>
          {/* Внутри-приложенческая «Назад» всегда видна: нативная BackButton Telegram есть
              только с Bot API 6.1, и на старых клиентах пользователь иначе застрянет. */}
          <button
            type="button"
            onClick={() => setView({ name: 'list' })}
            className="mb-4 self-start text-sm font-medium text-accent hover:text-accent-deep"
          >
            ← К списку
          </button>
          <ShipmentDetail shipmentId={view.id} />
        </>
      ) : listError ? (
        <CenteredState
          title="Ошибка загрузки"
          description={listError}
          action={<PrimaryButton onClick={() => setReloadCount((count) => count + 1)}>Повторить</PrimaryButton>}
        />
      ) : shipments === null ? (
        <div className="flex flex-1 items-center justify-center">
          <Spinner />
        </div>
      ) : (
        <ShipmentList
          client={client}
          shipments={shipments}
          onOpen={(id) => setView({ name: 'detail', id })}
        />
      )}
    </Screen>
  )
}
