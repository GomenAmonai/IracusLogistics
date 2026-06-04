import type { Client, Message, Shipment, ShipmentDetail } from './types'

// Клиентский API WebApp. Bearer-токен (client-JWT) держим в памяти модуля: его выдаёт
// /app/auth/telegram, дальше им подписываем каждый запрос.

// Базовый URL API. Пусто (локально) → относительный /api через прокси Vite. На задеплоенном
// фронте (Vercel) задаём VITE_API_BASE на публичный backend; запросы идут кросс-доменно
// (бэкенд отдаёт CORS *), авторизация по Bearer — куки не нужны, CSRF не возникает.
const API_BASE = (import.meta.env.VITE_API_BASE ?? '').replace(/\/$/, '')

let authToken = ''

export function setAuthToken(token: string): void {
  authToken = token
}

type ApiErrorBody = {
  error?: { code?: string; message?: string }
}

export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${API_BASE}/api${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
        ...init?.headers,
      },
    })
  } catch {
    throw new ApiError('network', 'Нет связи с сервером. Проверьте соединение.', 0)
  }

  if (!response.ok) {
    let body: ApiErrorBody = {}
    try {
      body = (await response.json()) as ApiErrorBody
    } catch {
      // Тело без JSON — отдадим обобщённую ошибку по статусу.
    }
    throw new ApiError(
      body.error?.code ?? 'error',
      body.error?.message ?? `Сервер вернул ${response.status}.`,
      response.status,
    )
  }

  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}

export async function authTelegram(initData: string): Promise<{ token: string; client: Client }> {
  return request('/app/auth/telegram', {
    method: 'POST',
    body: JSON.stringify({ init_data: initData }),
  })
}

export async function listShipments(): Promise<Shipment[]> {
  return request('/app/shipments')
}

export async function getShipment(id: string): Promise<ShipmentDetail> {
  return request(`/app/shipments/${id}`)
}

export async function listMessages(shipmentId: string): Promise<Message[]> {
  return request(`/app/shipments/${shipmentId}/messages`)
}

export async function sendMessage(shipmentId: string, text: string): Promise<Message> {
  return request(`/app/shipments/${shipmentId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  })
}
