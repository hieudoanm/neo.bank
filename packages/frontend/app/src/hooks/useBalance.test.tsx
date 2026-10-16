import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useAuth } from '@/context/AuthContext'
import { accountsApi } from '@/lib/api/client'
import { useBalance } from './useBalance'

vi.mock('@/lib/api/client', () => ({ accountsApi: { getBalance: vi.fn() } }))
vi.mock('@/context/AuthContext', () => ({ useAuth: vi.fn() }))

const mockedGetBalance = vi.mocked(accountsApi.getBalance)
const mockedUseAuth = vi.mocked(useAuth)

const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    {children}
  </QueryClientProvider>
)

const setAuth = (authenticated: boolean) =>
  mockedUseAuth.mockReturnValue({
    state: {
      user: authenticated ? { userId: 'u1', accountId: 'a1', email: 'e' } : null,
      accessToken: authenticated ? 'at' : null,
      refreshToken: authenticated ? 'rt' : null,
      isAuthenticated: authenticated,
    },
    dispatch: vi.fn(),
    login: vi.fn(),
    refresh: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>)

describe('useBalance', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches the balance for the authenticated account', async () => {
    setAuth(true)
    mockedGetBalance.mockResolvedValue({
      accountId: 'a1',
      balanceCents: 500,
      currencyCode: 'USD',
      updatedAt: '2026-01-01T00:00:00Z',
    })

    const { result } = renderHook(() => useBalance(), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedGetBalance).toHaveBeenCalledWith('a1', 'at')
    expect(result.current.data).toEqual({
      accountId: 'a1',
      balanceCents: 500,
      currencyCode: 'USD',
      updatedAt: '2026-01-01T00:00:00Z',
    })
  })

  it('does not fetch while logged out', async () => {
    setAuth(false)

    const { result } = renderHook(() => useBalance(), { wrapper })

    expect(result.current.fetchStatus).toBe('idle')
    expect(mockedGetBalance).not.toHaveBeenCalled()
  })
})
