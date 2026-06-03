// Контракт публичной формы. Поля 1:1 с доменной моделью Lead на бэкенде.
// Пустые опциональные поля (weight/volume/cargo_type/comment) в тело НЕ включаем.

export type CreateLeadInput = {
  name: string
  phone: string
  from_city: string
  to_city: string
  weight?: number
  volume?: number
  cargo_type?: string
  comment?: string
}

export type Lead = {
  id: string
  name: string
  phone: string
  from_city: string
  to_city: string
  status: string
  created_at: string
}

// Единый формат ошибок API: {"error": {"code", "message"}}.
type ApiErrorBody = {
  error?: {
    code?: string
    message?: string
  }
}

export class ApiError extends Error {
  readonly code: string

  constructor(code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

export async function createLead(input: CreateLeadInput): Promise<Lead> {
  let response: Response
  try {
    response = await fetch('/api/leads', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
  } catch {
    throw new ApiError('network', 'Не удалось связаться с сервером. Проверьте соединение.')
  }

  if (!response.ok) {
    let body: ApiErrorBody = {}
    try {
      body = (await response.json()) as ApiErrorBody
    } catch {
      // Тело без JSON — отдадим обобщённую ошибку по статусу ниже.
    }
    const code = body.error?.code ?? 'error'
    const message =
      body.error?.message ?? `Сервер вернул ${response.status}. Попробуйте позже.`
    throw new ApiError(code, message)
  }

  return (await response.json()) as Lead
}
