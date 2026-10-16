import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/context/AuthContext'
import { RegisterPage } from './RegisterPage'
import * as authApi from '@/lib/api/client'

vi.mock('@/lib/api/client', () => ({
  authApi: {
    register: vi.fn(),
  },
}))

const mockedRegister = vi.mocked(authApi.authApi.register)

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })

const renderPage = () =>
  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/register']}>
        <AuthProvider>
          <RegisterPage />
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )

const getInputs = () => ({
  email: screen.getByLabelText(/email/i),
  password: screen.getByLabelText(/^password$/i),
  confirmPassword: screen.getByLabelText(/confirm password/i),
})

describe('RegisterPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the form with email, password, confirm password, and create account button', () => {
    renderPage()
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
  })

  it('shows field validation errors for empty submit and does not call API', async () => {
    renderPage()
    fireEvent.submit(screen.getByRole('button', { name: /create account/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeInTheDocument()
      expect(screen.getByText('Password is required')).toBeInTheDocument()
      expect(screen.getByText('Please confirm your password')).toBeInTheDocument()
    })
    expect(mockedRegister).not.toHaveBeenCalled()
  })

  it('shows validation error when passwords do not match', async () => {
    renderPage()
    const { email, password, confirmPassword } = getInputs()
    fireEvent.change(email, { target: { value: 'user@example.com' } })
    fireEvent.change(password, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.change(confirmPassword, { target: { value: 'Different!Passw0rd' } })
    fireEvent.submit(screen.getByRole('button', { name: /create account/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
    })
    expect(mockedRegister).not.toHaveBeenCalled()
  })

  it('calls register API with email and password on valid submit', async () => {
    mockedRegister.mockResolvedValue({
      userId: 'u1',
      accountId: 'a1',
      accessToken: 'at',
      refreshToken: 'rt',
      tokenType: 'Bearer',
      expiresIn: 900,
    })
    renderPage()
    const { email, password, confirmPassword } = getInputs()
    fireEvent.change(email, { target: { value: 'user@example.com' } })
    fireEvent.change(password, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.change(confirmPassword, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.submit(screen.getByRole('button', { name: /create account/i }).closest('form')!)

    await waitFor(() => {
      expect(mockedRegister).toHaveBeenCalledTimes(1)
      expect(mockedRegister.mock.calls[0][0]).toMatchObject({
        email: 'user@example.com',
        password: 'Str0ng!Passw0rd',
      })
    })
  })

  it('shows server error when email already registered', async () => {
    const error = new Error('boom') as any
    error.status = 409
    error.error = { code: 'EMAIL_ALREADY_REGISTERED', message: 'Email is already registered', requestId: 'r1' }
    mockedRegister.mockRejectedValue(error)
    renderPage()
    const { email, password, confirmPassword } = getInputs()
    fireEvent.change(email, { target: { value: 'user@example.com' } })
    fireEvent.change(password, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.change(confirmPassword, { target: { value: 'Str0ng!Passw0rd' } })
    fireEvent.submit(screen.getByRole('button', { name: /create account/i }).closest('form')!)

    await waitFor(() => {
      expect(screen.getByRole('alert').textContent).toContain('Email is already registered')
    })
  })
})
