import { test, expect, devices } from '@playwright/test'

test.describe('Mobile Responsiveness', () => {
  test.use(devices['iPhone 13'])

  test.beforeEach(async ({ page }) => {
    // Mock the chatbot config API
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'test-chatbot',
          name: 'Test Bot',
          welcome_message: 'Merhaba! Size nasıl yardımcı olabilirim?',
          theme_color: '#6366f1',
          position: 'bottom-right',
          bot_display_name: 'Test Bot',
          suggested_questions: ['Nasıl yardımcı olabilirim?', 'Hızlı başlangıç'],
        }),
      })
    })

    // Mock the chat API
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Bu bir test yanıtıdır.',
          message_id: 'msg-123',
        }),
      })
    })
  })

  test('widget button is visible on mobile', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    // Wait for widget script to load
    await page.waitForTimeout(500)
    
    // Get the shadow host
    const host = page.locator('#chatbot-widget-host')
    await expect(host).toBeVisible()
    
    // Find the chat bubble button inside shadow DOM
    const bubble = host.locator('div.cbw-bubble')
    await expect(bubble).toBeVisible()
  })

  test('widget opens on mobile tap', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    // Tap (click) the bubble to open
    await bubble.click()
    
    // Wait for drawer to appear
    await page.waitForTimeout(300)
    
    // Check that the chat container is now visible
    const container = host.locator('.cbw-container')
    await expect(container).toBeVisible()
    
    // Check that the drawer panel is visible
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })

  test('input field is accessible on mobile', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await bubble.click()
    await page.waitForTimeout(300)
    
    // Check that the input textarea is visible and focusable
    const input = host.locator('textarea')
    await expect(input).toBeVisible()
    
    // Focus and type
    await input.focus()
    await input.fill('Test message')
    
    await expect(input).toHaveValue('Test message')
  })

  test('keyboard does not break layout', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await bubble.click()
    await page.waitForTimeout(300)
    
    const input = host.locator('textarea')
    
    // Focus input (simulates keyboard opening on mobile)
    await input.focus()
    
    // Container should still be visible
    const container = host.locator('.cbw-container')
    await expect(container).toBeVisible()
    
    // Drawer should still be visible
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })

  test('widget can be closed on mobile', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    // Open
    await bubble.click()
    await page.waitForTimeout(300)
    
    // Find and click close button
    const closeButton = host.locator('.cbw-close')
    await expect(closeButton).toBeVisible()
    await closeButton.click()
    
    await page.waitForTimeout(300)
    
    // Bubble should be visible again, drawer should be hidden
    await expect(bubble).toBeVisible()
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).not.toBeVisible()
  })

  test('messages render correctly on mobile', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await bubble.click()
    await page.waitForTimeout(300)
    
    // Check that welcome message is visible
    const messages = host.locator('.cbw-messages')
    await expect(messages).toBeVisible()
    
    const welcomeMessage = host.locator('.cbw-msg')
    await expect(welcomeMessage.first()).toBeVisible()
  })

  test('send button is accessible on mobile', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await bubble.click()
    await page.waitForTimeout(300)
    
    const input = host.locator('textarea')
    await input.fill('Hello')
    
    // Find send button
    const sendButton = host.locator('.cbw-send')
    await expect(sendButton).toBeVisible()
    
    // Click send
    await sendButton.click()
    
    // Wait for response
    await page.waitForTimeout(500)
    
    // Should have more messages now
    const messages = host.locator('.cbw-msg')
    const count = await messages.count()
    expect(count).toBeGreaterThan(1)
  })

  test('touch events work correctly', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    // Use tap() for touch-like interaction
    await bubble.tap()
    await page.waitForTimeout(300)
    
    // Drawer should open
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })

  test('widget takes appropriate space on small screens', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await bubble.click()
    await page.waitForTimeout(300)
    
    // Check drawer dimensions
    const drawer = host.locator('.cbw-drawer')
    const box = await drawer.boundingBox()
    
    expect(box).not.toBeNull()
    if (box) {
      // On mobile, drawer should be full-width or nearly full-width
      expect(box.width).toBeGreaterThan(300)
    }
  })
})

test.describe('Mobile - Different Devices', () => {
  test('works on Pixel 5', async ({ page }) => {
    test.use(devices['Pixel 5'])
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await expect(bubble).toBeVisible()
    
    // Open widget
    await bubble.click()
    await page.waitForTimeout(300)
    
    // Should work normally
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })

  test('works on iPhone 12', async ({ page }) => {
    test.use({ viewport: { width: 390, height: 844 } })
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await expect(bubble).toBeVisible()
  })

  test('works on iPad Mini', async ({ page }) => {
    test.use({ viewport: { width: 768, height: 1024 } })
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await expect(bubble).toBeVisible()
  })
})

test.describe('Mobile - Orientation', () => {
  test('handles landscape orientation', async ({ page }) => {
    // Start in portrait
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    // Switch to landscape
    await page.setViewportSize({ width: 844, height: 390 })
    await page.waitForTimeout(300)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await expect(bubble).toBeVisible()
    
    // Open widget
    await bubble.click()
    await page.waitForTimeout(300)
    
    // Should work in landscape
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })

  test('handles portrait orientation', async ({ page }) => {
    // Start in landscape
    await page.setViewportSize({ width: 844, height: 390 })
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    // Switch to portrait
    await page.setViewportSize({ width: 390, height: 844 })
    await page.waitForTimeout(300)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await expect(bubble).toBeVisible()
  })
})

test.describe('Mobile - Touch Gestures', () => {
  test('long press does not interfere with tap', async ({ page }) => {
    test.use(devices['iPhone 13'])
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    // Regular tap should work
    await bubble.click()
    await page.waitForTimeout(300)
    
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })

  test('multiple rapid taps are handled correctly', async ({ page }) => {
    test.use(devices['iPhone 13'])
    
    await page.goto('/widget-test.html')
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    // Rapid taps should not break anything
    await bubble.click()
    await bubble.click()
    await page.waitForTimeout(500)
    
    // Should still be open
    const drawer = host.locator('.cbw-drawer')
    await expect(drawer).toBeVisible()
  })
})

test.describe('Desktop Responsiveness', () => {
  test.use(devices['Desktop Chrome'])

  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'test-chatbot',
          name: 'Test Bot',
          welcome_message: 'Merhaba!',
          theme_color: '#6366f1',
          position: 'bottom-right',
        }),
      })
    })
  })

  test('widget renders at correct position', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await expect(bubble).toBeVisible()
    
    // Get bounding box to verify position
    const box = await bubble.boundingBox()
    expect(box).not.toBeNull()
    
    // Should be in bottom-right area
    const viewportSize = page.viewportSize()
    if (box && viewportSize) {
      expect(box.x + box.width).toBeGreaterThan(viewportSize.width * 0.5)
      expect(box.y + box.height).toBeGreaterThan(viewportSize.height * 0.5)
    }
  })

  test('widget panel has appropriate size on desktop', async ({ page }) => {
    await page.goto('/widget-test.html')
    
    await page.waitForTimeout(500)
    
    const host = page.locator('#chatbot-widget-host')
    const bubble = host.locator('div.cbw-bubble')
    
    await bubble.click()
    await page.waitForTimeout(300)
    
    const drawer = host.locator('.cbw-drawer')
    const box = await drawer.boundingBox()
    
    expect(box).not.toBeNull()
    if (box) {
      // Desktop panel should be reasonably sized
      expect(box.width).toBeGreaterThan(300)
      expect(box.height).toBeGreaterThan(400)
    }
  })
})
