# NEO Bank — Tailwind CSS Best Practices

_Frontend · Styles · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [styles index](./README.md)._

## 1. Configuration & Design Tokens

Design tokens come from `docs/07-ui-ux-design.md`. They live once, in the Tailwind theme, and are consumed as utilities — never re-declared per component.

```js
// tailwind.config.js
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#2563EB',
          hover: '#1D4ED8',
          active: '#1E40AF',
        },
        success: { DEFAULT: '#10B981', bg: '#D1FAE5' },
        error: { DEFAULT: '#EF4444', bg: '#FEE2E2' },
        warning: { DEFAULT: '#F59E0B', bg: '#FEF3C7' },
        info: { DEFAULT: '#3B82F6', bg: '#DBEAFE' },
        text: {
          primary: '#111827',
          secondary: '#6B7280',
          tertiary: '#9CA3AF',
        },
        border: { DEFAULT: '#E5E7EB', light: '#F3F4F6' },
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      borderRadius: {
        md: '0.375rem',
        lg: '0.5rem',
        xl: '0.75rem',
      },
    },
  },
  plugins: [],
};
```

Rules:

- Token names match `07-ui-ux-design.md`. A new colour, size, or radius is added to the theme first, then documented.
- `content` globs cover every template that contains class names. Missing globs mean missing styles in production.
- Extend the default theme rather than replacing it; keep Tailwind's defaults as the baseline.

## 2. Utility-First Conventions

Write utilities in markup; extract a component before you extract CSS.

```tsx
// components/features/BalanceCard.tsx
export const BalanceCard = ({ balance }: { balance: string }) => {
  return (
    <section className="rounded-xl border border-border bg-white p-6 shadow-sm transition-shadow hover:shadow-md">
      <h2 className="text-sm font-medium text-text-secondary">Current balance</h2>
      <p className="mt-2 text-3xl font-bold text-text-primary">{balance}</p>
    </section>
  );
};
```

Rules:

- Compose utilities directly; only reach for `@apply` when a pattern recurs and cannot be a component.
- Use semantic tokens (`bg-primary`, `text-text-secondary`, `border-border`) — never raw hex or arbitrary values like `bg-[#3f8cff]`.
- Keep a fixed scale for spacing (`p-4`, `gap-6`), radius (`rounded-xl` for cards, `rounded-full` for buttons and inputs), and shadows (`shadow-sm`).
- States are explicit: `hover:`, `focus-visible:`, `active:`, `disabled:`, `aria-[...]:`.

## 3. Component Extraction (not Class Soup)

When a combination repeats, extract a React component with a `className` escape hatch — do not create a CSS class per pattern.

```tsx
// components/ui/Button.tsx
import type { ButtonHTMLAttributes } from 'react';

type Variant = 'primary' | 'secondary' | 'danger';

const variants: Record<Variant, string> = {
  primary: 'bg-primary text-white hover:bg-primary-hover active:bg-primary-active',
  secondary: 'bg-white text-primary border border-border hover:bg-background-secondary',
  danger: 'bg-error text-white hover:opacity-90',
};

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
}

export const Button = ({ variant = 'primary', className = '', ...props }: ButtonProps) => {
  return (
    <button
      className={`inline-flex items-center justify-center gap-2 rounded-full px-6 py-3 font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 disabled:cursor-not-allowed disabled:opacity-50 ${variants[variant]} ${className}`}
      {...props}
    />
  );
};
```

Rules:

- Variant maps live in the component, not scattered across call sites.
- Always accept and append a `className` prop for layout tweaks at the call site.
- One component per file; combine with `clsx`/`tailwind-merge` if conditional classes grow.

## 4. Responsive Design

Mobile-first: unprefixed utilities are the smallest screen; add breakpoints upward.

```tsx
<div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
  <BalanceCard balance={balance} />
  <QuickActions />
  <RecentTransactions />
</div>
```

Breakpoints follow `07-ui-ux-design.md`: `sm 640`, `md 768`, `lg 1024`, `xl 1280`, `2xl 1536`.

Rules:

- Design single-column first; add columns at `md`/`lg`.
- Use container-level layout utilities (`grid`, `flex`, `gap`) rather than margins between siblings.
- Touch targets are at least 44px on touch breakpoints (`min-h-11`).
- Test at 375px, 768px, and 1440px before claiming a screen is responsive.

## 5. Theming & Consistency

- One light theme for banking; no dark mode unless `07-ui-ux-design.md` adds it.
- Semantic tokens only: `text-text-primary` for body text, `text-text-secondary` for helpers.
- Financial semantics are consistent everywhere: `success` for credit, `error` for debit/failure, `warning` for pending.
- Never encode meaning in colour alone — pair with text or an icon (accessibility).

## 6. Accessibility

- Maintain WCAG 2.1 AA contrast: normal text ≥ 4.5:1, large text ≥ 3:1, UI components ≥ 3:1.
- Visible focus on every interactive element: `focus-visible:ring-2 focus-visible:ring-primary/20`.
- Do not remove outlines without a replacement.
- Use utilities on semantic HTML (`<button>`, `<label>`, `<main>`), not on `<div>` click targets.
- Honour reduced motion: `motion-safe:` for transitions, `motion-reduce:transition-none` where needed.

## 7. Performance

- Purge is automatic via `content`; keep it accurate so unused utilities are dropped.
- **Never build class names dynamically** (`text-${color}`, `bg-${variant}`) — Tailwind cannot see them and they will not be generated. Use full class strings in a lookup map instead.

```tsx
// Good: full class names are statically visible
const badgeClass: Record<TransferStatus, string> = {
  pending: 'bg-warning-bg text-warning',
  completed: 'bg-success-bg text-success',
  failed: 'bg-error-bg text-error',
  reversed: 'bg-background-secondary text-text-secondary',
};
```

- Prefer a few component classes over long inline chains repeated hundreds of times.
- Enable the Vite bundle analyser in CI and watch CSS size.

## 8. Formatting & Class Order

- Use `prettier-plugin-tailwindcss` so classes are sorted consistently (layout → box → typography → colour → state).
- One `pnpm format` / `pnpm format:check` gate for the repo; do not hand-order classes.
- Long `className` strings wrap across lines; group by purpose with a comment only when non-obvious.

## 9. Testing

- Component tests assert rendered classes only where they encode behaviour (disabled, error, selected), not for pixel layout.
- Use Testing Library queries by role and label; do not query by Tailwind class.
- Snapshot tests are avoided for styling; a design change is not a test failure.

## 10. Code Style & Rules

- Semantic tokens only; no raw hex, no arbitrary `[...]` values.
- Components are arrow functions (`export const Button = (...) => { ... }`); no `function` declarations.
- Mobile-first; add breakpoints upward.
- Extract a component (with `className` passthrough) before extracting CSS.
- Never construct class names dynamically.
- Accessibility utilities (`focus-visible`, `motion-safe`, contrast) are part of the component, not an afterthought.
- `07-ui-ux-design.md` is the token source of truth; this file is the how.
