# NEO Bank — UI/UX Design — Responsive Design

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines how the UI adapts across mobile, tablet, and desktop breakpoints defined in [`layout.md`](./layout.md). NEO Bank uses a mobile-first approach: base styles target the smallest breakpoint, and enhancements are layered on at larger breakpoints using `min-width` media queries.

## Mobile (< 640px)

### Layout

- Single column layout (all content stacks vertically).
- Main content padding: `space-4` (horizontal), `space-6` (vertical) — reduced from desktop.
- Header: logo left, hamburger menu + user menu right.
- No side-by-side content; dashboard sections stack: balance card → quick actions → recent transactions.

### Components

- Full-width buttons (100% width, `space-6` padding vertical).
- Cards span full content width.
- Form fields full width.
- Navigation: hamburger menu opens a right-slide drawer with full-width menu items.
- Theme switcher: icon button in account menu on mobile.

### Touch targets

- Minimum 44px × 44px for all interactive elements.
- Generous spacing between targets (≥ 8px).
- Buttons and links have adequate hit areas even on small screens.

### Typography

- Headings: `text-2xl` max on mobile (prevents wrapping issues).
- Body: `text-base` (16px) for readability.
- Labels: `text-sm` with clear visual hierarchy.

### Examples

**Balance card on mobile:** Full width, balance amount `text-3xl`, currency inline, last updated below.

**Transfer form on mobile:** Single column, destination field → amount field → submit button (full width).

## Tablet (640px – 1024px)

### Layout

- Two-column layout where appropriate (dashboard: balance + quick actions left, transactions right).
- Main content padding: `space-6` (horizontal), `space-8` (vertical).
- Header: navigation may appear horizontally if space permits; otherwise hamburger.
- Max content width expands to 1024px.

### Components

- Buttons: auto width (not full width), `space-3` vertical, `space-6` horizontal padding.
- Cards: may sit side-by-side (2-column grid).
- Form fields: may be side-by-side for related fields (e.g., first name + last name).
- Navigation: horizontal menu or hamburger depending on available width.

### Touch targets

- Minimum 44px × 44px maintained (tablets are touch-first).
- Hover states optional on tablet (may not be reliable).

### Typography

- Headings: up to `text-3xl`.
- Body: `text-base`.

### Examples

**Dashboard on tablet:** Two-column grid — left: balance card + quick actions; right: recent transactions (up to 5 items visible).

**Transfer on tablet:** Centered card, max 500px width, two fields stacked vertically.

## Desktop (> 1024px)

### Layout

- Multi-column layout (dashboard: 3+ columns possible).
- Main content padding: `space-8` (horizontal), `space-6` (vertical).
- Header: full horizontal navigation (Dashboard, Transactions, Profile, Settings, Logout).
- Max content width: 1200px (centered).
- Footer: full width with multiple link columns.

### Components

- Buttons: auto width, `space-3` vertical, `space-6` horizontal.
- Cards: side-by-side grid (2-4 columns depending on content).
- Hover states: interactive elements change on hover (shadow, color, underline).
- Tooltips: appear on hover for icon buttons and truncated text.

### Touch targets

- Minimum 44px × 44px maintained (accessibility baseline).
- Hover states supplement (not replace) touch targets.

### Typography

- Headings: up to `text-3xl` for page headings.
- Body: `text-base`.
- Supporting text: `text-sm`.

### Examples

**Dashboard on desktop:** Three-column layout — left: balance card (large); center: quick actions + recent transactions; right: account summary + upcoming transfers.

**Transfer on desktop:** Centered card, max 500px width, destination and amount fields side-by-side if space permits.

## Breakpoint Reference

```css
/* Mobile-first base styles */
.component {
  /* Base mobile styles */
}

/* Tablet and up */
@media (min-width: 640px) {
  .component {
    /* sm breakpoint enhancements */
  }
}

/* Tablet landscape and up */
@media (min-width: 768px) {
  .component {
    /* md breakpoint enhancements */
  }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .component {
    /* lg breakpoint enhancements */
  }
}

/* Large desktop and up */
@media (min-width: 1280px) {
  .component {
    /* xl breakpoint enhancements */
  }
}

/* Extra large desktop */
@media (min-width: 1536px) {
  .component {
    /* 2xl breakpoint enhancements */
  }
}
```

## Cross-references

- Layout: [`layout.md`](./layout.md) (breakpoint definitions)
- Design tokens: [`design-tokens.md`](./design-tokens.md) (spacing, typography tokens)
- Screen designs: [`screens.md`](./screens.md) (per-screen responsive behavior)
- Components: [`components.md`](./components.md) (responsive component variants)
- Planned test: Visual regression tests at each breakpoint; Playwright E2E with mobile/tablet/desktop viewports