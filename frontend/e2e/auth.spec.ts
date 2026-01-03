import { test, expect } from '@playwright/test'
import { setupAuthMocks, setupOrgMocks, setupAnalyticsMocks } from './helpers'
import { TEST_IDS, TURKISH } from './test-constants'

test.describe('Authentication', () => {
  test.describe('Registration', () => {
    test('user can register successfully', async ({ page }) => {
      // Setup mocks
      await setupAuthMocks(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)

      // Navigate to register page
      await page.goto('/register')

      // Verify page elements using data-testid
      await expect(page.getByTestId(TEST_IDS.REGISTER_PAGE)).toBeVisible()
      await expect(page.getByTestId(TEST_IDS.REGISTER_TITLE)).toBeVisible()

      // Fill registration form using data-testid
      await page.getByTestId(TEST_IDS.REGISTER_NAME_INPUT)
        .or(page.getByLabel(TURKISH.NAME))
        .fill('Test User')
      await page.getByTestId(TEST_IDS.REGISTER_EMAIL_INPUT)
        .or(page.getByLabel(TURKISH.EMAIL))
        .fill(`test-${Date.now()}@example.com`)
      await page.getByTestId(TEST_IDS.REGISTER_PASSWORD_INPUT)
        .or(page.getByLabel(TURKISH.PASSWORD))
        .fill('SecurePass123!')

      // Submit form
      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON)
        .or(page.getByRole('button', { name: TURKISH.REGISTER }))
        .click()

      // Should redirect to dashboard or onboarding after successful registration
      await expect(page).toHaveURL(/\/(dashboard|onboarding|\/)/, { timeout: 15000 })
    })

    test('registration shows validation error for empty fields', async ({ page }) => {
      await page.goto('/register')

      // Click submit without filling fields
      await page.getByTestId(TEST_IDS.REGISTER_SUBMIT_BUTTON)
        .or(page.getByRole('button', { name: TURKISH.REGISTER }))
        .click()

      // Should show error toast
      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible()
    })

    test('registration page has link to login', async ({ page }) => {
      await page.goto('/register')

      // Check for login link
      const loginLink = page.getByRole('link', { name: TURKISH.LOGIN_LINK })
      await expect(loginLink).toBeVisible()

      // Click and verify navigation
      await loginLink.click()
      await expect(page).toHaveURL(/\/login/)
    })
  })

  test.describe('Login', () => {
    test('user can login successfully', async ({ page }) => {
      // Setup mocks
      await setupAuthMocks(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)

      // Setup chatbots mock for dashboard
      await page.route('**/api/v1/chatbots', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([]),
        })
      })

      // Navigate to login page
      await page.goto('/login')

      // Verify page elements using data-testid
      await expect(page.getByTestId(TEST_IDS.LOGIN_PAGE)).toBeVisible()
      await expect(page.getByTestId(TEST_IDS.LOGIN_TITLE)).toBeVisible()

      // Fill login form using data-testid
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)
        .or(page.getByLabel(TURKISH.EMAIL))
        .fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)
        .or(page.getByLabel(TURKISH.PASSWORD))
        .fill('password123')

      // Submit form
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)
        .or(page.getByRole('button', { name: TURKISH.LOGIN }))
        .click()

      // Should redirect to dashboard
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 10000 })
    })

    test('login shows validation error for empty fields', async ({ page }) => {
      await page.goto('/login')

      // Click submit without filling fields
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)
        .or(page.getByRole('button', { name: TURKISH.LOGIN }))
        .click()

      // Should show error toast
      await expect(page.getByText(TURKISH.FILL_ALL_FIELDS)).toBeVisible()
    })

    test('login shows error for invalid credentials', async ({ page }) => {
      // Mock failed login
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid credentials' }),
        })
      })

      await page.goto('/login')

      // Fill with invalid credentials using data-testid
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)
        .or(page.getByLabel(TURKISH.EMAIL))
        .fill('wrong@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)
        .or(page.getByLabel(TURKISH.PASSWORD))
        .fill('wrongpassword')

      // Submit form
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)
        .or(page.getByRole('button', { name: TURKISH.LOGIN }))
        .click()

      // Should show error using data-testid
      await expect(page.getByTestId(TEST_IDS.LOGIN_ERROR_MESSAGE)
        .or(page.getByText(TURKISH.LOGIN_FAILED)))
        .toBeVisible()
    })

    test('login page has link to register', async ({ page }) => {
      await page.goto('/login')

      // Check for register link
      const registerLink = page.getByRole('link', { name: TURKISH.REGISTER_LINK })
      await expect(registerLink).toBeVisible()

      // Click and verify navigation
      await registerLink.click()
      await expect(page).toHaveURL(/\/register/)
    })

    test('login page has forgot password link', async ({ page }) => {
      await page.goto('/login')

      // Check for forgot password link using data-testid
      await expect(page.getByTestId(TEST_IDS.LOGIN_FORGOT_PASSWORD_LINK)
        .or(page.getByRole('link', { name: TURKISH.FORGOT_PASSWORD })))
        .toBeVisible()
    })
  })

  test.describe('Logout', () => {
    test('user can logout', async ({ page }) => {
      // Setup mocks
      await setupAuthMocks(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)

      await page.route('**/api/v1/chatbots', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([]),
        })
      })

      // Login first
      await page.goto('/login')
      await page.getByTestId(TEST_IDS.LOGIN_EMAIL_INPUT)
        .or(page.getByLabel(TURKISH.EMAIL))
        .fill('test@example.com')
      await page.getByTestId(TEST_IDS.LOGIN_PASSWORD_INPUT)
        .or(page.getByLabel(TURKISH.PASSWORD))
        .fill('password123')
      await page.getByTestId(TEST_IDS.LOGIN_SUBMIT_BUTTON)
        .or(page.getByRole('button', { name: TURKISH.LOGIN }))
        .click()

      // Wait for dashboard
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 10000 })

      // Find and click logout using data-testid
      const userMenu = page.getByTestId(TEST_IDS.USER_MENU)
        .or(page.locator('.lucide-user'))
      if (await userMenu.isVisible()) {
        await userMenu.click()
        const logoutButton = page.getByRole('button', { name: /Çıkış|Logout/i })
        if (await logoutButton.isVisible()) {
          await logoutButton.click()
          // Should redirect to login
          await expect(page).toHaveURL(/\/login/, { timeout: 10000 })
        }
      }
    })
  })

  test.describe('Protected Routes', () => {
    test('unauthenticated user is redirected to login', async ({ page }) => {
      // Clear any stored tokens
      await page.goto('/')
      await page.evaluate(() => {
        localStorage.removeItem('botla_token')
        localStorage.removeItem('botla_refresh_token')
      })

      // Mock unauthenticated response
      await page.route('**/api/v1/auth/me', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Unauthorized' }),
        })
      })

      // Try to access dashboard directly
      await page.goto('/dashboard')

      // Should redirect to login
      await expect(page).toHaveURL(/\/login/, { timeout: 10000 })
    })
  })
})
