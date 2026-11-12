import { useEffect } from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useParams } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { TransferPage } from './TransferPage'
import { transfersApi } from '@/lib/api/client'

vi.mock('@/lib/api/client', () => ({
  transfersApi: {
    create: vi.fn(),
    get: vi.fn(),
  },
}))

const mockedCreate = vi.mocked(transfersApi.create)

const ACCOUNT_ID = '11111111-1111-1111-1111-111111111111'

const AutoLogin = () => {
  const { login } = useAuth()
  useEffect(() => {
    login({ userId: 'u1', accountId: ACCOUNT_ID, email: 'user@example.com' }, 'test-token', 'rt')
  }, [login])
  return null
}

const StatusStub = () => {
  const { transferId } = useParams<{ transferId: string }>()
  return <div>STATUS:{transferId}</div>
}

const renderPage = () =>
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={['/transfer']}>
        <AuthProvider>
          <AutoLogin />
          <Routes>
            <Route path="/transfer" element={<TransferPage />} />
            <Route path="/transfer/:transferId" element={<StatusStub />} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )

const fill = (destination: string, amount: string) => {
  fireEvent.change(screen.getByLabelText(/destination account id/i), {
    target: { value: destination },
  })
  fireEvent.change(screen.getByLabelText(/^amount$/i), { target: { value: amount } })
}

describe('TransferPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the transfer form', () => {
    renderPage()
    expect(screen.getByLabelText(/destination account id/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^amount$/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /confirm transfer/i })).toBeInTheDocument()
  })

  it('validates required fields without calling the API', async () => {
    renderPage()
    fireEvent.submit(screen.getByRole('button', { name: /confirm transfer/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('Destination account ID is required')).toBeInTheDocument()
      expect(screen.getByText('Amount is required')).toBeInTheDocument()
    })
    expect(mockedCreate).not.toHaveBeenCalled()
  })

  it('rejects a malformed destination account ID', async () => {
    renderPage()
    fill('not-a-uuid', '10.00')
    fireEvent.submit(screen.getByRole('button', { name: /confirm transfer/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('Enter a valid account ID')).toBeInTheDocument()
    })
    expect(mockedCreate).not.toHaveBeenCalled()
  })

  it('rejects a transfer to the same account', async () => {
    renderPage()
    fill(ACCOUNT_ID, '10.00')
    fireEvent.submit(screen.getByRole('button', { name: /confirm transfer/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('You cannot transfer to your own account')).toBeInTheDocument()
    })
    expect(mockedCreate).not.toHaveBeenCalled()
  })

  it('submits in minor units with an idempotency key and navigates to the status page', async () => {
    mockedCreate.mockResolvedValue({
      transferId: 'tr-9',
      status: 'pending',
      createdAt: '2026-10-15T20:00:00Z',
      amountCents: 25000,
      sourceAccountId: ACCOUNT_ID,
      destinationAccountId: '550e8400-e29b-41d4-a716-446655440001',
      updatedAt: '2026-10-15T20:00:00Z',
      failureReason: null,
    })
    renderPage()
    fill('550e8400-e29b-41d4-a716-446655440001', '250.00')
    fireEvent.submit(screen.getByRole('button', { name: /confirm transfer/i }).closest('form')!)

    expect(await screen.findByText('STATUS:tr-9')).toBeInTheDocument()
    expect(mockedCreate).toHaveBeenCalledTimes(1)
    const [body, token, key] = mockedCreate.mock.calls[0]
    expect(body).toEqual({
      destinationAccountId: '550e8400-e29b-41d4-a716-446655440001',
      amountCents: 25000,
    })
    expect(token).toBe('test-token')
    expect(key).toMatch(/^[a-zA-Z0-9_-]+$/)
  })

  it('shows a form error on insufficient funds', async () => {
    const error = new Error('nope') as Error & { status?: number; error?: { code?: string } }
    error.status = 400
    error.error = { code: 'INSUFFICIENT_FUNDS' }
    mockedCreate.mockRejectedValue(error)
    renderPage()
    fill('550e8400-e29b-41d4-a716-446655440001', '250.00')
    fireEvent.submit(screen.getByRole('button', { name: /confirm transfer/i }).closest('form')!)

    expect(await screen.findByText(/insufficient funds/i)).toBeInTheDocument()
  })
})
