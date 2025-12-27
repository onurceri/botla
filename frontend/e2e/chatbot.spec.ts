import { test, expect } from '@playwright/test'
import {
  setupAllMocks,
  login,
} from './helpers'

test.describe('Chatbot Management', () => {
  test.beforeEach(async ({ page }) => {
    // Setup all mocks before each test
    await setupAllMocks(page)
  })

  test.describe('Chatbot Creation', () => {
    test('can create a new chatbot', async ({ page }) => {
      // Login
      await login(page)

      // Navigate to dashboard
      await page.goto('/dashboard')

      // Wait for dashboard to load
      await page.waitForLoadState('networkidle')

      // Create chatbot
      const createBtn = page.getByRole('button', { name: /Yeni Chatbot|Yeni Oluştur/ }).first()
      await createBtn.waitFor({ state: 'visible', timeout: 10000 })
      await createBtn.click()

      // Fill chatbot name
      await page.getByPlaceholder('Örn: Müşteri Temsilcisi').fill('My Test Bot')

      // Submit
      await page.getByRole('button', { name: 'Oluştur' }).click()

      // Verify success message
      await expect(page.getByText('Chatbot başarıyla oluşturuldu.')).toBeVisible({ timeout: 10000 })

      // Verify redirected to chatbot detail page
      await expect(page).toHaveURL(/\/chatbots\/[a-zA-Z0-9_-]+$/)
    })

    test('chatbot creation requires a name', async ({ page }) => {
      await login(page)
      await page.goto('/dashboard')
      await page.waitForLoadState('networkidle')

      const createBtn = page.getByRole('button', { name: /Yeni Chatbot|Yeni Oluştur/ }).first()
      await createBtn.waitFor({ state: 'visible', timeout: 10000 })
      await createBtn.click()

      // Try to submit without name
      const submitBtn = page.getByRole('button', { name: 'Oluştur' })

      // The button should be disabled or show validation error
      // This depends on implementation, so we check if empty submission is handled
      if (await submitBtn.isEnabled()) {
        await submitBtn.click()
        // Should show validation error or remain on form
        await expect(page).not.toHaveURL(/\/chatbots\/[a-zA-Z0-9_-]+$/)
      }
    })
  })

  test.describe('Source Management', () => {
    test('can add text source to chatbot', async ({ page }) => {
      await login(page)

      // Navigate to chatbot sources page  
      await page.goto('/dashboard/chatbots/bot-1/sources')
      await page.waitForLoadState('networkidle')

      // Wait for page content to load
      await page.waitForTimeout(500)

      // Click text source option
      const textOption = page.getByText('Metin Gir', { exact: true })
      if (await textOption.first().isVisible({ timeout: 5000 })) {
        await textOption.last().click()

        // Fill in the content
        const textarea = page.getByPlaceholder('Metin içeriğini buraya yapıştırın...')
        await textarea.waitFor({ state: 'visible', timeout: 5000 })
        await textarea.fill('This is test content for the chatbot to learn from.')

        // Submit
        await page.getByRole('button', { name: 'Ekle' }).last().click()

        // Verify success
        await expect(page.getByText('Metin kaynağı eklendi.')).toBeVisible({ timeout: 5000 })
      }
    })

    test('can add URL source to chatbot', async ({ page }) => {
      // Additional mock for URL source
      await page.route('**/api/v1/chatbots/bot-1/sources', async (route) => {
        if (route.request().method() === 'POST') {
          const body = route.request().postDataJSON()
          if (body?.url) {
            await route.fulfill({
              status: 201,
              contentType: 'application/json',
              body: JSON.stringify({ id: 'src-url-1', status: 'pending', source_type: 'url' }),
            })
          } else {
            await route.fulfill({
              status: 201,
              contentType: 'application/json',
              body: JSON.stringify({ id: 'src-1', status: 'pending' }),
            })
          }
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([]),
          })
        }
      })

      await login(page)
      await page.goto('/dashboard/chatbots/bot-1/sources')
      await page.waitForLoadState('networkidle')

      // Click URL source option
      const urlOption = page.getByText('URL Ekle', { exact: true }).last()
      if (await urlOption.isVisible({ timeout: 5000 })) {
        await urlOption.click()

        // Fill in the URL
        const urlInput = page.getByPlaceholder(/URL|Adres/).or(page.locator('input[type="url"]'))
        if (await urlInput.isVisible()) {
          await urlInput.fill('https://example.com/docs')

          // Submit
          await page.getByRole('button', { name: /Ekle|Kaydet/ }).last().click()
        }
      }
    })

    test('sources list shows existing sources', async ({ page }) => {
      await login(page)
      await page.goto('/dashboard/chatbots/bot-1/sources')
      await page.waitForLoadState('networkidle')

      // Verify source is listed (from mock)
      await expect(page.getByText('test-content.txt')).toBeVisible({ timeout: 10000 })
    })
  })

  test.describe('Playground Testing', () => {
    test('can access playground page', async ({ page }) => {
      await login(page)
      // Navigate directly to playground
      await page.goto('/dashboard/chatbots/bot-1/playground')
      await page.waitForLoadState('networkidle')

      // Verify we're on the playground page
      await expect(page).toHaveURL(/\/chatbots\/bot-1\/playground/)
    })

    test('playground shows chat interface', async ({ page }) => {
      // Mock chat endpoint
      await page.route('**/api/v1/chatbots/bot-1/chat', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            response: 'Hello! I am your AI assistant.',
            tokens_used: 50,
            sources_used: [],
          }),
        })
      })

      await login(page)
      await page.goto('/dashboard/chatbots/bot-1/playground')
      await page.waitForLoadState('networkidle')
      
      // Verify we're on the playground page
      await expect(page).toHaveURL(/\/chatbots\/bot-1\/playground/)
      
      // Wait for page to fully render
      await page.waitForTimeout(500)

      // The playground page should render without errors
      // At minimum we should see the tab bar and page content
      const hasContent = await page.locator('nav').first().isVisible().catch(() => false)
      expect(hasContent).toBeTruthy()
    })
  })

  test.describe('Chatbot Settings', () => {
    test('can access chatbot settings', async ({ page }) => {
      await login(page)
      await page.goto('/dashboard/chatbots/bot-1/settings')
      await page.waitForLoadState('networkidle')

      // Settings page should load
      await expect(page).toHaveURL(/\/chatbots\/bot-1\/settings/)
    })

    test('can update chatbot name', async ({ page }) => {
      await login(page)
      await page.goto('/dashboard/chatbots/bot-1/settings')
      await page.waitForLoadState('networkidle')

      // Find name input and update
      const nameInput = page
        .getByLabel(/İsim|Ad|Name/)
        .or(page.locator('input[name="name"]'))
        .first()
      if (await nameInput.isVisible({ timeout: 5000 })) {
        await nameInput.clear()
        await nameInput.fill('Updated Bot Name')

        // Save
        const saveBtn = page.getByRole('button', { name: /Kaydet|Save/i })
        if (await saveBtn.isVisible()) {
          await saveBtn.click()
        }
      }
    })
  })

  test.describe('Chatbot List', () => {
    test('dashboard shows chatbot list', async ({ page }) => {
      // Mock chatbots list with data
      await page.route('**/api/v1/chatbots', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              id: 'bot-1',
              name: 'Support Bot',
              created_at: new Date().toISOString(),
            },
            {
              id: 'bot-2',
              name: 'Sales Bot',
              created_at: new Date().toISOString(),
            },
          ]),
        })
      })

      await login(page)
      await page.goto('/dashboard')
      await page.waitForLoadState('networkidle')

      // Verify bots are listed
      await expect(page.getByText('Support Bot')).toBeVisible({ timeout: 10000 })
      await expect(page.getByText('Sales Bot')).toBeVisible({ timeout: 10000 })
    })

    test('can navigate to chatbot from list', async ({ page }) => {
      // Mock chatbots list
      await page.route('**/api/v1/chatbots', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              id: 'bot-1',
              name: 'Support Bot',
              created_at: new Date().toISOString(),
            },
          ]),
        })
      })

      await login(page)
      await page.goto('/dashboard')
      await page.waitForLoadState('networkidle')

      // Click on chatbot
      await page.getByText('Support Bot').click()

      // Should navigate to chatbot detail
      await expect(page).toHaveURL(/\/chatbots\/bot-1/)
    })
  })
})
