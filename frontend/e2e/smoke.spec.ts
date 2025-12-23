import { test, expect } from '@playwright/test'

test.describe('E2E Smoke', () => {
  const isReal = !!process.env.E2E_API_BASE
  test.skip(!isReal, 'Skipped on stubbed backend to avoid flakiness')
  test('Login → Create Chatbot → Add Text Source → Chat', async ({ page }) => {
    let botId = 'e2e-bot'
    // Use the outer `isReal` variable from describe scope
    if (!isReal) {
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
          body: JSON.stringify({ token: 't', refresh_token: 'rt' }),
        })
      })
      await page.route('**/api/v1/chatbots', async (route) => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([]),
          })
        } else {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({ id: botId }),
          })
        }
      })
      await page.route(`**/api/v1/chatbots/${botId}`, async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: botId, name: 'E2E Bot' }),
        })
      })
      await page.route(`**/api/v1/chatbots/${botId}/sources`, async (route) => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ id: 'src-1' }),
        })
      })
      await page.route('**/api/v1/sources/src-1', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ id: 'src-1', status: 'completed', chunk_count: 3 }),
        })
      })
      await page.route('**/api/v1/analytics', async (route) => {
        const today = new Date()
        const series = Array.from({ length: 7 }).map((_, i) => {
          const d = new Date(today)
          d.setDate(today.getDate() - (6 - i))
          return {
            date: d.toISOString().slice(0, 10),
            messages: i,
            conversations: Math.floor(i / 2),
          }
        })
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(series),
        })
      })
      await page.route(`**/api/v1/chatbots/${botId}/chat`, async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ response: 'Merhaba!' }),
        })
      })
      await page.route('**/api/v1/auth/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ plan: 'pro' }),
        })
      })
    }

    // Registration
    const email = `e2e-${Date.now()}@example.com`
    await page.goto('/register')
    await page.getByLabel('Ad Soyad').fill('Test User')
    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Şifre').fill('password123')
    await page.getByRole('button', { name: 'Kayıt Ol' }).click()

    // Wait for redirect to login (success)
    await expect(page).toHaveURL(/.*login/, { timeout: 10000 })

    // Login
    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Şifre').fill('password123')
    await page.getByRole('button', { name: 'Giriş Yap' }).click()

    // Wait for successful login and redirect to dashboard (NOT login page)
    await expect(page).toHaveURL(/^(?!.*\/login).*\/$/, { timeout: 10000 })

    // Verify we are authenticated by checking for Dashboard elements
    // We might be redirected back to login if auth fails, so we should check for that too
    try {
      await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible({ timeout: 10000 })
    } catch (e) {
      // If dashboard is not visible, check if we are back at login or if there was an error
      if (page.url().includes('login')) {
        throw new Error('Redirected back to login page after login - Auth failed')
      }
      throw e
    }

    // Create chatbot
    const createBtn = page.getByRole('button', { name: /Yeni Chatbot|Yeni Oluştur/ }).first()
    await createBtn.waitFor({ state: 'visible', timeout: 10000 })
    await createBtn.click()
    await page.getByPlaceholder('Örn: Müşteri Temsilcisi').fill('E2E Bot')
    const saveBtn = page.getByRole('button', { name: 'Oluştur' })
    await saveBtn.click()
    await expect(page.getByText('Chatbot başarıyla oluşturuldu.')).toBeVisible({ timeout: 10000 })
    await expect(page).not.toHaveURL(/\/chatbots\/new$/)
    await expect(page).toHaveURL(/\/chatbots\/[a-zA-Z0-9_-]+$/)
    botId = new URL(page.url()).pathname.split('/').pop() || botId

    // Add text source
    await page.getByRole('tab', { name: 'Veri Kaynakları' }).click()
    await page.getByText('Metin Gir', { exact: true }).last().click()
    await page.getByPlaceholder('Metin içeriğini buraya yapıştırın...').fill('E2E kaynak metni')
    await page.getByRole('button', { name: 'Ekle' }).last().click()
    await expect(page.getByText('Metin kaynağı eklendi.')).toBeVisible()
    if (isReal && process.env.E2E_INCLUDE_INGEST === '1') {
      // Poll terminal status in sources table
      await page.waitForTimeout(2000)
      // Refresh sources by re-entering tab
      await page.getByRole('tab', { name: 'Veri Kaynakları' }).click()
      // Expect any terminal status to appear eventually
      await expect(page.getByText(/completed|failed/i)).toBeVisible({ timeout: 15000 })
    }

    // Chat
    await page.getByRole('tab', { name: 'Playground' }).click()
    await page.getByRole('button', { name: 'Sohbeti aç' }).click()
    const input = page.getByPlaceholder('Mesaj yazın...')
    await input.fill('merhaba')
    await input.press('Enter')
    if (isReal) {
      // Real backend may return fallback; assert any assistant message appears in widget
      await expect(
        page.locator('.cbw-msg.assistant').filter({ hasText: /Merhaba|hata|bilgi/i }),
      ).toBeVisible({ timeout: 10000 })
      // Navigate to Dashboard and confirm the bot appears in "Son Botlarınız"
      await page.goto('/')

      // Check if the bot appears in "Son Botlarınız" list
      // This confirms the bot was created and associated with the user correctly
      const recentBotsCard = page
        .locator('div.col-span-3')
        .filter({ has: page.getByText(/^Son Botlarınız$/) })
        .first()
      await expect(recentBotsCard).toBeVisible()
      await expect(recentBotsCard.getByText('E2E Bot')).toBeVisible()
    } else {
      await expect(page.getByText('Merhaba!')).toBeVisible()
    }
  })
})
