# NEO Bank — TypeScript Best Practices

_Frontend · Core · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [core index](./README.md)._

## 1. Type Safety

### Strict Mode

```json
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "exactOptionalPropertyTypes": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "forceConsistentCasingInFileNames": true,
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "module": "ESNext",
    "allowSyntheticDefaultImports": true,
    "esModuleInterop": true,
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist", "build"]
}
```

`strict` enables all strict type checks. `noUncheckedIndexedAccess` catches undefined array access. `noUnusedLocals`/`noUnusedParameters` keep the codebase clean. `skipLibCheck` avoids noise from third-party `.d.ts` files. `paths` maps `@/*` to `./src/*` (also declared in `vite.config.ts`).

### Avoid `any`

- Use `unknown` for values of uncertain type, then narrow with type guards.
- Use generics for reusable functions/components.
- Document exceptions with `// @ts-expect-error` and a comment explaining why.
- Never use `as any` to silence the compiler.

```typescript
// Good: unknown + type guard
function parseBody(body: unknown): Record<string, string> | null {
  if (typeof body === 'object' && body !== null) {
    return body as Record<string, string>;
  }
  return null;
}

// Bad: any defeats type checking
function bad(body: any): string {
  return body.mightNotExist;
}
```

### Discriminated Unions for State

State is modeled as discriminated unions so `switch` exhaustiveness is checked:

```typescript
type TransferState =
  | { status: 'pending'; transferId: string }
  | { status: 'processing'; transferId: string; startedAt: string }
  | { status: 'completed'; transferId: string; completedAt: string }
  | { status: 'failed'; transferId: string; error: string }
  | { status: 'reversed'; transferId: string; reversedAt: string };

const describeTransfer = (state: TransferState): string => {
  switch (state.status) {
    case 'pending': return `Pending: ${state.transferId}`;
    case 'processing': return `Processing since ${state.startedAt}`;
    case 'completed': return `Completed at ${state.completedAt}`;
    case 'failed': return `Failed: ${state.error}`;
    case 'reversed': return `Reversed at ${state.reversedAt}`;
    default:
      const _exhaustive: never = state;
      return `Unknown state: ${_exhaustive}`;
  }
};
```

The `default` branch with `_exhaustive: never` ensures that adding a new status without updating `describeTransfer` is a type error.

### Branded Types for IDs

Account IDs, transfer IDs, and user IDs are branded types to prevent accidental mixing:

```typescript
type AccountId = string & { readonly __brand: unique symbol };
type TransferId = string & { readonly __brand: unique symbol };
type UserId = string & { readonly __brand: unique symbol };

const createAccountId = (id: string): AccountId => id as AccountId;
const createTransferId = (id: string): TransferId => id as TransferId;
// createAccountId(createTransferId('abc')) // Type error — different brands
```

## 2. API Layer

### Centralised API Client

All API calls go through a single client wrapper so auth headers, error handling, and base URL are centralised:

```typescript
// lib/api/client.ts
import { fetchWrapper } from './fetchWrapper';

export const api = {
  auth: {
    register: (data: RegisterRequest) =>
      fetchWrapper.post<AuthResponse>('/register', data),
    login: (data: LoginRequest) =>
      fetchWrapper.post<AuthResponse>('/login', data),
    refresh: (refreshToken: string) =>
      fetchWrapper.post<TokenResponse>('/refresh', { refreshToken }),
    logout: (refreshToken: string) =>
      fetchWrapper.post<{ ok: boolean }>('/logout', { refreshToken }),
  },
  accounts: {
    getBalance: (accountId: AccountId) =>
      fetchWrapper.get<BalanceResponse>(`/accounts/${accountId}/balance`),
  },
  transfers: {
    create: (data: CreateTransferRequest) =>
      fetchWrapper.post<TransferResponse>('/transfers', data),
    getStatus: (transferId: TransferId) =>
      fetchWrapper.get<TransferStatusResponse>(`/transfers/${transferId}`),
  },
  transactions: {
    list: (accountId: AccountId, cursor?: string) =>
      fetchWrapper.get<TransactionListResponse>(`/accounts/${accountId}/transactions`, { cursor }),
  },
};
```

### Error Handling

Every API failure is normalised into a single `ApiError` shape at the boundary:

```typescript
// lib/api/errors.ts
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    public readonly message: string,
    public readonly details?: Record<string, string[]>,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}
```

Rules:

- Normalise every failure into a single `ApiError` shape at the API boundary.
- Callers branch on `status`/`code`, never on raw `fetch` rejections.
- Never surface raw server messages to the user; map them to safe copy.
- Preserve `requestId` so a support ticket can be traced to a log line.
- TanStack Query catches `ApiError` and exposes it via `error` on the query result.

## 3. Runtime Validation

TypeScript types are erased at runtime. Every external boundary (API responses, form input) is validated with Zod:

```typescript
// lib/api/schemas/transfer.ts
import { z } from 'zod';

export const transferSchema = z.object({
  destinationAccountId: z.string().uuid('Invalid account ID'),
  amountCents: z.number().int().positive('Amount must be positive'),
});

export type CreateTransferRequest = z.infer<typeof transferSchema>;
// The Idempotency-Key is sent as a request header (required for transfers),
// not as a body field — see docs/06-api-specification.md §7.
```

Rules:

- Validate every external boundary (API responses, form input) with Zod.
- Derive the type from the schema: `type X = z.infer<typeof schema>`.
- Keep each schema beside the type it guards; one source of truth per shape.
- Reject unknown fields on writes, matching the contract in `06-api-specification.md`.
- Model state with discriminated unions and validate with `z.discriminatedUnion`.
- Types alone are not enough — the schema is what actually protects the boundary.

## 4. TypeScript & API Consistency

Frontend types mirror backend contracts so mismatches are caught at compile time:

- `AccountId`, `TransferId`, `UserId` — branded types matching backend UUID v7.
- `BalanceResponse` — `{ accountId, balanceCents, currencyCode, updatedAt }`.
- `TransferResponse` — `{ transferId, status, amountCents, sourceAccountId, destinationAccountId, createdAt }`.
- `TransactionListResponse` — `{ transactions: Transaction[], nextCursor?: string, hasMore: boolean }`.

If the backend adds a field, the frontend type is updated before the UI consumes it.

## 5. Related Documents

| Concern | Document |
|---------|----------|
| Components, UI states | [`react.md`](./react.md) |
| Routing | [`react-router.md`](./react-router.md) |
| Server state / fetching | [`tanstack-query.md`](./tanstack-query.md) |
| Build config, env vars, aliases | [`vite.md`](./vite.md) |
| Unit & component tests | [`vitest.md`](../testing/vitest.md) |
| End-to-end tests | [`playwright.md`](../testing/playwright.md) |

## 6. Code Style

- **Arrow functions** for all declarations
- **`const`** by default; no `var`, no `let` unless reassigned
- **Named exports** for utilities/components; default exports for pages/layouts
- **Colocate tests** in `__tests__/` or `*.test.tsx`
- **Separate domain logic** from visual components
- **Use Zod** for runtime validation matching TypeScript types
- **No `any`** — use `unknown` and narrow with type guards
- **Discriminated unions** for state machines (transfer status, UI state)
- **Branded types** for IDs (AccountId, TransferId, UserId)
