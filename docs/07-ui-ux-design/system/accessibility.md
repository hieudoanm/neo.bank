# NEO Bank — UI/UX Design — Accessibility

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines accessibility requirements aligned to WCAG 2.1 AA. NEO Bank must be usable by people with visual, auditory, motor, and cognitive impairments. Accessibility is not an add-on — it is built into every component and screen from the start.

## Keyboard Navigation

All interactive elements must be keyboard accessible:

| Requirement | Specification |
|-------------|---------------|
| Tab order | Logical and predictable (DOM order matches visual order) |
| Focus indicators | Visible focus ring (2px `color-primary`) on all interactive elements |
| Skip link | "Skip to main content" link, visible on focus, at top of page |
| Escape key | Closes modals, dropdowns, and overlays |
| Enter/Space | Activates buttons and links |
| Arrow keys | Navigates within menus, tabs, and radio groups |
| No keyboard traps | Focus never trapped in any component |

## Screen Reader Support

| Requirement | Specification |
|-------------|---------------|
| Semantic HTML | `<header>`, `<nav>`, `<main>`, `<footer>`, `<section>`, `<article>`, `<button>`, `<a>`, `<input>`, `<label>` |
| ARIA labels | `aria-label` on icon buttons, `aria-labelledby` on complex components |
| Alt text | Decorative images: `alt=""`; informative images: descriptive alt text |
| Live regions | `aria-live="polite"` for status updates; `aria-live="assertive"` for errors |
| Role attributes | `role="alert"` for error banners, `role="dialog"` for modals |
| Landmarks | All content within landmark regions |

## Color Contrast

| Element | Minimum ratio | WCAG level |
|---------|--------------|------------|
| Normal text | 4.5:1 | AA |
| Large text (18pt+ or 14pt+ bold) | 3:1 | AA |
| UI components (borders, inputs) | 3:1 | AA |
| Graphical objects | 3:1 | AA |

All color tokens in [`design-tokens.md`](./design-tokens.md) are contrast-checked in both `neo-light` and `neo-dark` themes. Text never sits directly on semantic accent colors (`color-success`, `color-error`, `color-warning`, `color-info`) — it sits on the matching `-bg` using the `-fg` value.

## Focus Management

| Requirement | Specification |
|-------------|---------------|
| Visible focus indicators | 2px ring `color-primary` on all focusable elements |
| Focus trapped in modals | Focus cycles within modal; cannot tab out |
| Focus returned after modal close | Focus returns to the element that opened the modal |
| Focus moved after navigation | Focus moves to first focusable element or page heading |
| Focus visible indicator | Never suppressed via CSS (`:focus-visible` respected) |

## Error Handling

| Requirement | Specification |
|-------------|---------------|
| Errors announced | `aria-live="assertive"` for error messages |
| Error-field association | `aria-describedby` links field to error message |
| Error summary | For multiple errors, summary at top of form with links to each field |
| Non-revealing | Security errors generic ("Account not found") — no enumeration |
| Actionable | Error messages tell user how to fix the problem |

## Touch Targets

| Element | Minimum size |
|---------|-------------|
| Buttons | 44px × 44px |
| Links | 44px × 44px |
| Input fields | 44px height |
| Icon buttons | 44px × 44px |
| Spacing between targets | 8px minimum |

## Motion and Animation

| Requirement | Specification |
|-------------|---------------|
| Respect `prefers-reduced-motion` | Disable or reduce animations when enabled |
| No autoplay | No auto-playing video or audio > 5 seconds |
| No flashing | Content does not flash more than 3 times per second |
| Reduced motion fallback | Static state instead of animated transitions |

## Text and Typography

| Requirement | Specification |
|-------------|---------------|
| Font sizes | Minimum 14px for body text (`text-sm` and up) |
| Line height | Minimum 1.5 for body text (`leading-normal`) |
| Letter spacing | Not tighter than default |
| Text resizing | Supports 200% zoom without loss of content |
| No images of text | Text is text, not images (except logos) |

## Forms

| Requirement | Specification |
|-------------|---------------|
| Labels | Every input has a visible `<label>` |
| Required fields | Marked with `aria-required="true"` and visual indicator |
| Error association | `aria-describedby` links input to error message |
| Fieldset/legend | Group related fields with `<fieldset>` and `<legend>` |
| Autocomplete | `autocomplete` attributes on email/password fields |

## Testing

| Method | Tool | Frequency |
|--------|------|-----------|
| Automated axe-core scan | axe-core / Deque | Every PR |
| Keyboard-only navigation | Manual | Every PR |
| Screen reader test | NVDA / VoiceOver / TalkBack | Every release |
| Color contrast check | Automated token check | Every token change |
| Zoom test (200%) | Manual | Every release |

## Cross-references

- UI/UX design index: [`07-ui-ux-design.md`](../../07-ui-ux-design.md#7-accessibility)
- Design tokens: [`design-tokens.md`](./design-tokens.md) (contrast-checked tokens)
- Components: [`components.md`](./components.md) (accessible component specs)
- Planned test: `test_accessibility` (axe-core automated scan + manual keyboard/screen reader)