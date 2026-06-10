import type {
  Client,
  Lead,
  LeadStatus,
  Message,
  Payment,
  PaymentChannel,
  PaymentStatus,
  Shipment,
  ShipmentDetail,
  ShipmentStatus,
} from './types'

// API панели менеджера. Manager-JWT живёт в sessionStorage: переживает перезагрузку
// вкладки, но не закрытие браузера. XSS-риск осознан и зафиксирован в tech-debt (#23) —
// httpOnly-cookie потребовал бы перестройки auth на бэкенде.

const API_BASE = (import.meta.env.VITE_API_BASE ?? '').replace(/\/$/, '')

const TOKEN_KEY = 'icaris_manager_token'

export function getStoredToken(): string {
  return sessionStorage.getItem(TOKEN_KEY) ?? ''
}

export function setStoredToken(token: string): void {
  sessionStorage.setItem(TOKEN_KEY, token)
}

export function clearStoredToken(): void {
  sessionStorage.removeItem(TOKEN_KEY)
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

export function isAuthError(err: unknown): boolean {
  return err instanceof ApiError && err.status === 401
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${API_BASE}/api${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(getStoredToken() ? { Authorization: `Bearer ${getStoredToken()}` } : {}),
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

export async function login(email: string, password: string): Promise<string> {
  const result = await request<{ token: string }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  setStoredToken(result.token)

  return result.token
}

export async function listLeads(): Promise<Lead[]> {
  return request('/leads')
}

export async function updateLeadStatus(id: string, status: LeadStatus): Promise<Lead> {
  return request(`/leads/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}

export async function listClients(): Promise<Client[]> {
  return request('/clients')
}

export async function listShipments(): Promise<Shipment[]> {
  return request('/shipments')
}

// CreateShipmentInput повторяет service.CreateShipmentInput: decimal-поля шлём строками
// (бэкенд парсит их без потери точности), пустые строки не отправляем вовсе.
export type CreateShipmentInput = {
  client_id: string
  lane: string
  from_city?: string
  to_city?: string
  weight?: string
  volume?: string
  price?: string
  currency?: string
  status_note?: string
}

export async function createShipment(input: CreateShipmentInput): Promise<Shipment> {
  return request('/shipments', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function getShipment(id: string): Promise<ShipmentDetail> {
  return request(`/shipments/${id}`)
}

export async function updateShipmentStatus(
  id: string,
  status: ShipmentStatus,
  comment: string,
): Promise<Shipment> {
  return request(`/shipments/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status, comment }),
  })
}

export async function listMessages(shipmentId: string): Promise<Message[]> {
  return request(`/shipments/${shipmentId}/messages`)
}

export async function sendMessage(shipmentId: string, text: string): Promise<Message> {
  return request(`/shipments/${shipmentId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  })
}

export async function listPayments(shipmentId: string): Promise<Payment[]> {
  return request(`/shipments/${shipmentId}/payments`)
}

export type CreatePaymentInput = {
  amount: string
  currency?: string
  channel: PaymentChannel
  status?: PaymentStatus
  comment?: string
}

export async function createPayment(shipmentId: string, input: CreatePaymentInput): Promise<Payment> {
  return request(`/shipments/${shipmentId}/payments`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function updatePaymentStatus(
  shipmentId: string,
  paymentId: string,
  status: PaymentStatus,
): Promise<Payment> {
  return request(`/shipments/${shipmentId}/payments/${paymentId}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}
