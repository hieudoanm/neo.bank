import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { authApi, accountsApi, transactionsApi, transfersApi, HttpError } from './client'
import { logger } from '@/lib/log'
import { refreshSession } from '@/lib/authSession'

vi.mock('@/lib/log', () => ({
  logger: { debug: vi.fn(), info: vi.fn(), warn: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/authSession', () => ({
  refreshSession: vi.fn(),
}))

const okJson = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const tokenResponse = {
  userId: 'u1',
  accountId: 'a1',
  accessToken: 'at',
  refreshToken: 'rt',
  tokenType: 'Bearer',
  expiresIn: 900,
}

const transferResponse = {
  transferId: 't1',
  status: 'pending',
  createdAt: '2026-01-01T00:00:00Z',
  amountCents: 100,
  sourceAccountId: 'a1',
  destinationAccountId: 'a2',
  updatedAt: '2026-01-01T00:00:00Z',
  failureReason: null,
}

describe('api client', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('posts credentials to /api/v1/login and parses the token response', async () => {
    fetchMock.mockResolvedValueOnce(okJson(tokenResponse))

    const result = await authApi.login({ email: 'user@example.com', password: 'Str0ng!Passw0rd' })

    expect(result).toEqual(tokenResponse)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/login')
    expect(init.method).toBe('POST')
    expect(init.headers['Content-Type']).toBe('application/json')
    expect(JSON.parse(init.body)).toEqual({
      email: 'user@example.com',
      password: 'Str0ng!Passw0rd',
    })
  })

  it('posts to /api/v1/register', async () => {
    fetchMock.mockResolvedValueOnce(okJson(tokenResponse))
    await authApi.register({ email: 'user@example.com', password: 'Str0ng!Passw0rd' })
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/register')
  })

  it('posts the refresh token to /api/v1/refresh', async () => {
    fetchMock.mockResolvedValueOnce(okJson({ accessToken: 'newat', refreshToken: 'newrt' }))
    const result = await authApi.refresh({ refreshToken: 'rt' })
    expect(result).toEqual({ accessToken: 'newat', refreshToken: 'newrt' })
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/refresh')
  })

  it('returns undefined for an empty 204 logout response', async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }))
    await expect(authApi.logout({ refreshToken: 'rt' })).resolves.toBeUndefined()
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/logout')
  })

  it('adds the bearer token to balance requests', async () => {
    fetchMock.mockResolvedValueOnce(
      okJson({
        accountId: 'a1',
        balanceCents: 500,
        currencyCode: 'USD',
        updatedAt: '2026-01-01T00:00:00Z',
      }),
    )
    const result = await accountsApi.getBalance('a1', 'tok')
    expect(result).toEqual({
      accountId: 'a1',
      balanceCents: 500,
      currencyCode: 'USD',
      updatedAt: '2026-01-01T00:00:00Z',
    })
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/accounts/a1/balance')
    expect(init.headers.Authorization).toBe('Bearer tok')
  })

  it('omits the auth header when no access token is supplied', async () => {
    fetchMock.mockResolvedValueOnce(
      okJson({ accountId: 'a1', balanceCents: 0, currencyCode: 'USD' }),
    )
    await accountsApi.getBalance('a1', '')
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBeUndefined()
  })

  it('encodes cursor and limit as query parameters', async () => {
    fetchMock.mockResolvedValueOnce(okJson({ transactions: [], nextCursor: null, hasMore: false }))
    await transactionsApi.list('a1', 'tok', 'cur', 20)
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/accounts/a1/transactions?cursor=cur&limit=20')
  })

  it('omits the query string when no cursor or limit is given', async () => {
    fetchMock.mockResolvedValueOnce(okJson({ transactions: [], nextCursor: null, hasMore: false }))
    await transactionsApi.list('a1', 'tok')
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/accounts/a1/transactions')
  })

  it('sends the Idempotency-Key and bearer token on transfer creation', async () => {
    fetchMock.mockResolvedValueOnce(okJson(transferResponse, 202))
    await transfersApi.create({ destinationAccountId: 'a2', amountCents: 100 }, 'tok', 'key-1')
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/transfers')
    expect(init.headers['Idempotency-Key']).toBe('key-1')
    expect(init.headers.Authorization).toBe('Bearer tok')
  })

  it('fetches a transfer by id', async () => {
    fetchMock.mockResolvedValueOnce(okJson(transferResponse))
    const result = await transfersApi.get('t1', 'tok')
    expect(result.transferId).toBe('t1')
    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/transfers/t1')
  })

  it('throws an HttpError carrying the error envelope for an API error', async () => {
    fetchMock.mockResolvedValue(
      okJson(
        {
          error: {
            code: 'INVALID_CREDENTIALS',
            message: 'Invalid email or password',
            requestId: 'r1',
          },
        },
        401,
      ),
    )

    const err = await authApi
      .login({ email: 'a@b.com', password: 'Str0ng!Passw0rd' })
      .catch((e) => e)

    expect(err).toBeInstanceOf(HttpError)
    expect(err).toMatchObject({
      status: 401,
      error: { code: 'INVALID_CREDENTIALS', message: 'Invalid email or password', requestId: 'r1' },
    })
  })

  it('falls back to INTERNAL_ERROR when an error body is not JSON', async () => {
    fetchMock.mockResolvedValueOnce(new Response('<html>gateway error</html>', { status: 502 }))

    await expect(authApi.logout({ refreshToken: 'rt' })).rejects.toMatchObject({
      status: 502,
      error: { code: 'INTERNAL_ERROR', message: 'Unexpected error', requestId: '' },
    })
    expect(logger.warn).toHaveBeenCalledWith('unexpected api error response', {
      url: '/api/v1/logout',
      status: 502,
    })
  })

  it('refreshes the session and retries once with the new token on 401', async () => {
    vi.mocked(refreshSession).mockResolvedValue({ accessToken: 'newat', refreshToken: 'newrt' })
    fetchMock
      .mockResolvedValueOnce(
        okJson({ error: { code: 'UNAUTHORIZED', message: 'token expired', requestId: 'r1' } }, 401),
      )
      .mockResolvedValueOnce(okJson({ accountId: 'a1', balanceCents: 1, currencyCode: 'USD' }))

    const result = await accountsApi.getBalance('a1', 'stale')

    expect(result.balanceCents).toBe(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(fetchMock.mock.calls[1][1].headers.Authorization).toBe('Bearer newat')
  })

  it('surfaces the 401 when the session cannot be refreshed', async () => {
    vi.mocked(refreshSession).mockResolvedValue(null)
    fetchMock.mockResolvedValueOnce(
      okJson({ error: { code: 'UNAUTHORIZED', message: 'token expired', requestId: 'r1' } }, 401),
    )

    await expect(accountsApi.getBalance('a1', 'stale')).rejects.toMatchObject({
      status: 401,
      error: { code: 'UNAUTHORIZED' },
    })
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('logs and rethrows network failures', async () => {
    const failure = new TypeError('Failed to fetch')
    fetchMock.mockRejectedValueOnce(failure)

    await expect(accountsApi.getBalance('a1', 'tok')).rejects.toBe(failure)
    expect(logger.error).toHaveBeenCalledWith('api request failed', {
      url: '/api/v1/accounts/a1/balance',
      error: failure,
    })
  })
})
