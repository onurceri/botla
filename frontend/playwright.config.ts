import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  fullyParallel: true,
  reporter: [['list']],
  webServer: process.env.E2E_BASE_URL ? undefined : {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
    timeout: 60_000,
    env: {
      ...(process.env.E2E_API_BASE ? { VITE_API_BASE_URL: process.env.E2E_API_BASE } : {}),
      VITE_E2E: '1',
    },
  },
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    storageState: {
      origins: [
        {
          origin: 'http://localhost:5173',
          localStorage: [
            { name: 'botla_token', value: 'tok' },
            { name: 'botla_refresh_token', value: 'ref' },
          ],
        },
      ],
    },
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
