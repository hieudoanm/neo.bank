# NEO Bank — UI/UX Design — Design Tokens

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines the design tokens (colors, typography, spacing, border radius, shadows, themes) used across all screens. Every component must reference tokens, never hardcoded values.

## Colors

Every colour token has a **light** and a **dark** value. Light is the default theme; both are contrast-checked to WCAG 2.1 AA.

### Brand

| Token                   | Light   | Dark    | Usage                           |
| ----------------------- | ------- | ------- | ------------------------------- |
| `color-primary`         | #2563EB | #3B82F6 | Primary buttons, links, accents |
| `color-primary-hover`   | #1D4ED8 | #60A5FA | Primary hover                   |
| `color-primary-active`  | #1E40AF | #2563EB | Primary pressed                 |
| `color-primary-content` | #FFFFFF | #0B1220 | Text/icons on `color-primary`   |
| `color-primary-muted`   | #EFF6FF | #172554 | Primary tint background         |

### Surfaces & borders

| Token                        | Light   | Dark    | Usage                         |
| ---------------------------- | ------- | ------- | ----------------------------- |
| `color-background`           | #FFFFFF | #0B1220 | Page background               |
| `color-background-secondary` | #F9FAFB | #111827 | Secondary page background     |
| `color-surface`              | #FFFFFF | #111827 | Cards, panels, tables, modals |
| `color-surface-raised`       | #F9FAFB | #1F2937 | Menus, popovers, dropdowns    |
| `color-border`               | #E5E7EB | #374151 | Borders, dividers             |
| `color-border-light`         | #F3F4F6 | #1F2937 | Subtle dividers               |
| `color-border-strong`        | #D1D5DB | #4B5563 | Emphasised borders            |

### Text

| Token                  | Light   | Dark    | Usage                                 |
| ---------------------- | ------- | ------- | ------------------------------------- |
| `color-text-primary`   | #111827 | #F9FAFB | Primary text                          |
| `color-text-secondary` | #6B7280 | #D1D5DB | Secondary text                        |
| `color-text-tertiary`  | #9CA3AF | #9CA3AF | Tertiary text — large/decorative only |
| `color-text-inverse`   | #FFFFFF | #0B1220 | Text on solid dark/saturated fills    |

### Semantic (status)

| Token              | Light   | Dark    | Usage                            |
| ------------------ | ------- | ------- | -------------------------------- |
| `color-success`    | #10B981 | #34D399 | Success accents, positive amount |
| `color-success-bg` | #D1FAE5 | #064E3B | Success tint                     |
| `color-success-fg` | #047857 | #6EE7B7 | Success text on tint             |
| `color-error`      | #EF4444 | #F87171 | Error accents, negative amount   |
| `color-error-bg`   | #FEE2E2 | #450A0A | Error tint                       |
| `color-error-fg`   | #B91C1C | #FCA5A5 | Error text on tint               |
| `color-warning`    | #F59E0B | #FBBF24 | Warning accents                  |
| `color-warning-bg` | #FEF3C7 | #451A03 | Warning tint                     |
| `color-warning-fg` | #92400E | #FCD34D | Warning text on tint             |
| `color-info`       | #3B82F6 | #60A5FA | Information accents              |
| `color-info-bg`    | #DBEAFE | #172554 | Information tint                 |
| `color-info-fg`    | #1D4ED8 | #93C5FD | Information text on tint         |

### Rules

1. Text never sits directly on `color-success`/`error`/`warning`/`info`; it sits on the matching `-bg` using the `-fg` value (the base accent is for icons, borders, and fills only).
2. Positive and negative amounts use `color-success-fg` / `color-error-fg` as text, so they stay legible on either surface.
3. `color-text-tertiary` is reserved for large text and non-essential decoration; it does not meet the 4.5:1 target on the light surface.

## Typography

### Font family

- Primary: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif
- Monospace: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", monospace

### Font sizes

| Token       | Value           | Usage                  |
| ----------- | --------------- | ---------------------- |
| `text-xs`   | 0.75rem (12px)  | Labels, captions       |
| `text-sm`   | 0.875rem (14px) | Body text, helper text |
| `text-base` | 1rem (16px)     | Default body text      |
| `text-lg`   | 1.125rem (18px) | Large body text        |
| `text-xl`   | 1.25rem (20px)  | Headings               |
| `text-2xl`  | 1.5rem (24px)   | Section headings       |
| `text-3xl`  | 1.875rem (30px) | Page headings          |

### Font weights

