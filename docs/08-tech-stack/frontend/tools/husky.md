# NEO Bank — Husky Best Practices

_Frontend · Tools · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md)._

Husky is the Git-hook manager at the repository root (see [tech stack §13.2](../../../08-tech-stack.md#13-package-manager--git-hooks)). Hooks live under `.husky/` and run on commit/push.

## 1. Setup

```sh
pnpm add -D husky lint-staged   # at repo root
pnpm exec husky init            # scaffolds .husky/pre-commit
```

The `prepare` script in the root `package.json` (`"prepare": "husky"`) runs `husky` on `pnpm install`, installing hooks via the lifecycle. The `core.hooksPath` Git config points at `.husky`.

## 2. Hooks

| Hook         | Purpose                                         |
| ------------ | ----------------------------------------------- |
| `pre-commit` | Typecheck the frontend; keep it fast and quiet. |
| `pre-push`   | Run the full test suite plus backend vet.       |

Rules:

- Hooks are versioned in `.husky/`; never put hooks in `.git/hooks` (not shared, not reviewable).
- Bypass with `git commit --no-verify` / `git push --no-verify` only in emergencies; never as a convenience.
- No hook may weaken or skip a quality gate (no `--no-verify` wrappers, no disabling checks).
- Keep hooks fast: seconds, not minutes. Heavy checks belong in CI.

## 3. lint-staged

Compose lint-staged for staged-file-only checks to keep `pre-commit` fast:

```json
{
  "lint-staged": {
    "packages/frontend/app/src/**/*.{ts,tsx}": [
      "pnpm -C packages/frontend/app exec tsc --noEmit --skipLibCheck"
    ]
  }
}
```

Rules:

- lint-staged runs only on staged files; CI re-runs everything against the whole tree.
- A hook failure blocks the commit; fix the cause rather than bypassing.
- New hooks must be documented here and added to `assets/` linkage in CI.
