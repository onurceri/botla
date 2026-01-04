import { Locator, Page, expect } from '@playwright/test'

export class RegisterPage {
  readonly page: Page

  // Page container and title
  readonly pageContainer: Locator
  readonly title: Locator

  // Form elements (matching actual component test IDs)
  readonly nameInput: Locator
  readonly emailInput: Locator
  readonly passwordInput: Locator
  readonly submitButton: Locator

  // Error and feedback
  readonly errorMessage: Locator

  // Navigation
  readonly loginLink: Locator

  constructor(page: Page) {
    this.page = page

    // Page container
    this.pageContainer = page.locator('[data-testid="register-page"]')
    this.title = page.locator('[data-testid="register-page-title"]')

    // Form inputs matching actual component
    this.nameInput = page.getByTestId('register-page-name-input')
    this.emailInput = page.getByTestId('register-page-email-input')
    this.passwordInput = page.getByTestId('register-page-password-input')
    this.submitButton = page.getByTestId('register-page-submit-button')

    // Error message
    this.errorMessage = page.locator('[data-testid="register-page-error-message"]')

    // Login link (text-based fallback)
    this.loginLink = page.getByRole('link', { name: /giriş yap|login/i })
  }

  // Navigation

  async goto(): Promise<void> {
    await this.page.goto('/register')
    await this.expectToBeLoaded()
  }

  async expectToBeLoaded(): Promise<void> {
    await expect(this.pageContainer).toBeVisible()
    await expect(this.nameInput).toBeVisible()
    await expect(this.emailInput).toBeVisible()
    await expect(this.passwordInput).toBeVisible()
    await expect(this.submitButton).toBeVisible()
  }

  // Form Interactions

  async fillName(name: string): Promise<void> {
    await this.nameInput.fill(name)
  }

  async clearName(): Promise<void> {
    await this.nameInput.clear()
  }

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

  async fillForm(name: string, email: string, password: string): Promise<void> {
    await this.fillName(name)
    await this.fillEmail(email)
    await this.fillPassword(password)
  }

  async clearForm(): Promise<void> {
    await this.clearName()
    await this.clearEmail()
    await this.clearPassword()
  }

  // Button Interactions

  async clickSubmitButton(): Promise<void> {
    await this.submitButton.click()
  }

  async submit(): Promise<void> {
    await this.submitButton.click()
  }

  async register(name: string, email: string, password: string): Promise<void> {
    await this.fillForm(name, email, password)
    await this.submit()
  }

  // Navigation Links

  async clickLoginLink(): Promise<void> {
    await this.loginLink.click()
  }

  async goToLogin(): Promise<void> {
    await this.clickLoginLink()
  }

  // Focus States

  async focusNameInput(): Promise<void> {
    await this.nameInput.click()
  }

  async focusEmailInput(): Promise<void> {
    await this.emailInput.click()
  }

  async focusPasswordInput(): Promise<void> {
    await this.passwordInput.click()
  }

  async expectNameInputToBeFocused(): Promise<void> {
    await expect(this.nameInput).toBeFocused()
  }

  async expectEmailInputToBeFocused(): Promise<void> {
    await expect(this.emailInput).toBeFocused()
  }

  async expectPasswordInputToBeFocused(): Promise<void> {
    await expect(this.passwordInput).toBeFocused()
  }

  // Hover States

  async hoverSubmitButton(): Promise<void> {
    await this.submitButton.hover()
  }

  async hoverLoginLink(): Promise<void> {
    await this.loginLink.hover()
  }

  // Keyboard Navigation

  async pressTab(): Promise<void> {
    await this.page.keyboard.press('Tab')
  }

  async pressEnter(): Promise<void> {
    await this.page.keyboard.press('Enter')
  }

  async tabThroughFields(): Promise<void> {
    await this.nameInput.focus()
    await this.page.keyboard.press('Tab')
    await expect(this.emailInput).toBeFocused()
    await this.page.keyboard.press('Tab')
    await expect(this.passwordInput).toBeFocused()
    await this.page.keyboard.press('Tab')
    await expect(this.submitButton).toBeFocused()
  }

  // Assertions

  async expectToBeVisible(): Promise<void> {
    await expect(this.pageContainer).toBeVisible()
  }

  async expectTitleToBeVisible(): Promise<void> {
    await expect(this.title).toBeVisible()
  }

  async expectNameInputToBeVisible(): Promise<void> {
    await expect(this.nameInput).toBeVisible()
  }

  async expectEmailInputToBeVisible(): Promise<void> {
    await expect(this.emailInput).toBeVisible()
  }

  async expectPasswordInputToBeVisible(): Promise<void> {
    await expect(this.passwordInput).toBeVisible()
  }

  async expectSubmitButtonToBeVisible(): Promise<void> {
    await expect(this.submitButton).toBeVisible()
  }

  async expectErrorMessage(message: string | RegExp): Promise<void> {
    await expect(this.errorMessage).toBeVisible()
    await expect(this.errorMessage).toContainText(message)
  }

  async expectNoErrorMessage(): Promise<void> {
    await expect(this.errorMessage).toBeHidden()
  }

  async expectSubmitButtonToBeDisabled(): Promise<void> {
    await expect(this.submitButton).toBeDisabled()
  }

  async expectSubmitButtonToBeEnabled(): Promise<void> {
    await expect(this.submitButton).toBeEnabled()
  }

  async expectNameInputToHaveError(): Promise<void> {
    await expect(this.nameInput).toHaveClass(/error|invalid/i)
  }

  async expectEmailInputToHaveError(): Promise<void> {
    await expect(this.emailInput).toHaveClass(/error|invalid/i)
  }

  async expectPasswordInputToHaveError(): Promise<void> {
    await expect(this.passwordInput).toHaveClass(/error|invalid/i)
  }

  async expectUrlToContain(path: string | RegExp): Promise<void> {
    await expect(this.page).toHaveURL(path)
  }

  async expectLoginLinkToBeVisible(): Promise<void> {
    await expect(this.loginLink).toBeVisible()
  }

  // State Verification

  getCurrentUrl(): string {
    return this.page.url()
  }

  async expectNameToHaveValue(value: string): Promise<void> {
    await expect(this.nameInput).toHaveValue(value)
  }

  async expectEmailToHaveValue(value: string): Promise<void> {
    await expect(this.emailInput).toHaveValue(value)
  }

  async expectPasswordToHaveValue(value: string): Promise<void> {
    await expect(this.passwordInput).toHaveValue(value)
  }

  // Loading State

  async expectSubmitButtonToBeLoading(): Promise<boolean> {
    return await this.submitButton.isDisabled()
  }

  getSubmitButtonLocator(): Locator {
    return this.submitButton
  }
}
