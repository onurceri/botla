import { test, expect, type Page, BrowserContext } from '@playwright/test'
import { setAuthCookies, setupAuthenticatedMocks, DEFAULT_TEST_USER } from './utils/cookie-auth'
import { setupOrgMocks, setupAnalyticsMocks } from './helpers'

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
 * - Primary: CSS class selectors for responsive behavior (checking transform classes, viewport-dependent elements)
 * - Fallback: Semantic selectors (getByRole, getByText)
 * - Note: Responsive tests require checking CSS classes like -translate-x-full for sidebar state
 *
 * @see TESTING_STANDARDS.md for naming conventions
 * @see selectors.ts for element ID constants
 */
test.describe('Mobile Responsiveness', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set up HttpOnly cookies and mock auth endpoints
    await setAuthCookies(context, DEFAULT_TEST_USER)
    await setupAuthenticatedMocks(page)
    await setupOrgMocks(page)
    await setupAnalyticsMocks(page)
    
    // Mock chatbot detail endpoint
    await page.route('**/api/v1/chatbots/1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: '1', name: 'Test Bot' }),
      })
    })
  })

  test('should show bottom navigation bar on mobile in Chatbot Detail', async ({ page }) => {
    await page.goto('http://localhost:5173/dashboard/chatbots/1/settings')

    // Check for the bottom navigation bar
    // The component has a class "fixed bottom-0" and "lg:hidden"
    const bottomNav = page.locator('nav.fixed.bottom-0')
    await expect(bottomNav).toBeVisible()

    // Check if navigation items are visible (be flexible about which items exist)
    const navItems = bottomNav.locator('a, button')
    const itemCount = await navItems.count()
    expect(itemCount).toBeGreaterThan(0)
  })

  test('should adapt dashboard navigation for mobile viewport', async ({ page }) => {
    await page.goto('http://localhost:5173/dashboard')

    // Verify sidebar is hidden (hamburger menu should be visible)
    // The hamburger menu is a button with a Menu icon
    // In DashboardLayout.tsx: <button className="lg:hidden ..." onClick={() => setIsMobileMenuOpen(true)}>

    // Check for hamburger menu icon
    await expect(page.locator('.lucide-menu')).toBeVisible()

    // Sidebar should be hidden initially (has -translate-x-full class)
    const sidebar = page.locator('aside')
    await expect(sidebar).toHaveClass(/-translate-x-full/)
  })
})
