# NEO Bank — React Router Best Practices

_Frontend · Core · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [core index](./README.md)._

## 1. Route Structure

Routes describe the banking flows: public auth pages, a protected application shell, and a not-found fallback.

```tsx
// App.tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/context/AuthContext';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { Layout } from '@/components/layout/Layout';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { DashboardPage } from '@/pages/DashboardPage';
import { TransferPage } from '@/pages/TransferPage';
import { TransactionsPage } from '@/pages/TransactionsPage';

export const App = () => {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/transfer" element={<TransferPage />} />
            <Route path="/transfers/:transferId" element={<TransferStatusPage />} />
            <Route path="/transactions" element={<TransactionsPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
};
```

Rules:

- Public routes (`/login`, `/register`) are declared outside the protected shell.
- One shell route wraps every authenticated page; pages render into its `<Outlet />`.
- The catch-all redirects to `/` rather than rendering a blank screen.

## 2. Protected Routes

Authorization is enforced by the shell, not by each page.

```tsx
// components/auth/ProtectedRoute.tsx
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';

export const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { state } = useAuth();
  const location = useLocation();

  if (!state.isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
};
```

Rules:

- Redirect unauthenticated users to `/login` with `replace` so back does not loop.
- Preserve the attempted location in `state.from` and return there after login.
- Client routing is for usability only; the server remains the authority on access.
- A `404`-style route is preferred over revealing that a resource exists but is forbidden.

## 3. Layout Routes

Nested routes share a layout through `<Outlet />`.

```tsx
// components/layout/Layout.tsx
import { Outlet } from 'react-router-dom';
import { Header } from './Header';
import { Footer } from './Footer';

export const Layout = () => {
  return (
    <>
      <Header />
      <main className="mx-auto max-w-5xl px-6 py-8">
        <Outlet />
      </main>
      <Footer />
    </>
  );
};
```

Rules:

- Header, navigation, and footer live in the layout route, never duplicated per page.
- Pages are leaf routes; they render content only.
- Nest deeper only when a section genuinely shares chrome.

## 4. Navigation

```tsx
// components/layout/Header.tsx
import { NavLink, useNavigate } from 'react-router-dom';

const Nav = () => {
  return (
    <nav className="flex gap-6 text-sm font-medium">
      <NavLink to="/" end>Dashboard</NavLink>
      <NavLink to="/transfer">Transfer</NavLink>
      <NavLink to="/transactions">Transactions</NavLink>
    </nav>
  );
};
```

Rules:

- Use `<Link>`/`<NavLink>` for in-app navigation; never a bare `<a href>` (full reload).
- `NavLink` supplies active styling; mark the index route with `end`.
- Use `useNavigate()` for programmatic navigation after a confirmed mutation.
- Never navigate on form submit before the server confirms success.

## 5. Code Splitting

Load route components lazily so the initial bundle stays small.

```tsx
import { lazy, Suspense } from 'react';

const TransactionsPage = lazy(() => import('@/pages/TransactionsPage'));

<Route
  path="/transactions"
  element={
    <Suspense fallback={<PageSkeleton />}>
      <TransactionsPage />
    </Suspense>
  }
/>
```

Rules:

- `React.lazy` per route; wrap in a `Suspense` fallback that matches the page shape.
- Keep auth pages eager (they are the first paint); split heavier authenticated pages.

## 6. Route Parameters & Redirects

```tsx
// pages/TransferStatusPage.tsx
import { useParams } from 'react-router-dom';

export const TransferStatusPage = () => {
  const { transferId } = useParams<{ transferId: string }>();
  if (!transferId) return <Navigate to="/transactions" replace />;
  const { data, isLoading } = useTransferStatus(transferId);
  // ...
};
```

Rules:

- Validate route params before use; redirect when absent or malformed.
- Treat a param as an identifier, not as authorization — the server checks ownership.
- Keep redirects declarative (`<Navigate>`) where possible.

## 7. Accessibility

- On route change, move focus to the page heading or main landmark so screen readers announce the new view.
- Announce route transitions with a polite live region when content changes without a full reload.
- Keep the document title in sync with the current route.
- Ensure skip links and landmarks (`<main>`, `<nav>`) are part of the layout route.

## 8. Testing

```tsx
// components/auth/__tests__/ProtectedRoute.test.tsx
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

test('redirects unauthenticated users to login', () => {
  render(
    <MemoryRouter initialEntries={['/']}>
      <AuthProvider value={{ isAuthenticated: false }}>
        <Routes>
          <Route path="/" element={<ProtectedRoute>Secret</ProtectedRoute>} />
          <Route path="/login" element={<div>Login</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>
  );
  expect(screen.getByText('Login')).toBeInTheDocument();
});
```

Rules:

- Wrap components under test in `MemoryRouter`; set `initialEntries` per case.
- Assert on navigation outcomes (rendered page, redirect), not on router internals.
- Test protected routes negatively: unauthenticated access must redirect.

## 9. Code Style & Rules

- Declare routes centrally in `App.tsx`; keep page components free of route tables.
- Arrow functions only; no `function` declarations.
- Public routes outside the shell, authenticated routes inside it.
- `<Link>`/`<NavLink>`/`useNavigate` only — no full-page reloads.
- Lazy-load authenticated pages behind a `Suspense` fallback.
- Focus and title follow the active route.
- Related: [`react.md`](./react.md) for components and state, [`vitest.md`](../testing/vitest.md) for unit tests, [`playwright.md`](../testing/playwright.md) for end-to-end flows.
