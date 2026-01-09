import { Locator, Page, expect } from '@playwright/test'

/**
 * Dashboard Sidebar Page Object
 * Handles all sidebar interactions including navigation, collapse/expand, and user avatar.
 */
export class Sidebar {
  readonly page: Page

  // Main sidebar container
  readonly container: Locator
  readonly sidebarGlass: Locator

  // Logo
  readonly logo: Locator
  readonly logoLink: Locator

  // Navigation items - Main section
  readonly navDashboard: Locator
  readonly navChatbots: Locator

  // Navigation items - Settings section
  readonly navSettingsProfile: Locator
  readonly navSettingsPlan: Locator
  readonly navSettingsPrivacy: Locator

  // Admin navigation
  readonly navAdmin: Locator

  // Toggle buttons
  readonly sidebarToggle: Locator
  readonly mobileMenuButton: Locator
  readonly mobileCloseButton: Locator

  // User section
  readonly userAvatar: Locator
  readonly userProfileCard: Locator
  readonly userName: Locator

  // Logout button (from user-menu.page.ts)
  readonly logoutButton: Locator

  constructor(page: Page) {
    this.page = page

    // Main sidebar container - using class selector since no data-testid
    this.container = page.locator('.h-screen')
    this.sidebarGlass = page.locator('.sidebar-glass')

    // Logo
    this.logo = page.locator('.logo-glow')
    this.logoLink = page.locator('.logo-glow')

    // Navigation items - Main section (Platform)
    this.navDashboard = page.locator('.sidebar-nav-item').filter({ hasText: /Panel/i })
    this.navChatbots = page.locator('.sidebar-nav-item').filter({ hasText: /Chatbotlar/i })

    // Settings navigation items
    this.navSettingsProfile = page.locator('.sidebar-nav-item').filter({ hasText: /Profil/i })
    this.navSettingsPlan = page.locator('.sidebar-nav-item').filter({ hasText: /Plan/i })
    this.navSettingsPrivacy = page.locator('.sidebar-nav-item').filter({ hasText: /Gizlilik/i })

    // Admin navigation (only visible for platform admins)
    this.navAdmin = page.locator('.sidebar-nav-item').filter({ hasText: /Yönetim/i })

    // Toggle buttons
    this.sidebarToggle = page.locator('[title*="Sabit"]').or(page.locator('[title*="Hover"]'))
    this.mobileMenuButton = page.locator('header button.lg\\:hidden').first()
    this.mobileCloseButton = page.locator('aside button.lg\\:hidden').first()

    // User section
    this.userAvatar = page.locator('.avatar-ring').first()
    this.userProfileCard = page.locator('.user-profile-card')
    this.userName = page.locator('.user-profile-card').locator('.text-sm.font-semibold')

    // Logout button (shared with UserMenu)
    this.logoutButton = page.locator('.logout-btn')
  }

  // ============================================================================
  // Assertions
  // ============================================================================

  /**
   * Assert sidebar is visible
   */
  async expectVisible(): Promise<void> {
    await expect(this.sidebarGlass).toBeVisible({ timeout: 10000 })
  }

  /**
   * Assert sidebar is hidden (mobile)
   * Sidebar is hidden via -translate-x-full (off-screen), not display:none
   * So we check if it's NOT in the viewport
   */
  async expectHidden(): Promise<void> {
    // Use negative ratio to detect off-screen element (sidebar slides off to the left)
    await expect(this.sidebarGlass).not.toBeInViewport({ ratio: 0.1, timeout: 5000 })
  }

  /**
   * Assert sidebar is collapsed (hover mode)
   * When collapsed, sidebar has lg:w-[72px] width but logo link is still visible (only text hides).
   * We check localStorage mode and verify the sidebar has the collapsed width class.
   */
  async expectCollapsed(): Promise<void> {
    // Check localStorage for hover mode
    const mode = await this.page.evaluate(() => localStorage.getItem('botla_sidebar_mode'))
    expect(mode).toBe('hover')
    // In collapsed mode, sidebar has w-[72px] width class on lg screens
    await expect(this.sidebarGlass).toHaveClass(/lg:w-\[72px\]/)
  }

  /**
   * Assert sidebar is expanded (pinned mode)
   * When expanded, sidebar has lg:w-72 width and logo text is visible.
   */
  async expectExpanded(): Promise<void> {
    // Check localStorage for pinned mode (or default which shows expanded)
    const mode = await this.page.evaluate(() => localStorage.getItem('botla_sidebar_mode'))
    // Mode is either 'pinned' or null (default shows as expanded)
    expect(mode === 'pinned' || mode === null).toBe(true)
    // In expanded mode, sidebar has explicit w-72 class
    await expect(this.sidebarGlass).toHaveClass(/lg:w-72/)
    // Logo link should be visible in expanded state
    await expect(this.logo).toBeVisible()
  }

  /**
   * Assert navigation item is active
   */
  async expectNavItemActive(locator: Locator): Promise<void> {
    await expect(locator).toHaveClass(/active/)
  }

  /**
   * Assert navigation item is not active
   */
  async expectNavItemInactive(locator: Locator): Promise<void> {
    await expect(locator).not.toHaveClass(/active/)
  }

