import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/context/AuthContext'
import { LoginPage } from './LoginPage'
import * as authApi from '@/lib/api/client'

vi.mock('@/lib/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/api/client')>()
  return {
    ...actual,
    authApi: {
      login: vi.fn(),
    },
  }
})

const mockedLogin = vi.mocked(authApi.authApi.login)

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

const renderPage = () =>
  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/login']}>
        <AuthProvider>
          <LoginPage />
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )

const getInputs = () => ({
  email: screen.getByLabelText(/email/i),
  password: screen.getByLabelText(/^password$/i),
})

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders the form with email, password, and sign in button', () => {
    renderPage()
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('shows field validation errors for empty submit and does not call API', async () => {
    renderPage()
    fireEvent.submit(screen.getByRole('button', { name: /sign in/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeInTheDocument()
      expect(screen.getByText('Password is required')).toBeInTheDocument()
    })
    expect(mockedLogin).not.toHaveBeenCalled()
  })

  it('shows validation errors for weak password', async () => {
    renderPage()
    const { email, password } = getInputs()
    fireEvent.change(email, { target: { value: 'user@example.com' } })
    fireEvent.change(password, { target: { value: 'weak' } })
    fireEvent.submit(screen.getByRole('button', { name: /sign in/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('Password must be at least 12 characters')).toBeInTheDocument()
    })
    expect(mockedLogin).not.toHaveBeenCalled()
  })

  it('calls login API with credentials on valid submit', async () => {
    mockedLogin.mockResolvedValue({
      userId: 'u1',
      accountId: 'a1',
      accessToken: 'at',
      refreshToken: 'rt',
      tokenType: 'Bearer',
      expiresIn: 900,
    })
    renderPage()
    const { email, password } = getInputs()
    fireEvent.change(email, { target: { value: 'user@example.com' } })
    fireEvent.change(password, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.submit(screen.getByRole('button', { name: /sign in/i }).closest('form')!)

    await waitFor(() => {
      expect(mockedLogin).toHaveBeenCalledTimes(1)
      expect(mockedLogin.mock.calls[0][0]).toMatchObject({
        email: 'user@example.com',
        password: 'Str0ng!Passw0rd',
      })
    })
  })

  it('shows server error message on invalid credentials', async () => {
    const error = new authApi.HttpError(
      401,
      { code: 'INVALID_CREDENTIALS', message: 'Invalid email or password', requestId: 'r1' },
      'Invalid email or password',
    )
    mockedLogin.mockRejectedValue(error)
    renderPage()
    const { email, password } = getInputs()
    fireEvent.change(email, { target: { value: 'user@example.com' } })
    fireEvent.change(password, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.submit(screen.getByRole('button', { name: /sign in/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByRole('alert').textContent).toContain('Invalid email or password')
    })
  })
})
