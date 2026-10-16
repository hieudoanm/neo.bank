# NEO Bank — UI/UX Design — Iconography

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines icon style, sizing, semantics, and accessibility requirements. Icons are used throughout the UI for actions, navigation, status indicators, and visual hierarchy. Consistency in icon style and sizing ensures a polished, professional appearance.

## Icon Set

NEO Bank uses a single, consistent icon set (recommended: Heroicons or Lucide Icons). Mixing icon sets is prohibited.

| Property | Value |
|----------|-------|
| Style | Outline (default) or filled (for active/selected states only) |
| Stroke width | 2px (consistent across all icons) |
| Stroke linecap | round |
| Stroke linejoin | round |
| Fill | none (outline), currentColor (filled active state) |
| Color | `color-text-primary` (default), `color-primary` (active), `color-error` / `color-success` (status) |

## Icon Sizes

| Size | Value | Usage |
|------|-------|-------|
| `icon-xs` | 16px | Small actions, badges, compact UI |
| `icon-sm` | 20px | Button icons, list items |
| `icon-md` | 24px | Standard action icons, navigation |
| `icon-lg` | 32px | Section headers, feature callouts |
| `icon-xl` | 48px | Empty state illustrations, onboarding |

## Common Icons

| Icon | Name | Size | Usage |
|------|------|------|-------|
| User | `User` | 20px | Profile, account menu |
| Lock | `Lock` | 20px | Security, password fields |
| Balance | `Wallet` or `Banknote` | 24px | Balance display, financial actions |
| Transfer | `ArrowLeftRight` or `Send` | 20px | Transfer initiation, transfer status |
| Transaction | `List` or `Receipt` | 20px | Transaction history |
| Success | `CheckCircle` | 20px | Success alerts, completed transfers |
| Error | `XCircle` | 20px | Error alerts, failed transfers |
| Warning | `ExclamationTriangle` | 20px | Warning alerts, validation |
| Info | `InformationCircle` | 20px | Info alerts, help text |
| Home | `Home` | 20px | Navigation, dashboard link |
| Logout | `LogOut` | 20px | Sign out action |
| Search | `Search` | 20px | Search actions (future) |
| Filter | `Filter` | 20px | Filter controls (future) |
| Menu | `Menu` | 24px | Hamburger menu (mobile navigation) |
| Chevron down | `ChevronDown` | 16px | Dropdown indicators |
| Refresh | `RefreshCw` | 20px | Retry, refresh actions |
| Theme | `Sun` / `Moon` | 20px | Theme switcher |

## Accessibility

| Requirement | Specification |
|-------------|---------------|
| Decorative icons | `aria-hidden="true"` (screen readers skip them) |
| Functional icons | Visible text label or `aria-label` on the parent element |
| Icon buttons | `aria-label` describing the action (e.g., "Toggle theme", "Close modal") |
| Color meaning | Never conveyed by color alone; always accompanied by text or icon shape |
| Size | Minimum 24px × 24px touch target (or 44px with padding) |
| Focus | Icon buttons are focusable and show focus ring |

### Examples

```html
<!-- Decorative icon (hidden from screen readers) -->
<span aria-hidden="true">
  <svg icon="lock" size="icon-sm" />
</span>
<label>Password</label>

<!-- Functional icon button (screen reader accessible) -->
<button aria-label="Toggle colour theme">
  <span aria-hidden="true">
    <svg icon="sun" size="icon-md" />
  </span>
</button>

<!-- Status icon with text (redundant encoding) -->
<span class="status-badge status-success">
  <span aria-hidden="true">
    <svg icon="check-circle" size="icon-sm" />
  </span>
  <span class="visually-hidden">Success</span>
  Completed
</span>
```

## Rules

1. **One icon set only** — do not mix Heroicons, Lucide, FontAwesome, etc.
2. **Consistent stroke width** — 2px across all icons, all sizes.
3. **Semantic meaning** — each icon has a defined meaning; do not repurpose icons.
4. **Accessible by default** — decorative icons hidden; functional icons labeled.
5. **No text in icons** — icons are graphical, never contain readable text.
6. **Color by token** — icon colors use design tokens, never hardcoded hex values.
7. **Size by token** — use `icon-xs` through `icon-xl` tokens, never arbitrary pixel values.

## Cross-references

- Design tokens: [`design-tokens.md`](./design-tokens.md) (color tokens used for icon colors)
- Components: [`components.md`](./components.md) (button, badge, alert component specs)
- States: [`states.md`](./states.md) (status badge variants)
- Accessibility: [`accessibility.md`](./accessibility.md) (aria-label requirements)
- Planned test: Visual tests verify icon consistency; axe-core scan verifies icon accessibility