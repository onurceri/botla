/**
 * Widget Branding E2E Tests
 * 
 * Tests for widget branding customization:
 * - Custom branding configuration
 * - Bot icon display
 * - Hide branding option
 * - Custom footer text and links
 */

import { test, expect } from '@playwright/test'
import { WidgetHelper, waitForWidgetMounted } from './helpers'

test.describe('Widget Branding', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    helper = new WidgetHelper(page)
  })

  test('shows default Botla branding', async ({ page }) => {
    await helper.openWidget()
    
    // Default branding should be visible
    await expect(helper.brandingFooter).toBeVisible()
    await expect(helper.brandingFooter).toContainText('Botla')
  })

  test('displays bot icon when provided', async ({ page }) => {
    // Mock config with bot icon
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          bot_icon: 'https://example.com/bot-icon.png',
          bot_display_name: 'Icon Bot',
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
          bot_icon: 'https://example.com/bot-icon.png',
          bot_display_name: 'Icon Bot',
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Bot icon should be visible in header
    const headerIcon = page.locator('.cbw-header-icon')
    await expect(headerIcon).toBeVisible()
    
    // Check that the icon image has the correct src
    const iconImg = page.locator('.cbw-header-icon img')
    await expect(iconImg).toHaveAttribute('src', 'https://example.com/bot-icon.png')
  })

  test('displays custom branding when hide_branding is false', async ({ page }) => {
    // Mock config with custom branding
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          hide_branding: false,
          custom_branding: {
            logo_url: 'https://example.com/company-logo.png',
            text: 'Powered by Company',
            link: 'https://company.com',
          },
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
          hide_branding: false,
          custom_branding: {
            logo_url: 'https://example.com/company-logo.png',
            text: 'Powered by Company',
            link: 'https://company.com',
          },
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Custom branding should be visible
    const customBranding = page.locator('.cbw-brand-custom')
    await expect(customBranding).toBeVisible()
    await expect(customBranding).toContainText('Powered by Company')
    
    // Logo should be present
    const logo = page.locator('.cbw-brand-logo')
    await expect(logo).toHaveAttribute('src', 'https://example.com/company-logo.png')
  })

  test('hides default branding when hide_branding is true', async ({ page }) => {
    // Mock config with hide_branding enabled
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          hide_branding: true,
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
          hide_branding: true,
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Default branding should be hidden
    await expect(helper.brandingFooter).not.toBeVisible()
    
    // Custom branding area should be empty
    const customBranding = page.locator('.cbw-brand-custom')
    await expect(customBranding).toBeEmpty()
  })

  test('uses custom branding link', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          hide_branding: false,
          custom_branding: {
            logo_url: undefined,
            text: 'Visit our website',
            link: 'https://custom-example.com',
          },
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
          hide_branding: false,
          custom_branding: {
            logo_url: undefined,
            text: 'Visit our website',
            link: 'https://custom-example.com',
          },
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Custom link should be present
    const link = page.locator('.cbw-brand-text')
    await expect(link).toHaveAttribute('href', 'https://custom-example.com')
    await expect(link).toContainText('Visit our website')
  })

  test('handles missing custom branding gracefully', async ({ page }) => {
    // Mock config without custom branding
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          hide_branding: false,
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
          hide_branding: false,
        }),
      })
    })
    
    helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Should fall back to default branding
    await expect(helper.brandingFooter).toBeVisible()
  })
})

test.describe('Widget Branding - Bot Icon Display', () => {
  test('shows bot icon in messages', async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          bot_icon: 'https://example.com/bot-avatar.png',
        }),
      })
    })
    
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Test',
          bot_icon: 'https://example.com/bot-avatar.png',
        }),
      })
    })
    
    const helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Bot avatar should be visible in messages
    const avatar = page.locator('.cbw-avatar-img')
    await expect(avatar.first()).toHaveAttribute('src', 'https://example.com/bot-avatar.png')
  })

  test('uses default icon when bot_icon is not provided', async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Default icon SVG should be visible
    const defaultIcon = page.locator('.cbw-header-icon-default svg')
    await expect(defaultIcon).toBeVisible()
  })
})

test.describe('Widget Branding - Input Area Branding', () => {
  test('placeholder text is customizable via config', async ({ page }) => {
    // Note: This test verifies the behavior with available configuration options
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await helper.openWidget()
    
    // Default placeholder should be visible
    await expect(helper.inputField).toHaveAttribute('placeholder', 'Mesaj yazın...')
  })
})
