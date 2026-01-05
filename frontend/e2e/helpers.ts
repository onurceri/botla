import { Page, expect } from '@playwright/test'

/**
 * Sets up API mocks for authentication-related endpoints.
 * This enables testing without a real backend.
 */
export async function setupAuthMocks(page: Page) {
  await page.route('**/api/v1/auth/register', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ ok: true }),
    })
  })

  await page.route('**/api/v1/auth/login', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ token: 'test-token', refresh_token: 'test-refresh' }),
    })
  })

  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id: 'user-1', email: 'test@example.com', plan: 'pro' }),
    })
  })

  await page.route('**/api/v1/me/onboarding', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ completed: true, skipped: false }),
    })
  })
}

/**
 * Sets up API mocks for organization and workspace endpoints.
 */
export async function setupOrgMocks(page: Page) {
  await page.route('**/api/v1/organizations', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'org-1',
          name: 'Test Org',
          slug: 'test-org',
          owner_id: 'user-1',
          plan_id: 'pro',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ]),
    })
  })

  await page.route('**/api/v1/organizations/org-1/workspaces', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'ws-1',
          organization_id: 'org-1',
          name: 'Test Workspace',
          slug: 'test-ws',
          created_at: new Date().toISOString(),
        },
      ]),
    })
  })

  await page.route('**/workspaces', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'ws-1',
          organization_id: 'org-1',
          name: 'Test Workspace',
          slug: 'test-ws',
          created_at: new Date().toISOString(),
        },
      ]),
    })
  })
}

/**
 * Sets up API mocks for chatbot endpoints.
 */
export async function setupChatbotMocks(page: Page, botId: string = 'bot-1') {
  await page.route('**/api/v1/chatbots', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([]),
      })
    } else {
      // POST - create chatbot
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ id: botId, name: 'Test Bot' }),
      })
    }
  })

  await page.route(`**/api/v1/chatbots/${botId}`, async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: botId,
          name: 'Test Bot',
          welcome_message: 'Hello!',
          created_at: new Date().toISOString(),
        }),
      })
    } else if (route.request().method() === 'PUT' || route.request().method() === 'PATCH') {
      const body = route.request().postDataJSON()
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: botId, ...body }),
      })
    } else if (route.request().method() === 'DELETE') {
      await route.fulfill({
        status: 204,
        contentType: 'application/json',
        body: JSON.stringify({}),
      })
    }
  })
}

/**
 * Sets up API mocks for source endpoints.
 */
export async function setupSourceMocks(page: Page, botId: string = 'bot-1') {
  await page.route(`**/api/v1/chatbots/${botId}/sources`, async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ id: 'src-1', status: 'pending' }),
      })
    } else {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'src-1',
            source_type: 'text',
            original_filename: 'test-content.txt',
            status: 'completed',
            chunk_count: 5,
            created_at: new Date().toISOString(),
          },
        ]),
      })
    }
  })

  await page.route('**/api/v1/sources/src-1', async (route) => {
    if (route.request().method() === 'DELETE') {
      await route.fulfill({
        status: 204,
        contentType: 'application/json',
        body: JSON.stringify({}),
      })
    } else {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'src-1',
          status: 'completed',
          chunk_count: 5,
          source_type: 'text',
        }),
      })
    }
  })

  await page.route('**/api/v1/sources/src-1/chunks*', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        chunks: Array.from({ length: 5 }, (_, i) => ({
          id: `chunk-${i}`,
          score: 0.9,
          payload: {
            source_id: 'src-1',
            original_text: `This is chunk ${i} content for testing.`,
            chunk_index: i,
            created_at: new Date().toISOString(),
          },
        })),
        next_cursor: null,
      }),
    })
  })
}

/**
 * Sets up API mocks for analytics endpoints.
 */
export async function setupAnalyticsMocks(page: Page) {
  await page.route('**/api/v1/analytics', async (route) => {
    const today = new Date()
    const series = Array.from({ length: 7 }).map((_, i) => {
      const d = new Date(today)
      d.setDate(today.getDate() - (6 - i))
      return {
        date: d.toISOString().slice(0, 10),
        messages: i * 5,
        conversations: Math.floor(i * 2),
      }
    })
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(series),
    })
  })
}

/**
 * Sets up all common API mocks for authenticated tests.
 */
export async function setupAllMocks(page: Page, botId: string = 'bot-1') {
  await setupAuthMocks(page)
  await setupOrgMocks(page)
  await setupChatbotMocks(page, botId)
  await setupSourceMocks(page, botId)
  await setupAnalyticsMocks(page)
}

/**
 * Sets up session-related API mocks for token refresh and session management tests.
 * Includes mocks for:
 * - Token refresh endpoint
 * - Session status endpoint
 * - User info endpoint
 */
