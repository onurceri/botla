/**
 * E2E Test Helpers for Widget Testing
 * 
 * Provides helper methods for common widget interactions and assertions.
 * Handles Shadow DOM traversal automatically.
 */

import { type Locator, type Page, expect } from '@playwright/test'

/**
 * Helper class for widget interactions
 */
export class WidgetHelper {
  private page: Page
  private host: Locator

  constructor(page: Page, hostSelector: string = '#chatbot-widget-host') {
    this.page = page
    this.host = page.locator(hostSelector)
  }

  // ========================================
  // Shadow DOM Element Getters
  // ========================================

  /**
   * Gets the chat bubble button (launcher)
   */
  get bubble(): Locator {
    return this.host.locator('.cbw-bubble')
  }

  /**
   * Gets the close button
   */
  get closeButton(): Locator {
    return this.host.locator('.cbw-close-btn')
  }

  /**
   * Gets the chat panel/drawer
   */
  get panel(): Locator {
    return this.host.locator('.cbw-panel')
  }

  /**
   * Gets the messages container
   */
  get messagesContainer(): Locator {
    return this.host.locator('.cbw-messages')
  }

  /**
   * Gets the input textarea
   */
  get inputField(): Locator {
    return this.host.locator('.cbw-input-field')
  }

  /**
   * Gets the send button
   */
  get sendButton(): Locator {
    return this.host.locator('.cbw-send-btn')
  }

  /**
   * Gets the chat header
   */
  get header(): Locator {
    return this.host.locator('.cbw-header')
  }

  /**
   * Gets the header title
   */
  get headerTitle(): Locator {
    return this.host.locator('.cbw-header-title')
  }

  /**
   * Gets the suggestions container
   */
  get suggestionsContainer(): Locator {
    return this.host.locator('.cbw-suggestions-container')
  }

  /**
   * Gets all message rows
   */
  get messageRows(): Locator {
    return this.host.locator('.cbw-msg-row')
  }

  /**
   * Gets all user messages
   */
  get userMessages(): Locator {
    return this.host.locator('.cbw-msg-row.user')
  }

  /**
   * Gets all assistant messages
   */
  get assistantMessages(): Locator {
    return this.host.locator('.cbw-msg-row.assistant')
  }

  /**
   * Gets the loading indicator
   */
  get loadingIndicator(): Locator {
    return this.host.locator('.cbw-loading-row')
  }

  /**
   * Gets the character limit display
   */
  get charLimit(): Locator {
    return this.host.locator('.cbw-char-limit')
  }

  /**
   * Gets the branding footer
   */
  get brandingFooter(): Locator {
    return this.host.locator('.cbw-brand-default')
  }

  /**
   * Gets the unread badge
   */
  get unreadBadge(): Locator {
    return this.host.locator('.cbw-badge')
  }

  // ========================================
  // Action Methods
  // ========================================

  /**
   * Opens the widget by clicking the bubble
   */
  async openWidget(): Promise<void> {
    await expect(this.bubble).toBeVisible()
    await this.bubble.click()
    await expect(this.panel).toBeVisible({ timeout: 5000 })
  }

  /**
   * Closes the widget by clicking the close button
   */
  async closeWidget(): Promise<void> {
    await expect(this.closeButton).toBeVisible()
    await this.closeButton.click()
    await expect(this.panel).not.toBeVisible()
  }

  /**
   * Sends a message
   */
  async sendMessage(message: string): Promise<void> {
    await expect(this.inputField).toBeVisible()
    await this.inputField.fill(message)
    await expect(this.sendButton).toBeEnabled()
    await this.sendButton.click()
  }

  /**
   * Sends a message by pressing Enter
   */
  async sendMessageWithEnter(message: string): Promise<void> {
    await expect(this.inputField).toBeVisible()
    await this.inputField.fill(message)
    await this.inputField.press('Enter')
  }

  /**
   * Gets the current input value
   */
  async getInputValue(): Promise<string> {
    return await this.inputField.inputValue()
  }

  /**
   * Clears the input field
   */
  async clearInput(): Promise<void> {
    await this.inputField.clear()
  }

  /**
   * Clicks on a suggestion
   */
  async clickSuggestion(index: number): Promise<void> {
    const suggestion = this.host.locator('.cbw-suggestion-item').nth(index)
    await expect(suggestion).toBeVisible()
    await suggestion.click()
  }

  /**
   * Clicks a feedback button (thumbs up/down)
   */
  async clickFeedback(messageIndex: number, isPositive: boolean): Promise<void> {
    const message = this.messageRows.nth(messageIndex)
    const feedbackClass = isPositive ? 'positive' : 'negative'
    const button = message.locator(`.cbw-feedback-btn.${feedbackClass}`)
    await expect(button).toBeVisible()
    await button.click()
  }

  // ========================================
  // Assertion Methods
  // ========================================

  /**
   * Asserts the widget bubble is visible
   */
  async expectBubbleVisible(): Promise<void> {
    await expect(this.bubble).toBeVisible()
  }

  /**
   * Asserts the widget panel is visible
   */
  async expectPanelVisible(): Promise<void> {
    await expect(this.panel).toBeVisible()
  }

  /**
   * Asserts the widget panel is hidden
   */
  async expectPanelHidden(): Promise<void> {
    await expect(this.panel).not.toBeVisible()
  }

  /**
   * Asserts the panel contains specific text
   */
  async expectPanelToContainText(text: string): Promise<void> {
    await expect(this.panel).toContainText(text)
  }

  /**
   * Asserts a message with specific content exists
   */
  async expectMessageWithContent(content: string): Promise<void> {
    await expect(this.messagesContainer).toContainText(content)
  }

