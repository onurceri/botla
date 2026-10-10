import { defineConfig } from '@playwright/test'

// README screenshots. Run `make screenshots` (seeds first) against a running
// stack: API on :8080, dashboard on :5173, widget dev server on :5174.
export default defineConfig({
  testDir: '.',
  testMatch: 'screenshots.spec.ts',
  timeout: 360_000,
  workers: 1,
  reporter: 'list',
  outputDir: '../../test-results/screenshots',
  use: {
    baseURL: process.env.DASHBOARD_URL || 'http://localhost:5173',
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 1.5,
    colorScheme: 'light',
    locale: 'en-US',
    // Set PW_CHANNEL=chrome to use an installed Chrome instead of `npx playwright install chromium`.
    channel: process.env.PW_CHANNEL || undefined,
    // The sample host page is served through page.route(); newer Chrome would
    // otherwise block its calls to the local API (Local Network Access checks).
    launchOptions: { args: ['--disable-features=LocalNetworkAccessChecks'] },
  },
})
