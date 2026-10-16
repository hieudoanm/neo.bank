import { test, expect, type Page } from '@playwright/test'

// A single end-to-end journey through the NEO Bank UI:
//   register a recipient -> register a sender -> transfer funds ->
//   watch the async saga complete -> verify both balances and histories.
//
// It relies on the stack being up (docker compose) with the demo opening
// balance configured (INITIAL_BALANCE_CENTS=100000 -> $1,000.00). See
// ../../../../demo/demo.md.

const PASSWORD = 'Demo-Passw0rd!'
const OPENING_BALANCE = '$1,000.00'
const AMOUNT_INPUT = '25.00'
const DEBIT_LINE = '-$25.00'
const CREDIT_LINE = '+$25.00'
const SENDER_BALANCE = '$975.00'
const RECIPIENT_BALANCE = '$1,025.00'

const uniqueEmail = (role: string): string =>
  `${role}.${Date.now().toString(36)}.${Math.random().toString(36).slice(2, 8)}@example.com`

const register = async (page: Page, email: string, password: string): Promise<void> => {
  await expect(page.getByRole('heading', { name: 'Create your account' })).toBeVisible()
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByLabel('Confirm password').fill(password)
  await page.getByRole('button', { name: 'Create account' }).click()
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
}

const signIn = async (page: Page, email: string, password: string): Promise<void> => {
  await expect(page.getByRole('heading', { name: 'Sign in to NEO Bank' })).toBeVisible()
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
}

const readAccountId = async (page: Page): Promise<string> => {
  const line = await page.getByText(/^Account: [0-9a-f-]{36}$/).innerText()
  return line.replace(/^Account:\s*/, '').trim()
}

test('a customer can send money to another account end to end', async ({ page }) => {
  const recipientEmail = uniqueEmail('recipient')
  const senderEmail = uniqueEmail('sender')

  // 1. Register the recipient. The balance appearing proves the account was
  //    provisioned from the auth.user.created event before we transfer to it.
  await page.goto('/register')
  await register(page, recipientEmail, PASSWORD)
  await expect(page.getByText(OPENING_BALANCE)).toBeVisible()
  const recipientAccountId = await readAccountId(page)

  // 2. Register the sender. Registration signs the sender in and provisions an
  //    account funded with the configured opening balance.
  await page.goto('/register')
  await register(page, senderEmail, PASSWORD)
  const senderAccountId = await readAccountId(page)
  expect(senderAccountId).not.toEqual(recipientAccountId)
  await expect(page.getByText(OPENING_BALANCE)).toBeVisible()

  // 3. Send $25.00 to the recipient.
  await page.getByRole('link', { name: 'Transfer' }).click()
  await page.getByLabel('Destination account ID').fill(recipientAccountId)
  await page.getByLabel('Amount').fill(AMOUNT_INPUT)
  await page.getByRole('button', { name: 'Confirm transfer' }).click()

  // 4. The Transfer Service saga runs asynchronously; the page polls until done.
  await expect(page.getByRole('heading', { name: 'Transfer complete' })).toBeVisible({
    timeout: 120_000,
  })

  // 5. The balance reflects the debit.
  await page.getByRole('button', { name: 'Back to dashboard' }).click()
  await expect(page.getByText(SENDER_BALANCE)).toBeVisible()

  // 6. The debit appears in the sender's transaction history.
  await page.getByRole('link', { name: 'Transactions' }).click()
  await expect(page.getByText(DEBIT_LINE)).toBeVisible()

  // 7. Sign in as the recipient and confirm the credit landed.
  await page.goto('/login')
  await signIn(page, recipientEmail, PASSWORD)
  await expect(page.getByText(RECIPIENT_BALANCE)).toBeVisible()
  await page.getByRole('link', { name: 'Transactions' }).click()
  await expect(page.getByText(CREDIT_LINE)).toBeVisible()
})
