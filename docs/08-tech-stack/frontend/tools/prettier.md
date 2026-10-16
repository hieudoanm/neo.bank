# NEO Bank — Prettier Best Practices

_Frontend · Tools · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [tools index](./README.md)._

## 1. Setup

Prettier formats TypeScript, React, Tailwind, JSON, Markdown, and YAML with the Tailwind CSS plugin for class ordering.

```json
// .prettierrc
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "arrowParens": "always",
  "printWidth": 100,
  "plugins": ["prettier-plugin-tailwindcss"]
}
```

Rules:

- `printWidth: 100` — wrap long `className` strings across lines; group by purpose.
- `prettier-plugin-tailwindcss` sorts Tailwind classes (layout → box → typography → colour → state).
- No hand-ordered classes; run `pnpm format` before committing.
- Format on save is enabled in the editor; CI enforces with `pnpm format:check`.

## 2. Tailwind Class Ordering

Classes follow the `prettier-plugin-tailwindcss` order automatically:

```tsx
// Good — sorted by plugin
<section className="mx-auto max-w-6xl px-4 py-8 lg:px-8">

// Bad — hand-ordered, not plugin-sorted
<section className="max-w-6xl mx-auto lg:px-8 px-4 py-8">
```

## 3. CI

Prettier runs in CI on every pull request:

```yaml
- run: pnpm format:check
```

Any formatting difference fails the build; fix with `pnpm format` and re-commit.

## 4. Definition of Done (Prettier)

1. `pnpm format:check` exits with code 0 in CI.
2. `prettier-plugin-tailwindcss` is the only formatting plugin; no custom overrides.
3. `pnpm format` is run before every commit; no unstaged formatting drift.

## 5. Cross-references

- ESLint: [`./eslint.md`](./eslint.md) — lint runs after format.
- Tailwind: [`../styles/tailwindcss.md`](../styles/tailwindcss.md) — class conventions.
- Shared rules: [`../README.md`](../README.md#2-shared-rules).

---
↑ Part of the [NEO Bank tech stack decision record](../../../08-tech-stack.md).
