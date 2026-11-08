import { logger } from '@/lib/log'
import { refreshSession } from '@/lib/authSession'

export type ApiError = {
  code: string
  message: string
  details?: Record<string, string>
  requestId: string
}

export class HttpError extends Error {
  constructor(
    public status: number,
    public error: ApiError,
    message: string,
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

// Same-origin by default: the vite dev server and the production nginx image
// both proxy /api to the gateway. Set VITE_API_BASE only to point at a
// different absolute origin.
const API_BASE = import.meta.env.VITE_API_BASE ?? ''

function authHeader(accessToken?: string): Record<string, string> {
  return accessToken ? { Authorization: `Bearer ${accessToken}` } : {}
}

const AUTH_PATHS = ['/api/v1/login', '/api/v1/register', '/api/v1/refresh', '/api/v1/logout']

async function request<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const url = `${API_BASE}${typeof input === 'string' ? input : input.toString()}`
  const send = (headers?: Record<string, string>) =>
    fetch(url, {
      ...init,
      headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}), ...headers },
    })

  let res: Response
  try {
    res = await send()
  } catch (error) {
    logger.error('api request failed', { url, error })
    throw error
  }

  if (res.status === 401 && !AUTH_PATHS.some((p) => url.includes(p))) {
    const tokens = await refreshSession()
    if (tokens) {
      try {
        res = await send({ Authorization: `Bearer ${tokens.accessToken}` })
      } catch (error) {
        logger.error('api retry failed', { url, error })
        throw error
      }
    }
  }

  const body = (await res.text()) ?? ''
  if (!res.ok) {
    let parsed: ApiError
    try {
      parsed = JSON.parse(body).error as ApiError
    } catch {
      logger.warn('unexpected api error response', { url, status: res.status })
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
  currencyCode: string
  updatedAt: string
}

export const accountsApi = {
  getBalance: (accountId: string, accessToken: string) =>
    request<BalanceResponse>(`/api/v1/accounts/${accountId}/balance`, {
      headers: authHeader(accessToken),
    }),
}

export type TransactionType = 'debit' | 'credit'
export type TransactionStatus = 'pending' | 'completed' | 'failed' | 'reversed'

export type TransactionResponse = {
  transactionId: string
  transferId: string | null
  type: TransactionType
  amountCents: number
  counterpartyAccountId: string
  status: TransactionStatus | null
  createdAt: string
}

export type TransactionListResponse = {
  transactions: TransactionResponse[]
  nextCursor: string | null
  hasMore: boolean
}

export const transactionsApi = {
  list: (accountId: string, accessToken: string, cursor?: string, limit?: number) => {
    const params = new URLSearchParams()
    if (cursor) params.set('cursor', cursor)
    if (limit) params.set('limit', String(limit))
    const query = params.toString()
    return request<TransactionListResponse>(
      `/api/v1/accounts/${accountId}/transactions${query ? `?${query}` : ''}`,
      { headers: authHeader(accessToken) },
    )
  },
}

export type TransferStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'reversed'

export type TransferRequest = {
  destinationAccountId: string
  amountCents: number
}

export type TransferResponse = {
  transferId: string
  status: TransferStatus
  createdAt: string
  amountCents: number
  sourceAccountId: string
  destinationAccountId: string
  updatedAt: string
  failureReason: string | null
}

export const transfersApi = {
  create: (body: TransferRequest, accessToken: string, idempotencyKey: string) =>
    request<TransferResponse>('/api/v1/transfers', {
      method: 'POST',
      headers: { ...authHeader(accessToken), 'Idempotency-Key': idempotencyKey },
      body: JSON.stringify(body),
    }),
  get: (transferId: string, accessToken: string) =>
    request<TransferResponse>(`/api/v1/transfers/${transferId}`, {
      headers: authHeader(accessToken),
    }),
}
