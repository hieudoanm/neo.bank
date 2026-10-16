# NEO Bank — Frontend Best Practices

This is the entry point for the frontend best-practice documents. The documents are grouped by concern — **core** (the application framework), **styles** (the token layer), and **testing** (the two test layers). Each group has its own index with group-specific rules; this README is the top-level index and the shared rules that apply across all of them.

The **why** behind these choices (trade-offs and rejected alternatives) lives in the [tech stack decision record](../../08-tech-stack.md). These documents are the **how**.

## 1. Groups

| Group                              | Covers                                                           |
| ---------------------------------- | ---------------------------------------------------------------- |
| [`core/`](./core/README.md)        | React, routing, server state, TypeScript, the Vite build         |
| [`styles/`](./styles/README.md)    | Tailwind token layer and DaisyUI component classes               |
| [`testing/`](./testing/README.md)  | Vitest component/hook tests and Playwright end-to-end flows      |

### 1.1 Core ([`core/README.md`](./core/README.md))

| Document                                   | Covers                                                |
| ------------------------------------------ | ----------------------------------------------------- |
| [`react.md`](./core/react.md)              | Components, client state, UI states, accessibility    |
| [`react-router.md`](./core/react-router.md) | Route structure, protected routes, lazy loading      |
| [`tanstack-query.md`](./core/tanstack-query.md) | Server state, caching, mutations, polling        |
| [`typescript.md`](./core/typescript.md)    | Strict typing, API client, runtime validation         |
| [`vite.md`](./core/vite.md)                | Build config, env vars, aliases, code splitting       |

### 1.2 Styles ([`styles/README.md`](./styles/README.md))

| Document                                   | Covers                                                |
| ------------------------------------------ | ----------------------------------------------------- |
| [`tailwindcss.md`](./styles/tailwindcss.md) | Design tokens, utility conventions, responsive, a11y |
| [`daisyui.md`](./styles/daisyui.md)        | Theme mapping, component classes, shared UI wrappers  |

### 1.3 Testing ([`testing/README.md`](./testing/README.md))

| Document                                   | Covers                                                |
| ------------------------------------------ | ----------------------------------------------------- |
| [`vitest.md`](./testing/vitest.md)         | Component/hook tests, mocking, coverage               |
| [`playwright.md`](./testing/playwright.md) | End-to-end flows, auth state, CI                      |

### 1.4 Tools ([`tools/README.md`](./tools/README.md))

| Document                     | Covers                    |
| ---------------------------- | ------------------------- |
| [`eslint.md`](./tools/eslint.md)  | Lint rules for TS/React   |
| [`prettier.md`](./tools/prettier.md) | Format rules for all frontend files |

## 2. Shared Rules

These apply to every document in every group and to the group indices themselves:

1. **Arrow functions only.** Declare functions with `const` and an arrow function. No `function` declarations.

   ```tsx
   // Good
   export const TransferForm = ({ onSubmit }: TransferFormProps) => { /* ... */ };

   // Bad
   export function TransferForm({ onSubmit }: TransferFormProps) { /* ... */ }
   ```

2. **Named exports** for components and utilities; **default exports** only for pages and layouts.
3. **`const` by default**; never `var`.
4. **Strict TypeScript**, no `any` (see [`typescript.md`](./core/typescript.md)).
5. **Design tokens are the source of truth** ([`tailwindcss.md`](./styles/tailwindcss.md), [`daisyui.md`](./styles/daisyui.md)); no raw hex or arbitrary values.
6. **Server state via TanStack Query** ([`tanstack-query.md`](./core/tanstack-query.md)); never call the backend with raw secrets.
7. **Every data screen** implements loading, empty, validation-error, permission-error, network-error, and unexpected-error states.
8. **Never show success before the server confirms it.**
9. **Colocate** a component with its test; **separate domain logic** from presentational components.
10. **Behaviour, not implementation,** in tests (see [`vitest.md`](./testing/vitest.md), [`playwright.md`](./testing/playwright.md)).

## 3. Toolchain

| Area           | Tool                                              |
| -------------- | ------------------------------------------------- |
| Language       | TypeScript (strict)                               |
| Framework      | React 18                                          |
| Build          | Vite                                              |
| Data           | TanStack Query                                    |
| Routing        | React Router                                      |
| Styling        | Tailwind CSS + DaisyUI                            |
| Lint / format  | ESLint, Prettier (`prettier-plugin-tailwindcss`)  |
| Types          | `tsc --noEmit`                                    |
| Tests          | Vitest + Testing Library, Playwright (E2E)        |

Run the frontend checks through the root Makefile or CI pipeline. Never weaken a check to make CI pass.

## 4. Definition of Done (Frontend)

1. Builds, typechecks (`tsc --noEmit`), lints, and formats cleanly in CI.
2. All functions are arrow functions; no `function` declarations.
3. Every data screen implements its full set of UI states.
4. No raw hex or arbitrary utility values; tokens come from the theme.
5. Server access goes through TanStack Query; no direct `fetch` in components.
6. Component and E2E tests cover the critical journeys at the network boundary.

## 5. Cross-references

- Tech stack decision: [`../../08-tech-stack.md`](../../08-tech-stack.md)
- UI/UX design: [`../../07-ui-ux-design.md`](../../07-ui-ux-design.md)
- Requirements: [`../../01-requirements.md`](../../01-requirements.md)
- Frontend conventions: `AGENTS.md`
- Planned test: `test_frontend_index` (verifies all frontend documents exist and cross-reference correctly)

---

↑ Part of the [NEO Bank tech stack decision record](../../08-tech-stack.md).