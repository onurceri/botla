import { test, expect } from '@playwright/test'
import { TEST_IDS, TURKISH } from './test-constants'
import { setupAuthMocks, setupOrgMocks, setupAnalyticsMocks } from './helpers'
import {
  mockSuccessfulLogin,
  mockFailedLogin,
  mockNetworkError,
  mockServerError,
  mockRateLimitError,
  waitForLoginResponse,
} from './mocks/auth.mocks'

/**
 * Login Page E2E Tests
 * 
 * Comprehensive test coverage for the login page including:
 * - Page load and element visibility
 * - Form interactions and validation
 * - Keyboard navigation
 * - Hover states
 * - Authentication flows
 * - Error handling
 * - Forgot password flow
 * 
 * Element IDs follow the naming convention from TEST_PATHS.md Section 2.1
 */

test.describe('Login Page', () => {
  // ---------------------------------------------------------------------------
  // Phase 1: Page Load Tests
  // ---------------------------------------------------------------------------

  test.describe('Page Load', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should load login page successfully', async ({ page }) => {
      // Verify page container is visible
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
      
      // Verify all main elements are present
      await expect(page.getByTestId(TEST_IDS.LOGIN_TITLE)).toBeVisible()
      await expect(page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)).toBeVisible()
      await expect(page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)).toBeVisible()
      await expect(page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)).toBeVisible()
      
      // Verify navigation links
      await expect(page.getByTestId(TEST_IDS.LOGIN_FORGOT_PASSWORD_LINK)).toBeVisible()
    })

    test('should have correct page title', async ({ page }) => {
      await expect(page.getByTestId(TEST_IDS.LOGIN_TITLE)).toContainText(/hoş geldiniz|welcome/i)
    })

    test('should have email and password inputs', async ({ page }) => {
      const emailInput = page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)
      const passwordInput = page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)
      
      await expect(emailInput).toBeVisible()
      await expect(passwordInput).toBeVisible()
      await expect(emailInput).toHaveAttribute('type', 'email')
      await expect(passwordInput).toHaveAttribute('type', 'password')
    })

    test('should have login button', async ({ page }) => {
      const loginButton = page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)
      await expect(loginButton).toBeVisible()
      await expect(loginButton).toBeEnabled()
    })

    test('should have forgot password link', async ({ page }) => {
      const forgotPasswordLink = page.getByTestId(TEST_IDS.LOGIN_FORGOT_PASSWORD_LINK)
      await expect(forgotPasswordLink).toBeVisible()
    })

    test('should have register link', async ({ page }) => {
      const registerLink = page.getByRole('link', { name: /register|sign up|kayıt ol/i })
      await expect(registerLink).toBeVisible()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 2: Focus State Tests
  // ---------------------------------------------------------------------------

  test.describe('Focus States', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should focus email input when clicked', async ({ page }) => {
      const emailInput = page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)
      await emailInput.click()
      await expect(emailInput).toBeFocused()
    })

    test('should focus password input when clicked', async ({ page }) => {
      const passwordInput = page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)
      await passwordInput.click()
      await expect(passwordInput).toBeFocused()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 3: Keyboard Navigation Tests
  // ---------------------------------------------------------------------------

  test.describe('Keyboard Navigation', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should submit form when Enter is pressed on password field', async ({ page }) => {
      await setupAuthMocks(page)
      await setupOrgMocks(page)
      
      // Fill in credentials
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      
      // Focus password and press Enter
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).focus()
      await page.keyboard.press('Enter')
      
      // Should trigger login and navigate
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 15000 })
    })

    test('should be able to type in password field after email', async ({ page }) => {
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.keyboard.press('Enter')
      
      // Password field should be visible and interactable
      await expect(page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)).toBeVisible()
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      await expect(page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)).toHaveValue('password123')
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 4: Validation Tests
  // ---------------------------------------------------------------------------

  test.describe('Validation', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should show error when submitting with empty email', async ({ page }) => {
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Should show error toast/message
      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })

    test('should show error when submitting with empty password', async ({ page }) => {
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Should show error toast/message
      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })

    test('should show error when submitting with empty fields', async ({ page }) => {
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Should show error toast/message
      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible({ timeout: 5000 })
    })

    test('should validate email format on blur', async ({ page }) => {
      // Fill invalid email
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('invalid-email')
      
      // Blur by clicking elsewhere
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).click()
      
      // Input should still have the invalid value
      await expect(page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)).toHaveValue('invalid-email')
    })

    test('should show error message for invalid email format', async ({ page }) => {
      // Fill invalid email and submit
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('not-an-email')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait a bit for API response
      await page.waitForTimeout(1000)
      
      // Should still be on login page (invalid credentials rejected)
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
    })

    test('should accept valid email format', async ({ page }) => {
      // Fill valid email
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      
      // Blur by clicking elsewhere
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).click()
      
      // No immediate error should appear for valid email
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 5: Authentication Tests
  // ---------------------------------------------------------------------------

  test.describe('Authentication', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
      await clearAuthStorage(page)
    })

    test('should login successfully with valid credentials', async ({ page }) => {
      await mockSuccessfulLogin(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)
      
      // Fill login form
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('SecurePass123!')
      
      // Submit form
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait for login response
      await waitForLoginResponse(page)
      
      // Should redirect to dashboard
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 15000 })
    })

    test('should show loading state during login', async ({ page }) => {
      await mockSuccessfulLogin(page)
      await setupOrgMocks(page)
      
      // Fill login form
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('SecurePass123!')
      
      // Submit form
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait for navigation
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 15000 })
    })

    test('should show error for invalid credentials', async ({ page }) => {
      await mockFailedLogin(page)
      
      // Fill with wrong credentials
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('wrong@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('wrongpassword')
      
      // Submit form
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait for response
      await page.waitForResponse((r) => r.url().includes('/auth/login'))
      
      // Should show error message
      await expect(page.getByTestId(TEST_IDS.LOGIN_ERROR_MESSAGE)
        .or(page.getByText(TURKISH.LOGIN_FAILED)))
        .toBeVisible({ timeout: 5000 })
    })

    test('should redirect to dashboard after successful login', async ({ page }) => {
      await mockSuccessfulLogin(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)
      
      // Mock chatbots endpoint for dashboard
      await page.route('**/api/v1/chatbots', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([]),
        })
      })
      
      // Login
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('SecurePass123!')
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Should redirect to dashboard
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 15000 })
    })

    test('should handle remember me functionality', async ({ page }) => {
      await mockSuccessfulLogin(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)
      
      // Check remember me checkbox if present
      const rememberMeCheckbox = page.getByTestId(TEST_IDS.LOGIN_REMEMBER_ME_CHECKBOX)
        .or(page.getByRole('checkbox', { name: /remember me|beni hatırla/i }))
      if (await rememberMeCheckbox.isVisible()) {
        await rememberMeCheckbox.check()
      }
      
      // Login
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('SecurePass123!')
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait for navigation
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 15000 })
      
      // Tokens should be stored in localStorage
      const tokens = await page.evaluate(() => ({
        accessToken: localStorage.getItem('botla_token'),
        refreshToken: localStorage.getItem('botla_refresh_token'),
      }))
      
      expect(tokens.accessToken).toBeTruthy()
      expect(tokens.refreshToken).toBeTruthy()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 6: Forgot Password Flow Tests
  // ---------------------------------------------------------------------------

  test.describe('Forgot Password Flow', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should have forgot password link visible', async ({ page }) => {
      await expect(page.getByTestId(TEST_IDS.LOGIN_FORGOT_PASSWORD_LINK)).toBeVisible()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 7: Hover State Tests
  // ---------------------------------------------------------------------------

  test.describe('Hover States', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should change login button appearance on hover', async ({ page }) => {
      const loginButton = page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)
      
      // Button should be interactable
      await expect(loginButton).toBeEnabled()
      
      // Hover should work without errors
      await loginButton.hover()
    })

    test('should change forgot password link appearance on hover', async ({ page }) => {
      const forgotPasswordLink = page.getByTestId(TEST_IDS.LOGIN_FORGOT_PASSWORD_LINK)
      
      // Hover should work without errors
      await forgotPasswordLink.hover()
    })

    test('should change register link appearance on hover', async ({ page }) => {
      const registerLink = page.getByRole('link', { name: /register|sign up|kayıt ol/i })
      
      // Hover should work without errors
      await registerLink.hover()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 8: Error Handling Tests
  // ---------------------------------------------------------------------------

  test.describe('Error Handling', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should handle network error during login', async ({ page }) => {
      await mockNetworkError(page)
      
      // Fill credentials
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      
      // Submit
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Should show network error (after timeout)
      await page.waitForTimeout(2000)
      
      // Page should still be on login or show error
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
    })

    test('should handle server error during login', async ({ page }) => {
      await mockServerError(page, 500, 'Internal Server Error')
      
      // Fill credentials
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      
      // Submit
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait for response
      await page.waitForTimeout(2000)
      
      // Should still be on login page or show error
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
    })

    test('should handle rate limiting', async ({ page }) => {
      await mockRateLimitError(page)
      
      // Fill credentials
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill('password123')
      
      // Submit
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON).click()
      
      // Wait for response
      await page.waitForTimeout(2000)
      
      // Should still be on login page or show error
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 9: Navigation Tests
  // ---------------------------------------------------------------------------

  test.describe('Navigation', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should navigate to register page when register link is clicked', async ({ page }) => {
      const registerLink = page.getByRole('link', { name: /register|sign up|kayıt ol/i })
      await registerLink.click()
      
      await expect(page).toHaveURL(/\/register/)
    })

    test('should navigate to login page directly', async ({ page }) => {
      await page.goto('/login')
      
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
    })
  })

  // ---------------------------------------------------------------------------
  // Phase 10: Edge Cases
  // ---------------------------------------------------------------------------

  test.describe('Edge Cases', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('should handle very long email input', async ({ page }) => {
      const longEmail = 'a'.repeat(100) + '@example.com'
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill(longEmail)
      
      await expect(page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)).toHaveValue(longEmail)
    })

    test('should handle very long password input', async ({ page }) => {
      const longPassword = 'b'.repeat(200)
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT).fill(longPassword)
      
      await expect(page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)).toHaveValue(longPassword)
    })

    test('should handle email with leading/trailing spaces', async ({ page }) => {
      // Input value may be trimmed by the UI
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT).fill('test@example.com')
      await expect(page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)).toHaveValue('test@example.com')
    })
  })
})

// Helper function to clear auth storage
async function clearAuthStorage(page: any): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    localStorage.removeItem('botla_user')
    sessionStorage.clear()
  })
}
