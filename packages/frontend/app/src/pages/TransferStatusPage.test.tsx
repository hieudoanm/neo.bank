import { useEffect } from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { TransferStatusPage } from './TransferStatusPage'
import { transfersApi, type TransferResponse } from '@/lib/api/client'

vi.mock('@/lib/api/client', () => ({
  transfersApi: {
    create: vi.fn(),
    get: vi.fn(),
  },
}))

const mockedGet = vi.mocked(transfersApi.get)

const AutoLogin = () => {
  const { login } = useAuth()
  useEffect(() => {
    login({ userId: 'u1', accountId: 'acc-1', email: 'user@example.com' }, 'test-token', 'rt')
  }, [login])
  return null
}

const renderPage = (transferId = 'tr-1') =>
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[`/transfer/${transferId}`]}>
        <AuthProvider>
          <AutoLogin />
          <Routes>
            <Route path="/transfer/:transferId" element={<TransferStatusPage />} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )

const transfer = (overrides: Partial<TransferResponse> = {}): TransferResponse => ({
  transferId: 'tr-1',
  status: 'pending',
  createdAt: '2026-10-15T20:00:00Z',
  amountCents: 25000,
  sourceAccountId: 'acc-1',
  destinationAccountId: '11111111-1111-1111-1111-111111111234',
  updatedAt: '2026-10-15T20:00:00Z',
  failureReason: null,
  ...overrides,
})

describe('TransferStatusPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows a processing state while pending', async () => {
    mockedGet.mockResolvedValue(transfer())
    renderPage()

    expect(await screen.findByText('Processing your transfer')).toBeInTheDocument()
    expect(mockedGet).toHaveBeenCalledWith('tr-1', 'test-token')
  })

  it('shows the completed state with details', async () => {
    mockedGet.mockResolvedValue(transfer({ status: 'completed' }))
    renderPage()

    expect(await screen.findByText('Transfer complete')).toBeInTheDocument()
    expect(screen.getByText('ACC-1234')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /view transactions/i })).toBeInTheDocument()
  })

  it('shows the failed state with a mapped reason', async () => {
    mockedGet.mockResolvedValue(transfer({ status: 'failed', failureReason: 'INSUFFICIENT_FUNDS' }))
    renderPage()

    expect(await screen.findByText('Transfer failed')).toBeInTheDocument()
    expect(screen.getByText(/insufficient funds/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument()
  })

  it('shows a non-retryable not-found error on 404', async () => {
    const error = new Error('not found') as Error & { status?: number }
    error.status = 404
    mockedGet.mockRejectedValue(error)
    renderPage()

    expect(await screen.findByText('Transfer not found.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /try again/i })).not.toBeInTheDocument()
  })

  it('offers re-authentication on 401', async () => {
    const error = new Error('unauthorized') as Error & { status?: number }
    error.status = 401
    mockedGet.mockRejectedValue(error)
    renderPage()

    expect(await screen.findByText(/session has expired/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('retries after a transient error', async () => {
    const error = new Error('offline') as Error & { status?: number }
    error.status = 0
    mockedGet.mockRejectedValueOnce(error).mockResolvedValueOnce(transfer({ status: 'completed' }))
    renderPage()

    expect(await screen.findByText(/unable to load this transfer/i)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /try again/i }))
    expect(await screen.findByText('Transfer complete')).toBeInTheDocument()
  })
})