  /**
   * Assert user avatar is visible
   */
  async expectUserAvatarVisible(): Promise<void> {
    await expect(this.userAvatar).toBeVisible({ timeout: 5000 })
  }

  /**
   * Assert logout button is visible
   */
  async expectLogoutButtonVisible(): Promise<void> {
    await expect(this.logoutButton).toBeVisible({ timeout: 5000 })
  }

  /**
   * Assert admin navigation is visible
   */
  async expectAdminNavVisible(): Promise<void> {
    await expect(this.navAdmin).toBeVisible({ timeout: 5000 })
  }

  /**
   * Assert admin navigation is hidden
   */
  async expectAdminNavHidden(): Promise<void> {
    await expect(this.navAdmin).toBeHidden({ timeout: 5000 })
  }

  // ============================================================================
  // Navigation Actions
  // ============================================================================

  /**
   * Click on Dashboard navigation
   */
  async navigateToDashboard(): Promise<void> {
    await this.navDashboard.click()
    await expect(this.page).toHaveURL(/\/dashboard$/, { timeout: 10000 })
  }

  /**
   * Click on Chatbots navigation
   */
  async navigateToChatbots(): Promise<void> {
    await this.navChatbots.click()
    await expect(this.page).toHaveURL(/\/dashboard\/chatbots/, { timeout: 10000 })
  }

  /**
   * Click on Profile settings navigation
   */
  async navigateToProfile(): Promise<void> {
    await this.navSettingsProfile.click()
    await expect(this.page).toHaveURL(/\/dashboard\/settings\/profile/, { timeout: 10000 })
  }

  /**
   * Click on Plan settings navigation
   */
  async navigateToPlan(): Promise<void> {
    await this.navSettingsPlan.click()
    await expect(this.page).toHaveURL(/\/dashboard\/settings\/plan/, { timeout: 10000 })
  }

  /**
   * Click on Privacy settings navigation
   */
  async navigateToPrivacy(): Promise<void> {
    await this.navSettingsPrivacy.click()
    await expect(this.page).toHaveURL(/\/dashboard\/settings\/privacy/, { timeout: 10000 })
  }

  /**
   * Click on Admin navigation
   */
  async navigateToAdmin(): Promise<void> {
    await this.navAdmin.click()
    await expect(this.page).toHaveURL(/\/admin/, { timeout: 10000 })
  }

  /**
   * Click on logo to navigate to dashboard
   */
  async clickLogo(): Promise<void> {
    await this.logoLink.click()
    await expect(this.page).toHaveURL(/\/dashboard$/, { timeout: 10000 })
  }

  // ============================================================================
  // Sidebar Toggle Actions
  // ============================================================================

  /**
   * Click sidebar toggle button
   */
  async clickToggle(): Promise<void> {
    await this.sidebarToggle.click()
  }

  /**
   * Set sidebar mode to pinned (expanded)
   */
  async setPinnedMode(): Promise<void> {
    await this.page.evaluate(() => {
      localStorage.setItem('botla_sidebar_mode', 'pinned')
    })
    await this.page.reload()
    await this.expectExpanded()
  }

  /**
   * Set sidebar mode to hover (collapsed)
   */
  async setHoverMode(): Promise<void> {
    await this.page.evaluate(() => {
      localStorage.setItem('botla_sidebar_mode', 'hover')
    })
    await this.page.reload()
    await this.expectCollapsed()
  }

  /**
   * Open mobile menu
   */
  async openMobileMenu(): Promise<void> {
    await this.mobileMenuButton.scrollIntoViewIfNeeded()
    await this.mobileMenuButton.click()
    await expect(this.sidebarGlass).toBeVisible()
  }

  /**
   * Close mobile menu
   */
  async closeMobileMenu(): Promise<void> {
    await this.mobileCloseButton.scrollIntoViewIfNeeded()
    await this.mobileCloseButton.click()
    // Mobile sidebar is hidden via -translate-x-full (off-screen), not display:none
    await expect(this.sidebarGlass).not.toBeInViewport({ ratio: 0.1, timeout: 5000 })
  }

  // ============================================================================
  // User Section Actions
  // ============================================================================

  /**
   * Click on user avatar
   */
  async clickUserAvatar(): Promise<void> {
    await this.userAvatar.click()
  }

  /**
   * Click on user profile card
   */
  async clickUserProfileCard(): Promise<void> {
    await this.userProfileCard.click()
    await expect(this.page).toHaveURL(/\/dashboard\/settings\/profile/, { timeout: 10000 })
  }

  // ============================================================================
  // Utility Methods
  // ============================================================================

  /**
   * Check if admin nav is visible (for conditional testing)
   */
  async isAdminNavVisible(): Promise<boolean> {
    return await this.navAdmin.isVisible()
  }

  /**
   * Get current sidebar mode
   */
  async getSidebarMode(): Promise<'pinned' | 'hover'> {
    return await this.page.evaluate(() => {
      return (localStorage.getItem('botla_sidebar_mode') as 'pinned' | 'hover') || 'hover'
    })
  }
}
