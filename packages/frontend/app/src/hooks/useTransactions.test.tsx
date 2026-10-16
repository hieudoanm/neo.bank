import { describe, it, expect, vi, beforeEach } from 'vitest'
import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useAuth } from '@/context/AuthContext'
import { transactionsApi } from '@/lib/api/client'
import { TRANSACTIONS_PAGE_SIZE, useTransactions } from './useTransactions'

vi.mock('@/lib/api/client', () => ({ transactionsApi: { list: vi.fn() } }))
vi.mock('@/context/AuthContext', () => ({ useAuth: vi.fn() }))

const mockedList = vi.mocked(transactionsApi.list)
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

describe('useTransactions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches the first page and follows the cursor to the next page', async () => {
    setAuth(true)
    mockedList
      .mockResolvedValueOnce({ transactions: [], nextCursor: 'c1', hasMore: true })
      .mockResolvedValueOnce({ transactions: [], nextCursor: null, hasMore: false })

    const { result } = renderHook(() => useTransactions(), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedList).toHaveBeenCalledWith('a1', 'at', undefined, TRANSACTIONS_PAGE_SIZE)
    expect(result.current.hasNextPage).toBe(true)

    await act(async () => {
      await result.current.fetchNextPage()
    })

    await waitFor(() => expect(mockedList).toHaveBeenCalledWith('a1', 'at', 'c1', TRANSACTIONS_PAGE_SIZE))
    await waitFor(() => expect(result.current.hasNextPage).toBe(false))
  })

  it('does not fetch while logged out', () => {
    setAuth(false)

    const { result } = renderHook(() => useTransactions(), { wrapper })

    expect(result.current.fetchStatus).toBe('idle')
    expect(mockedList).not.toHaveBeenCalled()
  })
})
