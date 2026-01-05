import { Locator, Page, expect } from '@playwright/test'

/**
 * Session Page - Page Object for session management UI elements
 * Includes session expired modal, token refresh indicators, and session status elements
 */
export class SessionPage {
  readonly page: Page

  // Session expired modal elements
  readonly sessionExpiredModal: Locator
  readonly sessionExpiredTitle: Locator
  readonly sessionExpiredMessage: Locator
  readonly sessionExpiredReloginButton: Locator
  readonly sessionExpiredCancelButton: Locator

  // Token refresh indicator
  readonly tokenRefreshingIndicator: Locator
  readonly tokenRefreshSuccess: Locator
  readonly tokenRefreshError: Locator

  // Session status elements
  readonly sessionStatusIndicator: Locator
  readonly sessionExpiryTime: Locator
  readonly sessionInfoTooltip: Locator

  // User menu with session options
  readonly userMenu: Locator
  readonly userAvatar: Locator
  readonly userDropdown: Locator
  readonly menuItemProfile: Locator
  readonly menuItemSettings: Locator
  readonly menuItemLogout: Locator

  // Remember me checkbox (on login page)
  readonly rememberMeCheckbox: Locator
  readonly rememberMeLabel: Locator

  constructor(page: Page) {
    this.page = page

    // Session expired modal
    this.sessionExpiredModal = page.locator('[data-testid="modal-session-expired"]')
    this.sessionExpiredTitle = page.locator('[data-testid="session-expired-title"]')
    this.sessionExpiredMessage = page.locator('[data-testid="session-expired-message"]')
    this.sessionExpiredReloginButton = page.locator('[data-testid="btn-relogin"]')
    this.sessionExpiredCancelButton = page.locator('[data-testid="btn-session-cancel"]')

    // Token refresh indicator
    this.tokenRefreshingIndicator = page.locator('[data-testid="token-refreshing"]')
    this.tokenRefreshSuccess = page.locator('[data-testid="token-refresh-success"]')
    this.tokenRefreshError = page.locator('[data-testid="token-refresh-error"]')

    // Session status
    this.sessionStatusIndicator = page.locator('[data-testid="session-status"]')
    this.sessionExpiryTime = page.locator('[data-testid="session-expiry-time"]')
    this.sessionInfoTooltip = page.locator('[data-testid="session-info-tooltip"]')

    // User menu
    this.userMenu = page.locator('[data-testid="user-menu"]')
    this.userAvatar = page.locator('[data-testid="user-avatar"]')
    this.userDropdown = page.locator('[data-testid="user-menu-dropdown"]')
    this.menuItemProfile = page.locator('[data-testid="menu-item-profile"]')
    this.menuItemSettings = page.locator('[data-testid="menu-item-settings"]')
    this.menuItemLogout = page.locator('[data-testid="menu-item-logout"]')

    // Remember Me (fallback selectors)
    this.rememberMeCheckbox = page
      .getByTestId('login-page-remember-me-checkbox')
      .or(page.getByRole('checkbox', { name: /remember me|beni hatırla/i }))

    this.rememberMeLabel = page.getByText(/remember me|beni hatırla/i)
  }

  // Session Expired Modal Operations

  async expectSessionExpiredModalVisible(): Promise<void> {
    await expect(this.sessionExpiredModal).toBeVisible()
  }

  async expectSessionExpiredModalHidden(): Promise<void> {
    await expect(this.sessionExpiredModal).toBeHidden()
  }

  async expectSessionExpiredTitle(message: string | RegExp): Promise<void> {
    await expect(this.sessionExpiredTitle).toContainText(message)
  }

  async clickReloginButton(): Promise<void> {
    await this.sessionExpiredReloginButton.click()
  }

  async clickCancelButton(): Promise<void> {
    await this.sessionExpiredCancelButton.click()
  }

  // Token Refresh Indicator Operations

  async expectTokenRefreshingIndicatorVisible(): Promise<void> {
    await expect(this.tokenRefreshingIndicator).toBeVisible()
  }

  async expectTokenRefreshingIndicatorHidden(): Promise<void> {
    await expect(this.tokenRefreshingIndicator).toBeHidden()
  }

  async expectTokenRefreshSuccessVisible(): Promise<void> {
    await expect(this.tokenRefreshSuccess).toBeVisible()
  }

  async expectTokenRefreshErrorVisible(): Promise<void> {
    await expect(this.tokenRefreshError).toBeVisible()
  }

