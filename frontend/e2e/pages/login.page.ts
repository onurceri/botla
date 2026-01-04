import { Locator, Page, expect } from '@playwright/test'

export class LoginPage {
  readonly page: Page

  // Page elements with data-testid attributes
  readonly pageContainer: Locator
  readonly title: Locator
  readonly subtitle: Locator

  // Form elements
  readonly emailInput: Locator
  readonly emailLabel: Locator
  readonly passwordInput: Locator
  readonly passwordLabel: Locator
  readonly loginButton: Locator

  // Additional form elements
  readonly rememberMeCheckbox: Locator
  readonly rememberMeLabel: Locator
  readonly forgotPasswordLink: Locator
  readonly registerLink: Locator

  // Feedback elements
  readonly errorMessage: Locator
  readonly successMessage: Locator
  readonly loadingSpinner: Locator

  constructor(page: Page) {
    this.page = page

    // Page container and title
    this.pageContainer = page.locator('[data-testid="login-page"]')
    this.title = page.locator('[data-testid="login-page-title"]')
    this.subtitle = page.locator('[data-testid="login-page-subtitle"]')

    // Form inputs with multiple selector fallbacks
    this.emailInput = page
      .getByTestId('login-page-email-input')
      .or(page.getByLabel(/email/i, { exact: false }))
      .or(page.locator('input[type="email"]'))

    this.emailLabel = page.locator('label[for="email"]')

    this.passwordInput = page
      .getByTestId('login-page-password-input')
      .or(page.getByLabel(/password|şifre/i, { exact: false }))
      .or(page.locator('input[type="password"]'))

    this.passwordLabel = page.locator('label[for="password"]')

    // Submit button
    this.loginButton = page
      .getByTestId('login-page-submit-button')
      .or(page.getByRole('button', { name: /login|giriş yap/i }))

    // Additional form elements
    this.rememberMeCheckbox = page
      .getByTestId('login-page-remember-me-checkbox')
      .or(page.getByRole('checkbox', { name: /remember me|beni hatırla/i }))

    this.rememberMeLabel = page.getByText(/remember me|beni hatırla/i)

    this.forgotPasswordLink = page
      .getByTestId('login-page-forgot-password-link')
      .or(page.getByRole('link', { name: /forgot password|şifremi unuttum/i }))

    this.registerLink = page
      .getByTestId('login-page-register-link')
      .or(page.getByRole('link', { name: /register|sign up|kayıt ol/i }))

    // Feedback elements
    this.errorMessage = page.getByTestId('login-page-error-message')
    this.successMessage = page.getByTestId('login-page-success-message')
    this.loadingSpinner = page.getByTestId('loading-spinner')
  }

  // Navigation

  async goto(): Promise<void> {
    await this.page.goto('/login')
    await this.expectToBeLoaded()
  }

  async expectToBeLoaded(): Promise<void> {
    await expect(this.pageContainer).toBeVisible()
    await expect(this.emailInput).toBeVisible()
    await expect(this.passwordInput).toBeVisible()
    await expect(this.loginButton).toBeVisible()
  }

  // Form Interactions

  async fillEmail(email: string): Promise<void> {
    await this.emailInput.fill(email)
  }

  async clearEmail(): Promise<void> {
    await this.emailInput.clear()
  }

  async fillPassword(password: string): Promise<void> {
    await this.passwordInput.fill(password)
  }

  async clearPassword(): Promise<void> {
    await this.passwordInput.clear()
  }

  async fillCredentials(email: string, password: string): Promise<void> {
    await this.fillEmail(email)
    await this.fillPassword(password)
  }

  async clearCredentials(): Promise<void> {
    await this.clearEmail()
    await this.clearPassword()
  }

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

  async clickLoginButton(): Promise<void> {
    await this.loginButton.click()
  }

  async submit(): Promise<void> {
    await this.loginButton.click()
  }

  async login(email: string, password: string, rememberMe = false): Promise<void> {
    await this.fillCredentials(email, password)
    if (rememberMe) {
      await this.checkRememberMe()
    }
    await this.submit()
  }

  // Navigation Links

  async clickForgotPasswordLink(): Promise<void> {
    await this.forgotPasswordLink.click()
  }

  async goToForgotPassword(): Promise<void> {
    await this.clickForgotPasswordLink()
  }

  async clickRegisterLink(): Promise<void> {
    await this.registerLink.click()
  }

  async goToRegister(): Promise<void> {
    await this.clickRegisterLink()
  }

  // Focus States

  async focusEmailInput(): Promise<void> {
    await this.emailInput.click()
  }

  async focusPasswordInput(): Promise<void> {
    await this.passwordInput.click()
  }

  async expectEmailInputToBeFocused(): Promise<void> {
    await expect(this.emailInput).toBeFocused()
  }

  async expectPasswordInputToBeFocused(): Promise<void> {
    await expect(this.passwordInput).toBeFocused()
  }

  // Hover States

