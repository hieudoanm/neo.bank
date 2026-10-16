# NEO Bank — UI/UX Design

## 1. Purpose

This document defines the UI/UX design for NEO Bank, including design tokens, layout, states, and accessibility guidelines.

The design must be:

- Clean and professional (banking context).
- Accessible (WCAG 2.1 AA compliant).
- Responsive (mobile, tablet, desktop).
- Consistent across all screens.
- Clear in feedback (loading, empty, validation-error, permission-error, network-error, unexpected-error, and success states).
- Themeable: every screen renders correctly in both the light and dark theme (see the Themes & Implementation Mapping section of [`design-tokens.md`](./07-ui-ux-design/system/design-tokens.md#themes--implementation-mapping)).

Standalone HTML prototypes of the screens live in [`07-ui-ux-design/prototypes/`](./07-ui-ux-design/prototypes/README.md); they use Tailwind CSS and DaisyUI via CDN and implement the theme switcher exactly as specified here.

## 2. Design Tokens

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`design-tokens.md`](./07-ui-ux-design/system/design-tokens.md) | Colors, typography, spacing, border radius, shadows, themes |

## 3. Components

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`components.md`](./07-ui-ux-design/system/components.md) | Buttons, form fields, cards, badges, alerts, theme switcher  |

## 4. Layout

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`layout.md`](./07-ui-ux-design/system/layout.md) | Page layout, breakpoints, navigation |

## 5. Screen Designs

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`screens.md`](./07-ui-ux-design/system/screens.md) | Login, Register, Dashboard (includes balance), Transfer, Transaction History, Transfer Status |

## 6. States

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`states.md`](./07-ui-ux-design/system/states.md) | loading, empty, validation-error, permission-error, network-error, unexpected-error |

## 7. Accessibility

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`accessibility.md`](./07-ui-ux-design/system/accessibility.md) | Keyboard, screen reader, contrast, focus, error handling |

## 8. Responsive Design

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`responsive.md`](./07-ui-ux-design/system/responsive.md) | Mobile, tablet, desktop layouts |

## 9. Animation

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`animation.md`](./07-ui-ux-design/system/animation.md) | Transitions, loading animations, micro-interactions |

## 10. Iconography

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`iconography.md`](./07-ui-ux-design/system/iconography.md) | Icon style, sizing, semantics, accessibility |

## 11. Prototypes

| Document | Covers |
| -------- | ------ |
| [`prototypes/README.md`](./07-ui-ux-design/prototypes/README.md) | Static HTML prototypes with Tailwind CSS and DaisyUI, theme switcher, screenshots |
