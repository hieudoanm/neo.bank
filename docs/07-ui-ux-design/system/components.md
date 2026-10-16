# NEO Bank — UI/UX Design — Components

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines reusable UI components: buttons, form fields, cards, badges, alerts, and the theme switcher. All components reference design tokens defined in [`design-tokens.md`](./design-tokens.md).

## Buttons

### Primary button

- Background: `color-primary`
- Text: `color-primary-content`
- Border radius: `radius-full`
- Padding: `space-3` (vertical), `space-6` (horizontal)
- Font weight: `font-semibold`
- Hover: `color-primary-hover`
- Active: `color-primary-active`
- Focus: 2px ring `color-primary`
- Disabled: opacity 0.5, cursor not-allowed

### Secondary button

- Background: `color-surface`
- Text: `color-text-primary`
- Border: 1px solid `color-border`
- Border radius: `radius-full`
- Padding: `space-3` (vertical), `space-6` (horizontal)
- Font weight: `font-semibold`
- Hover: `color-background-secondary`
- Active: `color-surface-muted`

### Danger button

- Background: `color-error`
- Text: `color-text-inverse`
- Border radius: `radius-full`
- Padding: `space-3` (vertical), `space-6` (horizontal)
- Font weight: `font-semibold`
- Hover: darker red

## Form Fields

### Input field

- Background: `color-surface`
- Text: `color-text-primary`
- Placeholder: `color-text-tertiary`
- Border: 1px solid `color-border`
- Border radius: `radius-full`
- Padding: `space-3` (vertical), `space-4` (horizontal)
- Font size: `text-base`
- Focus: border `color-primary`, ring 2px `color-primary` (20% opacity)
- Validation-error: border `color-error`, text `color-error-fg` for the message
- Disabled: background `color-background-secondary`, text `color-text-tertiary`, cursor not-allowed

### Label

- Font size: `text-sm`
- Font weight: `font-medium`
- Color: `color-text-primary`
- Margin bottom: `space-2`

### Helper text

- Font size: `text-sm`
- Color: `color-text-secondary`
- Margin top: `space-2`

### Error message

- Font size: `text-sm`
- Color: `color-error-fg`
- Margin top: `space-2`

## Cards

### Card

- Background: `color-surface`
- Border: 1px solid `color-border`
- Border radius: `radius-xl`
- Padding: `space-6`
- Shadow: `shadow-sm`
- Hover: `shadow-md`

### Card header

- Font size: `text-lg`
- Font weight: `font-semibold`
- Color: `color-text-primary`
- Margin bottom: `space-4`

## Badges

- Padding: `space-1` (vertical), `space-3` (horizontal)
- Border radius: `radius-full`
- Font size: `text-xs`
- Font weight: `font-medium`

### Badge variants

| Variant  | Background          | Text             |
| -------- | ------------------- | ---------------- |
| Success  | `color-success-bg`  | `color-success-fg` |
| Error    | `color-error-bg`    | `color-error-fg`   |
| Warning  | `color-warning-bg`  | `color-warning-fg` |
| Info     | `color-info-bg`     | `color-info-fg`    |
| Neutral  | `color-background-secondary` | `color-text-secondary` |

## Alerts

- Padding: `space-4`
- Border radius: `radius-full`
- Border: 1px solid (variant accent)
- Background: variant tint (`color-*-bg`)
- Text: variant foreground (`color-*-fg`)
- Display: flex, align items start, gap `space-3`

### Alert variants

| Variant  | Border            | Background         | Text             |
| -------- | ----------------- | ------------------ | ---------------- |
| Success  | `color-success`   | `color-success-bg` | `color-success-fg` |
| Error    | `color-error`     | `color-error-bg`   | `color-error-fg`   |
| Warning  | `color-warning`   | `color-warning-bg` | `color-warning-fg` |
| Info     | `color-info`      | `color-info-bg`    | `color-info-fg`    |

## Theme Switcher

A small control that toggles between `neo-light` and `neo-dark`.

- Placement: header, right-hand cluster on desktop; in the account menu on mobile.
- Form: a two-state segmented control (sun / moon icons) or an icon button; either is acceptable.
- Size: 32px tall, icons 16–20px, min touch target 44px.
- Active theme is visually selected and marked with `aria-pressed="true"`.
- Changing the theme applies immediately, persists to `localStorage` (`neo-theme`), and updates `color-scheme` so native controls follow.
- Accessible label: "Colour theme"; the control is keyboard operable.