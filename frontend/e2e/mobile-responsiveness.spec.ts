import { test, expect } from '@playwright/test'

test.use({ viewport: { width: 375, height: 812 } }) // iPhone X

/**
 * Mobile Responsiveness E2E Tests
 * 
 * Tests cover:
 * - Mobile viewport adaptation
 * - Bottom navigation visibility
 * - Sidebar behavior on mobile
 * - Hamburger menu functionality
 * 
 * Element Selection Strategy:
 * - Primary: CSS class selectors for responsive components
 * - Fallback: Semantic selectors
 * 
 * @see TESTING_STANDARDS.md for naming conventions
 * @see selectors.ts for element ID constants
 */
test.describe('Mobile Responsiveness', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API
    await page.route('**/api/v1/auth/me', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ plan: 'pro' }),
      })
    })
    await page.route('**/api/v1/organizations', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ id: 'org-1', name: 'Test Org' }]),
      })
    })
    await page.route('**/api/v1/organizations/org-1/workspaces', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ id: 'ws-1', name: 'Test WS' }]),
      })
    })
    await page.route('**/api/v1/chatbots/1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: '1', name: 'Test Bot' }),
      })
    })
    // Mock analytics for dashboard
    await page.route('**/api/v1/analytics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([]),
      })
    })
    // Mock chatbots list
    await page.route('**/api/v1/chatbots', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([]),
      })
    })
  })

  test('Bottom navigation bar appears on mobile in Chatbot Detail', async ({ page }) => {
    await page.goto('/dashboard/chatbots/1/settings')

    // Check for the bottom navigation bar
    // The component has a class "fixed bottom-0" and "lg:hidden"
    const bottomNav = page.locator('nav.fixed.bottom-0')
    await expect(bottomNav).toBeVisible()

    // Check if "Ayarlar" is visible in it
    await expect(bottomNav.getByText('Ayarlar')).toBeVisible()
    await expect(bottomNav.getByText('Raporlar')).toBeVisible()
  })

  test('Dashboard navigation adapts to mobile', async ({ page }) => {
    await page.goto('/dashboard')
    // Verify sidebar is hidden (hamburger menu should be visible)
    // The hamburger menu is a button with a Menu icon
    // In DashboardLayout.tsx: <button className="lg:hidden ..." onClick={() => setIsMobileMenuOpen(true)}> <Menu ... /> </button>

    // We can just check for the Menu icon
    await expect(page.locator('.lucide-menu')).toBeVisible()

    // Sidebar should be hidden initially
    const sidebar = page.locator('aside')
    // It has -translate-x-full class
    await expect(sidebar).toHaveClass(/-translate-x-full/)
  })
})
