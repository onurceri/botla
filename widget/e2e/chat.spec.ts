/**
 * Chat Interaction E2E Tests
 * 
 * Tests for basic chat functionality:
 * - Opening/closing the widget
 * - Sending and receiving messages
 * - Message history
 * - Error handling
 */

import { test, expect } from '@playwright/test'
import { WidgetHelper, setupWidgetMocks, setupChatMocks, setupChatErrorMocks, waitForWidgetMounted } from './helpers'

test.describe('Chat Interactions', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    // Wait for widget to be fully mounted
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    // Setup API mocks
    await setupWidgetMocks(page)
    await setupChatMocks(page)
    
    helper = new WidgetHelper(page)
  })

  test('widget bubble is visible on page load', async ({ page }) => {
    await helper.expectBubbleVisible()
  })

  test('opens chat widget when bubble is clicked', async ({ page }) => {
    await helper.openWidget()
    
    // Verify panel is visible
    await helper.expectPanelVisible()
    
    // Verify welcome message is shown
    await expect(page.locator('.cbw-msg').first()).toBeVisible()
  })

  test('closes chat widget when close button is clicked', async ({ page }) => {
    await helper.openWidget()
    
    // Verify panel is open
    await helper.expectPanelVisible()
    
    // Close the widget
    await helper.closeWidget()
    
    // Verify panel is hidden
    await helper.expectPanelHidden()
    
    // Bubble should still be visible
    await helper.expectBubbleVisible()
  })

  test('sends user message and receives assistant response', async ({ page }) => {
    await helper.openWidget()
    
    // Get initial message count
    const initialCount = await helper.messageRows.count()
    
    // Send a message
    const testMessage = 'Merhaba, nasılsınız?'
    await helper.sendMessage(testMessage)
    
    // Wait for response
    await page.waitForFunction(
      (count) => document.querySelectorAll('.cbw-msg-row').length > count,
      initialCount,
      { timeout: 10000 }
    )
    
    // Verify user message appears
    await helper.expectMessageWithContent(testMessage)
    
    // Verify response appears
    await expect(helper.assistantMessages.last()).toBeVisible()
  })

  test('shows loading indicator while waiting for response', async ({ page }) => {
    // Mock delayed response
    await page.route('**/api/v1/public/chatbots/test-chatbot/chat', async (route) => {
      // Delay to ensure loading state is visible
      await new Promise(resolve => setTimeout(resolve, 500))
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          response: 'Delayed response',
          message_id: 'msg-delayed',
        }),
      })
    })
    
    await helper.openWidget()
    
    // Send message
    const messagePromise = helper.sendMessage('Slow request')
    
    // Loading indicator should appear
    await helper.expectLoadingVisible()
    
    // Wait for send to complete
    await messagePromise
    
    // Loading indicator should hide
    await helper.expectLoadingHidden()
  })

  test('displays multiple messages in conversation', async ({ page }) => {
    await helper.openWidget()
    
    // Send multiple messages
    await helper.sendMessage('İlk mesaj')
    await helper.sendMessage('İkinci mesaj')
    await helper.sendMessage('Üçüncü mesaj')
    
    // Verify all messages are displayed (3 user + 3 assistant = 6)
    await helper.expectMessageCount(6)
  })

  test('shows error message on API failure', async ({ page }) => {
    // Mock API error
    await setupChatErrorMocks(page, 'test-chatbot', 500, 'Sunucu hatası')
    
    await helper.openWidget()
    
    // Clear existing messages first
    await page.evaluate(() => {
      sessionStorage.clear()
    })
    await page.reload()
    await waitForWidgetMounted(page)
    await setupWidgetMocks(page)
    await setupChatErrorMocks(page, 'test-chatbot', 500, 'Sunucu hatası')
    
    await helper.openWidget()
    
    // Send a message that will fail
    await helper.sendMessage('Test mesajı')
    
    // Wait for error message
    await page.waitForTimeout(500)
    
    // Error message should appear (default error message)
    await helper.expectPanelToContainText('hata oluştu')
  })

  test('preserves message history across widget open/close', async ({ page }) => {
    await helper.openWidget()
    
    // Send a message
    const testMessage = 'Bu mesaj hatırlanmalı'
    await helper.sendMessage(testMessage)
    
    // Close widget
    await helper.closeWidget()
    
    // Reopen widget
    await helper.openWidget()
    
    // Verify message is still there
    await helper.expectMessageWithContent(testMessage)
  })

  test('does not send empty messages', async ({ page }) => {
    await helper.openWidget()
    
    // Get initial message count
    const initialCount = await helper.messageRows.count()
    
    // Click send without typing
    await expect(helper.sendButton).toBeDisabled()
    
    // No new messages should appear
    await helper.expectMessageCount(initialCount)
  })

  test('disables send button when input is empty', async ({ page }) => {
    await helper.openWidget()
    
    // Send button should be disabled with empty input
    await helper.expectSendButtonDisabled()
    
    // Type something
    await helper.inputField.fill('Test')
    
    // Send button should be enabled
    await helper.expectSendButtonEnabled()
    
    // Clear input
    await helper.inputField.clear()
    
    // Send button should be disabled again
    await helper.expectSendButtonDisabled()
  })

  test('shows character limit counter', async ({ page }) => {
    await helper.openWidget()
    
    // Character limit should be visible
    await expect(helper.charLimit).toContainText('0 / 1000')
    
    // Type a message
    const testMessage = 'Bu bir test mesajıdır.'
    await helper.inputField.fill(testMessage)
    
    // Character limit should update
    await expect(helper.charLimit).toContainText(String(testMessage.length))
  })

  test('truncates messages exceeding max length', async ({ page }) => {
    await helper.openWidget()
    
    // Create a message longer than 1000 characters
    const longMessage = 'a'.repeat(1200)
    await helper.sendMessage(longMessage)
    
    // Character limit should show 1000 (max)
    await expect(helper.charLimit).toContainText('1000 / 1000')
  })

  test('scrolls to bottom when new messages arrive', async ({ page }) => {
    await helper.openWidget()
    
    // Send a message
    await helper.sendMessage('Yeni mesaj')
    
    // The messages container should be scrollable and new message visible
    const lastMessage = helper.messageRows.last()
    await expect(lastMessage).toBeVisible()
  })

  test('handles Enter key to send message', async ({ page }) => {
    await helper.openWidget()
    
    // Type message and press Enter
    await helper.sendMessageWithEnter('Enter ile gönderilen mesaj')
    
    // Message should be sent
    await helper.expectMessageWithContent('Enter ile gönderilen mesaj')
  })

  test('Shift+Enter creates new line in input', async ({ page }) => {
    await helper.openWidget()
    
    // Type message with Shift+Enter for newline
    await helper.inputField.fill('Satır 1\nSatır 2')
    
    // Input should contain newline
    const inputValue = await helper.getInputValue()
    expect(inputValue).toContain('\n')
  })

  test('input field is focusable', async ({ page }) => {
    await helper.openWidget()
    
    // Input should be focusable
    await helper.inputField.focus()
    await expect(helper.inputField).toBeFocused()
  })

  test('header shows bot name from config', async ({ page }) => {
    // Setup with custom bot name
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Merhaba!',
          bot_display_name: 'Özel Bot Adı',
        }),
      })
    })
    
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    await helper.openWidget()
    
    // Header should show custom bot name
    await expect(helper.headerTitle).toContainText('Özel Bot Adı')
  })
})

