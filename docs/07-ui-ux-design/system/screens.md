# NEO Bank — UI/UX Design — Screen Designs

_Part of [07-ui-ux-design.md](../../07-ui-ux-design.md)._

## Purpose

Defines the design of each screen: Login, Register, Dashboard (includes balance), Transfer, Transaction History, and Transfer Status.

## Login Screen (Login)

**Layout:**

- Centered card on page
- Card width: 400px max
- Card padding: `space-8`

**Content:**

- Page heading: "Sign in to NEO Bank"
- Email field
- Password field
- "Forgot password?" link (future enhancement)
- "Sign in" button (primary, full width)
- "Don't have an account? Sign up" link

**States:**

- loading: button shows spinner, disabled
- validation-error: field-specific error messages
- network-error: alert with retry guidance
- unexpected-error: alert above the form with an error message
- success: redirect to dashboard

## Register Screen (Register)

**Layout:**

- Centered card on page
- Card width: 400px max
- Card padding: `space-8`

**Content:**

- Page heading: "Create your account"
- Email field
- Password field
- Confirm password field
- Password strength indicator
- "Create account" button (primary, full width)
- "Already have an account? Sign in" link

**States:**

- loading: button shows spinner, disabled
- validation-error: field-specific error messages (email format, password strength)
- network-error: alert with retry guidance
- unexpected-error: alert above the form with an error message
- success: redirect to dashboard

## Dashboard Screen (Dashboard, includes balance)

**Layout:**

- Two-column grid (desktop), single column (mobile)
- Left column: balance card, quick actions
- Right column: recent transactions

**Balance card:**

- Card with shadow
- Label: "Current balance"
- Balance amount: large text, `text-3xl`, bold
- Currency: smaller text beside amount
- Last updated: `text-sm`, `color-text-secondary`

**Quick actions:**

- "Transfer money" button (primary)
- "View transactions" link

**Recent transactions:**

- Card with shadow
- Heading: "Recent transactions"
- Transaction list (max 5)
- Each transaction: type badge, amount, date, counterparty
- "View all transactions" link

**States:**

- loading: skeleton loaders for balance and transactions
- empty: "No transactions yet" message
- network-error: alert with error message and retry button
- unexpected-error: alert with error message and retry button

## Transfer Screen

**Layout:**

- Centered card on page
- Card width: 500px max
- Card padding: `space-8`

**Content:**

- Page heading: "Transfer money"
- Destination account ID field
- Amount field (with currency symbol)
- "Available balance" helper text
- "Transfer" button (primary, full width)
- "Cancel" link

**States:**

- loading: button shows spinner, disabled
- validation-error: field-specific error messages
- network-error: alert with retry guidance
- unexpected-error: alert with error message
- success: redirect to transfer status screen

## Transfer Status Screen (Transfer Status)

**Layout:**

- Centered card on page
- Card width: 500px max
- Card padding: `space-8`

**Content:**

- Page heading: "Transfer status"
- Status badge (pending, processing, completed, failed, reversed)
- Transfer details:
  - Transfer ID
  - Destination account ID
  - Amount
  - Created at
  - Updated at
- Failure reason (if failed)
- "Back to dashboard" button

**States:**

- loading: spinner
- network-error: alert with retry guidance
- unexpected-error: alert with error message
- success: status badge green
- failed: status badge red, failure reason displayed
- polling: status updates automatically every 2 seconds

## Transaction History Screen (Transaction History)

**Layout:**

- Page heading: "Transaction history"
- Filter controls (future enhancement)
- Transaction list
- Pagination controls

**Transaction list item:**

- Type badge (debit/credit)
- Amount (debit: `color-error-fg`, credit: `color-success-fg`)
- Counterparty account ID
- Date
- Hover: highlight row (`color-background-secondary`)

**States:**

- loading: skeleton loaders
- empty: "No transactions" message
- network-error: alert with error message and retry button
- unexpected-error: alert with error message and retry button
- loading-more: spinner at bottom of list