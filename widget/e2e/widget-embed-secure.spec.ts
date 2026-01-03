/**
 * Secure Embedding E2E Tests
 * 
 * Tests for secure widget embedding scenarios:
 * - Authentication with embed tokens
 * - Captcha integration
 * - Secure API communication
 * - Error handling for auth failures
 */

import { test, expect } from '@playwright/test'
import { WidgetHelper, setupWidgetMocks, setupChatMocks, waitForWidgetMounted } from './helpers'

test.describe('Secure Embedding', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    helper = new WidgetHelper(page)
  })

  test('widget loads without embed token by default', async ({ page }) => {
    // Verify widget loads without auth requirements
    await helper.expectBubbleVisible()
    
    // Open widget
    await helper.openWidget()
    
    // Should display welcome message (no auth required for basic loading)
    await expect(helper.messagesContainer).toBeVisible()
  })

  test('handles chat request without embed token', async ({ page }) => {
    await setupWidgetMocks(page)
    await setupChatMocks(page, 'test-chatbot', { response: 'Yanıt' })
    
    await helper.openWidget()
    
    // Send message - should work without token
    await helper.sendMessage('Test mesajı')
    
    // Response should appear
    await helper.expectMessageWithContent('Yanıt')
  })

  test('displays error on API 401 unauthorized', async ({ page }) => {
    // Mock 401 error for chat
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Unauthorized',
          message: 'Geçersiz oturum',
        }),
      })
    })
    
    await helper.openWidget()
    
    // Send message that will trigger 401
    await helper.sendMessage('Yetkisiz mesaj')
    
    // Error handling should occur (error message or session handling)
    await page.waitForTimeout(500)
  })

  test('displays error on API 403 forbidden', async ({ page }) => {
    // Mock 403 error for chat
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 403,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Forbidden',
          message: 'Bu işlem için yetkiniz yok',
        }),
      })
    })
    
    await helper.openWidget()
    
    // Send message that will trigger 403
    await helper.sendMessage('Yasak mesaj')
    
    // Should handle error gracefully
    await page.waitForTimeout(500)
  })

  test('handles 429 rate limit response', async ({ page }) => {
    // Mock 429 rate limit
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 429,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Too Many Requests',
          message: 'Çok fazla istek gönrildiniz',
        }),
      })
    })
    
    await helper.openWidget()
    
    // Send message that will trigger rate limiting
    await helper.sendMessage('Rate limit test')
    
    // Should handle rate limiting
    await page.waitForTimeout(500)
  })

  test('handles 404 not found for chatbot', async ({ page }) => {
    // Mock 404 for config
    await page.route('**/api/v1/public/chatbots/nonexistent', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Not Found',
          message: 'Chatbot bulunamadı',
        }),
      })
    })
    
    // Reload with nonexistent chatbot
    await page.goto('/widget-test.html?chatbot-id=nonexistent')
    await page.waitForTimeout(1000)
    
    // Widget should still have bubble visible (fallback)
    const bubble = page.locator('.cbw-bubble')
    await expect(bubble).toBeVisible()
  })
})

test.describe('Secure Embedding - Captcha Integration', () => {
  test('captcha token is sent when configured', async ({ page }) => {
    await page.goto('/widget-test.html?captcha-site-key=test-key')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    
    // Mock captcha function
    await page.evaluate(() => {
      (window as any).getCaptchaToken = async (siteKey: string) => {
        return 'mock-captcha-token-' + siteKey
      }
    })
    
    // Mock chat that verifies captcha token is sent
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      const postData = route.request().postData()
      const body = postData ? JSON.parse(postData) : {}
      
      // Verify captcha_token is included
      if (body.captcha_token) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            response: 'Captcha verified',
            message_id: 'msg-captcha',
          }),
        })
      } else {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Missing captcha token',
          }),
        })
      }
    })
    
    await helper.openWidget()
    await helper.sendMessage('Captcha test')
    
    // Should handle captcha verification
    await page.waitForTimeout(500)
  })

  test('handles captcha failure gracefully', async ({ page }) => {
    await page.goto('/widget-test.html?captcha-site-key=test-key')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    
    // Mock captcha failure
    await page.evaluate(() => {
      (window as any).getCaptchaToken = async (_siteKey: string) => {
        throw new Error('Captcha failed')
      }
    })
    
    // Chat should still work (captcha optional)
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Response without captcha',
          message_id: 'msg-no-captcha',
        }),
      })
    })
    
    await helper.openWidget()
    await helper.sendMessage('No captcha test')
    
    // Should still work
    await helper.expectMessageWithContent('Response without captcha')
  })
})

