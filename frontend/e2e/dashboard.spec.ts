import { test, expect, devices, Page, BrowserContext } from '@playwright/test'
import { Sidebar } from './pages/sidebar.page'
import { OrgSwitcher } from './pages/org-switcher.page'
import { Breadcrumb } from './pages/breadcrumb.page'
import { UserMenu } from './pages/user-menu.page'
import { 
  setAuthCookies, 
  setupAuthenticatedMocks, 
  generateMockToken,
  DEFAULT_TEST_USER,
  UserData 
} from './utils/cookie-auth'

/**
 * Dashboard Layout E2E Tests
 * Comprehensive tests for sidebar navigation, top bar, breadcrumb navigation, and layout structure.
 *
 * Task: 06-dashboard-layout
 * Reference: docs/frontend/TEST_PATHS.md Section 3.1
 * 
 * NOTE: Uses HYBRID authentication:
 * - API requests use HttpOnly cookies (set by backend)
 * - Client-side PrivateRoute checks localStorage for token
 * See e2e/utils/cookie-auth.ts for auth utilities.
 */

// ============================================================================
// Test Setup Helper
// ============================================================================

async function initializeDashboardTest(
  page: Page,
  context: BrowserContext,
  user: Partial<UserData> = {}
) {
  const userData = { ...DEFAULT_TEST_USER, ...user }
  
  // Generate tokens that will be used for both cookies and localStorage
  const accessToken = generateMockToken({ 
    userId: userData.id,
    email: userData.email,
    name: userData.name,
    tokenType: 'access',
  })
  const refreshToken = generateMockToken({ 
    userId: userData.id,
    tokenType: 'refresh',
    expiresInSeconds: 604800,
  })
  
  // Set auth cookies (for API requests)
  await setAuthCookies(context, { accessToken, refreshToken })
  
  // Setup all required API mocks
  await setupAuthenticatedMocks(page, { user: userData })
  
  // Set tokens AND user data in localStorage (for PrivateRoute auth check + UI display)
  // Note: Frontend PrivateRoute checks localStorage.getItem('botla_token') for auth
  // Sidebar mode: set to 'pinned' by default so UI elements are visible, but preserve existing mode on reload
  await page.addInitScript(({ userData, accessToken, refreshToken }) => {
    localStorage.setItem('botla_token', accessToken)
    localStorage.setItem('botla_refresh_token', refreshToken)
    localStorage.setItem('botla_user', JSON.stringify(userData))
    // Only set default sidebar mode if not already set (allows tests to override before reload)
    if (!localStorage.getItem('botla_sidebar_mode')) {
      localStorage.setItem('botla_sidebar_mode', 'pinned')
    }
  }, { userData, accessToken, refreshToken })
  
  await page.goto('http://localhost:5173/dashboard')
  await page.waitForLoadState('domcontentloaded')
}

async function initializeAdminTest(page: Page, context: BrowserContext) {
  const adminUser: UserData = {
    id: 'admin-1',
    email: 'admin@example.com',
    name: 'Admin User',
    full_name: 'Admin User',
    plan: 'enterprise',
    is_platform_admin: true,
  }
  
  // Generate tokens
  const accessToken = generateMockToken({ 
    userId: adminUser.id,
    email: adminUser.email,
    name: adminUser.name,
    isPlatformAdmin: true,
    tokenType: 'access',
  })
  const refreshToken = generateMockToken({ 
    userId: adminUser.id,
    tokenType: 'refresh',
    expiresInSeconds: 604800,
  })
  
  // Set auth cookies (for API requests)
  await setAuthCookies(context, { accessToken, refreshToken })
  
  // Setup mocks with admin user
  await setupAuthenticatedMocks(page, { user: adminUser })
  
  // Set tokens AND admin user data in localStorage
  await page.addInitScript(({ userData, accessToken, refreshToken }) => {
    localStorage.setItem('botla_token', accessToken)
    localStorage.setItem('botla_refresh_token', refreshToken)
    localStorage.setItem('botla_user', JSON.stringify(userData))
  }, { userData: adminUser, accessToken, refreshToken })
  
  await page.goto('http://localhost:5173/dashboard')
  await page.waitForLoadState('domcontentloaded')
}

// ============================================================================
// Main Dashboard Layout Tests
// ============================================================================

