import { Locator, Page, expect } from '@playwright/test'

/**
 * User Menu Page Object
 * Handles user profile, settings navigation, and logout from dashboard context.
 * Expanded with dashboard-specific methods from task 06-dashboard-layout.
 */
export class UserMenu {
  readonly page: Page

  // Avatar and profile
  readonly userAvatar: Locator
  readonly userProfileCard: Locator
  readonly userName: Locator
  readonly userPlanBadge: Locator

  // Dropdown menu
  readonly dropdown: Locator
  readonly menuItems: Locator

  // Menu item links
  readonly profileLink: Locator
  readonly settingsLink: Locator
  readonly helpLink: Locator

  // Logout button - main interaction point
  readonly logoutButton: Locator

  // Session expired modal
  readonly sessionExpiredModal: Locator
  readonly sessionExpiredTitle: Locator
  readonly sessionExpiredMessage: Locator
  readonly reloginButton: Locator
  readonly cancelButton: Locator

  // Loading indicator
  readonly loadingSpinner: Locator

  constructor(page: Page) {
    this.page = page

    // Avatar and profile
    this.userAvatar = page.locator('.avatar-ring').first()
    this.userProfileCard = page.locator('.user-profile-card')
    this.userName = page.locator('.user-profile-card .text-sm.font-semibold.truncate')
    this.userPlanBadge = page.locator('.user-profile-card .rounded-full').filter({ hasText: /PRO|FREE|ENTERPRISE/i })

    // Dropdown menu
    this.dropdown = page.locator('[data-testid="dropdown-user-menu"]').or(
      page.locator('[class*="user-menu-dropdown"]')
    )
    this.menuItems = page.locator('[data-testid="menu-item"]').or(
      this.dropdown.locator('[class*="item"]')
    )

    // Menu item links
    this.profileLink = page.locator('.sidebar-nav-item').filter({ hasText: /Profil/i })
    this.settingsLink = page.locator('[href*="settings"]').first()
    this.helpLink = page.locator('a:has-text("Help")').or(page.locator('a:has-text("Yardım")'))

    // Logout button with specific class - main way to logout
    this.logoutButton = page.locator('.logout-btn')

    // Session expired modal
    this.sessionExpiredModal = page
      .getByTestId('modal-session-expired')
      .or(page.getByRole('dialog', { name: /session expired|oturum|süre/i }))
      .or(page.locator('[class*="session"][class*="expired"]'))

    this.sessionExpiredTitle = page
      .getByTestId('session-expired-title')
      .or(page.getByRole('heading', { name: /session expired|oturum/i }))

    this.sessionExpiredMessage = page
      .getByTestId('session-expired-message')
      .or(page.getByText(/session expired|süresi doldu/i))

    // Relogin button in session expired modal
    this.reloginButton = page
      .getByTestId('btn-relogin')
      .or(page.getByRole('button', { name: /relogin|tekrar|giriş yap/i }))
      .or(page.getByText(/tekrar giriş|relogin/i))

    // Cancel button (if present)
    this.cancelButton = page
      .getByRole('button', { name: /cancel|vazgeç|iptal/i })
      .or(page.getByTestId('btn-cancel'))

    // Loading spinner
    this.loadingSpinner = page
      .getByTestId('loading-spinner')
      .or(page.locator('[class*="loading"][class*="spinner"]'))
  }

  // ============================================================================
  // Dashboard-Specific Methods (from task 06-dashboard-layout)
  // ============================================================================

  /**
   * Open user menu from dashboard by clicking avatar
   */
  async openFromDashboard(): Promise<void> {
    await this.page.goto('/dashboard')
    await this.page.waitForLoadState('networkidle')
    await this.clickUserAvatar()
  }

  /**
   * Click on user avatar to open menu
   */
  async clickUserAvatar(): Promise<void> {
    await this.userAvatar.click()
  }

  /**
   * Click on user profile card (navigates to profile settings)
   */
  async clickUserProfileCard(): Promise<void> {
    await this.userProfileCard.click()
    await expect(this.page).toHaveURL(/\/dashboard\/settings\/profile/, { timeout: 10000 })
  }

  /**
   * Navigate to Profile settings
   */
  async navigateToProfile(): Promise<void> {
    await this.clickUserAvatar()
    await this.profileLink.click()
    await expect(this.page).toHaveURL(/\/dashboard\/settings\/profile/, { timeout: 10000 })
  }

