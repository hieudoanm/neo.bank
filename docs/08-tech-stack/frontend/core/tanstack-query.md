# NEO Bank — TanStack Query Best Practices

_Frontend · Core · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [core index](./README.md)._

## 1. Server State vs Client State

TanStack Query owns **server state**: anything that lives in the API — balance, transfers, transactions, transfer status. React state (context, `useState`, `useReducer`) owns **client state**: the current user, UI toggles, form drafts.

Rules:

- Never copy server data into `useState` or context; read it from the query cache.
- Never call `fetch` from a component; go through the query/mutation layer.
- If a value comes from the backend, it has a query key.

## 2. QueryClient Setup

```tsx
// lib/queryClient.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
});
```

Rules:

- One `QueryClient` per app, provided once at the root.
- Financial reads are not aggressively cached: keep `staleTime` short (balance ~10s, lists ~30s).
- Retries are bounded (2) so a failing backend surfaces quickly instead of hanging.

## 3. Query Keys

Keys are the cache identity; they must be stable and structured.

```ts
// lib/api/keys.ts
export const keys = {
  balance: ['balance'] as const,
  transfers: ['transfers'] as const,
  transfer: (id: string) => ['transfers', id] as const,
  transactions: ['transactions'] as const,
};
```

Rules:

- One module owns every key; never inline string arrays in components.
- List keys are prefixes of detail keys (`['transfers']` vs `['transfers', id]`) so invalidation cascades.
- Parameterised keys use a factory function, not string concatenation.

## 4. Queries

```ts
// hooks/useBalance.ts
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { keys } from '@/lib/api/keys';

export const useBalance = () => {
  return useQuery({
    queryKey: keys.balance,
    queryFn: () => api.accounts.getBalance(),
    staleTime: 10_000,
    refetchInterval: 30_000, // poll for balance changes
  });
};
```

Rules:

- `queryFn` returns the typed API response; the hook exposes `data`, `isLoading`, `isError`.
- Give every query a sensible `staleTime`; polling uses `refetchInterval`.
- Let the hook own the query; components own the rendering of its states.

## 5. Mutations & Invalidation

```ts
// hooks/useCreateTransfer.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { keys } from '@/lib/api/keys';

export const useCreateTransfer = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: api.transfers.create,
    onSuccess: (transfer) => {
      queryClient.invalidateQueries({ queryKey: keys.balance });
      queryClient.invalidateQueries({ queryKey: keys.transactions });
      queryClient.setQueryData(keys.transfer(transfer.id), transfer);
    },
  });
};
```

Rules:

- After a mutation, invalidate the affected keys; do not hand-edit unrelated caches.
- Seed the new resource into the cache (`setQueryData`) when the response already contains it.
- Never show success before `onSuccess` runs — the server is the authority.
- Send an idempotency key on financial mutations; a retry must not double-charge.

## 6. Infinite Queries

Transaction history uses cursor pagination.

```ts
// hooks/useTransactions.ts
import { useInfiniteQuery } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { keys } from '@/lib/api/keys';

export const useTransactions = () => {
  return useInfiniteQuery({
    queryKey: keys.transactions,
    queryFn: ({ pageParam }) => api.transactions.list(pageParam),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
  });
};
```

Rules:

- The cursor is the opaque `nextCursor` from the API; never construct it client-side.
- `getNextPageParam` returns `undefined` when there is no next page.
- Flatten pages for rendering; keep ordering deterministic.

## 7. Polling

Transfer status is asynchronous; poll until it reaches a terminal state.

```ts
// hooks/useTransferStatus.ts
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { keys } from '@/lib/api/keys';

export const useTransferStatus = (transferId: string) => {
  return useQuery({
    queryKey: keys.transfer(transferId),
    queryFn: () => api.transfers.getStatus(transferId),
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status === 'completed' || status === 'failed' ? false : 2_000;
    },
  });
};
```

Rules:

- Poll with `refetchInterval`; stop when the status is terminal (`completed`, `failed`, `reversed`).
- Keep the interval short (2s) only while a transfer is in flight.
- A future WebSocket/SSE push replaces polling without changing the hook's shape.

## 8. Error Handling & Retries

- Normalise API failures into `ApiError` in the client (`typescript.md`); components branch on it.
- Retry only idempotent reads; mutations use an explicit idempotency key instead of blind retries.
- Do not retry `4xx`; surface validation and permission errors immediately.
- Every data screen renders `isLoading`, `isError`, and empty states (see `react.md`).

## 9. Optimistic Updates

- Use optimistic updates sparingly and never for balance or transfer confirmation.
- Only optimistically update reversible UI (for example, a filter selection), with a rollback in `onError`.
- Financial truth always comes from a server response; show "pending" until confirmed.

## 10. Testing

```tsx
// hooks/__tests__/useBalance.test.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';

const wrapper = ({ children }: { children: React.ReactNode }) => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
};

test('returns the account balance', async () => {
  const { result } = renderHook(() => useBalance(), { wrapper });
  await waitFor(() => expect(result.current.isSuccess).toBe(true));
  expect(result.current.data?.balanceCents).toBe(100_000);
});
```

Rules:

- Wrap hook tests in a fresh `QueryClientProvider`; disable retries.
- Mock the API layer (MSW) at the network boundary, not the hook internals.
- Assert on query state and eventual data, not on cache internals.

## 11. Code Style & Rules

- TanStack Query owns all server state; React state owns client state only.
- Hooks are arrow functions (`export const useFoo = () => { ... }`); no `function` declarations.
- Query keys live in one module; lists are prefixes of details.
- Mutations invalidate affected keys; financial mutations carry an idempotency key.
- Pagination uses the server cursor; polling stops at terminal status.
- No optimistic updates for balances or transfer confirmation.
- Related: [`react.md`](./react.md) for components, [`typescript.md`](./typescript.md) for the typed API client, [`vitest.md`](../testing/vitest.md) for tests.