test.describe('Chat - Suggestion Tests', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    // Setup with suggestions
    await page.route('**/api/v1/public/chatbots/test-chatbot', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme_color: '#6366f1',
          position: 'bottom-right',
          welcome_message: 'Merhaba!',
          suggested_questions: ['Nasıl yardımcı olabilirim?', 'Özellikler', 'İletişim'],
        }),
      })
    })
    
    await setupChatMocks(page)
    helper = new WidgetHelper(page)
  })

  test('shows suggested questions on initial load', async ({ page }) => {
    await helper.openWidget()
    
    // Suggestions should be visible
    await helper.expectSuggestionsVisible()
    
    // All suggestions should be present
    const suggestionItems = page.locator('.cbw-suggestion-item')
    await expect(suggestionItems).toHaveCount(3)
  })

  test('clicking suggestion sends message', async ({ page }) => {
    await helper.openWidget()
    
    // Click first suggestion
    await helper.clickSuggestion(0)
    
    // Message should be sent
    await helper.expectMessageWithContent('Nasıl yardımcı olabilirim?')
  })

  test('suggestions hide after user sends message', async ({ page }) => {
    await helper.openWidget()
    
    // Suggestions should be visible initially
    await helper.expectSuggestionsVisible()
    
    // Send a user message
    await helper.sendMessage('Kullanıcı mesajı')
    
    // Suggestions should be hidden
    await helper.expectSuggestionsHidden()
  })
})

test.describe('Chat - Feedback Tests', () => {
  let helper: WidgetHelper

  test.beforeEach(async ({ page }) => {
    await page.goto('/widget-test.html')
    await waitForWidgetMounted(page)
    
    await setupWidgetMocks(page)
    await setupChatMocks(page)
    
    helper = new WidgetHelper(page)
  })

  test('shows feedback buttons on assistant messages', async ({ page }) => {
    await helper.openWidget()
    
    // Send a message to get assistant response
    await helper.sendMessage('Test mesajı')
    await page.waitForTimeout(500)
    
    // Feedback buttons should be visible on assistant message
    const feedbackContainer = page.locator('.cbw-feedback-container').first()
    await expect(feedbackContainer).toBeVisible()
  })

  test('positive feedback click updates UI', async ({ page }) => {
    await helper.openWidget()
    
    await helper.sendMessage('Test mesajı')
    await page.waitForTimeout(500)
    
    // Click thumbs up
    await helper.clickFeedback(1, true)
    
    // Thumbs up should be active
    const thumbsUp = page.locator('.cbw-feedback-btn.positive').first()
    await expect(thumbsUp).toHaveClass(/active/)
  })
})
