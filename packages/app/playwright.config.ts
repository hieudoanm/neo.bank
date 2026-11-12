import { defineConfig } from '@playwright/test'

// The demo drives the real stack (docker compose) through the browser:
// the frontend at :3000 talks to the API gateway at :8080, which fans out to
// the Go services, Cassandra and Kafka. See ../../../../demo/demo.md.
const baseURL = process.env.DEMO_BASE_URL ?? 'http://localhost:3000'
const slowMo = Number(process.env.DEMO_SLOWMO ?? '300')

export default defineConfig({
  testDir: './e2e',
  outputDir: './test-results',
  fullyParallel: false,
  workers: 1,
  forbidOnly: Boolean(process.env.CI),
  retries: 0,
  timeout: 5 * 60_000,
  expect: { timeout: 30_000 },
  reporter: [['list']],
  use: {
    baseURL,
    viewport: { width: 1280, height: 720 },
    video: { mode: 'on', size: { width: 1280, height: 720 } },
    trace: 'retain-on-failure',
    launchOptions: { slowMo },
  },
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
})
