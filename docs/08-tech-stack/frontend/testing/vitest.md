# NEO Bank — Vitest Best Practices

_Frontend · Testing · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [testing index](./README.md)._

## 1. Setup

Vitest runs unit and component tests with React Testing Library and jsdom.

```ts
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin/react';
import path from 'node:path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary', 'html'],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 80,
        statements: 80,
      },
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['**/*.test.tsx', '**/*.test.ts', '**/*.d.ts', 'src/main.tsx', 'src/setupTests.ts'],
    },
  },
});
```

```ts
// src/test/setup.ts
import '@testing-library/jest-dom';
import { vi } from 'vitest';

vi.mock('@/lib/api/client', () => ({
  api: { auth: {}, accounts: {}, transfers: {}, transactions: {} },
}));
```

Rules:

- One `setupFiles` entry for shared matchers (`@testing-library/jest-dom`) and cleanup.
- Coverage thresholds match `09-testing-strategy.md` (frontend 80% minimum for lines/functions/branches/statements).
- Tests live next to code (`*.test.tsx`) or in `__tests__/`.
- `globals: true` so `describe`/`it`/`expect` are available without imports.
- `jsdom` environment provides DOM APIs for component rendering.

## 2. Component Tests

Test observable behaviour: what the user sees and does.

```tsx
// components/features/__tests__/TransferForm.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TransferForm } from '../TransferForm';

test('submits a valid transfer', async () => {
  const onSubmit = vi.fn().mockResolvedValue(undefined);
  const user = userEvent.setup();
  render(<TransferForm onSubmit={onSubmit} />);

  await user.type(screen.getByLabelText('Destination account ID'), 'acc-123');
  await user.type(screen.getByLabelText('Amount (cents)'), '10000');
  await user.click(screen.getByRole('button', { name: 'Transfer' }));

  await waitFor(() =>
    expect(onSubmit).toHaveBeenCalledWith({ destinationAccountId: 'acc-123', amount: 10000 }),
  );
});

test('shows validation error for invalid amount', async () => {
  const onSubmit = vi.fn().mockResolvedValue(undefined);
  const user = userEvent.setup();
  render(<TransferForm onSubmit={onSubmit} />);

  await user.type(screen.getByLabelText('Amount (cents)'), '-500');
  await user.click(screen.getByRole('button', { name: 'Transfer' }));

  expect(screen.getByText('Amount must be positive')).toBeInTheDocument();
  expect(onSubmit).not.toHaveBeenCalled();
});

test('disables submit button while loading', async () => {
  const onSubmit = vi.fn().mockResolvedValue(new Promise(() => {}));
  const user = userEvent.setup();
  render(<TransferForm onSubmit={onSubmit} />);

  await user.type(screen.getByLabelText('Destination account ID'), 'acc-123');
  await user.type(screen.getByLabelText('Amount (cents)'), '1000');
  await user.click(screen.getByRole('button', { name: 'Transfer' }));

  expect(screen.getByRole('button', { name: 'Transfer' })).toBeDisabled();
});
```

Rules:

- Query by role, label, or text — never by class name or test id unless unavoidable.
- Drive interactions with `userEvent`, not `fireEvent`, for realistic events.
- Assert the user-visible outcome (validation message, called handler), not internals.
- Test each data screen's states: loading, empty, error, and success.
- Use `findBy*` / `waitFor` for async assertions; never a bare `setTimeout`.

## 3. Hook Tests

```tsx
// hooks/__tests__/useBalance.test.tsx
import { renderHook, waitFor } from '@testing-library/react';
import { useBalance } from '../useBalance';
import { QueryClientProvider } from '@tanstack/react-query';
import { QueryClient } from '@tanstack/react-query';

const wrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
};

test('exposes the balance once loaded', async () => {
  const { result } = renderHook(() => useBalance(), { wrapper });
  await waitFor(() => expect(result.current.isSuccess).toBe(true));
  expect(result.current.data?.balanceCents).toBe(100_000);
});

test('shows error state when API fails', async () => {
  const { result } = renderHook(() => useBalance(), { wrapper });
  await waitFor(() => expect(result.current.isError).toBe(true));
  expect(result.current.error).toBeInstanceOf(Error);
});
```

Rules:

- Give hooks the wrappers they need (QueryClient, Router, Auth provider).
- Assert on returned state, not on re-render counts.
- Reset QueryClient between tests to avoid cross-test pollution.

## 4. Mocking

Mock at the network boundary so query hooks exercise their real logic.

```ts
// src/test/server.ts
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

export const server = setupServer(
  http.get('/api/v1/accounts/:accountId/balance', () =>
    HttpResponse.json({
      accountId: '018b4a2e-1111-7000-8000-000000000001',
      balanceCents: 100_000,
      currencyCode: 'USD',
      updatedAt: '2026-10-15T20:00:00Z',
    }),
  ),
  http.post('/api/v1/transfers', () =>
    HttpResponse.json({ transferId: 'transfer-1', status: 'pending' }, { status: 202 }),
  ),
  http.get('/api/v1/accounts/:accountId/transactions', () =>
    HttpResponse.json({ transactions: [], nextCursor: null, hasMore: false }),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

Rules:

- Use MSW to intercept HTTP; do not mock the API client or the query library.
- Reset handlers between tests (`server.resetHandlers()`) in `afterEach`.
- Never mock what you are testing; mock the boundary around it.
- `onUnhandledRequest: 'error'` fails tests that make unexpected real HTTP calls.

## 5. Async, Timers & Queries

- Use `findBy*` / `waitFor` for async assertions; never a bare `setTimeout`.
- Use `vi.useFakeTimers()` for polling/intervals and advance explicitly.
- Assert the absence of an error with `queryBy*`, presence with `getBy*`/`findBy*`.
- For TanStack Query, use `waitFor` from `@tanstack/react-query` or React Testing Library's `waitFor`.
- Mock `matchMedia` if components use responsive breakpoints.

## 6. Coverage & Quality Gates

- Enforce the frontend threshold in CI; do not weaken it to pass.
- No flaky tests: a timing-dependent test is fixed or removed, not retried.
- Unit tests run in under 2 minutes (see `09-testing-strategy.md`).
- Coverage is enforced for lines (80%), functions (80%), branches (80%), statements (80%).

## 7. Code Style & Rules

- Behaviour over implementation; roles and labels over selectors.
- `userEvent` for interaction, `findBy`/`waitFor` for async.
- MSW at the network boundary; no mocking of the code under test.
- Coverage threshold enforced in CI; zero tolerance for flaky tests.
- Test file naming: `*.test.tsx` colocated with source or in `__tests__/`.
- Related: [`react.md`](../core/react.md) for component structure, [`tanstack-query.md`](../core/tanstack-query.md) for query wrappers, [`playwright.md`](./playwright.md) for end-to-end coverage.