export async function setupSessionMocks(page: Page) {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          access_token: 'mock-refreshed-token-' + Date.now(),
          refresh_token: 'mock-refreshed-refresh-' + Date.now(),
          expires_in: 3600,
          token_type: 'Bearer',
        }),
      })
    }
  })

  await page.route('**/api/v1/auth/session', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        active: true,
        user: {
          id: 'user-123',
          email: 'test@example.com',
          name: 'Test User',
          plan: 'pro',
        },
        expires_at: new Date(Date.now() + 3600000).toISOString(),
      }),
    })
  })

  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        plan: 'pro',
      }),
    })
  })
}

/**
 * Sets up session mocks with expired token handling.
 * Used for testing token refresh flows.
 */
export async function setupSessionMocksWithExpiry(page: Page) {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          access_token: 'mock-refreshed-token-' + Date.now(),
          expires_in: 3600,
          token_type: 'Bearer',
        }),
      })
    }
  })

  await page.route('**/api/v1/**', async (route) => {
    const url = route.request().url()
    if (url.includes('/api/v1/') && !url.includes('/auth/')) {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'TOKEN_EXPIRED',
          message: 'Session has expired',
          code: 'TOKEN_EXPIRED',
          requiresRelogin: true,
        }),
      })
    }
  })
}

/**
 * Performs login via the login page.
 * If mocks are not set up, will attempt real login.
 */
export async function login(
  page: Page,
  email: string = 'test@example.com',
  password: string = 'password123',
) {
  await page.goto('/login')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Şifre').fill(password)
  await page.getByRole('button', { name: 'Giriş Yap' }).click()
  await page.waitForURL(/\/(dashboard|onboarding|\/)/, { timeout: 10000 })
}

/**
 * Performs registration via the register page.
 */
export async function register(
  page: Page,
  name: string = 'Test User',
  email?: string,
  password: string = 'SecurePass123!',
) {
  const userEmail = email || `test-${Date.now()}@example.com`
  await page.goto('/register')
  await page.getByLabel('Ad Soyad').fill(name)
  await page.getByLabel('Email').fill(userEmail)
  await page.getByLabel('Şifre').fill(password)
  await page.getByRole('button', { name: 'Kayıt Ol' }).click()
  return userEmail
}

/**
 * Waits for the dashboard to be fully loaded.
 */
export async function waitForDashboard(page: Page) {
  await expect(page).toHaveURL(/\/(dashboard)?\/?$/, { timeout: 10000 })
  // Wait for any dashboard element to confirm load
  await page.waitForLoadState('networkidle')
}

/**
 * Creates a chatbot through the UI.
 */
export async function createChatbot(page: Page, name: string = 'Test Bot'): Promise<string> {
  // Click create button
  const createBtn = page.getByRole('button', { name: /Yeni Chatbot|Yeni Oluştur/ }).first()
  await createBtn.waitFor({ state: 'visible', timeout: 10000 })
  await createBtn.click()

  // Fill in the name
  await page.getByPlaceholder('Örn: Müşteri Temsilcisi').fill(name)

  // Submit
  await page.getByRole('button', { name: 'Oluştur' }).click()

  // Wait for navigation to chatbot detail page
  await expect(page).toHaveURL(/\/chatbots\/[a-zA-Z0-9_-]+$/, { timeout: 10000 })

  // Extract and return the bot ID
  const url = new URL(page.url())
  const botId = url.pathname.split('/').pop() || ''
  return botId
}

/**
 * Adds a text source to the current chatbot.
 */
export async function addTextSource(page: Page, content: string = 'Test content for the chatbot.') {
  // Navigate to sources tab
  await page.getByRole('tab', { name: 'Veri Kaynakları' }).click()

  // Click text source option
  await page.getByText('Metin Gir', { exact: true }).last().click()

  // Fill in the content
  await page.getByPlaceholder('Metin içeriğini buraya yapıştırın...').fill(content)

  // Submit
  await page.getByRole('button', { name: 'Ekle' }).last().click()

  // Wait for success message
  await expect(page.getByText('Metin kaynağı eklendi.')).toBeVisible({ timeout: 5000 })
}

/**
 * Adds a URL source to the current chatbot.
 */
export async function addUrlSource(page: Page, url: string = 'https://example.com') {
  // Navigate to sources tab
  await page.getByRole('tab', { name: 'Veri Kaynakları' }).click()

  // Click URL source option
  await page.getByText('URL Ekle', { exact: true }).last().click()

  // Fill in the URL
  await page.getByPlaceholder(/URL|Adres/).fill(url)

  // Submit
  await page.getByRole('button', { name: /Ekle|Kaydet/ }).last().click()
}

/**
 * Opens the playground/test area for the current chatbot.
 */
export async function openPlayground(page: Page) {
  await page.getByRole('tab', { name: 'Test Alanı' }).click()
  await page.getByRole('button', { name: 'Sohbeti aç' }).click()
}

/**
 * Sends a message in the playground chat.
 */
export async function sendChatMessage(page: Page, message: string) {
  const input = page.getByPlaceholder('Mesaj yazın...')
  await input.fill(message)
  await input.press('Enter')
}
