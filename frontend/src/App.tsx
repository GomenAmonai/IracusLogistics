import { useState } from 'react'
import './App.css'

type HealthState = 'idle' | 'loading' | 'ok' | 'error'

const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

function App() {
  const [healthState, setHealthState] = useState<HealthState>('idle')
  const [healthMessage, setHealthMessage] = useState('Нажмите кнопку проверки')

  async function checkHealth() {
    setHealthState('loading')
    setHealthMessage('Проверяем Go API...')

    try {
      const response = await fetch(`${apiUrl}/api/health`)
      const payload = (await response.json()) as {
        status?: string
        database?: string
      }

      if (!response.ok) {
        throw new Error(`API returned ${response.status}`)
      }

      setHealthState('ok')
      setHealthMessage(`API: ${payload.status}, DB: ${payload.database}`)
    } catch (error) {
      setHealthState('error')
      setHealthMessage(error instanceof Error ? error.message : 'API недоступен')
    }
  }

  return (
    <main className="app-shell">
      <section className="hero">
        <div className="hero__content">
          <p className="eyebrow">Iracus Logistic</p>
          <h1>Сервис заявок на доставку из Китая</h1>
          <p className="lead">
            Первый MVP: клиент оставляет заявку, Go API сохраняет данные,
            менеджер видит запрос в админке.
          </p>
        </div>
      </section>

      <section className="workspace">
        <article className="panel">
          <h2>Текущий фокус</h2>
          <ul className="checklist">
            <li>Описать карточку заявки</li>
            <li>Сделать таблицы PostgreSQL</li>
            <li>Добавить endpoint создания заявки</li>
            <li>Собрать простую форму и админку</li>
          </ul>
        </article>

        <article className="panel">
          <h2>Backend status</h2>
          <dl className="status-list">
            <div>
              <dt>API URL</dt>
              <dd>{apiUrl}</dd>
            </div>
            <div>
              <dt>Health</dt>
              <dd className={`health health--${healthState}`}>
                {healthMessage}
              </dd>
            </div>
          </dl>
          <button type="button" onClick={checkHealth}>
            Проверить API
          </button>
        </article>
      </section>
    </main>
  )
}

export default App
