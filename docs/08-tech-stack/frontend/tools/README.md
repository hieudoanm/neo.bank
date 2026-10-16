# NEO Bank — Frontend Tools

Linting and formatting tooling for the frontend: ESLint and Prettier.

Back to the [frontend index](../README.md) · [shared frontend rules](../README.md#2-shared-rules) · [tech stack decision record](../../../08-tech-stack.md).

## 1. Documents

| Document                       | Covers                                |
| ------------------------------ | ------------------------------------- |
| [`eslint.md`](./eslint.md)     | Lint rules for TS/React               |
| [`prettier.md`](./prettier.md) | Format rules for all frontend files   |
| [`pnpm.md`](./pnpm.md)         | Package manager, workspaces, installs |
| [`husky.md`](./husky.md)       | Git hooks, lint-staged, pre-commit    |

## 2. Toolchain Rules

1. **Format before lint** — `pnpm format` then `pnpm eslint`; never the reverse.
2. **Prettier owns formatting**; ESLint owns correctness rules; no overlap.
3. **`prettier-plugin-tailwindcss`** sorts Tailwind classes; ESLint's `tailwindcss` plugin warns on arbitrary values.
4. Both run in CI and must exit code 0.

## 3. Definition of Done (Tools)

1. `pnpm format:check` exits with code 0 in CI.
2. `pnpm eslint src --ext .ts,.tsx` exits with code 0 in CI.
3. Format-on-save is enabled in the editor.

## 4. Cross-references

- ESLint practices: [`./eslint.md`](./eslint.md)
- Prettier practices: [`./prettier.md`](./prettier.md)
- pnpm practices: [`./pnpm.md`](./pnpm.md)
- Husky practices: [`./husky.md`](./husky.md)
- TypeScript types: [`../core/typescript.md`](../core/typescript.md)
- Tailwind classes: [`../styles/tailwindcss.md`](../styles/tailwindcss.md)

---
↑ Part of the [NEO Bank tech stack decision record](../../../08-tech-stack.md).
