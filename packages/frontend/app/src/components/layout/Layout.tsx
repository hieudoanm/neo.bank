import { Outlet } from 'react-router-dom'

export const Layout = () => {
  return (
    <>
      <header className="bg-surface border-b border-border px-4 py-3 flex items-center justify-between">
        <span className="text-text-primary font-semibold tracking-tight text-lg">NEO Bank</span>
        <nav className="flex gap-3 text-sm font-medium">
          <a href="/" className="text-text-secondary hover:text-text-primary">
            Dashboard
          </a>
        </nav>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-8">
        <Outlet />
      </main>
    </>
  )
}