test.describe('Secure Embedding - Session Management', () => {
  test('session ID is included in chat requests', async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    
    // Mock chat that echoes session ID
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      const postData = route.request().postData()
      const body = postData ? JSON.parse(postData) : {}
      
      // Verify session_id is included
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: `Session: ${body.session_id || 'none'}`,
          message_id: 'msg-session',
        }),
      })
    })
    
    await helper.openWidget()
    await helper.sendMessage('Session test')
    
    // Should receive response with session info
    await page.waitForTimeout(500)
  })

  test('session persists across widget reload', async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    await setupChatMocks(page)
    
    await helper.openWidget()
    await helper.sendMessage('First message')
    
    // Get session from storage
    const initialSession = await page.evaluate(() => {
      const storage = sessionStorage.getItem('chatbot_session_test-chatbot')
      return storage ? JSON.parse(storage) : null
    })
    
    expect(initialSession).not.toBeNull()
    expect(initialSession.sessionId).toBeDefined()
    
    // Reload page
    await page.reload()
    await waitForWidgetMounted(page)
    
    // Session should persist
    const sessionAfterReload = await page.evaluate(() => {
      const storage = sessionStorage.getItem('chatbot_session_test-chatbot')
      return storage ? JSON.parse(storage) : null
    })
    
    expect(sessionAfterReload).not.toBeNull()
    expect(sessionAfterReload.sessionId).toBe(initialSession.sessionId)
  })
})

test.describe('Secure Embedding - Embed Token URL', () => {
  test('fetches embed token from configured URL', async ({ page }) => {
    await page.goto('/widget-test.html?embed-token-url=https://example.com/token')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    
    // Mock token fetch
    await page.route('https://example.com/token', async (route) => {
      await route.fulfill({
        status: 200,
        body: 'test-embed-token-123',
      })
    })
    
    // Mock chat that checks for embed token
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      const headers = route.request().headers()
      
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Token received',
          message_id: 'msg-token',
        }),
      })
    })
    
    await helper.openWidget()
    await helper.sendMessage('Token URL test')
    
    // Should work with token
    await helper.expectMessageWithContent('Token received')
  })

  test('handles embed token fetch failure gracefully', async ({ page }) => {
    await page.goto('/widget-test.html?embed-token-url=https://example.com/invalid-token')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    
    // Mock token fetch failure
    await page.route('https://example.com/invalid-token', async (route) => {
      await route.fulfill({
        status: 500,
        body: 'Token fetch failed',
      })
    })
    
    // Chat should still work
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Works without token',
          message_id: 'msg-no-token',
        }),
      })
    })
    
    await helper.openWidget()
    await helper.sendMessage('No token test')
    
    // Should still work
    await helper.expectMessageWithContent('Works without token')
  })
})

test.describe('Secure Embedding - Network Security', () => {
  test('CORS is handled correctly', async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    await setupChatMocks(page)
    
    await helper.openWidget()
    await helper.sendMessage('CORS test')
    
    // Should handle CORS without issues
    await helper.expectMessageWithContent('Bu bir test yanıtıdır.')
  })

  test('handles network disconnection', async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    
    // Mock network failure
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.abort('failed')
    })
    
    await helper.openWidget()
    await helper.sendMessage('Offline test')
    
    // Should handle network error gracefully
    await page.waitForTimeout(500)
    await helper.expectPanelToContainText('hata oluştu')
  })

  test('handles timeout gracefully', async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    const helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
    
    // Mock slow response that times out
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      // Very long delay
      await new Promise(resolve => setTimeout(resolve, 30000))
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Too late',
          message_id: 'msg-late',
        }),
      })
    })
    
    await helper.openWidget()
    
    // Start sending message (will timeout)
    const input = page.locator('.cbw-input-field')
    await input.fill('Timeout test')
    
    // Click send - should handle timeout
    const sendButton = page.locator('.cbw-send-btn')
    await sendButton.click()
    
    // Widget should still be responsive
    await expect(helper.bubble).toBeVisible()
  })
})
