import { test, expect } from '@playwright/test'
import { fileURLToPath } from 'url'
import fs from 'fs'
import path from 'path'

/**
 * Widget Branding E2E Tests
 *
 * Tests cover:
 * - Custom branding configuration
 * - Logo display
 * - Branding text visibility
 *
 * Element Selection Strategy:
 * - Widget components use Shadow DOM, so CSS selectors within shadow root are required
 * - CSS selectors for shadow DOM are documented in TESTING_STANDARDS.md
 *
 * @see TESTING_STANDARDS.md for naming conventions
 * @see selectors.ts for element ID constants
 */
test.describe('Widget Branding', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API with custom branding configuration
    await page.route('http://api.test/api/v1/public/chatbots/bot1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'bot1',
          theme_color: '#3b82f6',
          welcome_message: 'Merhaba!',
          hide_branding: true,
          custom_branding: {
            logo_url: 'http://example.com/logo.png',
            text: 'Custom Power',
            link: 'http://example.com',
          },
        }),
      })
    })
  })

  test('should display custom branding when configured', async ({ page }) => {
    const __filename = fileURLToPath(import.meta.url)
    const __dirname = path.dirname(__filename)
    const widgetPath = path.resolve(__dirname, '../../widget/dist/widget.js')
    const code = fs.readFileSync(widgetPath, 'utf-8')

    await page.goto('about:blank')
    await page.addScriptTag({
      content: `window.__CBW_PARAMS={"chatbot-id":"bot1","api-base":"http://api.test","auto-open":"1"}`,
    })
    await page.addScriptTag({ content: code })

    // Wait for widget host and shadow DOM
    const host = await page.waitForSelector('#chatbot-widget-host', { state: 'attached' })
    await page.waitForFunction(() => {
      const h = document.getElementById('chatbot-widget-host')
      return !!h && !!h.shadowRoot
    })

    // Verify custom branding element exists
    const brand = await host.evaluateHandle((el) => el.shadowRoot!.querySelector('.cbw-brand'))
    expect(await brand.asElement()!.isVisible()).toBeTruthy()

    // Verify branding text
    const text = await brand.evaluate((el) => el?.textContent)
    expect(text).toContain('Custom Power')

    // Verify logo URL
    const img = await brand.evaluate((el) => el?.querySelector('img')?.src)
    expect(img).toBe('http://example.com/logo.png')
  })
})
