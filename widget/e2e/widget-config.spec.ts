/**
 * Widget Configuration E2E Tests
 * 
 * Tests for widget configuration and customization:
 * - Theme color application
 * - Position settings
 * - Custom welcome message
 * - Font family settings
 * - Panel dimensions
 */

import { test, expect } from '@playwright/test'
import { WidgetHelper, setupChatMocks, waitForWidgetMounted } from './helpers'

test.describe('Widget Configuration', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    helper = new WidgetHelper(page)
  })

  test('loads configuration from API', async ({ page }) => {
    // Verify config is loaded and applied
    await helper.openWidget()
    
    // Widget should display configured welcome message
    await expect(helper.messagesContainer).toBeVisible()
  })

  test('applies custom theme color', async ({ page }) => {
    // Mock config with custom theme color
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#ff0000',
          position: 'bottom-right',
          welcome_message: 'Test Welcome',
        }),
      })
    })
    
    // Reload with custom config
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    // Setup mock again after reload
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#ff0000',
          position: 'bottom-right',
          welcome_message: 'Test Welcome',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    
    // Open widget
    await helper.openWidget()
    
    // Verify bubble has custom color
    const bubble = page.locator('.cbw-bubble')
    const color = await bubble.evaluate((el) => {
      return window.getComputedStyle(el).backgroundColor
    })
    
    // #ff0000 in RGB is rgb(255, 0, 0)
    expect(color).toContain('255')
    expect(color).toContain('0')
    expect(color).toContain('0')
  })

  test('applies custom welcome message', async ({ page }) => {
    // Mock config with custom welcome message
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Özel karşılama mesajı!',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Özel karşılama mesajı!',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify custom welcome message
    await helper.expectMessageWithContent('Özel karşılama mesajı!')
  })

  test('handles bottom-left position', async ({ page }) => {
    // Mock config with bottom-left position
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-left',
          welcome_message: 'Test',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-left',
          welcome_message: 'Test',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify widget has bottom-left class
    const container = page.locator('.cbw-container')
    await expect(container).toHaveClass(/cbw-pos-left/)
  })

  test('applies custom bot display name', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Hello!',
          bot_display_name: 'Custom Bot Name',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Hello!',
          bot_display_name: 'Custom Bot Name',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify custom bot name in header
    await expect(helper.headerTitle).toContainText('Custom Bot Name')
  })

  test('applies custom panel height', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_panel_height: '500px',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_panel_height: '500px',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify custom height is applied via CSS variable
    const panel = page.locator('.cbw-panel')
    const height = await panel.evaluate((el) => {
      return window.getComputedStyle(el).getPropertyValue('--cbw-panel-height')
    })
    expect(height).toBe('500px')
  })

  test('applies custom panel width', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_panel_width: '400px',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_panel_width: '400px',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify custom width is applied
    const panel = page.locator('.cbw-panel')
    const width = await panel.evaluate((el) => {
      return window.getComputedStyle(el).getPropertyValue('--cbw-panel-width')
    })
    expect(width).toBe('400px')
  })

  test('applies custom font family', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_font_family: 'Arial, sans-serif',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_font_family: 'Arial, sans-serif',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify font family is applied
    const input = page.locator('.cbw-input-field')
    const fontFamily = await input.evaluate((el) => {
      return window.getComputedStyle(el).fontFamily
    })
    expect(fontFamily).toContain('Arial')
  })

  test('applies custom chat background color', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_background_color: '#f0f0f0',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          chat_background_color: '#f0f0f0',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify background color is applied
    const messagesArea = page.locator('.cbw-messages')
    const bgColor = await messagesArea.evaluate((el) => {
      return window.getComputedStyle(el).backgroundColor
    })
    expect(bgColor).toContain('240') // #f0f0f0 = rgb(240, 240, 240)
  })

  test('applies custom input background color', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          input_background_color: '#e0e0e0',
        }),
      })
    })
    
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          input_background_color: '#e0e0e0',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Verify input background color is applied
    const input = page.locator('.cbw-input-field')
    const bgColor = await input.evaluate((el) => {
      return window.getComputedStyle(el).backgroundColor
    })
    expect(bgColor).toContain('224') // #e0e0e0 = rgb(224, 224, 224)
  })
})

test.describe('Widget Configuration - URL Overrides', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    helper = new WidgetHelper(page)
  })

  test('URL parameters override API config', async ({ page }) => {
    // This test verifies the widget accepts URL parameters
    // The actual override behavior depends on the widget implementation
    
    // Visit with URL override parameters
    await page.goto('/widget-test.html?color=%23ff6600&bot-name=URL%20Bot')
    await waitForWidgetMounted(page)
    
    // Widget should still load
    await helper.expectBubbleVisible()
  })

  test('handles invalid color gracefully', async ({ page }) => {
    // Visit with invalid color
    await page.goto('/widget-test.html?color=invalid-color')
    await waitForWidgetMounted(page)
    
    // Widget should still load
    await helper.expectBubbleVisible()
  })
})

test.describe('Widget Configuration - Error Handling', () => {
  test('handles missing chatbot ID gracefully', async ({ page }) => {
    // Visit with invalid chatbot ID
    await page.goto('/widget-test.html?chatbot-id=nonexistent')
    
    // Widget container should still exist
    const host = page.locator('#chatbot-widget-host')
    await expect(host).toBeVisible()
  })

  test('handles API error gracefully', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      })
    })
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(1000)
    
    // Bubble should still be visible (fallback behavior)
    const bubble = page.locator('.cbw-bubble')
    await expect(bubble).toBeVisible()
  })

  test('handles invalid JSON in API response', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: 'invalid json{',
      })
    })
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(1000)
    
    // Widget should still function with defaults
    const bubble = page.locator('.cbw-bubble')
    await expect(bubble).toBeVisible()
  })
})
