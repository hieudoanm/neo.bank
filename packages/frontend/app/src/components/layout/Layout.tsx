import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'

export const Layout = () => {
  const { logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <>
      <header className="bg-surface border-b border-border px-4 py-3 flex items-center justify-between">
        <span className="text-text-primary font-semibold tracking-tight text-lg">NEO Bank</span>
        <nav className="flex items-center gap-3 text-sm font-medium">
          <Link to="/" className="text-text-secondary hover:text-text-primary">
            Dashboard
          </Link>
          <Link to="/transfer" className="text-text-secondary hover:text-text-primary">
            Transfer
          </Link>
          <Link to="/transactions" className="text-text-secondary hover:text-text-primary">
            Transactions
          </Link>
          <button
            type="button"
            onClick={handleLogout}
            className="text-text-secondary hover:text-text-primary ml-2"
          >
            Sign out
          </button>
        </nav>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-8">
        <Outlet />
      </main>
    </>
  )
}
