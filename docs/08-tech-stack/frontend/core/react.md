# NEO Bank — React Best Practices

_Frontend · Core · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [core index](./README.md)._

This document covers React-proper concerns: components, client state, UI states, accessibility, and performance. Sibling documents own the surrounding stack:

| Concern                 | Document                                                           |
| ----------------------- | ------------------------------------------------------------------ |
| Routing                 | [`react-router.md`](./react-router.md)                             |
| Server state / fetching | [`tanstack-query.md`](./tanstack-query.md)                         |
| Unit & component tests  | [`vitest.md`](../testing/vitest.md)                                |
| End-to-end tests        | [`playwright.md`](../testing/playwright.md)                        |
| Styling utilities       | [`tailwindcss.md`](../styles/tailwindcss.md)                       |
| Component classes       | [`daisyui.md`](../styles/daisyui.md)                               |
| Icons                   | [`iconography.md`](../../../07-ui-ux-design/system/iconography.md) |
| Types & API client      | [`typescript.md`](./typescript.md)                                 |

## 1. Component Architecture

### Component Categories

```
components/
├── ui/              # Primitive, reusable UI components (Button, Input, Card, Badge)
├── forms/           # Form-specific components (FormField, FormError, FormSubmit)
├── features/        # Feature-specific components (TransferForm, BalanceDisplay, TransactionList)
├── layout/          # Layout components (Header, Sidebar, PageContainer, Grid)
└── charts/          # Data visualisation components (future)
```

Rules:

- `ui/` components are presentational and unopinionated about data.
- `features/` components may use hooks and domain logic.
- Keep a component's tests and styles colocated with it.

### Component Structure

```tsx
// components/features/TransferForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { FormField } from '@/components/forms/FormField';
import { transferSchema, type CreateTransferRequest } from '@/lib/validation/transfer';

interface TransferFormProps {
  onSubmit: (data: CreateTransferRequest) => Promise<void>;
  isLoading?: boolean;
}

export const TransferForm = ({ onSubmit, isLoading }: TransferFormProps) => {
  const form = useForm<CreateTransferRequest>({
    resolver: zodResolver(transferSchema),
    defaultValues: { destinationAccountId: '', amount: 0 },
  });

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
      <FormField
        control={form.control}
        name="destinationAccountId"
        render={({ field }) => (
          <Input
            {...field}
            label="Destination account ID"
            error={form.formState.errors.destinationAccountId?.message}
          />
        )}
      />
      <FormField
        control={form.control}
        name="amount"
        render={({ field }) => (
          <Input
            {...field}
            type="number"
            label="Amount (cents)"
            min={1}
            error={form.formState.errors.amount?.message}
          />
        )}
      />
      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading ? 'Processing...' : 'Transfer'}
      </Button>
    </form>
  );
};
```

Rules:

- Arrow-function components with hooks only; no class components, no `function` declarations.
- Type every prop with an explicit interface.
- One responsibility per component; extract a hook when logic grows.
- Presentational components receive data and callbacks as props; they do not fetch.

## 2. Client State (React Context + useReducer)

Server data belongs to [`tanstack-query.md`](./tanstack-query.md). React state holds the current user, UI flags, and form drafts.

```tsx
// context/AuthContext.tsx
interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
}

type AuthAction =
  | { type: 'LOGIN'; payload: { user: User; accessToken: string } }
  | { type: 'LOGOUT' }
  | { type: 'REFRESH'; payload: { accessToken: string } };

const AuthContext = createContext<{ state: AuthState; dispatch: React.Dispatch<AuthAction> } | null>(null);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);
  return <AuthContext.Provider value={{ state, dispatch }}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
};
```

Rules:

- Reach for context only for genuinely global client state (auth, theme); local `useState` otherwise.
- A context hook throws when used outside its provider — fail loudly, not silently.
- Keep reducers pure and actions explicit; never mutate state.
- Never store tokens in `localStorage`; keep them in memory and refresh via the API layer.

## 3. UI States

Every data screen renders all of these states. Data itself comes from the query layer.

```tsx
// components/features/TransactionList.tsx
export const TransactionList = () => {
  const { data, isLoading, isError, error, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useTransactions();

  if (isLoading) return <TransactionListSkeleton />;
  if (isError) return <ErrorState error={error} />;
  if (!data?.pages.length) return <EmptyState message="No transactions yet" />;

  return (
    <div className="space-y-3">
      {data.pages.flatMap((page) => page.transactions).map((tx) => (
        <TransactionItem key={tx.id} transaction={tx} />
      ))}
      {hasNextPage && (
        <Button onClick={() => fetchNextPage()} disabled={isFetchingNextPage} variant="outline">
          {isFetchingNextPage ? 'Loading...' : 'Load more'}
        </Button>
      )}
    </div>
  );
};
```

### Required States

| State      | Component              | Purpose                        |
| ---------- | ---------------------- | ------------------------------ |
| Loading    | `<Skeleton />`         | Initial load, show structure   |
| Empty      | `<EmptyState />`       | No data, guide to action       |
| Error      | `<ErrorState />`       | Recoverable error with retry   |
| Permission | `<PermissionDenied />` | 404 only (never 403)           |
| Network    | `<NetworkError />`     | Offline/connection issues      |
| Success    | Data display           | Confirmed server response      |
| Disabled   | Disabled inputs        | Explain why action unavailable |

Rules:

- Never show success before the server confirms it.
- Error states must offer a retry where the operation is safe to repeat.
- Hide unavailable actions for usability, but trust the server as the authority.

## 4. Accessibility

- Semantic HTML (`<main>`, `<section>`, `<button>`, `<label>`).
- Visible focus rings (`focus-visible:ring-2`); never remove focus without a replacement.
- ARIA labels for icon-only buttons.
- Colour contrast meets WCAG 2.1 AA.
- Keyboard navigation reaches every interactive element.
- Announce async updates with `aria-live="polite"`.

## 5. Performance

- **Code splitting:** `React.lazy` at the route level (see [`react-router.md`](./react-router.md)).
- **Memoisation:** `React.memo` for list items; `useMemo`/`useCallback` only for measured hot paths.
- **Virtualisation:** `react-window` for long transaction lists (future).
- **Bundle analysis:** `vite-bundle-analyzer` in CI.
- Keep server data in the query cache; do not mirror it into component state.

## 6. Code Style & Rules

- Arrow-function components with hooks; explicit prop interfaces.
- Named exports for components and utilities; default exports for pages.
- Colocate components with their tests; keep domain logic out of presentational components.
- Client state via context/reducer; server state via TanStack Query.
- Every data screen implements loading, empty, error, permission, and network states.
- Avoid inline styles; style via the token layer ([`tailwindcss.md`](../styles/tailwindcss.md), [`daisyui.md`](../styles/daisyui.md)).
- Icons come from `react-icons/pi` (Phosphor); never hand-author SVG icons and never add a second icon library. Size and colour them with the `icon-*` utilities and token classes ([`iconography.md`](../../../07-ui-ux-design/system/iconography.md)).