  async expectTokenRefreshErrorHidden(): Promise<void> {
    await expect(this.tokenRefreshError).toBeHidden()
  }

  // Session Status Operations

  async expectSessionActive(): Promise<void> {
    await expect(this.sessionStatusIndicator).toHaveAttribute('data-status', 'active')
  }

  async expectSessionExpiringSoon(): Promise<void> {
    await expect(this.sessionStatusIndicator).toHaveAttribute('data-status', 'expiring')
  }

  async expectSessionExpired(): Promise<void> {
    await expect(this.sessionStatusIndicator).toHaveAttribute('data-status', 'expired')
  }

  async getSessionExpiryTime(): Promise<string | null> {
    return await this.sessionExpiryTime.textContent()
  }

  async hoverSessionInfoTooltip(): Promise<void> {
    await this.sessionInfoTooltip.hover()
  }

  // User Menu Operations

  async clickUserAvatar(): Promise<void> {
    await this.userAvatar.click()
  }

  async openUserMenu(): Promise<void> {
    await this.userMenu.click()
  }

  async expectUserMenuOpen(): Promise<void> {
    await expect(this.userDropdown).toBeVisible()
  }

  async expectUserMenuClosed(): Promise<void> {
    await expect(this.userDropdown).toBeHidden()
  }

  async clickMenuItemProfile(): Promise<void> {
    await this.menuItemProfile.click()
  }

  async clickMenuItemSettings(): Promise<void> {
    await this.menuItemSettings.click()
  }

  async clickMenuItemLogout(): Promise<void> {
    await this.menuItemLogout.click()
  }

  // Remember Me Operations

  async checkRememberMe(): Promise<void> {
    if (!(await this.rememberMeCheckbox.isChecked())) {
      await this.rememberMeCheckbox.check()
    }
  }

  async uncheckRememberMe(): Promise<void> {
    if (await this.rememberMeCheckbox.isChecked()) {
      await this.rememberMeCheckbox.uncheck()
    }
  }

  async toggleRememberMe(): Promise<void> {
    await this.rememberMeCheckbox.click()
  }

  async expectRememberMeChecked(): Promise<void> {
    await expect(this.rememberMeCheckbox).toBeChecked()
  }

  async expectRememberMeUnchecked(): Promise<void> {
    await expect(this.rememberMeCheckbox).not.toBeChecked()
  }

  async isRememberMeChecked(): Promise<boolean> {
    return await this.rememberMeCheckbox.isChecked()
  }

  // Navigation and State Verification

  async expectOnLoginPage(): Promise<void> {
    await expect(this.page).toHaveURL(/\/login/)
  }

  async expectOnDashboard(): Promise<void> {
    await expect(this.page).toHaveURL(/\/dashboard/)
  }

  async expectAuthenticated(): Promise<void> {
    const accessToken = await this.page.evaluate(() => localStorage.getItem('botla_token'))
    const refreshToken = await this.page.evaluate(() => localStorage.getItem('botla_refresh_token'))
    expect(accessToken).toBeTruthy()
    expect(refreshToken).toBeTruthy()
  }

  async expectUnauthenticated(): Promise<void> {
    const accessToken = await this.page.evaluate(() => localStorage.getItem('botla_token'))
    expect(accessToken).toBeNull()
  }

  // Wait for specific session states

  async waitForSessionCleared(timeout: number = 10000): Promise<void> {
    await this.page.waitForFunction(
      () => {
        const token = localStorage.getItem('botla_token')
        const refreshToken = localStorage.getItem('botla_refresh_token')
        return token === null && refreshToken === null
      },
      { timeout }
    )
  }

  async waitForSessionValid(timeout: number = 10000): Promise<void> {
    await this.page.waitForFunction(
      () => {
        const token = localStorage.getItem('botla_token')
        const refreshToken = localStorage.getItem('botla_refresh_token')
        if (!token || !refreshToken) return false

        try {
          const payload = JSON.parse(atob(token.split('.')[1]))
          if (payload.exp) {
            const now = Math.floor(Date.now() / 1000)
            if (payload.exp < now) return false
          }
        } catch {
          return false
        }

        return true
      },
      { timeout }
    )
  }

  async waitForSessionExpiredModal(timeout: number = 10000): Promise<void> {
    await expect(this.sessionExpiredModal).toBeVisible({ timeout })
  }
}