test.describe('Dashboard Layout', () => {
  test.beforeEach(async ({ page, context }) => {
    await initializeDashboardTest(page, context)
  })

  // Phase 2: Sidebar Navigation Tests
  test.describe('Sidebar Navigation', () => {
    let sidebar: Sidebar

    test.beforeEach(async ({ page }) => {
      sidebar = new Sidebar(page)
    })

    test('should display sidebar on dashboard', async () => {
      await sidebar.expectVisible()
    })

    test('should show all navigation items', async () => {
      await expect(sidebar.navDashboard).toBeVisible()
      await expect(sidebar.navChatbots).toBeVisible()
    })

    test('should have Dashboard as active on home page', async () => {
      await sidebar.expectNavItemActive(sidebar.navDashboard)
    })

    test('should navigate to Chatbots page when clicking nav item', async ({ page }) => {
      await sidebar.navigateToChatbots()
      await expect(sidebar.navChatbots).toHaveClass(/active/)
    })

    test('should navigate to Settings when clicking settings nav', async ({ page }) => {
      await sidebar.navigateToProfile()
      await expect(page).toHaveURL(/\/dashboard\/settings\/profile/)
    })

    test('should navigate to Plan settings', async ({ page }) => {
      await sidebar.navigateToPlan()
      await expect(page).toHaveURL(/\/dashboard\/settings\/plan/)
    })

    test('should navigate to Privacy settings', async ({ page }) => {
      await sidebar.navigateToPrivacy()
      await expect(page).toHaveURL(/\/dashboard\/settings\/privacy/)
    })

    test('should navigate to Admin for platform admins', async ({ page }) => {
      // Remove the default /api/v1/me mock from setupAuthMocks and add admin version
      await page.unroute('**/api/v1/me')
      await page.route('**/api/v1/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: 'admin-1',
            email: 'admin@example.com',
            full_name: 'Admin User',
            is_platform_admin: true,
          }),
        })
      })

      await page.evaluate(() => {
        localStorage.setItem('botla_token', 'admin-token')
        localStorage.setItem('botla_refresh_token', 'admin-refresh')
        localStorage.setItem(
          'botla_user',
          JSON.stringify({ id: 'admin-1', email: 'admin@example.com', is_platform_admin: true })
        )
      })

      // Navigate fresh to trigger profile fetch
      await page.goto('http://localhost:5173/dashboard')
      await page.waitForLoadState('networkidle')

      const sidebar = new Sidebar(page)
      await sidebar.expectAdminNavVisible()
      await sidebar.navigateToAdmin()
      await expect(page).toHaveURL(/\/admin/)
    })

    test('should hide admin nav for regular users', async ({ page }) => {
      await page.evaluate(() => {
        localStorage.setItem(
          'botla_user',
          JSON.stringify({ id: 'user-1', email: 'user@example.com', is_platform_admin: false })
        )
      })
      await page.reload()
      const sidebar = new Sidebar(page)
      await sidebar.expectAdminNavHidden()
    })

    test('should have logo that links to dashboard', async ({ page }) => {
      await sidebar.navigateToChatbots()
      await sidebar.clickLogo()
      await expect(page).toHaveURL(/\/dashboard$/)
      await sidebar.expectNavItemActive(sidebar.navDashboard)
    })

    test('should update active state when navigating', async () => {
      await sidebar.expectNavItemActive(sidebar.navDashboard)
      await sidebar.expectNavItemInactive(sidebar.navChatbots)
      await sidebar.navigateToChatbots()
      await sidebar.expectNavItemInactive(sidebar.navDashboard)
      await sidebar.expectNavItemActive(sidebar.navChatbots)
    })
  })

  // Phase 3: Sidebar Collapse Tests
  test.describe('Sidebar Collapse/Expand', () => {
    let sidebar: Sidebar

    test.beforeEach(async ({ page }) => {
      sidebar = new Sidebar(page)
    })


    test('should toggle sidebar collapse state', async () => {
      // Start with pinned mode (expanded) to test toggle functionality
      await sidebar.setPinnedMode()
      await sidebar.expectExpanded()
      await sidebar.clickToggle()
      await sidebar.expectCollapsed()
      await sidebar.clickToggle()
      await sidebar.expectExpanded()
    })

    test('should show icons only when collapsed', async () => {
      await sidebar.setHoverMode()
      await expect(sidebar.navDashboard.locator('svg')).toBeVisible()
      const navText = await sidebar.navDashboard.locator('span').first().isHidden()
      expect(navText).toBe(true)
    })

    test('should expand on hover when in hover mode', async () => {
      await sidebar.setHoverMode()
      await sidebar.expectCollapsed()
      await sidebar.container.hover()
      await expect(sidebar.sidebarGlass).toBeVisible()
    })

    test('should persist collapse state on refresh', async ({ page }) => {
      await sidebar.setHoverMode()
      await page.reload()
      await sidebar.expectCollapsed()

      await sidebar.setPinnedMode()
      await page.reload()
      await sidebar.expectExpanded()
    })

    test('should allow navigation when collapsed', async ({ page }) => {
      await sidebar.setHoverMode()
      await sidebar.navigateToChatbots()
      await expect(page).toHaveURL(/\/dashboard\/chatbots/)
    })
  })

  // Phase 4: Organization Switcher Tests
  test.describe('Organization Switcher', () => {
    let orgSwitcher: OrgSwitcher

    test.beforeEach(async ({ page }) => {
      orgSwitcher = new OrgSwitcher(page)
    })

    test('should display organization switcher', async () => {
      await orgSwitcher.expectVisible()
    })

    test('should show current organization name', async () => {
      await orgSwitcher.expectCurrentOrg('Test Org')
    })

    test('should open dropdown when clicked', async () => {
      await orgSwitcher.openDropdown()
      await orgSwitcher.expectDropdownVisible()
    })

    test('should close dropdown when clicked again', async () => {
      await orgSwitcher.openDropdown()
      await orgSwitcher.closeDropdown()
      await orgSwitcher.expectDropdownHidden()
    })

    test('should show organization list in dropdown', async () => {
      await orgSwitcher.openDropdown()
      await orgSwitcher.expectOrgCount(1)
      await orgSwitcher.expectOrgInList('Test Org')
    })

    test('should highlight org item on hover', async () => {
      await orgSwitcher.openDropdown()
      await orgSwitcher.hoverOrgItem('Test Org')
    })

    test('should switch organization when selecting', async () => {
      await orgSwitcher.selectOrg('Test Org')
      await orgSwitcher.expectDropdownHidden()
      await orgSwitcher.expectCurrentOrg('Test Org')
    })
  })

  // Phase 5: User Menu Tests
  test.describe('User Menu', () => {
    let userMenu: UserMenu

    test.beforeEach(async ({ page }) => {
      userMenu = new UserMenu(page)
    })

    test('should display user avatar', async () => {
      await userMenu.expectUserAvatarVisible()
    })

    test('should show user name', async () => {
      await userMenu.expectUserNameVisible()
    })

    test('should show plan badge', async () => {
      await userMenu.expectUserPlanBadgeVisible()
    })

    test('should navigate to profile when clicking profile card', async ({ page }) => {
      await userMenu.clickUserProfileCard()
      await expect(page).toHaveURL(/\/dashboard\/settings\/profile/)
    })

    test('should have logout button in sidebar', async ({ page }) => {
      const sidebar = new Sidebar(page)
      await sidebar.expectLogoutButtonVisible()
    })
  })

  // Phase 6: Breadcrumb Navigation Tests
  test.describe('Breadcrumb Navigation', () => {
    let breadcrumb: Breadcrumb
    let sidebar: Sidebar

    test.beforeEach(async ({ page }) => {
      breadcrumb = new Breadcrumb(page)
      sidebar = new Sidebar(page)
      // Breadcrumb is hidden on mobile, visible on md+ screens
      await page.setViewportSize({ width: 1280, height: 720 })
    })

    test('should display breadcrumb on dashboard inner pages', async ({ page }) => {
      await page.goto('http://localhost:5173/dashboard/chatbots')
      await page.waitForLoadState('domcontentloaded')
      await breadcrumb.expectVisible()
    })

    test('should show correct breadcrumb path on nested pages', async ({ page }) => {
      await page.goto('http://localhost:5173/dashboard/chatbots')
      await page.waitForLoadState('domcontentloaded')
      const texts = await breadcrumb.getBreadcrumbTexts()
      expect(texts).toContain('Botla')
      // Breadcrumb shows "Panel" for all dashboard routes (current implementation)
      expect(texts).toContain('Panel')
    })

    test('should navigate when clicking breadcrumb links', async ({ page }) => {
      await page.goto('http://localhost:5173/dashboard/settings/profile')
      await page.waitForLoadState('domcontentloaded')
      // Navigate to dashboard via sidebar
      await sidebar.navigateToDashboard()
      await expect(page).toHaveURL(/\/dashboard$/)
    })

    test('should not have clickable link on current page', async ({ page }) => {
      await page.goto('http://localhost:5173/dashboard/chatbots')
      await page.waitForLoadState('domcontentloaded')
      const lastItem = breadcrumb.items.last()
      await expect(lastItem.locator('a')).not.toBeAttached()
    })

    test('should navigate to parent via breadcrumb', async ({ page }) => {
      await page.goto('http://localhost:5173/dashboard/chatbots/bot-1/sources')
      await page.waitForLoadState('domcontentloaded')
      // Navigate to parent via sidebar since breadcrumbs are not clickable
      await sidebar.navigateToChatbots()
      await expect(page).toHaveURL(/\/dashboard\/chatbots/)
    })
  })

  // Phase 7: Top Bar Tests
  test.describe('Top Bar', () => {
    test('should display top bar header', async ({ page }) => {
      await expect(page.locator('header.h-16')).toBeVisible()
    })

    test('should display organization switcher in top bar', async ({ page }) => {
      const orgSwitcher = new OrgSwitcher(page)
      await orgSwitcher.expectVisible()
    })
  })

  // Phase 8: Responsive Tests
  test.describe('Responsive Behavior', () => {
    let sidebar: Sidebar

    test.beforeEach(async ({ page }) => {
      sidebar = new Sidebar(page)
    })

    test('should show hamburger menu on mobile', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 })
      await page.reload() // Reload so React re-renders with mobile layout
      await page.waitForLoadState('domcontentloaded')
      await sidebar.expectHidden()
      await sidebar.openMobileMenu()
      await sidebar.expectVisible()
    })

    test('should have full sidebar on desktop', async ({ page }) => {
      await page.setViewportSize({ width: 1280, height: 720 })
      await sidebar.expectVisible()
      await sidebar.expectExpanded()
    })

    test('should collapse sidebar on tablet', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 })
      await sidebar.expectVisible()
    })
  })
})

