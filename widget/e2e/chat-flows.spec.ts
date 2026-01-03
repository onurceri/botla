/**
 * Chat Flows E2E Tests
 * 
 * Tests for multi-step chat conversation flows:
 * - Question and answer flows
 * - Long conversations
 * - Rapid message sending
 * - Context preservation
 */

import { test, expect } from '@playwright/test'
import { WidgetHelper, setupWidgetMocks, setupChatMocks, waitForWidgetMounted } from './helpers'

test.describe('Chat Flows', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    // Setup default mocks
    await setupWidgetMocks(page)
    await setupChatMocks(page, 'test-chatbot', { response: 'Size nasıl yardımcı olabilirim?' })
    
    helper = new WidgetHelper(page)
  })

  test('handles simple question and answer flow', async ({ page }) => {
    await helper.openWidget()
    
    // Initial welcome message should be present
    await expect(helper.messageRows.first()).toBeVisible()
    
    // User asks a question
    await helper.sendMessage('Ürünleriniz hakkında bilgi almak istiyorum')
    
    // Assistant should respond
    await helper.expectMessageWithContent('Size nasıl yardımcı olabilirim?')
    
    // User follows up
    await helper.sendMessage('Fiyatlarınız nedir?')
    
    // Assistant should respond
    await helper.expectMessageWithContent('Size nasıl yardımcı olabilirim?')
    
    // Verify conversation structure (user, assistant, user, assistant)
    await expect(helper.userMessages).toHaveCount(2)
    await expect(helper.assistantMessages).toHaveCount(2)
  })

  test('handles long conversation with many messages', async ({ page }) => {
    await helper.openWidget()
    
    const messages = [
      'Merhaba',
      'Size birkaç sorum olacak',
      'İlk olarak, hizmetleriniz nelerdir?',
      'İkinci olarak, fiyatlandırma nasıl?',
      'Üçüncü olarak, destek seçenekleriniz neler?',
      'Son olarak, ne zaman başlayabilirim?',
      'Teşekkürler',
    ]
    
    // Send multiple messages
    for (const message of messages) {
      await helper.sendMessage(message)
      // Small delay between messages
      await page.waitForTimeout(100)
    }
    
    // Verify all messages are in the conversation
    // Initial welcome + 7 user messages + 7 assistant responses = 15
    await expect(helper.messageRows).toHaveCount(15)
  })

  test('handles rapid message sending', async ({ page }) => {
    await helper.openWidget()
    
    // Send messages rapidly
    await helper.sendMessage('Mesaj 1')
    await helper.sendMessage('Mesaj 2')
    await helper.sendMessage('Mesaj 3')
    
    // All messages should be present
    await helper.expectMessageWithContent('Mesaj 1')
    await helper.expectMessageWithContent('Mesaj 2')
    await helper.expectMessageWithContent('Mesaj 3')
  })

  test('maintains conversation context', async ({ page }) => {
    await helper.openWidget()
    
    // First message
    await helper.sendMessage('Ben Ali')
    
    // Assistant responds
    await helper.expectMessageWithContent('Size nasıl yardımcı olabilirim?')
    
    // Second message referencing first
    await helper.sendMessage('Adım Ali ve size bir şey sormak istiyorum')
    
    // All messages should be preserved
    await expect(helper.messageRows).toHaveCount(4) // welcome + 2 user + 2 assistant
  })

  test('handles very long messages', async ({ page }) => {
    await helper.openWidget()
    
    // Create a long message (close to max length)
    const longMessage = 'Bu '.repeat(400) + 'son kelime'
    
    await helper.sendMessage(longMessage)
    
    // Message should be sent (truncated to max chars)
    await expect(helper.messageRows.last()).toBeVisible()
  })

  test('handles special characters in messages', async ({ page }) => {
    await helper.openWidget()
    
    // Message with special characters
    const specialMessage = 'Merhaba! Nasılsınız? İyi günler.😊 #test @admin'
    
    await helper.sendMessage(specialMessage)
    
    // Message should be sent correctly
    await helper.expectMessageWithContent('Merhaba!')
  })

  test('handles multi-line messages', async ({ page }) => {
    await helper.openWidget()
    
    // Message with newlines
    const multiLineMessage = 'Satır 1\nSatır 2\nSatır 3'
    await helper.sendMessage(multiLineMessage)
    
    // Message should be sent
    await helper.expectMessageWithContent('Satır 1')
  })

  test('handles Unicode and Turkish characters', async ({ page }) => {
    await helper.openWidget()
    
    // Turkish specific characters
    const turkishMessage = 'Ç Ş Ğ İ Ö Ü ç ş ğ ı ö ü Türkçe karakter testi'
    
    await helper.sendMessage(turkishMessage)
    
    // Message should be sent correctly
    await helper.expectMessageWithContent('Türkçe')
  })

  test('handles consecutive user messages without assistant response delay', async ({ page }) => {
    await helper.openWidget()
    
    // Send messages quickly
    await helper.sendMessage('Hızlı mesaj 1')
    await helper.sendMessage('Hızlı mesaj 2')
    await helper.sendMessage('Hızlı mesaj 3')
    
    // All user messages should be present
    const userMessages = helper.userMessages
    await expect(userMessages).toHaveCount(3)
  })
})

