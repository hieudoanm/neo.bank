# NEO Bank — pnpm Best Practices

_Frontend · Tools · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md)._

pnpm is the root-level package manager for the repository (see [tech stack §13.1](../../../08-tech-stack.md#13-package-manager--git-hooks)). It manages the frontend app and the root tooling (husky, lint-staged) through a content-addressable store.

## 1. Commands

| Task           | Command                                                        |
| -------------- | -------------------------------------------------------------- |
| Install (root) | `pnpm install`                                                 |
| Frontend dev   | `pnpm -C packages/frontend/app dev`                            |
| Frontend build | `pnpm -C packages/frontend/app build`                          |
| Typecheck      | `pnpm -C packages/frontend/app typecheck`                      |
| Tests          | `pnpm -C packages/frontend/app test`                           |
| Add dependency | `pnpm -C packages/frontend/app add <pkg>` (runtime) / `add -D` |
| Add at root    | `pnpm add -D <pkg>` (run from repo root)                       |

Rules:

- Always use the lockfile; run `pnpm install --frozen-lockfile` in CI.
- Never edit `node_modules` or the store manually; let pnpm own it.
- Prefer workspace/filter usage (`pnpm -C <dir>`) over `cd` + pnpm for scripts.
- Strict dependency resolution surfaces missing/hoisted dependencies as errors; do not add `shamefully-hoist` to silence them — fix the import instead.

## 2. Workspaces

A root `pnpm-workspace.yaml` (if added) declares `packages/*`; the frontend app is `packages/frontend/app`. Until then, the frontend is a standalone package and dev tooling lives at the root via `pnpm add -D` (run from repo root).

## 3. CI Caching

Cache `~/.local/share/pnpm/store` in GitHub Actions for fast installs; pin pnpm via `pnpm/action-setup` using the version in `packageManager`.
