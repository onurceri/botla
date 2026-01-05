import { test as base, Page } from '@playwright/test'
import { UserMenu } from '../pages/user-menu.page'
import { setValidSession, clearSessionStorage, setExpiredToken } from '../utils/session.utils'

// Test fixtures interface
interface LogoutFixtures {
  // Page objects
  userMenu: UserMenu

  // Auth state helpers
  authenticatedPage: Page
  expiredSessionPage: Page

  // Session test data
  mockAccessToken: string
  mockRefreshToken: string
}

// Custom test implementation
export const test = base.extend<LogoutFixtures>({
  // UserMenu fixture - creates and provides UserMenu instance
  userMenu: async ({ page }: { page: Page }, use: (userMenu: UserMenu) => Promise<void>) => {
    const userMenu = new UserMenu(page)
    await use(userMenu)
  },

  // Authenticated page fixture - navigates to dashboard with valid session
  authenticatedPage: async ({ page }: { page: Page }, use: (page: Page) => Promise<void>) => {
    await page.goto('/dashboard')
    await setValidSession(page)
    await page.waitForLoadState('networkidle')
    await use(page)
    // Cleanup
    await clearSessionStorage(page)
  },

  // Expired session page fixture - navigates with expired token
  expiredSessionPage: async ({ page }: { page: Page }, use: (page: Page) => Promise<void>) => {
    await page.goto('/dashboard')
    await setExpiredToken(page)
    await page.waitForLoadState('networkidle')
    await use(page)
    // Cleanup
    await clearSessionStorage(page)
  },

  // Mock tokens for testing
  mockAccessToken: async () => 'mock-access-token-' + Date.now(),

  mockRefreshToken: async () => 'mock-refresh-token-' + Date.now(),
})

// Re-export for convenience
export { expect } from '@playwright/test'
