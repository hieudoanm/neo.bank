# NEO Bank — ESLint Best Practices

_Frontend · Tools · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [tools index](./README.md)._

## 1. Setup

ESLint lints TypeScript and React source files with strict rules and Tailwind class ordering.

```js
// eslint.config.mjs
import eslintPluginReact from 'eslint-plugin-react';
import eslintPluginReactHooks from 'eslint-plugin-react-hooks';
import eslintPluginTailwindcss from 'eslint-plugin-tailwindcss';
import typescriptEslint from '@typescript-eslint/eslint-plugin';
import typescriptParser from '@typescript-eslint/parser';

export default [
  {
    ignores: ['dist/', 'node_modules/', '*.config.*', 'playwright.config.ts'],
  },
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: typescriptParser,
      parserOptions: { ecmaVersion: 'latest', sourceType: 'module' },
    },
    plugins: {
      '@typescript-eslint': typescriptEslint,
      react: eslintPluginReact,
      'react-hooks': eslintPluginReactHooks,
      tailwindcss: eslintPluginTailwindcss,
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      'react/display-name': 'off',
      'tailwindcss/no-custom-classname': 'warn',
    },
  },
];
```

Rules:

- `no-explicit-any` is an error — strict TypeScript has no escape hatches.
- `react-hooks/exhaustive-deps` is a warning; fix missing deps, never disable with `// eslint-disable-next-line` without justification.
- Tailwind class names must match a theme token; arbitrary values like `bg-[#2563EB]` trigger `no-custom-classname`.
- Never disable rules silently; every `eslint-disable` needs an inline comment explaining why.

## 2. React Rules

Use the React plugin's recommended config for component correctness:

- Components are PascalCase; props interfaces are named `ComponentNameProps`.
- Event handlers are named `on<Event>` (`onSubmit`, `onChange`).
- Fragment shorthand (`<>...</>`) is preferred unless a `key` is needed.
- No inline function definitions in render output that break referential equality for memoised children.

## 3. CI

ESLint runs in CI on every pull request:

```yaml
- run: pnpm eslint src --ext .ts,.tsx
```

Exit non-zero on any error; warnings do not fail the build but must be addressed before merge.

## 4. Definition of Done (ESLint)

1. `pnpm eslint src --ext .ts,.tsx` exits with code 0 in CI.
2. No `eslint-disable` comments without justification.
3. New files pass lint on first commit; no deferred fixes.

## 5. Cross-references

- Prettier: [`./prettier.md`](./prettier.md) — formatting runs before lint.
- TypeScript: [`../core/typescript.md`](../core/typescript.md) — types are ESLint's input.
- Shared rules: [`../README.md`](../README.md#2-shared-rules) — arrow functions, named exports.

---
↑ Part of the [NEO Bank tech stack decision record](../../../08-tech-stack.md).
