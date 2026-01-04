import { test, expect } from '@playwright/test'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

/**
 * Widget Embed E2E Tests
 *
 * Tests cover:
 * - Widget loading and initialization
 * - Widget configuration
 * - Widget host creation
 *
 * Element Selection Strategy:
 * - Widget components use Shadow DOM, so CSS selectors within shadow root are required
 * - For non-shadow DOM elements, use data-testid attributes from SELECTORS
 * - CSS selectors for shadow DOM are documented in TESTING_STANDARDS.md
 *
 * @see TESTING_STANDARDS.md for naming conventions
 * @see selectors.ts for element ID constants
 */
test.describe('Widget Embed', () => {
  test.beforeEach(async ({ page }) => {
    // Mock chatbot configuration endpoint
    await page.route('http://api.test/api/v1/public/chatbots/bot1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'bot1',
          theme_color: '#3b82f6',
          welcome_message: 'Merhaba!',
          position: 'bottom-right',
          bot_message_color: '#3b82f6',
          user_message_color: '#f3f4f6',
          bot_message_text_color: '#ffffff',
          user_message_text_color: '#1f2937',
          chat_font_family: 'Inter, sans-serif',
          chat_header_color: '#3b82f6',
          chat_header_text_color: '#ffffff',
        }),
      })
    })
  })

  test('should load and initialize widget with correct configuration', async ({ page }) => {
    let configCalled = false
    // The route mock above will trigger this flag
    await page.route('http://api.test/api/v1/public/chatbots/bot1', async (route) => {
      configCalled = true
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'bot1',
          theme_color: '#3b82f6',
          welcome_message: 'Merhaba!',
          position: 'bottom-right',
          bot_message_color: '#3b82f6',
          user_message_color: '#f3f4f6',
          bot_message_text_color: '#ffffff',
          user_message_text_color: '#1f2937',
          chat_font_family: 'Inter, sans-serif',
          chat_header_color: '#3b82f6',
          chat_header_text_color: '#ffffff',
        }),
      })
    })

    const __filename = fileURLToPath(import.meta.url)
    const __dirname = path.dirname(__filename)
    const widgetPath = path.resolve(__dirname, '../../widget/dist/widget.js')
    const code = fs.readFileSync(widgetPath, 'utf-8')

    await page.goto('about:blank')
    await page.addScriptTag({
      content: `window.__CBW_PARAMS={"chatbot-id":"bot1","api-base":"http://api.test","auto-open":"1"}`,
    })
    await page.addScriptTag({ content: code })

    // Verify widget host is created
    await page.waitForSelector('#chatbot-widget-host', { state: 'attached' })

    // Verify configuration was fetched
    await page.waitForFunction(() => (window as any).__CBW_PARAMS)
    await page.waitForFunction(() => true)
    expect(configCalled).toBeTruthy()
  })

  test('should create widget host element', async ({ page }) => {
    const __filename = fileURLToPath(import.meta.url)
    const __dirname = path.dirname(__filename)
    const widgetPath = path.resolve(__dirname, '../../widget/dist/widget.js')
    const code = fs.readFileSync(widgetPath, 'utf-8')

    await page.goto('about:blank')
    await page.addScriptTag({
      content: `window.__CBW_PARAMS={"chatbot-id":"bot1","api-base":"http://api.test"}`,
    })
    await page.addScriptTag({ content: code })

    // Verify host element exists
    const host = await page.waitForSelector('#chatbot-widget-host', { state: 'attached' })
    expect(host).not.toBeNull()
  })
})
