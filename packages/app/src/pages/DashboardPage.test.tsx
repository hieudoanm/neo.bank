import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { useAuth } from '@/context/AuthContext'
import { useBalance } from '@/hooks/useBalance'
import { DashboardPage } from './DashboardPage'

vi.mock('@/context/AuthContext', () => ({ useAuth: vi.fn() }))
vi.mock('@/hooks/useBalance', () => ({ useBalance: vi.fn() }))

const mockedUseAuth = vi.mocked(useAuth)
const mockedUseBalance = vi.mocked(useBalance)

const setAuth = () =>
  mockedUseAuth.mockReturnValue({
    state: {
      user: { userId: 'u1', accountId: 'a1', email: 'user@example.com' },
      accessToken: 'at',
      refreshToken: 'rt',
      isAuthenticated: true,
    },
    dispatch: vi.fn(),
    login: vi.fn(),
    refresh: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>)

const setBalance = (overrides: Record<string, unknown>) =>
  mockedUseBalance.mockReturnValue({
    data: undefined,
    isLoading: false,
    isError: false,
    isFetching: false,
    refetch: vi.fn(),
    ...overrides,
  } as unknown as ReturnType<typeof useBalance>)

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setAuth()
  })

  it('shows a loading placeholder while fetching', () => {
    setBalance({ isLoading: true })
    render(<DashboardPage />)
    expect(screen.getByText('Current balance')).toBeInTheDocument()
    expect(screen.queryByText(/\$/)).not.toBeInTheDocument()
  })

  it('renders the formatted balance on success', () => {
    setBalance({
      data: {
        accountId: 'a1',
        balanceCents: 123456,
        currencyCode: 'USD',
        updatedAt: '2026-01-01T00:00:00Z',
      },
    })
    render(<DashboardPage />)
    expect(screen.getByText('$1,234.56')).toBeInTheDocument()
    expect(screen.getByText('123456 minor units (USD)')).toBeInTheDocument()
    expect(screen.getByText('Account: a1')).toBeInTheDocument()
  })

  it('shows an error with retry and calls refetch on click', () => {
    const refetch = vi.fn()
    setBalance({ isError: true, refetch })
    render(<DashboardPage />)
    expect(screen.getByText(/unable to load your balance/i)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /retry/i }))
    expect(refetch).toHaveBeenCalledTimes(1)
  })
})
