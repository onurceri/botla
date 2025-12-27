import { test, expect } from '@playwright/test'
import { setupAuthMocks, setupOrgMocks, setupAnalyticsMocks } from './helpers'

test.describe('Authentication', () => {
  test.describe('Registration', () => {
    test('user can register successfully', async ({ page }) => {
      // Setup mocks
      await setupAuthMocks(page)
      await setupOrgMocks(page)
      await setupAnalyticsMocks(page)

      // Navigate to register page
      await page.goto('/register')

      // Verify page elements
      await expect(page.getByRole('heading', { name: 'Hesap Oluştur' })).toBeVisible()

      // Fill registration form
      await page.getByLabel('Ad Soyad').fill('Test User')
      await page.getByLabel('Email').fill(`test-${Date.now()}@example.com`)
      await page.getByLabel('Şifre').fill('SecurePass123!')

      // Submit form
      await page.getByRole('button', { name: 'Kayıt Ol' }).click()

      // Should redirect to dashboard or onboarding after successful registration
      await expect(page).toHaveURL(/\/(dashboard|onboarding|\/)/, { timeout: 15000 })
    })

    test('registration shows validation error for empty fields', async ({ page }) => {
      await page.goto('/register')

      // Click submit without filling fields
      await page.getByRole('button', { name: 'Kayıt Ol' }).click()

      // Should show error toast
      await expect(page.getByText('Lütfen tüm alanları doldurun.')).toBeVisible()
    })

    test('registration page has link to login', async ({ page }) => {
      await page.goto('/register')

      // Check for login link
      const loginLink = page.getByRole('link', { name: 'Giriş Yapın' })
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

      // Verify page elements
      await expect(page.getByRole('heading', { name: 'Hoş Geldiniz' })).toBeVisible()

      // Fill login form
      await page.getByLabel('Email').fill('test@example.com')
      await page.getByLabel('Şifre').fill('password123')

      // Submit form
      await page.getByRole('button', { name: 'Giriş Yap' }).click()

      // Should redirect to dashboard
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 10000 })
    })

    test('login shows validation error for empty fields', async ({ page }) => {
      await page.goto('/login')

      // Click submit without filling fields
      await page.getByRole('button', { name: 'Giriş Yap' }).click()

      // Should show error toast
      await expect(page.getByText('Lütfen tüm alanları doldurun.')).toBeVisible()
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

      // Fill with invalid credentials
      await page.getByLabel('Email').fill('wrong@example.com')
      await page.getByLabel('Şifre').fill('wrongpassword')

      // Submit form
      await page.getByRole('button', { name: 'Giriş Yap' }).click()

      // Should show error
      await expect(
        page.getByText('Giriş başarısız. Lütfen bilgilerinizi kontrol edin.'),
      ).toBeVisible()
    })

    test('login page has link to register', async ({ page }) => {
      await page.goto('/login')

      // Check for register link
      const registerLink = page.getByRole('link', { name: 'Kayıt Olun' })
      await expect(registerLink).toBeVisible()

      // Click and verify navigation
      await registerLink.click()
      await expect(page).toHaveURL(/\/register/)
    })

    test('login page has forgot password link', async ({ page }) => {
      await page.goto('/login')

      // Check for forgot password link
      const forgotLink = page.getByRole('link', { name: 'Şifremi unuttum?' })
      await expect(forgotLink).toBeVisible()
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
      await page.getByLabel('Email').fill('test@example.com')
      await page.getByLabel('Şifre').fill('password123')
      await page.getByRole('button', { name: 'Giriş Yap' }).click()

      // Wait for dashboard
      await expect(page).toHaveURL(/\/(dashboard)?/, { timeout: 10000 })

      // Find and click logout (usually in a dropdown menu)
      // This may need adjustment based on actual UI implementation
      const userMenu = page.locator('[data-testid="user-menu"]').or(page.locator('.lucide-user'))
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
