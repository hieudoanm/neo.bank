# FEATURE-09: UI States (Global)

## Priority

P0

## Description

This feature is cross-cutting: it defines the global UI state model that every screen and feature (FEATURE-01..FEATURE-08) must implement. Because it applies to all features, it has no Inputs, Outputs, or API Contract sections of its own; each state is driven by the canonical API responses (success payloads and the error envelope) defined elsewhere.

All primary screens implement a consistent set of UI states: loading (skeletons), empty (with next action), success, validation_error (safe), permission_error (non-revealing), network_error (retry), unexpected_error (safe message + request ID), partial_success, and disabled (explain why). Every data screen implements `loading`, `empty`, `validation_error`, `permission_error`, `network_error`, and `unexpected_error` states as mandated by the frontend conventions.

## User-visible behaviour

- **Loading states** (`loading`): Show skeletons or spinners while data is being fetched; no content jump when data arrives.
- **Empty states** (`empty`): Guide users to take the next action (e.g., "No transfers yet — make your first transfer").
- **Success states** (`success`): Confirm completed actions with relevant details (e.g., transfer ID, new balance).
- **Validation errors** (`validation_error`): Highlight specific fields with actionable feedback (e.g., "Password must be at least 12 characters").
- **Permission errors** (`permission_error`): Show generic messages to prevent information leakage (e.g., "Account not found" instead of "You don't own this account"; the API returns `404`, never `403`).
- **Network errors** (`network_error`): Offer retry options and explain connectivity issues.
- **Unexpected errors** (`unexpected_error`): Show generic server-failure messages with request IDs for support tracking.
- **Partial success** (`partial_success`): Indicate some operations succeeded while others failed (e.g., transfer completed but notification failed).
- **Disabled states** (`disabled`): Explain why an action is not currently available (e.g., "Transfer disabled — insufficient funds").

## States

| State | Trigger | User Action | Visual Pattern |
|-------|---------|-------------|----------------|
| `loading` | Data fetch in progress | Wait | Skeleton screens or spinner |
| `empty` | No data to display | Take suggested action | Empty illustration + primary button |
| `success` | Action completed | Acknowledge | Checkmark icon + confirmation message |
| `validation_error` | Invalid input | Correct field | Inline field error + red border |
| `permission_error` | Unauthorised access (API returns `404`, never `403`) | None (generic) | Generic message, no details |
| `network_error` | Connectivity failure | Retry | Error banner + retry button |
| `unexpected_error` | Backend failure (e.g., `500 INTERNAL_ERROR`, `503 SERVICE_UNAVAILABLE`) | Retry or contact support | Error banner + request ID |
| `partial_success` | Some operations succeeded | Review results | Mixed success/failure indicators |
| `disabled` | Action unavailable | None (explain why) | Disabled button + tooltip |

`permission_error` is the single state for all access-restriction cases (there is no separate permission-denied state), and `unexpected_error` is the single state for server-side failures (there is no separate server-error state).

## Acceptance conditions

- No internal paths, stack traces, or sensitive data leaked to the UI.
- States distinguish recoverable (retryable) vs non-recoverable errors.
- Focus management is predictable; error messages are accessible (ARIA announcements).
- Loading states use skeletons or spinners (not blank screens).
- Empty states guide the user to a next action (never a dead end).
- Success states confirm what was completed (transfer ID, new balance, etc.).
- Validation errors highlight specific fields with actionable feedback.
- Permission errors show generic messages (e.g., "Account not found") to prevent enumeration.
- Network errors offer retry options with exponential backoff guidance.
- Unexpected (server) errors show generic messages with request IDs for support tracking.
- Every data screen implements: `loading`, `empty`, `validation_error`, `permission_error`, `network_error`, `unexpected_error`.
- Never show success before the server confirms it.
- Hide unavailable actions for usability, but treat the server as the only authority.

## Cross-references

- UI/UX design: [`docs/07-ui-ux-design.md`](../07-ui-ux-design.md#6-states)
- States design: [`docs/07-ui-ux-design/system/states.md`](../07-ui-ux-design/system/states.md)
- Frontend conventions: `AGENTS.md`
- Error handling: [`FEATURE-08`](FEATURE-08-Error-Handling.md)
- Planned test: UI state tests in Vitest + Playwright E2E