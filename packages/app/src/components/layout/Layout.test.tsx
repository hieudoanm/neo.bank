import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/context/AuthContext'
import { Layout } from './Layout'

const renderLayout = () =>
  render(
    <AuthProvider>
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<div>Outlet content</div>} />
            <Route path="/login" element={<div>Login page</div>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </AuthProvider>,
  )

describe('Layout', () => {
  it('renders the brand, navigation links, and the outlet', () => {
    renderLayout()
    expect(screen.getByText('NEO Bank')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Dashboard' })).toHaveAttribute('href', '/')
    expect(screen.getByRole('link', { name: 'Transfer' })).toHaveAttribute('href', '/transfer')
    expect(screen.getByRole('link', { name: 'Transactions' })).toHaveAttribute(
      'href',
      '/transactions',
    )
    expect(screen.getByText('Outlet content')).toBeInTheDocument()
  })

  it('navigates to login when signing out', () => {
    renderLayout()
    fireEvent.click(screen.getByRole('button', { name: 'Sign out' }))
    expect(screen.getByText('Login page')).toBeInTheDocument()
  })
})