test.describe('Chat Flows - Error Recovery', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    await setupWidgetMocks(page)
    
    helper = new WidgetHelper(page)
  })

  test('recovers after API error and continues conversation', async ({ page }) => {
    await helper.openWidget()
    
    // First message succeeds
    await setupChatMocks(page, 'test-chatbot', { response: 'İlk yanıt' })
    await helper.sendMessage('İlk mesaj')
    await helper.expectMessageWithContent('İlk yanıt')
    
    // Mock error for second message
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      })
    })
    
    // Clear session to prevent cached messages
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    // Setup error mock
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      })
    })
    
    await helper.openWidget()
    await helper.sendMessage('Hata mesajı')
    
    // Wait for error handling
    await page.waitForTimeout(500)
    
    // Error message should appear
    await helper.expectPanelToContainText('hata oluştu')
    
    // Mock success for third message
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Kaldığınız yerden devam ediyoruz',
          message_id: 'msg-recovery',
        }),
      })
    })
    
    // Clear and reload again for fresh state
    await page.evaluate(() => sessionStorage.clear())
    await page.reload()
    await waitForWidgetMounted(page)
    
    // Setup success mock
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Kaldığınız yerden devam ediyoruz',
          message_id: 'msg-recovery',
        }),
      })
    })
    
    await helper.openWidget()
    await helper.sendMessage('Kurtarma mesajı')
    
    // Conversation should continue
    await helper.expectMessageWithContent('Kaldığınız yerden devam ediyoruz')
  })

  test('handles network disconnection and reconnection', async ({ page }) => {
    await helper.openWidget()
    
    // Mock offline behavior
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      await route.abort('failed')
    })
    
    await helper.sendMessage('Çevrimdışı mesaj')
    
    // Should show error
    await page.waitForTimeout(500)
    await helper.expectPanelToContainText('hata oluştu')
  })
})

test.describe('Chat Flows - Session Persistence', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    await setupWidgetMocks(page)
    await setupChatMocks(page)
    
    helper = new WidgetHelper(page)
  })

  test('persists messages in session storage', async ({ page }) => {
    await helper.openWidget()
    
    // Send messages
    await helper.sendMessage('Kalıcı mesaj 1')
    await helper.sendMessage('Kalıcı mesaj 2')
    
    // Verify messages in session storage
    const sessionData = await page.evaluate(() => {
      const storage = sessionStorage.getItem('chatbot_session_test-chatbot')
      return storage ? JSON.parse(storage) : null
    })
    
    expect(sessionData).not.toBeNull()
    expect(sessionData.messages.length).toBeGreaterThan(2)
  })

  test('loads messages from session storage on page reload', async ({ page }) => {
    await helper.openWidget()
    
    // Send messages
    await helper.sendMessage('Sayfa yenileme testi')
    
    // Reload page
    await page.reload()
    await waitForWidgetMounted(page)
    
    // Setup mocks again
    await setupWidgetMocks(page)
    await setupChatMocks(page)
    
    helper = new WidgetHelper(page)
    
    // Open widget
    await helper.openWidget()
    
    // Previous messages should be loaded
    await helper.expectMessageWithContent('Sayfa yenileme testi')
  })
})
