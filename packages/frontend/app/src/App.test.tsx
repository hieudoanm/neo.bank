import { describe, it, expect, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from './App'

const renderAt = (path: string) => {
  window.history.pushState({}, '', path)
  return render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <App />
    </QueryClientProvider>,
  )
}

describe('App', () => {
  afterEach(() => {
    window.history.pushState({}, '', '/')
  })

  it('renders the login page at /login', () => {
    renderAt('/login')
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('redirects unknown routes to the login page', () => {
    renderAt('/does-not-exist')
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('redirects a protected route to login when unauthenticated', () => {
    renderAt('/transactions')
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })
})
