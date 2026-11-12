import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import { ProtectedRoute } from './ProtectedRoute'

vi.mock('@/context/AuthContext', () => ({
  useAuth: vi.fn(),
}))

const mockedUseAuth = vi.mocked(useAuth)

const setAuth = (isAuthenticated: boolean) =>
  mockedUseAuth.mockReturnValue({
    state: {
      user: isAuthenticated ? { userId: 'u1', accountId: 'a1', email: 'e' } : null,
      accessToken: isAuthenticated ? 'at' : null,
      refreshToken: isAuthenticated ? 'rt' : null,
      isAuthenticated,
    },
    dispatch: vi.fn(),
    login: vi.fn(),
    refresh: vi.fn(),
    logout: vi.fn(),
  } as unknown as ReturnType<typeof useAuth>)

const renderAt = (path: string) =>
  render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/login" element={<div>Login screen</div>} />
        <Route
          path="/protected"
          element={
            <ProtectedRoute>
              <div>Protected content</div>
            </ProtectedRoute>
          }
        />
      </Routes>
    </MemoryRouter>,
  )

describe('ProtectedRoute', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders children when authenticated', () => {
    setAuth(true)
    renderAt('/protected')
    expect(screen.getByText('Protected content')).toBeInTheDocument()
  })

  it('redirects to /login when unauthenticated', () => {
    setAuth(false)
    renderAt('/protected')
    expect(screen.getByText('Login screen')).toBeInTheDocument()
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
  })
})