| Token           | Value | Usage             |
| --------------- | ----- | ----------------- |
| `font-normal`   | 400   | Body text         |
| `font-medium`   | 500   | Emphasized text   |
| `font-semibold` | 600   | Headings, buttons |
| `font-bold`     | 700   | Strong emphasis   |

### Line heights

| Token             | Value | Usage          |
| ----------------- | ----- | -------------- |
| `leading-tight`   | 1.25  | Headings       |
| `leading-normal`  | 1.5   | Body text      |
| `leading-relaxed` | 1.625 | Long-form text |

## Spacing

| Token      | Value          | Usage           |
| ---------- | -------------- | --------------- |
| `space-1`  | 0.25rem (4px)  | Tight spacing   |
| `space-2`  | 0.5rem (8px)   | Small spacing   |
| `space-3`  | 0.75rem (12px) | Default spacing |
| `space-4`  | 1rem (16px)    | Medium spacing  |
| `space-6`  | 1.5rem (24px)  | Large spacing   |
| `space-8`  | 2rem (32px)    | Section spacing |
| `space-12` | 3rem (48px)    | Page spacing    |

## Border Radius

| Token         | Value          | Usage                          |
| ------------- | -------------- | ------------------------------ |
| `radius-sm`   | 0.5rem (8px)   | Small elements                 |
| `radius-md`   | 0.75rem (12px) | Default elements               |
| `radius-lg`   | 1rem (16px)    | Medium elements                |
| `radius-xl`   | 1.5rem (24px)  | Cards                          |
| `radius-full` | 9999px         | Pills, badges, buttons, inputs |

## Shadows

| Token       | Value                        | Usage             |
| ----------- | ---------------------------- | ----------------- |
| `shadow-sm` | 0 1px 2px rgba(0,0,0,0.05)   | Subtle elevation  |
| `shadow-md` | 0 4px 6px rgba(0,0,0,0.1)    | Default elevation |
| `shadow-lg` | 0 10px 15px rgba(0,0,0,0.1)  | High elevation    |
| `shadow-xl` | 0 20px 25px rgba(0,0,0,0.15) | Modal, popover    |

In the dark theme, elevation is carried mostly by the lighter surface and border, not the shadow; when a shadow is used it is stronger and fully black (`rgba(0,0,0,0.5)`).

## Themes & Implementation Mapping

NEO Bank ships **two themes**: `neo-light` (default) and `neo-dark`. Everything above is expressed as a token, so a component written once renders correctly in both.

### Mapping to Tailwind CSS + DaisyUI

The frontend uses Tailwind CSS + DaisyUI. The token set maps to implementation as follows:

- **Tailwind CSS configuration:** All color, spacing, typography, radius, and shadow tokens are defined in the Tailwind config (via `@theme`/theme extension) so classes like `bg-surface`, `text-text-primary`, `gap-4`, `rounded-xl`, etc. reference tokens rather than arbitrary values.
- **DaisyUI theme:** A custom DaisyUI theme (extending DaisyUI) defines the semantic color roles (primary, secondary, success, warning, error, info, neutral) using the same token values for light (`neo-light`) and dark (`neo-dark`). The HTML `data-theme` attribute (e.g. `data-theme="neo-light"`) activates the theme.
- **Usage:** Components must use these token-based classes/utilities; hardcoded hex or pixel values must not be used. The `neo-theme` localStorage key and `prefers-color-scheme` fallback apply the active theme.

### Switching

- The active theme is set with `data-theme="neo-light" | "neo-dark"` on the `<html>` element.
- The choice is persisted in `localStorage` under the `neo-theme` key.
- When no choice is stored, the theme follows the operating system (`prefers-color-scheme`).
- The stored theme is applied by an inline script **before first paint** so there is no flash of the wrong theme.
- A visible switcher lets the user change it; the change applies immediately and is announced to assistive technology.

```html
<!-- before any paint -->
<script>
  (function () {
    try {
      var t = localStorage.getItem('neo-theme');
      if (t !== 'neo-light' && t !== 'neo-dark') {
        t = matchMedia('(prefers-color-scheme: dark)').matches ? 'neo-dark' : 'neo-light';
      }
      document.documentElement.setAttribute('data-theme', t);
    } catch (e) {}
  })();
</script>
```

### Rules

1. No component hardcodes a colour; every colour is a token, and every token has a dark value.
2. Both themes are designed together — a change to a light token requires its dark counterpart in the same change.
3. Contrast is verified in both themes against the colour-contrast requirements in [`accessibility.md`](./accessibility.md).
4. The HTML prototypes in [`prototypes/`](../prototypes/README.md) implement this switcher with Tailwind CSS and DaisyUI via CDN.