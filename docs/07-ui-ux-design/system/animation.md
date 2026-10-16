# NEO Bank — UI/UX Design — Animation

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines animation timing, easing, and motion preferences used across the UI. Animations are purposeful: they provide feedback, indicate state changes, and guide attention — never decorative or distracting. All animations respect the user's `prefers-reduced-motion` setting.

## Transitions

| Token | Duration | Easing | Usage |
|-------|----------|--------|-------|
| `transition-fast` | 200ms | ease-in-out | Button hover, badge state changes |
| `transition-default` | 300ms | ease-in-out | Card hover, form field focus |
| `transition-slow` | 500ms | ease-in-out | Modal open/close, drawer slide, page transitions |

### Standard transition property

```css
/* Default transition for interactive elements */
transition: all 300ms ease-in-out;

/* Fast transition for subtle feedback */
transition: color 200ms ease-in-out, background-color 200ms ease-in-out;

/* Slow transition for layout changes */
transition: transform 500ms ease-in-out, opacity 500ms ease-in-out;
```

## Loading Animations

### Skeleton loaders

- Shimmer effect moves left-to-right across placeholder content.
- Base color: `color-background-secondary`; shimmer color: `color-border-light`.
- Duration: 1500ms, infinite loop, ease-in-out.
- No distracting animation; subtle and professional.
- Clear indication that content is loading.

### Spinners

- Size: 24px (inline), 40px (page-level).
- Color: `color-primary`.
- Duration: 800ms per rotation, linear.
- Used inside buttons during submission and in empty content areas.

### Progress indicators

- For transfer status polling: subtle pulse on status badge.
- For long operations: determinate progress bar (when progress is known).
- Indeterminate spinner when progress is unknown.

## Micro-interactions

### Button hover

- Color change: `color-primary` → `color-primary-hover`.
- Shadow increase: `shadow-sm` → `shadow-md`.
- Transform: none (buttons don't move on hover).
- Duration: 200ms, ease-in-out.

### Button click (active state)

- Scale down slightly: `transform: scale(0.98)`.
- Background: `color-primary-active`.
- Duration: 100ms, ease-in-out (immediate feedback).

### Link hover

- Underline appears.
- Color: `color-text-primary` → `color-primary`.
- Duration: 200ms, ease-in-out.

### Card hover

- Shadow increase: `shadow-sm` → `shadow-md`.
- Transform: none (cards don't lift).
- Duration: 300ms, ease-in-out.

### Form field focus

- Border color: `color-border` → `color-primary`.
- Ring: 2px `color-primary` at 20% opacity.
- Duration: 200ms, ease-in-out.

### Alert appearance

- Slide down from top (300ms, ease-in-out).
- Fade in (opacity 0 → 1, 200ms, ease-in-out).
- Auto-dismiss after 5 seconds for non-critical alerts.

### Theme switcher

- Smooth color transition: 300ms, ease-in-out.
- Icon morph (sun ↔ moon): 300ms, ease-in-out.
- No flash of wrong theme (applied before first paint).

## Motion Preferences

### `prefers-reduced-motion: reduce`

When the user has requested reduced motion:

| Animation | Fallback |
|-----------|----------|
| Transitions | Instant (0ms) or 50ms fade |
| Skeletons | Static pulse (no shimmer movement) |
| Spinners | Static loading indicator |
| Modal/drawer | Fade in/out (no slide) |
| Micro-interactions | Color change only (no transform/scale) |

### Implementation

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
  
  .skeleton {
    animation: none;
    background: color-background-secondary;
  }
  
  .modal, .drawer {
    transition: opacity 50ms ease-in-out;
  }
}
```

## Accessibility

- No content flashes more than 3 times per second (seizure safety).
- No autoplaying video or audio longer than 5 seconds.
- Animations can be paused or hidden (respect `prefers-reduced-motion`).
- Meaning conveyed by animation is also available in static form (e.g., status text, not just a spinning icon).
- Reduced motion does not break functionality.

## Cross-references

- Design tokens: [`design-tokens.md`](./design-tokens.md) (spacing, shadow tokens)
- Components: [`components.md`](./components.md) (interactive component specs)
- States: [`states.md`](./states.md) (loading, empty, validation-error, permission-error, network-error, unexpected-error states)
- Accessibility: [`accessibility.md`](./accessibility.md) (motion preferences)
- Planned test: Visual tests verify no excessive flashing; manual test with reduced motion enabled