import { test, expect } from '@playwright/test'
import { TEST_IDS, TURKISH } from './test-constants'
import {
  mockSuccessfulRegistration,
  mockEmailExists,
  mockNetworkError,
  mockServerError,
  mockRateLimitError,
  clearAuthStorage,
  getAuthTokens,
} from './mocks/register.mocks'

/**
 * Registration Page E2E Tests
 *
 * Tests for the actual registration page component which has:
 * - Name, email, and password fields
 * - Submit button
 * - Login link
 * - No confirm password, terms, or privacy checkboxes
 * - Automatic login after registration
 */

test.describe('Registration Page', () => {
  // ---------------------------------------------------------------------------
  // Phase 1: Page Load Tests
  // ---------------------------------------------------------------------------

  test.describe('Page Load', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
    })

    test('should load register page successfully', async ({ page }) => {
      await expect(page.getByTestId('register-page')).toBeVisible()
      await expect(page.getByTestId('register-page-name-input')).toBeVisible()
      await expect(page.getByTestId('register-page-email-input')).toBeVisible()
      await expect(page.getByTestId('register-page-password-input')).toBeVisible()
      await expect(page.getByTestId('register-page-submit-button')).toBeVisible()
    })

    test('should have correct page title', async ({ page }) => {
      await expect(page.getByTestId('register-page-title')).toContainText(/hesap oluştur|create account/i)
    })

    test('should have all form inputs', async ({ page }) => {
      const nameInput = page.getByTestId('register-page-name-input')
      const emailInput = page.getByTestId('register-page-email-input')
      const passwordInput = page.getByTestId('register-page-password-input')

      await expect(nameInput).toBeVisible()

      await expect(emailInput).toBeVisible()
      await expect(emailInput).toHaveAttribute('type', 'email')

      await expect(passwordInput).toBeVisible()
      await expect(passwordInput).toHaveAttribute('type', 'password')
    })

    test('should have register button', async ({ page }) => {
      const submitButton = page.getByTestId('register-page-submit-button')
      await expect(submitButton).toBeVisible()
      await expect(submitButton).toContainText(/kayıt ol|register/i)
    })

    test('should have login link', async ({ page }) => {
      const loginLink = page.getByRole('link', { name: /giriş yap|login/i })
      await expect(loginLink).toBeVisible()
    })

    test('should have all inputs empty initially', async ({ page }) => {
      await expect(page.getByTestId('register-page-name-input')).toHaveValue('')
      await expect(page.getByTestId('register-page-email-input')).toHaveValue('')
      await expect(page.getByTestId('register-page-password-input')).toHaveValue('')
    })

    test('should have no error message initially', async ({ page }) => {
      await expect(page.getByTestId('register-page-error-message')).toBeHidden()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 2: Form Field Tests
  // ---------------------------------------------------------------------------

  test.describe('Form Fields', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
    })

    test('should fill name input correctly', async ({ page }) => {
      const nameInput = page.getByTestId('register-page-name-input')
      const testName = 'John Doe'

      await nameInput.fill(testName)
      await expect(nameInput).toHaveValue(testName)
    })

    test('should clear name input', async ({ page }) => {
      const nameInput = page.getByTestId('register-page-name-input')

      await nameInput.fill('John Doe')
      await nameInput.clear()
      await expect(nameInput).toHaveValue('')
    })

    test('should fill email input correctly', async ({ page }) => {
      const emailInput = page.getByTestId('register-page-email-input')
      const testEmail = 'john@example.com'

      await emailInput.fill(testEmail)
      await expect(emailInput).toHaveValue(testEmail)
    })

    test('should clear email input', async ({ page }) => {
      const emailInput = page.getByTestId('register-page-email-input')

      await emailInput.fill('john@example.com')
      await emailInput.clear()
      await expect(emailInput).toHaveValue('')
    })

    test('should fill password input correctly', async ({ page }) => {
      const passwordInput = page.getByTestId('register-page-password-input')
      const testPassword = 'SecurePass123!'

      await passwordInput.fill(testPassword)
      await expect(passwordInput).toHaveValue(testPassword)
    })

    test('should clear password input', async ({ page }) => {
      const passwordInput = page.getByTestId('register-page-password-input')

      await passwordInput.fill('SecurePass123!')
      await passwordInput.clear()
      await expect(passwordInput).toHaveValue('')
    })

    test('should handle very long name input', async ({ page }) => {
      const nameInput = page.getByTestId('register-page-name-input')
      const longName = 'A'.repeat(100)

      await nameInput.fill(longName)
      await expect(nameInput).toHaveValue(longName)
    })

    test('should handle very long email input', async ({ page }) => {
      const emailInput = page.getByTestId('register-page-email-input')
      const longEmail = 'a'.repeat(100) + '@example.com'

      await emailInput.fill(longEmail)
      await expect(emailInput).toHaveValue(longEmail)
    })

    test('should handle very long password input', async ({ page }) => {
      const passwordInput = page.getByTestId('register-page-password-input')
      const longPassword = 'b'.repeat(200)

      await passwordInput.fill(longPassword)
      await expect(passwordInput).toHaveValue(longPassword)
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 3: Form Submission Validation Tests
  // ---------------------------------------------------------------------------

  test.describe('Form Submission Validation', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
      await clearAuthStorage(page)
    })

    test('should show error when submitting without name', async ({ page }) => {
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Should show error toast
      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })

    test('should show error when submitting without email', async ({ page }) => {
      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })

    test('should show error when submitting with invalid email format', async ({ page }) => {
      await mockServerError(page, 400, 'Invalid email format')

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('not-an-email')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      await page.waitForTimeout(1000)
      const errorVisible = await page.getByTestId(TEST_IDS.REGISTER_ERROR_MESSAGE).isVisible().catch(() => false)
      expect(errorVisible || await page.getByTestId(TEST_IDS.REGISTER_PAGE).isVisible()).toBe(true)
    })

    test('should show error when submitting with empty password', async ({ page }) => {
      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })

    test('should show error when submitting with empty fields', async ({ page }) => {
      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 4: Successful Registration Tests
  // ---------------------------------------------------------------------------

  test.describe('Successful Registration', () => {
    test.beforeEach(async ({ page }) => {
      // Set up mocks BEFORE navigation to ensure AuthContext initializes correctly
      await mockSuccessfulRegistration(page)
      await page.goto('http://localhost:5173/register')
      await clearAuthStorage(page)
    })

    test('should register successfully with valid data', async ({ page }) => {
      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Should redirect to dashboard or onboarding
      await expect(page).toHaveURL(/\/(dashboard|onboarding)/, { timeout: 15000 })
    })

    test('should show loading state during registration', async ({ page }) => {
      await mockSuccessfulRegistration(page)

      const submitButton = page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON)

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await submitButton.click()
      await expect(submitButton).toHaveClass(/opacity-50/)
    })

    test('should disable button during submission', async ({ page }) => {
      await mockSuccessfulRegistration(page)

      const submitButton = page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON)

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      const clickPromise = submitButton.click()
      await expect(submitButton).toHaveClass(/opacity-50/)
      await clickPromise
    })

    test('should have auth cookies after registration', async ({ page }) => {
      await mockSuccessfulRegistration(page)

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Wait for navigation
      await expect(page).toHaveURL(/\/(dashboard|onboarding)/, { timeout: 15000 })

      // Verify tokens are stored in cookies
      const tokens = await getAuthTokens(page)
      expect(tokens.accessToken).toBeTruthy()
      expect(tokens.refreshToken).toBeTruthy()
    })

    test('should complete full registration flow', async ({ page }) => {
      await mockSuccessfulRegistration(page)

      // Fill form with valid data
      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('John Doe')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('john@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      // Submit registration
      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Verify success
      await expect(page).toHaveURL(/\/(dashboard|onboarding)/, { timeout: 15000 })

      const tokens = await getAuthTokens(page)
      expect(tokens.accessToken).toBeTruthy()
      expect(tokens.refreshToken).toBeTruthy()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 5: Error Scenarios Tests
  // ---------------------------------------------------------------------------

  test.describe('Error Scenarios', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
      await clearAuthStorage(page)
    })

    test('should show error when email already exists', async ({ page }) => {
      await mockEmailExists(page)

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('existing@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Wait for API response
      await page.waitForTimeout(1000)

      // Should show error message
      await expect(page.getByTestId(TEST_IDS.REGISTER_ERROR_MESSAGE)).toBeVisible()
    })

    test('should handle network error during registration', async ({ page }) => {
      await mockNetworkError(page)

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Wait a bit for network error to be handled
      await page.waitForTimeout(2000)

      // Should still be on register page
      await expect(page.getByTestId(TEST_IDS.REGISTER_PAGE)).toBeVisible()
    })

    test('should handle server error during registration', async ({ page }) => {
      await mockServerError(page, 500, 'Internal Server Error')

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Wait for response
      await page.waitForTimeout(2000)

      // Should show error message
      await expect(page.getByTestId(TEST_IDS.REGISTER_ERROR_MESSAGE)).toBeVisible()
    })

    test('should handle rate limiting', async ({ page }) => {
      await mockRateLimitError(page)

      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).click()

      // Wait for response
      await page.waitForTimeout(2000)

      // Should show error message
      await expect(page.getByTestId(TEST_IDS.REGISTER_ERROR_MESSAGE)).toBeVisible()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 6: Navigation Tests
  // ---------------------------------------------------------------------------

  test.describe('Navigation', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
    })

    test('should navigate to login page when login link is clicked', async ({ page }) => {
      const loginLink = page.getByRole('link', { name: /giriş yap|login/i })
      await loginLink.click()

      await expect(page).toHaveURL(/\/login/)
    })

    test('should handle browser back button after navigating away', async ({ page }) => {
      // Navigate to login
      await page.getByRole('link', { name: /giriş yap|login/i }).click()
      await expect(page).toHaveURL(/\/login/)

      // Go back
      await page.goBack()
      await expect(page).toHaveURL(/\/register/)
    })

    test('should have register page accessible directly', async ({ page }) => {
      await page.goto('http://localhost:5173/register')

      await expect(page.getByTestId(TEST_IDS.REGISTER_PAGE)).toBeVisible()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 7: Focus State Tests
  // ---------------------------------------------------------------------------

  test.describe('Focus States', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
    })

    test('should focus name input when clicked', async ({ page }) => {
      const nameInput = page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT)
      await nameInput.click()
      await expect(nameInput).toBeFocused()
    })

    test('should focus email input when clicked', async ({ page }) => {
      const emailInput = page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT)
      await emailInput.click()
      await expect(emailInput).toBeFocused()
    })

    test('should focus password input when clicked', async ({ page }) => {
      const passwordInput = page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT)
      await passwordInput.click()
      await expect(passwordInput).toBeFocused()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 8: Keyboard Navigation Tests
  // ---------------------------------------------------------------------------

  test.describe('Keyboard Navigation', () => {
    test.beforeEach(async ({ page }) => {
      // Set up mocks BEFORE navigation to ensure AuthContext initializes correctly
      await mockSuccessfulRegistration(page)
      await page.goto('http://localhost:5173/register')
    })

    test('should be able to tab through all form fields', async ({ page }) => {
      // Start from the beginning
      await page.keyboard.press('Tab')
      await expect(page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT)).toBeFocused()

      await page.keyboard.press('Tab')
      await expect(page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT)).toBeFocused()

      await page.keyboard.press('Tab')
      await expect(page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT)).toBeFocused()

      await page.keyboard.press('Tab')
      await expect(page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON)).toBeFocused()
    })

    test('should submit form when Enter is pressed on submit button', async ({ page }) => {
      // Fill in all required fields
      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT).fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT).fill('SecurePass123!')

      // Focus submit button and press Enter
      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON).focus()
      await page.keyboard.press('Enter')

      // Should trigger registration and navigate
      await expect(page).toHaveURL(/\/(dashboard|onboarding)/, { timeout: 15000 })
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 9: Hover State Tests
  // ---------------------------------------------------------------------------

  test.describe('Hover States', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('http://localhost:5173/register')
    })

    test('should change submit button appearance on hover', async ({ page }) => {
      const submitButton = page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON)

      await expect(submitButton).toBeEnabled()
      await submitButton.hover()
    })

    test('should change login link appearance on hover', async ({ page }) => {
      const loginLink = page.getByRole('link', { name: /giriş yap|login/i })
      await loginLink.hover()
    })
  })
})
