# NEO Bank — Frontend: Testing

The two frontend test layers: component/hook tests with Vitest, and end-to-end journeys with Playwright.

Back to the [frontend index](../README.md) · [shared frontend rules](../README.md#2-shared-rules) · [tech stack decision record](../../../08-tech-stack.md).

## 1. Documents

| Document                           | Covers                                  |
| ---------------------------------- | --------------------------------------- |
| [`vitest.md`](./vitest.md)         | Component/hook tests, mocking, coverage |
| [`playwright.md`](./playwright.md) | End-to-end flows, auth state, CI        |

## 2. Testing Rules

1. **Assert behaviour, not implementation** — no test reaches into component internals (see [`vitest.md`](./vitest.md)).
2. **Mock at the network boundary only** (MSW); components under test stay real.
3. **Cover the critical journeys end to end**: sign in, view balance, transfer, confirm status (see [`playwright.md`](./playwright.md)).
4. **Assume duplicate delivery** in any test that touches a transfer; idempotency is part of the behaviour.
5. **Never weaken a check to make CI pass.**

## 3. Layer Responsibilities

| Layer      | Tool         | Scope                                      | Speed target |
| ---------- | ------------ | ------------------------------------------ | ------------ |
| Component  | Vitest + RTL | Individual components, hooks, UI states    | < 2 minutes  |
| Hook       | Vitest + RTL | Custom hooks (query, mutation, auth logic) | < 2 minutes  |
| End-to-end | Playwright   | Critical user journeys through the real UI | < 10 minutes |

## 4. Mocking Strategy

- **Network boundary:** MSW intercepts HTTP at the transport layer; the API client and query hooks exercise their real logic.
- **No internal mocks:** do not mock the code under test; mock what lies outside it.
- **Unhandled requests fail:** `onUnhandledRequest: 'error'` fails tests that make unexpected real HTTP calls.

## 5. Test Data

- Synthetic accounts and users only; never real financial data (`09-testing-strategy.md`).
- Deterministic fixtures: same seed, same result — no random data.
- Unique emails per test run to avoid conflicts.
- Test isolation: each spec cleans up or uses uniquely-scoped data for parallel runs.

## 6. Definition of Done (Testing)

1. Component tests cover loading, empty, validation-error, permission-error, network-error, and unexpected-error states.
2. E2E tests run against the real API in CI and gate the deploy.
3. No test depends on execution order or sleeps; Playwright waits are auto-waiting, not fixed delays.
4. Coverage thresholds enforced in CI: lines 80%, functions 80%, branches 80%, statements 80%.
5. Zero tolerance for flaky tests — fix or disable, never retry into green.

## 7. Cross-references

- Testing strategy: [`../../../09-testing-strategy.md`](../../../09-testing-strategy.md)
- Component structure: [`../core/react.md`](../core/react.md)
- Query layer: [`../core/tanstack-query.md`](../core/tanstack-query.md)
- Vitest practices: [`./vitest.md`](./vitest.md)
- Playwright practices: [`./playwright.md`](./playwright.md)
- Planned test: `test_frontend_testing_index` (verifies testing documents exist and cross-reference correctly)

---

↑ Part of the [NEO Bank tech stack decision record](../../../08-tech-stack.md).