import { Locator, Page, expect } from '@playwright/test'

export class UserMenu {
  readonly page: Page

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

  /**
   * Navigate to dashboard first
   */
  async openFromDashboard(): Promise<void> {
    await this.page.goto('/dashboard')
    await this.page.waitForLoadState('networkidle')
    await this.page.waitForTimeout(500)
  }

  // Menu Interactions

  /**
   * Click the logout button directly
   */
  async clickLogout(): Promise<void> {
    // Use force to bypass visibility checks since elements might be in collapsed sidebar
    await this.logoutButton.click({ force: true })
  }

  /**
   * Perform logout
   */
  async logout(): Promise<void> {
    // First ensure we're on dashboard
    await this.page.goto('/dashboard')
    await this.page.waitForLoadState('networkidle')
    await this.page.waitForTimeout(300)

    // Click logout button with force
    await this.clickLogout()
  }

  // Modal Interactions

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