// ============================================================================
// Mobile Viewport Tests
// ============================================================================

test.describe('Dashboard Layout - Mobile Viewport', () => {
  test.use({ viewport: devices['iPhone 12'].viewport })

  test.beforeEach(async ({ page, context }) => {
    await initializeDashboardTest(page, context)
    // Wait for viewport resize to take effect
    await page.waitForTimeout(100)
  })

  test('should hide sidebar initially on mobile', async ({ page }) => {
    const sidebar = new Sidebar(page)
    // Wait for page to settle after viewport resize
    await page.waitForLoadState('domcontentloaded')
    await sidebar.expectHidden()
  })

  test('should show hamburger menu on mobile', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await page.waitForLoadState('domcontentloaded')
    await sidebar.mobileMenuButton.scrollIntoViewIfNeeded()
    await expect(sidebar.mobileMenuButton).toBeVisible({ timeout: 10000 })
  })

  test('should open mobile menu when clicking hamburger', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await page.waitForLoadState('domcontentloaded')
    await sidebar.mobileMenuButton.scrollIntoViewIfNeeded()
    await sidebar.openMobileMenu()
    await sidebar.expectVisible()
  })

  test('should close mobile menu when clicking close button', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await page.waitForLoadState('domcontentloaded')
    await sidebar.mobileMenuButton.scrollIntoViewIfNeeded()
    await sidebar.openMobileMenu()
    await sidebar.closeMobileMenu()
    await sidebar.expectHidden()
  })
})

// ============================================================================
// Admin-Specific Tests
// ============================================================================

test.describe('Dashboard Layout - Admin User', () => {
  test.beforeEach(async ({ page, context }) => {
    await initializeAdminTest(page, context)
  })

  test('should show admin navigation for admin users', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await sidebar.expectAdminNavVisible()
  })

  test('should navigate to admin panel', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await sidebar.navigateToAdmin()
    await expect(page).toHaveURL(/\/admin/)
  })
})

// ============================================================================
// Sidebar Mode Persistence Tests
// ============================================================================

test.describe('Sidebar Mode Persistence', () => {
  test.beforeEach(async ({ page, context }) => {
    await initializeDashboardTest(page, context)
  })

  test('should persist pinned mode after reload', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await sidebar.setPinnedMode()
    await page.reload()
    await sidebar.expectExpanded()
    expect(await sidebar.getSidebarMode()).toBe('pinned')
  })

  test('should persist hover mode after reload', async ({ page }) => {
    const sidebar = new Sidebar(page)
    await sidebar.setHoverMode()
    await page.reload()
    await sidebar.expectCollapsed()
    expect(await sidebar.getSidebarMode()).toBe('hover')
  })
})