  /**
   * Asserts the number of messages
   */
  async expectMessageCount(count: number): Promise<void> {
    await expect(this.messageRows).toHaveCount(count)
  }

  /**
   * Asserts the input has specific placeholder
   */
  async expectInputPlaceholder(placeholder: string): Promise<void> {
    await expect(this.inputField).toHaveAttribute('placeholder', placeholder)
  }

  /**
   * Asserts the send button is disabled
   */
  async expectSendButtonDisabled(): Promise<void> {
    await expect(this.sendButton).toBeDisabled()
  }

  /**
   * Asserts the send button is enabled
   */
  async expectSendButtonEnabled(): Promise<void> {
    await expect(this.sendButton).toBeEnabled()
  }

  /**
   * Asserts loading indicator is visible
   */
  async expectLoadingVisible(): Promise<void> {
    await expect(this.loadingIndicator).toBeVisible()
  }

  /**
   * Asserts loading indicator is hidden
   */
  async expectLoadingHidden(): Promise<void> {
    await expect(this.loadingIndicator).not.toBeVisible()
  }

  /**
   * Asserts unread badge shows specific count
   */
  async expectUnreadCount(count: number): Promise<void> {
    await expect(this.unreadBadge).toHaveText(String(count))
  }

  /**
   * Asserts suggestions are visible
   */
  async expectSuggestionsVisible(): Promise<void> {
    await expect(this.suggestionsContainer).toBeVisible()
  }

  /**
   * Asserts suggestions are hidden
   */
  async expectSuggestionsHidden(): Promise<void> {
    await expect(this.suggestionsContainer).not.toBeVisible()
  }

  /**
   * Asserts branding footer is visible
   */
  async expectBrandingVisible(): Promise<void> {
    await expect(this.brandingFooter).toBeVisible()
  }

  /**
   * Asserts character limit shows correct count
   */
  async expectCharLimit(current: number, max: number = 1000): Promise<void> {
    await expect(this.charLimit).toHaveText(`${current} / ${max}`)
  }
}

/**
 * Mock API response helpers
 */
export interface MockConfig {
  chatbotId?: string
  themeColor?: string
  welcomeMessage?: string
  position?: 'bottom-right' | 'bottom-left'
  botDisplayName?: string
  suggestedQuestions?: string[]
  botIcon?: string
  delay?: number
  errorStatus?: number
}

export interface MockChatResponse {
  response?: string
  messageId?: string
  handoffRequestId?: string
}

/**
 * Sets up mock API handlers for widget tests
 */
export async function setupWidgetMocks(
  page: Page,
  config: MockConfig = {}
): Promise<void> {
  const { chatbotId = 'test-chatbot', delay, errorStatus } = config

  // Mock config endpoint
  await page.route(`**/api/v1/public/chatbots/${chatbotId}`, async (route) => {
    if (delay) {
      await new Promise(resolve => setTimeout(resolve, delay))
    }

    if (errorStatus) {
      await route.fulfill({ status: errorStatus })
      return
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        theme_color: config.themeColor || '#6366f1',
        position: config.position || 'bottom-right',
        welcome_message: config.welcomeMessage || 'Merhaba! Size nasıl yardımcı olabilirim?',
        suggested_questions: config.suggestedQuestions || [],
        bot_display_name: config.botDisplayName || 'Test Bot',
        bot_icon: config.botIcon || undefined,
        hide_branding: false,
        max_chars: 1000,
      }),
    })
  })
}

/**
 * Sets up mock chat endpoint
 */
export async function setupChatMocks(
  page: Page,
  chatbotId: string = 'test-chatbot',
  response: MockChatResponse = {}
): Promise<void> {
  await page.route(`**/api/v1/public/chatbots/${chatbotId}/chat`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        response: response.response || 'Bu bir test yanıtıdır.',
        message_id: response.messageId || `msg-${Date.now()}`,
        handoff_request_id: response.handoffRequestId || undefined,
      }),
    })
  })
}

/**
 * Sets up mock chat endpoint with error
 */
export async function setupChatErrorMocks(
  page: Page,
  chatbotId: string = 'test-chatbot',
  status: number = 500,
  errorMessage: string = 'Internal Server Error'
): Promise<void> {
  await page.route(`**/api/v1/public/chatbots/${chatbotId}/chat`, async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({
        error: errorMessage,
      }),
    })
  })
}

/**
 * Sets up mock feedback endpoint
 */
export async function setupFeedbackMocks(
  page: Page,
  chatbotId: string = 'test-chatbot',
  success: boolean = true
): Promise<void> {
  await page.route(`**/api/v1/public/chatbots/${chatbotId}/feedback`, async (route) => {
    await route.fulfill({
      status: success ? 200 : 500,
      contentType: 'application/json',
      body: JSON.stringify({ success }),
    })
  })
}

/**
 * Waits for widget to be mounted
 */
export async function waitForWidgetMounted(page: Page): Promise<void> {
  await page.waitForSelector('#chatbot-widget-host', { timeout: 10000 })
  await page.waitForFunction(() => {
    const host = document.getElementById('chatbot-widget-host')
    return host && host.shadowRoot && host.shadowRoot.querySelector('.cbw-bubble')
  }, { timeout: 10000 })
}

/**
 * Gets computed style value for an element
 */
export async function getComputedStyleValue(
  page: Page,
  selector: string,
  property: string
): Promise<string> {
  const host = page.locator('#chatbot-widget-host')
  return await host.evaluate((el, { sel, prop }) => {
    const shadowEl = el.shadowRoot?.querySelector(sel)
    if (shadowEl) {
      return window.getComputedStyle(shadowEl).getPropertyValue(prop)
    }
    return ''
  }, { sel: selector, prop: property })
}
