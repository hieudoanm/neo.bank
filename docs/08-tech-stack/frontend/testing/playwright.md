# NEO Bank — Playwright Best Practices

_Frontend · Testing · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [testing index](./README.md)._

## 1. Configuration

End-to-end tests run the real browser against the full stack (frontend + API + Cassandra + Kafka).

```ts
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['html'], ['github']] : 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile',
      use: { ...devices['Pixel 5'] },
    },
  ],
  webServer: process.env.CI
    ? undefined
    : {
        command: 'pnpm dev',
        port: 3000,
        reuseExistingServer: !process.env.CI,
      },
});
```

Rules:

- One browser project in CI (Chromium) for speed; mobile project added for breakpoint coverage.
- `baseURL` comes from the environment so the same suite runs locally and in CI.
- Retries are a CI safety net, not a fix for flakiness — flaky specs are repaired, not retried into green.
- `trace: 'on-first-retry'` captures traces only on retry, keeping CI output manageable.
- `screenshot: 'only-on-failure'` captures screenshots on test failure for diagnosis.

## 2. Authentication State

Authenticate once, reuse the session; do not log in through the UI in every spec.

```ts
// e2e/auth.setup.ts
import { test as setup, expect } from '@playwright/test';

setup('authenticate', async ({ page }) => {
  await page.goto('/login');
  await page.getByLabel('Email').fill(process.env.E2E_EMAIL!);
  await page.getByLabel('Password').fill(process.env.E2E_PASSWORD!);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await expect(page).toHaveURL('/');
  await page.context().storageState({ path: 'e2e/.auth/user.json' });
});
```

```ts
// playwright.config.ts (projects using auth)
projects: [
  {
    name: 'authenticated',
    testMatch: /.*\.auth\.spec\.ts/,
    use: {
      ...devices['Desktop Chrome'],
      storageState: 'e2e/.auth/user.json',
    },
    dependencies: ['authenticate'],
  },
  {
    name: 'unauthenticated',
    testMatch: /.*\.unauth\.spec\.ts/,
    use: { ...devices['Desktop Chrome'], storageState: { cookies: [], origins: [] } },
  },
],
```

Rules:

- A setup project performs login and writes `storageState`; dependent projects reuse it.
- Credentials come from environment variables, never committed.
- Tests that assert unauthenticated behaviour use a fresh, empty state (no `storageState`).
- `e2e/.auth/` is git-ignored.

## 3. Test Structure & Fixtures

```ts
// e2e/transfer.auth.spec.ts
import { test, expect } from '@playwright/test';

test.describe('transfer flow', () => {
  test('completes a transfer end to end', async ({ page }) => {
    await page.goto('/transfer');
    await page.getByLabel('Destination account ID').fill('acc-test-2');
    await page.getByLabel('Amount (cents)').fill('10000');
    await page.getByRole('button', { name: 'Transfer' }).click();

    await expect(page.getByTestId('transfer-status')).toHaveText('pending');
    await expect(page.getByTestId('transfer-status')).toHaveText('completed', { timeout: 10_000 });
    await expect(page.getByTestId('balance')).toContainText('900.00');
  });

  test('rejects insufficient funds', async ({ page }) => {
    await page.goto('/transfer');
    await page.getByLabel('Destination account ID').fill('acc-test-2');
    await page.getByLabel('Amount (cents)').fill('999999999');
    await page.getByRole('button', { name: 'Transfer' }).click();

    await expect(page.getByText('Insufficient funds')).toBeVisible();
  });
});
```

Rules:

- Group related specs with `test.describe`; name tests as user outcomes.
- Extend `test` with custom fixtures for seeded accounts rather than repeating setup.
- Prefer role/label locators; use `data-testid` only where no accessible locator exists.
- Test file naming: `*.auth.spec.ts` (authenticated) or `*.unauth.spec.ts` (unauthenticated).

## 4. Critical Flows

Cover the journeys in `09-testing-strategy.md`:

| Journey | Steps |
|---------|-------|
| Core banking | Register → login → view balance → transfer → view transaction history |
| Transfer lifecycle | Login → transfer → monitor status → view completed transfer |
| Transaction history | Login → transaction history → paginate |
| Negative auth | Unauthenticated access redirects to login |
| Negative authorisation | Another account's data is not shown (404) |

Rules:

- E2E covers critical paths only; exhaustive state coverage belongs in Vitest.
- Always include the negative authorisation path.
- Synthetic accounts and users only; never real financial data.

## 5. Async, Polling & Assertions

- Use web-first assertions (`await expect(locator).toHaveText(...)`) that auto-retry.
- Never `page.waitForTimeout()`; wait on observable state.
- For asynchronous transfers, assert the terminal status with an explicit `timeout`.
- Assert balance changes against seeded starting balances so the arithmetic is deterministic.
- Use `test.step()` to break long flows into named steps for better trace reporting.

```ts
test.step('initiate transfer', async () => {
  await page.getByLabel('Destination account ID').fill('acc-test-2');
  await page.getByLabel('Amount (cents)').fill('10000');
  await page.getByRole('button', { name: 'Transfer' }).click();
});

test.step('monitor status', async () => {
  await expect(page.getByTestId('transfer-status')).toHaveText('completed', { timeout: 10_000 });
});
```

## 6. Test Data

- Synthetic accounts and users only; never real financial data (`09-testing-strategy.md`).
- Seed data through an API or fixture, not by clicking through unrelated screens.
- Make each spec independent; clean up or use uniquely-scoped data so parallel runs do not collide.
- Use unique emails per test run to avoid conflicts with existing users.

## 7. Debugging & CI

- Traces and screenshots are captured on failure; open the HTML report to diagnose.
- Run `--headed` and `--ui` locally when developing a spec.
- CI installs browsers with `pnpm exec playwright install --with-deps` and uploads the report artifact.
- E2E is the last CI gate and targets under 10 minutes.

## 8. Code Style & Rules

- One browser project in CI; environment-driven `baseURL`.
- Reuse authenticated state via `storageState`; never re-login through the UI per test.
- Role/label locators first; web-first assertions; no fixed sleeps.
- Synthetic, independent test data; negative authorisation path always covered.
- Failures are fixed, not retried into green.
- Related: [`vitest.md`](./vitest.md) for unit/component tests, [`react-router.md`](../core/react-router.md) for routes under test.