  async hoverLoginButton(): Promise<void> {
    await this.loginButton.hover()
  }

  async hoverForgotPasswordLink(): Promise<void> {
    await this.forgotPasswordLink.hover()
  }

  async hoverRegisterLink(): Promise<void> {
    await this.registerLink.hover()
  }

  async hoverEmailLabel(): Promise<void> {
    await this.emailLabel.hover()
  }

  async hoverPasswordLabel(): Promise<void> {
    await this.passwordLabel.hover()
  }

  // Keyboard Navigation

  async pressTab(): Promise<void> {
    await this.page.keyboard.press('Tab')
  }

  async pressShiftTab(): Promise<void> {
    await this.page.keyboard.press('Shift+Tab')
  }

  async pressEnter(): Promise<void> {
    await this.page.keyboard.press('Enter')
  }

  async pressEscape(): Promise<void> {
    await this.page.keyboard.press('Escape')
  }

  async tabThroughFields(): Promise<void> {
    await this.emailInput.focus()
    await this.page.keyboard.press('Tab')
    await expect(this.passwordInput).toBeFocused()
    await this.page.keyboard.press('Tab')
    await expect(this.loginButton).toBeFocused()
  }

  // Assertions

  async expectToBeVisible(): Promise<void> {
    await expect(this.pageContainer).toBeVisible()
  }

  async expectTitleToBeVisible(): Promise<void> {
    await expect(this.title).toBeVisible()
  }

  async expectEmailInputToBeVisible(): Promise<void> {
    await expect(this.emailInput).toBeVisible()
  }

  async expectPasswordInputToBeVisible(): Promise<void> {
    await expect(this.passwordInput).toBeVisible()
  }

  async expectLoginButtonToBeVisible(): Promise<void> {
    await expect(this.loginButton).toBeVisible()
  }

  async expectErrorMessage(message: string | RegExp): Promise<void> {
    await expect(this.errorMessage).toBeVisible()
    await expect(this.errorMessage).toContainText(message)
  }

  async expectNoErrorMessage(): Promise<void> {
    await expect(this.errorMessage).toBeHidden()
  }

  async expectSuccessMessage(message: string | RegExp): Promise<void> {
    await expect(this.successMessage).toBeVisible()
    await expect(this.successMessage).toContainText(message)
  }

  async expectLoadingSpinnerToBeVisible(): Promise<void> {
    await expect(this.loadingSpinner).toBeVisible()
  }

  async expectLoadingSpinnerToBeHidden(): Promise<void> {
    await expect(this.loadingSpinner).toBeHidden()
  }

  async expectLoginButtonToBeDisabled(): Promise<void> {
    await expect(this.loginButton).toBeDisabled()
  }

  async expectLoginButtonToBeEnabled(): Promise<void> {
    await expect(this.loginButton).toBeEnabled()
  }

  async expectEmailInputToHaveError(): Promise<void> {
    await expect(this.emailInput).toHaveClass(/error|invalid/)
  }

  async expectPasswordInputToHaveError(): Promise<void> {
    await expect(this.passwordInput).toHaveClass(/error|invalid/)
  }

  async expectEmailInputToNotHaveError(): Promise<void> {
    await expect(this.emailInput).not.toHaveClass(/error|invalid/)
  }

  async expectPasswordInputToNotHaveError(): Promise<void> {
    await expect(this.passwordInput).not.toHaveClass(/error|invalid/)
  }

  async expectUrlToContain(path: string | RegExp): Promise<void> {
    await expect(this.page).toHaveURL(path)
  }

  async expectRememberMeToBeChecked(): Promise<void> {
    await expect(this.rememberMeCheckbox).toBeChecked()
  }

  async expectRememberMeToNotBeChecked(): Promise<void> {
    await expect(this.rememberMeCheckbox).not.toBeChecked()
  }

  async expectForgotPasswordLinkToBeVisible(): Promise<void> {
    await expect(this.forgotPasswordLink).toBeVisible()
  }

  async expectRegisterLinkToBeVisible(): Promise<void> {
    await expect(this.registerLink).toBeVisible()
  }

  // Validation Helpers

  async validateEmailOnBlur(invalidEmail: string): Promise<void> {
    await this.fillEmail(invalidEmail)
    await this.passwordInput.click()
    await this.expectEmailInputToHaveError()
  }

  async validateRequiredFields(): Promise<void> {
    await this.loginButton.click()
    await this.expectErrorMessage(/required|zorunlu|doldurun/i)
  }

  // State Verification

  async isLoginButtonLoading(): Promise<boolean> {
    return await this.loginButton.isDisabled()
  }

  getCurrentUrl(): string {
    return this.page.url()
  }

  async expectEmailToHaveValue(value: string): Promise<void> {
    await expect(this.emailInput).toHaveValue(value)
  }

  async expectPasswordToHaveValue(value: string): Promise<void> {
    await expect(this.passwordInput).toHaveValue(value)
  }
}
