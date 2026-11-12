import { useEffect } from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { TransactionsPage } from './TransactionsPage'
import { transactionsApi, type TransactionListResponse } from '@/lib/api/client'

vi.mock('@/lib/api/client', () => ({
  transactionsApi: {
    list: vi.fn(),
  },
}))

const mockedList = vi.mocked(transactionsApi.list)

const usd = (cents: number) =>
  new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(cents / 100)

const AutoLogin = () => {
  const { login } = useAuth()
  useEffect(() => {
    login({ userId: 'u1', accountId: 'acc-1234', email: 'user@example.com' }, 'test-token', 'rt')
  }, [login])
  return null
}

const renderPage = () =>
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter>
        <AuthProvider>
          <AutoLogin />
          <TransactionsPage />
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )

const debit: TransactionListResponse = {
  transactions: [
    {
      transactionId: 't1',
      transferId: 'tr1',
      type: 'debit',
      amountCents: 320000,
      counterpartyAccountId: '11111111-1111-1111-1111-111111111234',
      status: 'completed',
      createdAt: '2026-10-14T20:30:00Z',
    },
  ],
  nextCursor: 'cursor-2',
  hasMore: true,
}

const credit: TransactionListResponse = {
  transactions: [
    {
      transactionId: 't2',
      transferId: 'tr2',
      type: 'credit',
      amountCents: 5000,
      counterpartyAccountId: '22222222-2222-2222-2222-222222222567',
      status: 'completed',
      createdAt: '2026-10-15T09:00:00Z',
    },
  ],
  nextCursor: null,
  hasMore: false,
}

describe('TransactionsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads and renders transactions for the signed-in account', async () => {
    mockedList.mockResolvedValue(debit)
    renderPage()

    expect(await screen.findByText('Transfer to ACC-1234')).toBeInTheDocument()
    expect(screen.getByText(`-${usd(320000)}`)).toBeInTheDocument()
    expect(screen.getByText('completed')).toBeInTheDocument()
    expect(screen.getByText('debit transaction')).toBeInTheDocument()
    await waitFor(() =>
      expect(mockedList).toHaveBeenCalledWith('acc-1234', 'test-token', undefined, 20),
    )
  })

  it('renders credits with a positive amount and IN marker', async () => {
    mockedList.mockResolvedValue(credit)
    renderPage()

    expect(await screen.findByText('Transfer from ACC-2567')).toBeInTheDocument()
    expect(screen.getByText(`+${usd(5000)}`)).toBeInTheDocument()
    expect(screen.getByText('credit transaction')).toBeInTheDocument()
  })

  it('shows the empty state when there are no transactions', async () => {
    mockedList.mockResolvedValue({ transactions: [], nextCursor: null, hasMore: false })
    renderPage()

    expect(await screen.findByText('No transactions yet')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /transfer money/i })).toBeInTheDocument()
  })

  it('shows a retryable error on network failure', async () => {
    const error = new Error('offline') as Error & { status?: number }
    error.status = 0
    mockedList.mockRejectedValueOnce(error).mockResolvedValueOnce(debit)
    renderPage()

    expect(await screen.findByText(/unable to load your transactions/i)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /try again/i }))
    expect(await screen.findByText('Transfer to ACC-1234')).toBeInTheDocument()
  })

  it('shows a non-retryable account-not-found error on 404', async () => {
    const error = new Error('not found') as Error & { status?: number }
    error.status = 404
    mockedList.mockRejectedValue(error)
    renderPage()

    expect(await screen.findByText('Account not found.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /try again/i })).not.toBeInTheDocument()
  })

  it('offers re-authentication on 401', async () => {
    const error = new Error('unauthorized') as Error & { status?: number }
    error.status = 401
    mockedList.mockRejectedValue(error)
    renderPage()

    expect(await screen.findByText(/session has expired/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('paginates forward with the cursor and back to loaded pages', async () => {
    mockedList.mockResolvedValueOnce(debit).mockResolvedValueOnce(credit)
    renderPage()

    expect(await screen.findByText('Transfer to ACC-1234')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /next/i }))

    expect(await screen.findByText('Transfer from ACC-2567')).toBeInTheDocument()
    await waitFor(() =>
      expect(mockedList).toHaveBeenCalledWith('acc-1234', 'test-token', 'cursor-2', 20),
    )

    fireEvent.click(screen.getByRole('button', { name: /previous/i }))
    expect(await screen.findByText('Transfer to ACC-1234')).toBeInTheDocument()
    expect(mockedList).toHaveBeenCalledTimes(2)
  })
})
