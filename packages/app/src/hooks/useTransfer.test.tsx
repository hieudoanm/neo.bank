import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useAuth } from '@/context/AuthContext'
import { transfersApi } from '@/lib/api/client'
import { useTransfer } from './useTransfer'

vi.mock('@/lib/api/client', () => ({ transfersApi: { get: vi.fn() } }))
vi.mock('@/context/AuthContext', () => ({ useAuth: vi.fn() }))

const mockedGet = vi.mocked(transfersApi.get)
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

const transfer = (status: string) => ({
  transferId: 't1',
  status,
  createdAt: '2026-01-01T00:00:00Z',
  amountCents: 100,
  sourceAccountId: 'a1',
  destinationAccountId: 'a2',
  updatedAt: '2026-01-01T00:00:00Z',
  failureReason: null,
})

describe('useTransfer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('fetches the transfer for the given id', async () => {
    setAuth(true)
    mockedGet.mockResolvedValue(transfer('pending') as never)

    const { result } = renderHook(() => useTransfer('t1'), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedGet).toHaveBeenCalledWith('t1', 'at')
    expect(result.current.data?.status).toBe('pending')
  })

  it('does not fetch when no transfer id is given', () => {
    setAuth(true)

    const { result } = renderHook(() => useTransfer(''), { wrapper })

    expect(result.current.fetchStatus).toBe('idle')
    expect(mockedGet).not.toHaveBeenCalled()
  })

  it('keeps polling while the transfer is pending', async () => {
    vi.useFakeTimers()
    setAuth(true)
    mockedGet.mockResolvedValue(transfer('pending') as never)

    const { result } = renderHook(() => useTransfer('t1'), { wrapper })
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })
    expect(result.current.isSuccess).toBe(true)
    expect(mockedGet).toHaveBeenCalledTimes(1)

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000)
    })
    expect(mockedGet).toHaveBeenCalledTimes(2)
  })

  it.each(['completed', 'failed', 'reversed'])(
    'stops polling for terminal status %s',
    async (status) => {
      vi.useFakeTimers()
      setAuth(true)
      mockedGet.mockResolvedValue(transfer(status) as never)

      const { result } = renderHook(() => useTransfer('t1'), { wrapper })
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0)
      })
      expect(result.current.isSuccess).toBe(true)

      await act(async () => {
        await vi.advanceTimersByTimeAsync(5000)
      })
      expect(mockedGet).toHaveBeenCalledTimes(1)
    },
  )
})
