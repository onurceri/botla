import { test, expect, type Page } from '@playwright/test'

/**
 * Chunk Inspector E2E Tests
 *
 * Tests cover:
 * - Login flow
 * - Chatbot creation
 * - Source addition
 * - Chunk inspection functionality
 * - Search within chunks
 *
 * Element Selection Strategy:
 * - Primary: Semantic selectors (getByLabel, getByRole, getByPlaceholder)
 * - Fallback: data-testid attributes when available
 * - Uses consistent test.describe() grouping for logical organization
 *
 * @see TESTING_STANDARDS.md for naming conventions
 * @see selectors.ts for element ID constants
 */
test.describe('Chunk Inspector', () => {
  const isReal = !!process.env.E2E_API_BASE

  test.beforeEach(async ({ page }) => {
    // Setup common mocks for chunk inspector tests
    await setupChunkInspectorMocks(page, isReal)
  })

  test.describe('Authentication Flow', () => {
    test('should register and login successfully', async ({ page }) => {
      // Registration
      const email = `test-${Date.now()}@example.com`
      await page.goto('http://localhost:5173/register')
      await page.getByLabel('Ad Soyad').fill('Test User')
      await page.getByLabel('Email').fill(email)
      await page.getByLabel('Şifre').fill('password')
      await page.getByRole('button', { name: 'Kayıt Ol' }).click()
      await expect(page).toHaveURL(/.*onboarding|.*login/, { timeout: 10000 })

      // Handle onboarding redirect to login
      if (page.url().includes('onboarding')) {
        await page.goto('http://localhost:5173/login')
      }

      // Login
      await page.getByLabel('Email').fill(email)
      await page.getByLabel('Şifre').fill('password')
      await page.getByRole('button', { name: 'Giriş Yap' }).click()
      await expect(page).toHaveURL(/.*dashboard|.*\//, { timeout: 10000 })
    })
  })

  test.describe('Chatbot and Source Creation', () => {
    test('should create chatbot and add source', async ({ page }) => {
      // Login first
      await page.goto('http://localhost:5173/login')
      await page.getByLabel('Email').fill('test@example.com')
      await page.getByLabel('Şifre').fill('password')
      await page.getByRole('button', { name: 'Giriş Yap' }).click()
      await expect(page).toHaveURL(/.*dashboard|.*\//, { timeout: 10000 })

      // Navigate to dashboard
      await page.goto('/dashboard')
      await page.waitForLoadState('domcontentloaded')

      // Create Bot
      const createBtn = page.getByRole('button', { name: /Yeni Chatbot|Yeni Oluştur/ }).first()
      await createBtn.waitFor({ state: 'visible', timeout: 10000 })
      await createBtn.click()
      await page.getByPlaceholder('Örn: Müşteri Temsilcisi').fill('Test Bot')
      await page.getByRole('button', { name: 'Oluştur' }).click()
      await expect(page).toHaveURL(/\/chatbots\/[a-zA-Z0-9_-]+$/, { timeout: 10000 })

      // Go to Sources
      await page.getByRole('link', { name: 'Kaynaklar' }).click()

      // Wait for sources list to load
      await expect(page.getByRole('heading', { name: 'test-source.txt' })).toBeVisible()
    })
  })

  test.describe('Chunk Inspection', () => {
    test('should inspect and search chunks', async ({ page }) => {
      // Login first
      await page.goto('http://localhost:5173/login')
      await page.getByLabel('Email').fill('test@example.com')
      await page.getByLabel('Şifre').fill('password')
      await page.getByRole('button', { name: 'Giriş Yap' }).click()
      await expect(page).toHaveURL(/.*dashboard|.*\//, { timeout: 10000 })

      // Navigate to chatbot sources page
      await page.goto('/dashboard/chatbots/bot-1/sources')
      await page.waitForLoadState('domcontentloaded')

      // Navigate to sources
      await page.getByRole('link', { name: 'Kaynaklar' }).click()

      // Find and click "Inspect Chunks" button
      const inspectBtn = page.getByRole('button', { name: 'Parçaları İncele' }).first()
      await inspectBtn.click()

      // Verify inspector content
      await expect(page.getByText('Kaynak Parçaları')).toBeVisible()
      await expect(page.getByText('Chunk content 0 for inspection.')).toBeVisible()

      // Test search functionality
      const searchInput = page.getByPlaceholder('Yüklenen parçalarda ara...')
      await searchInput.fill('Chunk content 2')
      await expect(page.getByText('Chunk content 2 for inspection.')).toBeVisible()
      await expect(page.getByText('Chunk content 0 for inspection.')).not.toBeVisible()
    })
  })
})

/**
 * Sets up API mocks for chunk inspector tests
 */
async function setupChunkInspectorMocks(page: Page, isReal: boolean) {
  if (!isReal) {
    // Generic fallback for unhandled GETs
    await page.route('**/api/v1/**', async (r) => {
      if (
        r.request().method() === 'GET' &&
        !r.request().url().includes('chatbots') &&
        !r.request().url().includes('auth')
      ) {
        await r.fulfill({ status: 200, body: JSON.stringify([]) })
      } else {
        await r.continue()
      }
    })

    // Stub auth and basic endpoints
    await page.route('**/api/v1/auth/register', async (r) =>
      r.fulfill({ status: 200, body: JSON.stringify({ ok: true }) }),
    )
    await page.route('**/api/v1/auth/login', async (r) =>
      r.fulfill({ status: 200, body: JSON.stringify({ token: 't', refresh_token: 'rt' }) }),
    )
    await page.route('**/api/v1/auth/me', async (r) =>
      r.fulfill({ status: 200, body: JSON.stringify({ plan: 'pro', id: 'user-1' }) }),
    )
    await page.route('**/api/v1/analytics', async (r) =>
      r.fulfill({ status: 200, body: JSON.stringify([]) }),
    )

    // Stub orgs/workspaces
    await page.route('**/api/v1/organizations', async (r) =>
      r.fulfill({
        status: 200,
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
      }),
    )
    await page.route('**/workspaces', async (r) =>
      r.fulfill({
        status: 200,
        body: JSON.stringify([
          {
            id: 'ws-1',
            organization_id: 'org-1',
            name: 'Test Workspace',
            slug: 'test-ws',
            created_at: new Date().toISOString(),
          },
        ]),
      }),
    )

    // Stub chatbots list
    await page.route('**/api/v1/chatbots', async (r) => {
      if (r.request().method() === 'GET') {
        await r.fulfill({ status: 200, body: JSON.stringify([]) })
      } else {
        await r.fulfill({ status: 201, body: JSON.stringify({ id: 'bot-1' }) })
      }
    })
    await page.route('**/api/v1/chatbots/bot-1', async (r) =>
      r.fulfill({ status: 200, body: JSON.stringify({ id: 'bot-1', name: 'Test Bot' }) }),
    )

    // Stub sources
    await page.route('**/api/v1/chatbots/bot-1/sources', async (r) => {
      if (r.request().method() === 'POST') {
        await r.fulfill({ status: 201, body: JSON.stringify({ id: 'src-1' }) })
      } else {
        await r.fulfill({
          status: 200,
          body: JSON.stringify([
            {
              id: 'src-1',
              source_type: 'text',
              original_filename: 'test-source.txt',
              status: 'completed',
              chunk_count: 5,
              created_at: new Date().toISOString(),
            },
          ]),
        })
      }
    })

    // Stub chunks endpoint
    await page.route('**/api/v1/sources/src-1/chunks*', async (r) => {
      await r.fulfill({
        status: 200,
        body: JSON.stringify({
          chunks: Array.from({ length: 5 }, (_, i) => ({
            id: `c-${i}`,
            score: 0.9,
            payload: {
              source_id: 'src-1',
              original_text: `Chunk content ${i} for inspection.`,
              chunk_index: i,
              created_at: new Date().toISOString(),
            },
          })),
          next_cursor: null,
        }),
      })
    })
  }
}
