# NEO Bank — UI/UX Design — States

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines the state patterns used consistently across all screens: the six data-screen states — loading, empty, validation-error, permission-error, network-error, unexpected-error — plus `success`, `disabled`, and `partial-success`. Every data screen implements these states: loading, empty, validation-error, permission-error, network-error, unexpected-error.

## Loading

### When to show

- Initial page load.
- Data fetch in progress (balance, transactions, transfer status).
- Form submission in progress (registration, login, transfer).
- Polling for transfer status updates.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Skeleton loader | Shimmer effect, `color-background-secondary` base, `color-border-light` shimmer |
| Spinner | 24px, `color-primary`, 200ms rotation |
| Button loading | Spinner inside button, button disabled |
| Text | "Loading…" (`text-sm`, `color-text-secondary`) |

### Behaviour

- Use skeleton loaders for lists and cards (preserves layout).
- Use spinners for buttons and inline loading.
- Disable interactive elements during loading (prevent double-submit).
- Show loading text where appropriate for screen readers (`aria-live="polite"`).

## Empty

### When to show

- No transactions in history.
- No transfers in progress.
- Search returns no results.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Illustration | Simple line icon, 64px, `color-text-tertiary` |
| Heading | `text-lg`, `font-semibold`, `color-text-primary` |
| Description | `text-sm`, `color-text-secondary` |
| Action button | Primary, full width on mobile, auto width on desktop |

### Examples

- **"No transactions yet"**: "You haven't made any transfers. Start by sending money to another account." + "Transfer money" button.
- **"No transfers in progress"**: "Your transfer history is empty. Initiate a transfer to see it here." + "Transfer money" button.

## Validation-error

### When to show

- Form field fails validation (invalid email, weak password, etc.).
- Input format is incorrect.
- Required field is missing.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Field border | 1px solid `color-error` |
| Field text | `color-text-primary` |
| Error message | `text-sm`, `color-error-fg`, below field |
| Helper text | `text-sm`, `color-text-secondary` (requirements) |

### Behaviour

- Field-specific error messages (not generic "form invalid").
- Red border on invalid fields.
- Helper text shows requirements (e.g., "Min 12 chars, mixed case, number, symbol").
- Clear indication of what needs to be fixed.
- Errors cleared when user starts correcting the field.
- Focus moves to first invalid field on form submit error.

## Permission-error

### When to show

- User attempts to access another account's data.
- User attempts an action without required permissions.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Message | Generic ("Account not found" or "Action not available") |
| No details | Never reveals why access was denied |

### Behaviour

- Generic message to prevent information leakage (no account enumeration).
- No retry option (permission won't change by retrying).
- Optional link to support or help documentation.

## Network-error

### When to show

- API request fails due to connectivity.
- Request times out.
- DNS or TLS failure.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Error banner | `color-error-bg`, `color-error-fg` |
| Message | "Unable to connect. Please check your internet connection." |
| Retry button | Secondary button, "Try again" |

### Behaviour

- Retry option with exponential backoff guidance.
- Offline indicator if applicable.
- Queue actions for sync when connection returns (future enhancement).

## Unexpected-error

### When to show

- API request fails with an unexpected server error (e.g. `500`, `503`).
- Unexpected exception occurs.
- General failure cases.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Error banner | `color-error-bg` background, `color-error` border, `color-error-fg` text |
| Error icon | Exclamation triangle, 20px, `color-error-fg` |
| Error message | `text-sm`, `color-text-primary`; generic message for unexpected cases |
| Retry button | Secondary button, "Try again" |
| Dismiss button | Icon button, `color-text-tertiary` |
| Request ID | `text-xs`, `color-text-tertiary` (for support) |

### Behaviour

- Clear error message explaining what went wrong.
- Actionable guidance for recovery where appropriate.
- Retry button where appropriate (network-error and transient server failures).
- Error code or request ID displayed for support (`requestId` from API).
- Non-revealing for security-sensitive errors (e.g., "Account not found" instead of "You don't own this account").
- Errors announced to screen readers (`aria-live="assertive"`).
- No internal details exposed.

## Success

### When to show

- Form submission succeeds (registration, login, transfer).
- Action completes (password change, logout).

### Visual pattern

| Element | Specification |
|---------|---------------|
| Success icon | Check circle, 24px, `color-success-fg` |
| Success message | `text-base`, `color-text-primary` |
| Contextual details | Transfer ID, new balance, etc. (`text-sm`, `color-text-secondary`) |
| Next action button | Primary button for next step |

### Behaviour

- Clear success message explaining what happened.
- Contextual feedback (what was completed, what's next).
- Action button for next step (e.g., "View transactions" after transfer).
- Auto-redirect where appropriate (e.g., after login, redirect to dashboard).
- Never show success before the server confirms it.



## Disabled

### When to show

- Action not currently available (e.g., transfer button disabled when insufficient funds).
- Form not yet valid (submit button disabled until all fields pass).

### Visual pattern

| Element | Specification |
|---------|---------------|
| Opacity | 0.5 |
| Cursor | not-allowed |
| Tooltip | Explains why disabled (e.g., "Insufficient funds for this transfer") |

### Behaviour

- Explains why the action is unavailable.
- Not clickable; no focus.
- Tooltip on hover/focus for screen readers.

## Partial Success

### When to show

- Some operations in a batch succeeded while others failed.
- Transfer completed but notification failed.

### Visual pattern

| Element | Specification |
|---------|---------------|
| Mixed icons | Check circle (success) + x circle (failure) |
| Message | "1 of 2 transfers completed" |
| Details | List of succeeded and failed items |

### Behaviour

- Clear distinction between succeeded and failed items.
- Actionable next steps for failed items.
- Overall status reflects partial completion.

## Cross-references

- UI/UX design index: [`07-ui-ux-design.md`](../../07-ui-ux-design.md#6-states)
- Feature-09 UI States: [`docs/02-features/FEATURE-09-UI-States-Global.md`](../../02-features/FEATURE-09-UI-States-Global.md)
- Error handling: [`FEATURE-08`](../../02-features/FEATURE-08-Error-Handling.md)
- Design tokens: [`design-tokens.md`](./design-tokens.md)
- Planned test: UI state tests in Vitest + Playwright E2E