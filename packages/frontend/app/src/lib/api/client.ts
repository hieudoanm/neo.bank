export type ApiError = {
  code: string
  message: string
  details?: Record<string, string>
  requestId: string
}

class HttpError extends Error {
  constructor(
    public status: number,
    public error: ApiError,
    message: string,
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080'

async function request<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<T> {
  const res = await fetch(`${API_BASE}${typeof input === 'string' ? input : input.toString()}`, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    ...init,
  })

  const body = (await res.text()) ?? ''
  if (!res.ok) {
    let parsed: ApiError
    try {
      parsed = JSON.parse(body).error as ApiError
    } catch {
      parsed = { code: 'INTERNAL_ERROR', message: 'Unexpected error', requestId: '' }
    }
    throw new HttpError(res.status, parsed, parsed.message)
  }

  if (body.length === 0) return undefined as T
  return JSON.parse(body) as T
}

type RegisterRequest = { email: string; password: string }
type LoginRequest = { email: string; password: string }
type RefreshRequest = { refreshToken: string }
type LogoutRequest = { refreshToken: string }

type TokenResponse = {
  userId: string
  accountId: string
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
}

export const authApi = {
  register: (body: RegisterRequest) =>
    request<TokenResponse>('/api/v1/register', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  login: (body: LoginRequest) =>
    request<TokenResponse>('/api/v1/login', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  refresh: (body: RefreshRequest) =>
    request<{ accessToken: string; refreshToken: string }>('/api/v1/refresh', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  logout: (body: LogoutRequest) =>
    request<void>('/api/v1/logout', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
}

type BalanceResponse = {
  accountId: string
  balanceCents: number
  currency: string
}

export const accountsApi = {
  getBalance: (accountId: string) =>
    request<BalanceResponse>(`/api/v1/accounts/${accountId}/balance`),
}
