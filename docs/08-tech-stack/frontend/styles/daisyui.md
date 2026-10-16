# NEO Bank — DaisyUI Best Practices

_Frontend · Styles · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [styles index](./README.md)._

## 1. Role of DaisyUI

DaisyUI sits on top of Tailwind and provides semantic component classes (`btn`, `input`, `card`, `badge`, `alert`). It removes long, repeated utility chains from JSX and keeps accessibility defaults in one place.

The rule: **DaisyUI first, Tailwind utilities second, custom CSS last.** Reach for raw utilities only to adjust layout or spacing that DaisyUI does not express.

DaisyUI does not replace the tokens in `07-ui-ux-design.md`; it is configured from them. See `docs/08-tech-stack/frontend/styles/tailwindcss.md` for the token layer.

## 2. Theme Configuration

Map DaisyUI's semantic names to the NEO Bank tokens so `btn-primary`, `badge-success`, and `text-error` resolve to the documented colours.

```js
// tailwind.config.js
import daisyui from 'daisyui';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  plugins: [daisyui],
  daisyui: {
    logs: false,
    themes: [
      {
        neobank: {
          primary: '#2563EB',
          'primary-content': '#FFFFFF',
          secondary: '#64748B',
          accent: '#0EA5E9',
          neutral: '#1E293B',
          'neutral-content': '#FFFFFF',
          'base-100': '#FFFFFF',
          'base-200': '#F9FAFB',
          'base-300': '#F3F4F6',
          'base-content': '#111827',
          info: '#3B82F6',
          success: '#10B981',
          warning: '#F59E0B',
          error: '#EF4444',
        },
      },
    ],
    darkTheme: false, // banking UI ships a single, audited light theme
  },
};
```

Rules:

- One custom theme named `neobank`; `darkTheme: false`.
- Set `default: true` on the theme (or apply `data-theme="neobank"` on `<html>`) so components never fall back to DaisyUI defaults.
- Values must match `07-ui-ux-design.md`; change the theme there and here together.
- Keep `logs: false` so CI output stays clean.

## 3. Component Usage

```tsx
// components/features/TransferForm.tsx — DaisyUI classes, no raw utility chains
<form className="card card-body gap-4" onSubmit={handleSubmit}>
  <label className="form-control w-full">
    <span className="label-text">Destination account ID</span>
    <input name="destinationAccountId" className="input input-bordered w-full" />
  </label>

  <label className="form-control w-full">
    <span className="label-text">Amount (cents)</span>
    <input name="amount" type="number" className="input input-bordered w-full" />
  </label>

  <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
    {isSubmitting ? <span className="loading loading-spinner loading-sm" /> : 'Transfer'}
  </button>
</form>
```

Rules:

- Prefer semantic classes: `btn btn-primary`, `input input-bordered`, `card card-body`, `badge badge-success`, `alert alert-error`.
- Compose with layout utilities only (`gap-4`, `w-full`, `mt-2`).
- Never invent a variant when DaisyUI has one; check the component class list first.
- Forms use `form-control` + `label-text` + `input-bordered`; do not rebuild inputs from scratch.

## 4. Wrapping DaisyUI in Components

Wrap DaisyUI classes in typed components so call sites do not repeat class strings and variant intent stays consistent.

```tsx
// components/ui/Badge.tsx
import type { ReactNode } from 'react';

type BadgeVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral';

const variantClass: Record<BadgeVariant, string> = {
  success: 'badge-success',
  warning: 'badge-warning',
  error: 'badge-error',
  info: 'badge-info',
  neutral: 'badge-neutral',
};

export const Badge = ({ variant, children }: { variant: BadgeVariant; children: ReactNode }) => {
  return <span className={`badge ${variantClass[variant]}`}>{children}</span>;
};

// components/ui/Card.tsx
export const Card = ({ className = '', children }: { className?: string; children: ReactNode }) => {
  return <div className={`card card-body bg-base-100 shadow-sm ${className}`}>{children}</div>;
};
```

Rules:

- Wrap each DaisyUI component once (`Button`, `Input`, `Card`, `Badge`, `Alert`); pages import the wrapper.
- Keep the `className` passthrough for layout adjustments.
- Variant maps are static full class strings (never `badge-${variant}`) so Tailwind/DaisyUI emits them.
- A wrapper is the place to add ARIA, focus handling, and loading states.

## 5. Status & Financial Semantics

Transfer and transaction states map to a fixed badge vocabulary, consistent everywhere.

```tsx
const statusBadge: Record<TransferStatus, BadgeVariant> = {
  pending: 'warning',
  processing: 'info',
  completed: 'success',
  failed: 'error',
  reversed: 'neutral',
};

const transactionBadge = (type: 'debit' | 'credit'): BadgeVariant =>
  type === 'credit' ? 'success' : 'error';
```

Rules:

- One mapping per domain concept, defined once and reused.
- Never colour alone: the badge text always names the state.
- Unexpected states fall back to `neutral`, never an undefined class.

## 6. Feedback, Loading & Empty States

- Errors: `alert alert-error`; warnings: `alert alert-warning`; success: `alert alert-success`.
- Loading: skeleton blocks (`skeleton`) for structure, `loading loading-spinner` for buttons and inline waits.
- Disabled: `disabled` attribute plus `btn-disabled` where a non-button is styled as one — never rely on colour alone.
- Empty: `card` with a heading, explanatory text, and an action button.

## 7. Accessibility

- DaisyUI components render semantic elements; keep them semantic (`<button>`, `<label>`, `<dialog>`).
- Associate every `input` with a `label` (or `label` wrapper); helper text goes in `label-text-alt`.
- Visible focus comes from DaisyUI; do not override it away without a replacement.
- Contrast follows the theme tokens (WCAG 2.1 AA from `07-ui-ux-design.md`).
- Announce async status changes with `aria-live` where a badge or alert updates.

## 8. Customisation Boundary

- Prefer DaisyUI classes and theme tokens over custom CSS.
- When a component needs a one-off look, extend with utilities on the wrapper, not a new global class.
- If a pattern is reused, add it to the wrapper component — never to a stylesheet.
- Keep `globals.css` limited to Tailwind directives, the theme reset, and true globals (font, body background).

## 9. Testing

- Test the wrapper's behaviour (disabled, error, loading), not DaisyUI's internal markup.
- Query by role and accessible name; DaisyUI class names are implementation detail.
- Assert accessibility attributes (`aria-live`, `aria-label`, `disabled`) explicitly.
- Do not snapshot DaisyUI markup — a library upgrade must not fail tests.

## 10. Code Style & Rules

- DaisyUI first, Tailwind utilities for layout, custom CSS last.
- One custom `neobank` light theme, derived from `07-ui-ux-design.md`.
- Wrap components once; variant maps use static full class strings.
- Wrappers are arrow functions (`export const Badge = (...) => { ... }`); no `function` declarations.
- Status and financial semantics use a single, shared badge mapping.
- Never construct DaisyUI class names dynamically.
- `tailwindcss.md` owns the token/utility layer; this file owns the component layer.
