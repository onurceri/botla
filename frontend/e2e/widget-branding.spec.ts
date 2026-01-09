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
    // hide_branding: false shows default branding (Powered by Botla)
    // hide_branding: true shows custom branding if custom_branding is configured
    await page.route('http://api.test/api/v1/public/chatbots/bot1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'bot1',
          theme_color: '#3b82f6',
          welcome_message: 'Merhaba!',
          hide_branding: false,
          custom_branding: {
            logo_url: 'http://example.com/logo.png',
            text: 'Custom Power',
            link: 'http://example.com',
          },
        }),
      })
    })
  })

  test('should display default branding when hide_branding is false', async ({ page }) => {
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

    // Wait for config to load and widget to render
    await page.waitForTimeout(1000)

    // Verify default branding element exists
    const brandHandle = await host.evaluateHandle((el) => el.shadowRoot!.querySelector('.cbw-brand-default'))
    const brandElement = brandHandle.asElement()
    
    if (!brandElement) {
      throw new Error('Default branding element not found in shadow DOM')
    }
    
    // Check visibility using handle
    const isVisible = await brandElement.isVisible()
    expect(isVisible).toBe(true)

    // Verify branding text shows "Powered by Botla"
    const text = await brandHandle.evaluate((el) => el?.textContent || '')
    expect(text).toContain('Powered by')
    expect(text).toContain('Botla')
  })
})
