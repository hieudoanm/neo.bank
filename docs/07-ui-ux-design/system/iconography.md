# NEO Bank — UI/UX Design — Iconography

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines icon style, sizing, semantics, and accessibility requirements. Icons are used throughout the UI for actions, navigation, status indicators, and visual hierarchy. Consistency in icon style and sizing ensures a polished, professional appearance.

## Icon Set

NEO Bank uses a single, consistent icon set: **Phosphor Icons**, consumed through **`react-icons`** (the `react-icons/pi` entry point, [react-icons.github.io/react-icons/icons/pi](https://react-icons.github.io/react-icons/icons/pi/)). Mixing icon sets is prohibited, and icons must never be hand-authored as bespoke inline SVG.

Install `react-icons` and import each icon as a React component (`Pi<Name>`); every icon renders `currentColor` and inherits its size from `font-size`:

```tsx
import { PiLock, PiSun } from 'react-icons/pi';

<PiLock aria-hidden="true" className="icon-sm text-text-tertiary" />
```

| Property  | Value                                                                                              |
| --------- | -------------------------------------------------------------------------------------------------- |
| Set       | Phosphor Icons, via `react-icons/pi`                                                               |
| Weight    | Regular everywhere — pick one Phosphor weight and do not mix weights                               |
| Rendering | SVG filled with `currentColor`                                                                     |
| Color     | `color-text-primary` (default), `color-primary` (active), `color-error` / `color-success` (status) |
| Source    | `react-icons/pi` only — no second icon library, no custom SVG                                      |

## Icon Sizes

Sizes are set with the `icon-*` font-size utilities (Phosphor icons scale with `font-size`); never pass the arbitrary `size` prop.

| Size      | Value | Usage                                 |
| --------- | ----- | ------------------------------------- |
| `icon-xs` | 16px  | Small actions, badges, compact UI     |
| `icon-sm` | 20px  | Button icons, list items              |
| `icon-md` | 24px  | Standard action icons, navigation     |
| `icon-lg` | 32px  | Section headers, feature callouts     |
| `icon-xl` | 48px  | Empty state illustrations, onboarding |

## Common Icons

| Icon         | Component (`react-icons/pi`)              | Size      | Usage                                |
| ------------ | ----------------------------------------- | --------- | ------------------------------------ |
| Profile      | `PiUser`                                  | `icon-sm` | Profile, account menu                |
| Lock         | `PiLockKey`                               | `icon-sm` | Security, password fields            |
| Balance      | `PiWallet` or `PiBank`                    | `icon-md` | Balance display, financial actions   |
| Transfer     | `PiArrowsLeftRight` or `PiPaperPlaneTilt` | `icon-sm` | Transfer initiation, transfer status |
| Transaction  | `PiReceipt` or `PiListBullets`            | `icon-sm` | Transaction history                  |
| Success      | `PiCheckCircle`                           | `icon-sm` | Success alerts, completed transfers  |
| Error        | `PiXCircle`                               | `icon-sm` | Error alerts, failed transfers       |
| Warning      | `PiWarning`                               | `icon-sm` | Warning alerts, validation           |
| Info         | `PiInfo`                                  | `icon-sm` | Info alerts, help text               |
| Home         | `PiHouse`                                 | `icon-sm` | Navigation, dashboard link           |
| Logout       | `PiSignOut`                               | `icon-sm` | Sign out action                      |
| Search       | `PiMagnifyingGlass`                       | `icon-sm` | Search actions (future)              |
| Filter       | `PiFunnel`                                | `icon-sm` | Filter controls (future)             |
| Menu         | `PiList`                                  | `icon-md` | Hamburger menu (mobile navigation)   |
| Chevron down | `PiCaretDown`                             | `icon-xs` | Dropdown indicators                  |
| Refresh      | `PiArrowsClockwise`                       | `icon-sm` | Retry, refresh actions               |
| Theme        | `PiSun` / `PiMoon`                        | `icon-sm` | Theme switcher                       |

## Accessibility

| Requirement      | Specification                                                            |
| ---------------- | ------------------------------------------------------------------------ |
| Decorative icons | `aria-hidden="true"` (screen readers skip them)                          |
| Functional icons | Visible text label or `aria-label` on the parent element                 |
| Icon buttons     | `aria-label` describing the action (e.g., "Toggle theme", "Close modal") |
| Color meaning    | Never conveyed by color alone; always accompanied by text or icon shape  |
| Size             | Minimum 24px × 24px touch target (or 44px with padding)                  |
| Focus            | Icon buttons are focusable and show focus ring                           |

### Examples

```tsx
import { PiCheckCircle, PiLock, PiSun } from 'react-icons/pi';

// Decorative icon next to a label (hidden from assistive tech)
<label className="flex items-center gap-2">
  <PiLock aria-hidden="true" className="icon-sm text-text-tertiary" />
  Password
</label>

// Icon-only button: the accessible name lives on the button
<button type="button" aria-label="Toggle colour theme">
  <PiSun aria-hidden="true" className="icon-md" />
</button>

// Status with redundant encoding (icon + text)
<span className="inline-flex items-center gap-1 rounded-full bg-success-bg px-3 py-1 text-xs font-medium text-success-fg">
  <PiCheckCircle aria-hidden="true" className="icon-sm" />
  Completed
</span>
```

## Rules

1. **One icon set only** — Phosphor Icons via `react-icons/pi`; never mix another library and never hand-author SVG icons.
2. **One weight** — use the Regular Phosphor weight everywhere; do not mix weights.
3. **Semantic meaning** — each icon has a defined meaning; do not repurpose icons.
4. **Accessible by default** — decorative icons hidden (`aria-hidden="true"`); functional icons labeled.
5. **No text in icons** — icons are graphical, never contain readable text.
6. **Color by token** — icons inherit `currentColor`; set it from design tokens, never a hardcoded hex.
7. **Size by token** — use the `icon-xs` through `icon-xl` font-size utilities, never the arbitrary `size` prop or pixel values.

## Cross-references

- Design tokens: [`design-tokens.md`](./design-tokens.md) (color tokens used for icon colors)
- Components: [`components.md`](./components.md) (button, badge, alert component specs)
- States: [`states.md`](./states.md) (status badge variants)
- Accessibility: [`accessibility.md`](./accessibility.md) (aria-label requirements)
- Planned test: Visual tests verify icon consistency; axe-core scan verifies icon accessibility
