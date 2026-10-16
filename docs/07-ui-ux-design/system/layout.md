# NEO Bank — UI/UX Design — Layout

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines page layout, responsive breakpoints, navigation patterns, and grid system used across all screens. Every screen follows the same layout shell so users build consistent muscle memory.

## Page Layout Shell

### Header

| Property | Value | Usage |
|----------|-------|-------|
| Height | 64px | Fixed header height |
| Background | `color-surface` | Card/surface background |
| Border bottom | 1px solid `color-border` | Separation from content |
| Display | flex, align-items center, justify-content space-between | Horizontal layout |
| Padding | `space-4` (horizontal) | Inner spacing |
| Z-index | 100 | Above content, below modals |

**Header content (left to right):**
- **Logo**: the NEO Bank mark plus the `NEO` wordmark, links to dashboard. Use the `light` brand asset on light surfaces and the `dark` asset on dark surfaces ([`brand/README.md`](../brand/README.md)).
- **Navigation**: Horizontal menu (desktop only); hidden on mobile.
- **User menu**: Profile, theme switcher, logout (right side).

### Main content

| Property | Value | Usage |
|----------|-------|-------|
| Max width | 1200px | Content container |
| Margin | 0 auto | Centered |
| Padding | `space-8` (horizontal), `space-6` (vertical) | Inner spacing |
| Min height | calc(100vh - header - footer) | Fills viewport |

### Footer

| Property | Value | Usage |
|----------|-------|-------|
| Background | `color-background-secondary` | Subtle background |
| Border top | 1px solid `color-border` | Separation from content |
| Padding | `space-8` (horizontal), `space-6` (vertical) | Inner spacing |
| Font size | `text-sm` | Small secondary text |
| Color | `color-text-secondary` | Muted text |

**Footer content:** version, terms, privacy, support links.

## Responsive Breakpoints

| Breakpoint | Min width | Max width | Usage |
|------------|-----------|-----------|-------|
| `sm` | 640px | — | Small tablets |
| `md` | 768px | — | Tablets |
| `lg` | 1024px | — | Small desktops |
| `xl` | 1280px | — | Desktops |
| `2xl` | 1536px | — | Large desktops |

Mobile-first approach: styles apply from the smallest breakpoint up. Each breakpoint override uses `min-width` media queries.

## Grid System

A 12-column grid is used for page layouts:

| Column count | Breakpoint | Usage |
|--------------|------------|-------|
| 1 column | < 640px | Mobile: single column, stacked |
| 2 columns | 640px – 1024px | Tablet: side-by-side where appropriate |
| 4+ columns | > 1024px | Desktop: multi-column dashboards |

Gutters: `space-4` between columns. Row gaps: `space-6`.

## Navigation

### Desktop navigation

- Horizontal menu in header (center).
- Font size: `text-sm`.
- Font weight: `font-medium`.
- Color: `color-text-secondary`.
- Hover: `color-text-primary`.
- Active: `color-primary`.
- Focus: 2px ring `color-primary`.
- Items: Dashboard, Transactions, Profile, Settings, Logout.

### Mobile navigation

- Hamburger menu button in header (right side, before user menu).
- Slide-in drawer from right on click.
- Full-width menu items (touch target ≥ 44px).
- Close button (X) at top right of drawer.
- Drawer overlay dims background (`rgba(0,0,0,0.5)`).
- Body scroll locked when drawer is open.

## Card Layout

All content cards follow a consistent pattern:

| Property | Value | Usage |
|----------|-------|-------|
| Background | `color-surface` | Card background |
| Border | 1px solid `color-border` | Subtle border |
| Border radius | `radius-xl` | Rounded corners |
| Padding | `space-6` | Inner spacing |
| Shadow | `shadow-sm` | Subtle elevation |
| Hover shadow | `shadow-md` | Interactive elevation |

## Form Layout

Forms follow a vertical stack pattern:

| Property | Value | Usage |
|----------|-------|-------|
| Label | `text-sm`, `font-medium`, `color-text-primary` | Field label |
| Label margin bottom | `space-2` | Space below label |
| Input | `text-base`, `color-text-primary`, `color-surface` | Input field |
| Input border | 1px solid `color-border` | Default border |
| Input border radius | `radius-full` | Rounded input |
| Input padding | `space-3` (vertical), `space-4` (horizontal) | Inner spacing |
| Input focus | border `color-primary`, 2px ring `color-primary` (20% opacity) | Focus state |
| Input error | border `color-error`, `color-error-fg` message | Validation-error |
| Helper text | `text-sm`, `color-text-secondary`, margin top `space-2` | Guidance text |
| Error message | `text-sm`, `color-error-fg`, margin top `space-2` | Error text |

## Cross-references

- Design tokens: [`design-tokens.md`](./design-tokens.md)
- Components: [`components.md`](./components.md)
- Screen designs: [`screens.md`](./screens.md)
- Responsive design: [`responsive.md`](./responsive.md)
- States: [`states.md`](./states.md)
- Planned test: Visual regression tests + Playwright E2E for layout on all breakpoints