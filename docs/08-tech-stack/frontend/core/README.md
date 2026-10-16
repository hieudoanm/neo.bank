# NEO Bank — Frontend: Core

The application framework: component model, routing, server state, language, and build.

Back to the [frontend index](../README.md) · [shared frontend rules](../README.md#2-shared-rules) · [tech stack decision record](../../../08-tech-stack.md).

## 1. Documents

| Document                                     | Covers                                          |
| -------------------------------------------- | ----------------------------------------------- |
| [`react.md`](./react.md)                     | Components, client state, UI states, a11y       |
| [`react-router.md`](./react-router.md)       | Route structure, protected routes, lazy loading |
| [`tanstack-query.md`](./tanstack-query.md)   | Server state, caching, mutations, polling       |
| [`typescript.md`](./typescript.md)           | Strict typing, API client, runtime validation   |
| [`vite.md`](./vite.md)                       | Build config, env vars, aliases, code splitting |

## 2. Core Rules

1. **One component, one concern.** Presentational components take props; domain logic lives in hooks and queries (see [`react.md`](./react.md)).
2. **Server state never lives in component state.** Balances, transfers, and transactions are fetched and cached by [`tanstack-query.md`](./tanstack-query.md).
3. **Routes are declared centrally** and protected at the route boundary (see [`react-router.md`](./react-router.md)).
4. **Strict TypeScript at every boundary**, no `any`, runtime validation for API payloads (see [`typescript.md`](./typescript.md)).
5. **Split at the route level**, not per component (see [`vite.md`](./vite.md)).

## 3. Definition of Done (Core)

1. Components render every documented UI state without touching the network directly.
2. `tsc --noEmit` and the linter pass with no suppressions added.
3. Every route is lazy-loaded and guarded where it needs authentication.
4. Every API payload is validated at the client boundary before use.

---

↑ Part of the [NEO Bank tech stack decision record](../../../08-tech-stack.md).