  /**
   * Navigate to Settings page
   */
  async navigateToSettings(): Promise<void> {
    await this.clickUserAvatar()
    await this.settingsLink.click()
    await expect(this.page).toHaveURL(/\/settings/, { timeout: 10000 })
  }

  /**
   * Navigate to Help page
   */
  async navigateToHelp(): Promise<void> {
    await this.clickUserAvatar()
    await this.helpLink.click()
  }

  /**
   * Perform logout from dashboard
   */
  async logout(): Promise<void> {
    await this.page.goto('/dashboard')
    await this.page.waitForLoadState('networkidle')
    await this.page.waitForTimeout(300)
    await this.clickLogout()
  }

  /**
   * Assert user avatar is visible on dashboard
   */
  async expectUserAvatarVisible(): Promise<void> {
    await expect(this.userAvatar).toBeVisible({ timeout: 10000 })
  }

  /**
   * Assert user name is displayed
   */
  async expectUserNameVisible(): Promise<void> {
    await expect(this.userName).toBeVisible()
  }

  /**
   * Assert user plan badge is visible
   */
  async expectUserPlanBadgeVisible(): Promise<void> {
    await expect(this.userPlanBadge).toBeVisible()
  }

  /**
   * Assert dropdown menu is visible
   */
  async expectDropdownVisible(): Promise<void> {
    await expect(this.dropdown).toBeVisible({ timeout: 5000 })
  }

  /**
   * Assert dropdown menu is hidden
   */
  async expectDropdownHidden(): Promise<void> {
    await expect(this.dropdown).toBeHidden({ timeout: 5000 })
  }

  // ============================================================================
  // Menu Interactions
  // ============================================================================

  /**
   * Click the logout button directly
   */
  async clickLogout(): Promise<void> {
    // Use force to bypass visibility checks since elements might be in collapsed sidebar
    await this.logoutButton.click({ force: true })
  }

  async clickRelogin(): Promise<void> {
    await this.reloginButton.click()
  }

  async clickCancel(): Promise<void> {
    await this.cancelButton.click()
  }

  // Assertions

  /**
   * Check if logout button exists in DOM (not necessarily visible)
   */
  async expectLogoutButtonExists(): Promise<void> {
    await expect(this.logoutButton).toBeAttached({ timeout: 10000 })
  }

  /**
   * Check if logout button is visible
   */
  async expectLogoutButtonVisible(): Promise<void> {
    await expect(this.logoutButton).toBeVisible({ timeout: 10000 })
  }

  /**
   * Check if logout button is enabled
   */
  async expectLogoutButtonEnabled(): Promise<void> {
    await expect(this.logoutButton).toBeEnabled()
  }

  // Session Expired Modal

  async expectSessionExpiredModalVisible(): Promise<void> {
    await expect(this.sessionExpiredModal).toBeVisible({ timeout: 5000 })
  }

  async expectSessionExpiredModalHidden(): Promise<void> {
    await expect(this.sessionExpiredModal).toBeHidden({ timeout: 5000 })
  }

  async expectReloginButtonVisible(): Promise<void> {
    await expect(this.reloginButton).toBeVisible()
  }

  async expectReloginButtonEnabled(): Promise<void> {
    await expect(this.reloginButton).toBeEnabled()
  }

  async handleSessionExpired(): Promise<void> {
    await this.expectSessionExpiredModalVisible()
    await this.clickRelogin()
  }

  // Loading States

  async expectLoadingSpinnerVisible(): Promise<void> {
    await expect(this.loadingSpinner).toBeVisible({ timeout: 5000 })
  }

  async expectLoadingSpinnerHidden(): Promise<void> {
    await expect(this.loadingSpinner).toBeHidden({ timeout: 10000 })
  }

  // State Verification

  async isLogoutButtonVisible(): Promise<boolean> {
    return await this.logoutButton.isVisible()
  }

  async isSessionExpiredModalVisible(): Promise<boolean> {
    return await this.sessionExpiredModal.isVisible()
  }

  // Keyboard Navigation

  async pressEscape(): Promise<void> {
    await this.page.keyboard.press('Escape')
  }

  async pressTab(): Promise<void> {
    await this.page.keyboard.press('Tab')
  }

  // Hover States

  async hoverLogoutButton(): Promise<void> {
    await this.logoutButton.hover()
  }
}
