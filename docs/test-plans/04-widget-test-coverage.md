# Test Plan: Widget Test Coverage Expansion

**Plan ID:** TP-WIDGET-COVERAGE-001  
**Priority:** HIGH  
**Estimated Duration:** 2-3 weeks  
**Target:** Comprehensive test coverage for widget functionality including chat, configuration, and theming  
**Status:** Draft  

---

## Executive Summary

The widget component currently has minimal test coverage - only mobile responsiveness tests exist. This plan outlines a comprehensive approach to adding:
1. E2E tests for chat interactions
2. E2E tests for widget configuration
3. E2E tests for theming/customization
4. Unit tests for core components
5. Integration tests with the main application

The goal is to ensure the widget is thoroughly tested across all user-facing features.

---

## Sisyphus Agent Prompt

```
You are Sisyphus, a senior frontend engineer with expertise in Playwright, React component testing, and widget development.

### Task Context
The Botla chat widget (widget/ directory) currently has minimal test coverage:
- Only mobile.spec.ts exists for mobile responsiveness
- No tests for chat interactions
- No tests for widget configuration
- No tests for theming/customization
- No unit tests for components
- No integration tests with backend

The widget is a critical user-facing component that embeds in customer websites. It needs comprehensive testing.

### Current Test Structure
widget/
├── e2e/
│   └── mobile.spec.ts          # Only mobile tests
├── src/
│   ├── __tests__/              # Empty or minimal
│   ├── App.tsx
│   ├── components/
│   │   ├── ChatWidget.tsx
│   │   ├── ChatWindow.tsx
│   │   ├── MessageBubble.tsx
│   │   ├── InputArea.tsx
│   │   └── Header.tsx
│   └── hooks/
│       └── useChat.ts

### Your Mission
Create comprehensive test coverage for the widget:

1. E2E Tests (Playwright):
   - Chat interactions (send message, receive response)
   - Widget configuration (theme, branding)
   - Widget embedding scenarios
   - Secure embedding with auth
   - Multi-step chat flows

2. Unit Tests (Vitest):
   - Component rendering
   - Hook functionality
   - State management
   - Helper functions

3. Integration Tests:
   - API integration with backend
   - WebSocket/real-time communication
   - Error handling

4. Test Infrastructure:
   - Test fixtures and utilities
   - Mock API responses
   - Test configuration

### Critical Rules
- Use the same patterns as frontend tests (frontend/e2e/helpers.ts)
- Create data-testid attributes for all interactive elements
- Mock external dependencies appropriately
- Follow Playwright best practices
- Create reusable test utilities

### Deliverables
1. E2E tests for chat interactions
2. E2E tests for widget configuration
3. E2E tests for theming
4. Unit tests for components
5. Test utilities and fixtures
6. Updated test configuration

Begin by analyzing the widget source code to understand its structure and functionality.
```

---

## Current Widget Analysis

### Widget Structure

```
widget/
├── src/
│   ├── App.tsx                    # Main app component
│   ├── index.tsx                  # Entry point
│   ├── components/
│   │   ├── ChatWidget.tsx         # Widget container
│   │   ├── ChatWindow.tsx         # Chat interface
│   │   ├── MessageBubble.tsx      # Individual messages
│   │   ├── InputArea.tsx          # Message input
│   │   ├── Header.tsx             # Chat header
│   │   ├── Launcher.tsx           # Open/close button
│   │   └── TypingIndicator.tsx    # "..." typing animation
│   ├── hooks/
│   │   ├── useChat.ts             # Chat state management
│   │   ├── useWidget.ts           # Widget visibility
│   │   └── useTheme.ts            # Theme management
│   ├── services/
│   │   ├── api.ts                 # API client
│   │   └── websocket.ts           # WebSocket client
│   ├── styles/
│   │   ├── theme.ts               # Theme definitions
│   │   └── animations.ts          # Animation utilities
│   ├── types/
│   │   └── index.ts               # TypeScript types
│   └── utils/
│       └── helpers.ts             # Helper functions
├── public/
│   └── index.html
├── package.json
├── playwright.config.ts
└── tailwind.config.js
```

### Component Dependency Graph

```
App.tsx
├── ChatWidget.tsx
│   ├── ChatWindow.tsx
│   │   ├── Header.tsx
│   │   ├── MessageBubble.tsx
│   │   └── InputArea.tsx
│   └── Launcher.tsx
├── useChat.ts (hook)
├── useWidget.ts (hook)
└── useTheme.ts (hook)
```

### Testable Units

| Component | Test Type | Priority | Complexity |
|-----------|-----------|----------|------------|
| ChatWidget.tsx | E2E + Unit | HIGH | Medium |
| ChatWindow.tsx | E2E + Unit | HIGH | Medium |
| MessageBubble.tsx | Unit | HIGH | Low |
| InputArea.tsx | E2E + Unit | HIGH | Medium |
| Header.tsx | E2E + Unit | MEDIUM | Low |
| Launcher.tsx | E2E | MEDIUM | Low |
| useChat.ts | Unit | HIGH | Medium |
| useWidget.ts | Unit | MEDIUM | Low |
| useTheme.ts | Unit | MEDIUM | Low |

---

## Step-by-Step Implementation Plan

### Phase 1: Test Infrastructure Setup (Days 1-3)

#### Step 1.1: Update Playwright Configuration

**File:** `widget/playwright.config.ts`

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'mobile-safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
  },
})
```

#### Step 1.2: Create Test Data Factories

**File:** `widget/src/__tests__/factories.ts`

```typescript
import { faker } from '@faker-js/faker'

export interface ChatMessage {
  id: string
  content: string
  role: 'user' | 'assistant'
  timestamp: Date
  isTyping?: boolean
}

export interface WidgetConfig {
  chatbotId: string
  theme: 'light' | 'dark' | 'auto'
  primaryColor: string
  accentColor: string
  fontFamily: string
  welcomeMessage: string
  placeholderText: string
  position: 'bottom-right' | 'bottom-left'
}

export interface UserSession {
  userId: string
  sessionToken: string
  expiresAt: Date
}

// Message Factory
export const createMessage = (overrides: Partial<ChatMessage> = {}): ChatMessage => ({
  id: faker.string.uuid(),
  content: faker.lorem.sentence(),
  role: faker.helpers.arrayElement(['user', 'assistant']),
  timestamp: new Date(),
  isTyping: false,
  ...overrides,
})

// Message List Factory
export const createMessageList = (count: number = 5): ChatMessage[] => {
  const messages: ChatMessage[] = []
  let role: 'user' | 'assistant' = 'user'
  
  for (let i = 0; i < count; i++) {
    messages.push(createMessage({
      id: `msg-${i}`,
      content: faker.lorem.sentence(),
      role,
    }))
    role = role === 'user' ? 'assistant' : 'user'
  }
  
  return messages
}

// Widget Config Factory
export const createWidgetConfig = (overrides: Partial<WidgetConfig> = {}): WidgetConfig => ({
  chatbotId: faker.string.uuid(),
  theme: 'light',
  primaryColor: '#3B82F6',
  accentColor: '#8B5CF6',
  fontFamily: 'Inter, sans-serif',
  welcomeMessage: 'Merhaba! Size nasıl yardımcı olabilirim?',
  placeholderText: 'Mesajınızı yazın...',
  position: 'bottom-right',
  ...overrides,
})

// User Session Factory
export const createUserSession = (overrides: Partial<UserSession> = {}): UserSession => ({
  userId: faker.string.uuid(),
  sessionToken: faker.string.alphanumeric(32),
  expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000),
  ...overrides,
})
```

#### Step 1.3: Create Mock API Handlers

**File:** `widget/e2e/mocks/api-handlers.ts`

```typescript
import { ChatMessage, WidgetConfig, UserSession } from '../../src/__tests__/factories'

export interface MockAPIConfig {
  messages?: ChatMessage[]
  config?: WidgetConfig
  session?: UserSession
  shouldFail?: boolean
  delay?: number
  errorMessage?: string
}

// Chat API Handlers
export const createChatHandlers = (config: MockAPIConfig = {}) => ({
  'POST /api/v1/chat': async (request, response) => {
    const { message, chatbotId, sessionId } = await request.body()
    
    if (config.delay) {
      await new Promise(resolve => setTimeout(resolve, config.delay))
    }
    
    if (config.shouldFail) {
      return response.status(500).send({
        error: 'Internal Server Error',
        message: config.errorMessage || 'Chat request failed',
      })
    }
    
    return response.send({
      message: {
        id: `msg-${Date.now()}`,
        content: `Echo: ${message}`,
        role: 'assistant',
        timestamp: new Date().toISOString(),
      },
      sessionId: sessionId || 'session-123',
    })
  },
  
  'GET /api/v1/chat/:sessionId/history': async (request, response) => {
    return response.send({
      messages: config.messages || [],
    })
  },
  
  'POST /api/v1/chat/:sessionId/feedback': async (request, response) => {
    return response.send({ success: true })
  },
})

// Config API Handlers
export const createConfigHandlers = (config: MockAPIConfig = {}) => ({
  'GET /api/v1/widget/config': async (request, response) => {
    return response.send(config.config || {
      chatbotId: 'chatbot-123',
      theme: 'light',
      primaryColor: '#3B82F6',
      accentColor: '#8B5CF6',
      welcomeMessage: 'Merhaba!',
      placeholderText: 'Mesajınızı yazın...',
    })
  },
})

// Auth API Handlers
export const createAuthHandlers = (config: MockAPIConfig = {}) => ({
  'POST /api/v1/widget/auth': async (request, response) => {
    return response.send(config.session || {
      userId: 'user-123',
      sessionToken: 'session-token-123',
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    })
  },
})
```

#### Step 1.4: Create Test Helpers

**File:** `widget/e2e/helpers.ts`

```typescript
import { Page, Locator, expect } from '@playwright/test'
import { ChatMessage, WidgetConfig } from '../../src/__tests__/factories'

export class WidgetHelper {
  private page: Page
  
  constructor(page: Page) {
    this.page = page
  }
  
  // Widget Launcher
  get launcherButton(): Locator {
    return this.page.getByTestId('widget-launcher-button')
  }
  
  get closeButton(): Locator {
    return this.page.getByTestId('widget-close-button')
  }
  
  // Chat Window
  get chatWindow(): Locator {
    return this.page.getByTestId('chat-window')
  }
  
  get messageList(): Locator {
    return this.page.getByTestId('message-list')
  }
  
  get inputArea(): Locator {
    return this.page.getByTestId('input-area')
  }
  
  get sendButton(): Locator {
    return this.page.getByTestId('send-button')
  }
  
  get messageInput(): Locator {
    return this.page.getByTestId('message-input')
  }
  
  // Messages
  get userMessages(): Locator {
    return this.page.getByTestId('message-user')
  }
  
  get assistantMessages(): Locator {
    return this.page.getByTestId('message-assistant')
  }
  
  get typingIndicator(): Locator {
    return this.page.getByTestId('typing-indicator')
  }
  
  // Actions
  async openWidget(): Promise<void> {
    await this.launcherButton.click()
    await expect(this.chatWindow).toBeVisible()
  }
  
  async closeWidget(): Promise<void> {
    await this.closeButton.click()
    await expect(this.chatWindow).toBeHidden()
  }
  
  async sendMessage(message: string): Promise<void> {
    await this.messageInput.fill(message)
    await this.sendButton.click()
  }
  
  async expectMessageCount(count: number): Promise<void> {
    await expect(this.messageList.locator('[data-testid^="message-"]')).toHaveCount(count)
  }
  
  async expectLastMessage(content: string): Promise<void> {
    const lastMessage = this.messageList.locator('[data-testid^="message-"]').last()
    await expect(lastMessage).toContainText(content)
  }
  
  async expectTypingVisible(): Promise<void> {
    await expect(this.typingIndicator).toBeVisible()
  }
  
  async expectTypingHidden(): Promise<void> {
    await expect(this.typingIndicator).toBeHidden()
  }
}

// Mock Setup Helper
export async function setupWidgetMocks(
  page: Page,
  config?: {
    messages?: ChatMessage[]
    config?: WidgetConfig
    delay?: number
  }
): Promise<void> {
  // Setup API mocking
  await page.route('**/api/v1/chat', async (route) => {
    if (config?.delay) {
      await new Promise(resolve => setTimeout(resolve, config.delay))
    }
    
    await route.fulfill({
      status: 200,
      body: JSON.stringify({
        message: {
          id: `msg-${Date.now()}`,
          content: 'Test response',
          role: 'assistant',
          timestamp: new Date().toISOString(),
        },
      }),
    })
  })
  
  await page.route('**/api/v1/widget/config', async (route) => {
    await route.fulfill({
      status: 200,
      body: JSON.stringify(config?.config || {
        chatbotId: 'test-bot',
        theme: 'light',
        primaryColor: '#3B82F6',
      }),
    })
  })
}
```

---

### Phase 2: Chat Interaction E2E Tests (Days 4-7)

#### Step 2.1: chat.spec.ts - Basic Chat

**File:** `widget/e2e/chat.spec.ts`

```typescript
import { test, expect } from '@playwright/test'
import { WidgetHelper } from './helpers'
import { setupWidgetMocks } from './helpers'
import { createMessageList } from '../../src/__tests__/factories'

test.describe('Chat Interactions', () => {
  let helper: WidgetHelper
  
  test.beforeEach(async ({ page }) => {
    helper = new WidgetHelper(page)
    await setupWidgetMocks(page)
  })
  
  test('should open chat widget when launcher is clicked', async ({ page }) => {
    // Widget should be closed initially
    await expect(helper.chatWindow).toBeHidden()
    
    // Click launcher to open
    await helper.openWidget()
    
    // Verify chat window is visible
    await expect(helper.chatWindow).toBeVisible()
    
    // Verify welcome message is shown
    await expect(helper.messageList).toContainText('Merhaba!')
  })
  
  test('should send user message and receive response', async ({ page }) => {
    await helper.openWidget()
    
    // Send a message
    await helper.sendMessage('Hello, I need help')
    
    // Verify user message appears
    await expect(helper.messageList).toContainText('Hello, I need help')
    
    // Verify assistant response appears
    await expect(helper.messageList).toContainText('Echo: Hello, I need help')
  })
  
  test('should show typing indicator while waiting for response', async ({ page }) => {
    await helper.openWidget()
    
    // Mock delayed response
    await page.route('**/api/v1/chat', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          message: {
            id: 'msg-delayed',
            content: 'Delayed response',
            role: 'assistant',
            timestamp: new Date().toISOString(),
          },
        }),
      })
    })
    
    // Start a slow request
    const responsePromise = page.request.post('/api/v1/chat', {
      data: { message: 'Slow request' },
    })
    
    // Typing indicator should appear
    await helper.expectTypingVisible()
    
    // Wait for response
    await responsePromise
    
    // Typing indicator should hide
    await helper.expectTypingHidden()
  })
  
  test('should display multiple messages in conversation', async ({ page }) => {
    await helper.openWidget()
    
    // Send multiple messages
    await helper.sendMessage('First message')
    await helper.sendMessage('Second message')
    await helper.sendMessage('Third message')
    
    // Verify all messages are displayed
    await helper.expectMessageCount(6) // 3 user + 3 assistant
  })
  
  test('should close chat widget when close button is clicked', async ({ page }) => {
    await helper.openWidget()
    
    // Verify chat window is open
    await expect(helper.chatWindow).toBeVisible()
    
    // Close the widget
    await helper.closeWidget()
    
    // Verify chat window is hidden
    await expect(helper.chatWindow).toBeHidden()
  })
  
  test('should show error message on API failure', async ({ page }) => {
    await helper.openWidget()
    
    // Mock API failure
    await page.route('**/api/v1/chat', async (route) => {
      await route.fulfill({
        status: 500,
        body: JSON.stringify({
          error: 'Internal Server Error',
          message: 'Failed to send message',
        }),
      })
    })
    
    // Send a message
    await helper.sendMessage('Test message')
    
    // Verify error message is shown
    await expect(helper.messageList).toContainText('Failed to send message')
  })
  
  test('should preserve message history across widget open/close', async ({ page }) => {
    await helper.openWidget()
    
    // Send a message
    await helper.sendMessage('Message before close')
    
    // Close widget
    await helper.closeWidget()
    
    // Reopen widget
    await helper.openWidget()
    
    // Verify message is still there
    await expect(helper.messageList).toContainText('Message before close')
  })
  
  test('should handle empty message input gracefully', async ({ page }) => {
    await helper.openWidget()
    
    // Click send without typing message
    await helper.sendButton.click()
    
    // Should not send empty message
    await expect(helper.messageList).not.toContainText('Echo:')
  })
  
  test('should show welcome message only on first open', async ({ page }) => {
    await helper.openWidget()
    
    // Welcome message should appear
    await expect(helper.messageList).toContainText('Merhaba!')
    
    // Close and reopen
    await helper.closeWidget()
    await helper.openWidget()
    
    // Welcome message should not appear again
    await expect(helper.messageList.locator('text=Merhaba!')).toHaveCount(1)
  })
})
```

#### Step 2.2: chat-flows.spec.ts - Multi-step Flows

**File:** `widget/e2e/chat-flows.spec.ts`

```typescript
import { test, expect } from '@playwright/test'
import { WidgetHelper } from './helpers'

test.describe('Chat Flows', () => {
  let helper: WidgetHelper
  
  test.beforeEach(async ({ page }) => {
    helper = new WidgetHelper(page)
    // Setup mocks
    await page.route('**/api/v1/chat', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          message: {
            id: `msg-${Date.now()}`,
            content: 'I can help with that!',
            role: 'assistant',
            timestamp: new Date().toISOString(),
          },
        }),
      })
    })
  })
  
  test('should handle question and answer flow', async ({ page }) => {
    await helper.openWidget()
    
    // User asks a question
    await helper.sendMessage('What are your business hours?')
    await expect(helper.messageList).toContainText('What are your business hours?')
    
    // Assistant responds
    await expect(helper.messageList).toContainText('I can help with that!')
    
    // User follows up
    await helper.sendMessage('Do you support English?')
    await expect(helper.messageList).toContainText('Do you support English?')
    
    // Assistant follows up
    await expect(helper.messageList).toContainText('I can help with that!')
  })
  
  test('should handle long conversation', async ({ page }) => {
    await helper.openWidget()
    
    const messages = [
      'Hello',
      'I need assistance with my order',
      'Order number is 12345',
      'When will it arrive?',
      'Can I change the delivery address?',
      'Thank you for your help',
    ]
    
    for (const message of messages) {
      await helper.sendMessage(message)
      await expect(helper.messageList).toContainText(message)
    }
    
    // Verify all messages are in the list
    await helper.expectMessageCount(messages.length * 2)
  })
  
  test('should handle rapid message sending', async ({ page }) => {
    await helper.openWidget()
    
    // Send messages quickly
    await helper.sendMessage('Message 1')
    await helper.sendMessage('Message 2')
    await helper.sendMessage('Message 3')
    
    // All messages should be in the list
    await expect(helper.messageList).toContainText('Message 1')
    await expect(helper.messageList).toContainText('Message 2')
    await expect(helper.messageList).toContainText('Message 3')
  })
})
```

---

### Phase 3: Widget Configuration E2E Tests (Days 8-10)

#### Step 3.1: widget-config.spec.ts - Configuration

**File:** `widget/e2e/widget-config.spec.ts`

```typescript
import { test, expect } from '@playwright/test'

test.describe('Widget Configuration', () => {
  test.beforeEach(async ({ page }) => {
    // Setup config endpoint
    await page.route('**/api/v1/widget/config', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          chatbotId: 'test-bot',
          theme: 'light',
          primaryColor: '#3B82F6',
          accentColor: '#8B5CF6',
          fontFamily: 'Inter, sans-serif',
          welcomeMessage: 'Welcome! How can I help?',
          placeholderText: 'Type your message...',
          position: 'bottom-right',
        }),
      })
    })
    
    await page.goto('/')
  })
  
  test('should load configuration from API', async ({ page }) => {
    // Open widget
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify custom welcome message
    await expect(page.getByTestId('message-list')).toContainText('Welcome! How can I help?')
    
    // Verify custom placeholder
    await expect(page.getByTestId('message-input')).toHaveAttribute(
      'placeholder',
      'Type your message...'
    )
  })
  
  test('should apply custom primary color', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify color is applied (check CSS)
    const launcherButton = page.getByTestId('widget-launcher-button')
    const backgroundColor = await launcherButton.evaluate(
      (el) => window.getComputedStyle(el).backgroundColor
    )
    
    // RGB for #3B82F6 is rgb(59, 130, 246)
    expect(backgroundColor).toContain('59')
    expect(backgroundColor).toContain('130')
    expect(backgroundColor).toContain('246')
  })
  
  test('should apply custom accent color', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify accent color is applied
    const sendButton = page.getByTestId('send-button')
    const color = await sendButton.evaluate(
      (el) => window.getComputedStyle(el).color
    )
    
    // RGB for #8B5CF6 is rgb(139, 92, 246)
    expect(color).toContain('139')
    expect(color).toContain('92')
    expect(color).toContain('246')
  })
  
  test('should handle dark theme configuration', async ({ page }) => {
    // Mock dark theme config
    await page.route('**/api/v1/widget/config', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          chatbotId: 'test-bot',
          theme: 'dark',
          primaryColor: '#1F2937',
          accentColor: '#6366F1',
          welcomeMessage: 'Dark mode enabled',
          placeholderText: 'Type here...',
          position: 'bottom-right',
        }),
      })
    })
    
    await page.goto('/')
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify dark mode is applied
    const chatWindow = page.getByTestId('chat-window')
    await expect(chatWindow).toHaveClass(/dark/)
  })
  
  test('should handle invalid configuration gracefully', async ({ page }) => {
    // Mock invalid config
    await page.route('**/api/v1/widget/config', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          // Missing required fields
          chatbotId: 'test-bot',
        }),
      })
    })
    
    // Should not crash
    await page.getByTestId('widget-launcher-button').click()
    await expect(page.getByTestId('chat-window')).toBeVisible()
  })
})
```

#### Step 3.2: widget-branding.spec.ts - Branding Customization

**File:** `widget/e2e/widget-branding.spec.ts`

```typescript
import { test, expect } from '@playwright/test'

test.describe('Widget Branding', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v1/widget/config', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          chatbotId: 'test-bot',
          theme: 'light',
          primaryColor: '#3B82F6',
          accentColor: '#8B5CF6',
          fontFamily: 'Inter, sans-serif',
          welcomeMessage: 'Branded Welcome!',
          placeholderText: 'Ask us anything...',
          position: 'bottom-right',
          companyLogo: 'https://example.com/logo.png',
          companyName: 'Test Company',
        }),
      })
    })
    
    await page.goto('/')
  })
  
  test('should display company name in header', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify company name is displayed
    await expect(page.getByTestId('chat-header-title')).toContainText('Test Company')
  })
  
  test('should display company logo', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify logo image is present
    const logo = page.getByTestId('chat-header-logo')
    await expect(logo).toHaveAttribute('src', 'https://example.com/logo.png')
  })
  
  test('should apply custom font family', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify font is applied
    const input = page.getByTestId('message-input')
    const fontFamily = await input.evaluate(
      (el) => window.getComputedStyle(el).fontFamily
    )
    
    expect(fontFamily).toContain('Inter')
  })
})
```

---

### Phase 4: Secure Embedding E2E Tests (Days 11-12)

#### Step 4.1: widget-embed-secure.spec.ts - Authentication

**File:** `widget/e2e/widget-embed-secure.spec.ts`

```typescript
import { test, expect } from '@playwright/test'

test.describe('Secure Widget Embedding', () => {
  test.beforeEach(async ({ page }) => {
    // Setup auth endpoint
    await page.route('**/api/v1/widget/auth', async (route) => {
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          userId: 'user-123',
          sessionToken: 'session-token-abc',
          expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
        }),
      })
    })
    
    await page.goto('/')
  })
  
  test('should authenticate user on widget open', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Verify auth request was made
    const authRequest = await page.waitForRequest('**/api/v1/widget/auth')
    expect(authRequest).toBeDefined()
    
    // Verify chat works after auth
    await expect(page.getByTestId('chat-window')).toBeVisible()
  })
  
  test('should handle authentication failure gracefully', async ({ page }) => {
    // Mock auth failure
    await page.route('**/api/v1/widget/auth', async (route) => {
      await route.fulfill({
        status: 401,
        body: JSON.stringify({
          error: 'Unauthorized',
          message: 'Invalid authentication token',
        }),
      })
    })
    
    await page.getByTestId('widget-launcher-button').click()
    
    // Widget should still open but show error state
    await expect(page.getByTestId('chat-window')).toBeVisible()
    await expect(page.getByTestId('auth-error-message')).toContainText('Please log in')
  })
  
  test('should include auth token in API requests', async ({ page }) => {
    await page.getByTestId('widget-launcher-button').click()
    
    // Send a message
    await page.getByTestId('message-input').fill('Test message')
    await page.getByTestId('send-button').click()
    
    // Verify auth token is included in request
    const chatRequest = await page.waitForRequest('**/api/v1/chat')
    const postData = chatRequest.postDataJSON()
    
    expect(postData.sessionToken).toBe('session-token-abc')
  })
})
```

---

### Phase 5: Mobile and Responsiveness Tests (Days 13-14)

#### Step 5.1: mobile.spec.ts - Enhanced Mobile Tests

**File:** `widget/e2e/mobile.spec.ts` (enhance existing)

```typescript
import { test, expect } from '@playwright/test'

test.describe('Mobile Responsiveness', () => {
  test.use({
    viewport: { width: 390, height: 844 }, // iPhone 12 Pro
  })
  
  test('should display launcher button on mobile', async ({ page }) => {
    await page.goto('/')
    
    const launcher = page.getByTestId('widget-launcher-button')
    await expect(launcher).toBeVisible()
    
    // Should be positioned correctly on mobile
    const box = await launcher.boundingBox()
    expect(box?.x).toBeGreaterThan(280) // Right side of screen
  })
  
  test('should open full-screen chat on mobile', async ({ page }) => {
    await page.goto('/')
    await page.getByTestId('widget-launcher-button').click()
    
    const chatWindow = page.getByTestId('chat-window')
    await expect(chatWindow).toBeVisible()
    
    // On mobile, chat should take full width
    const box = await chatWindow.boundingBox()
    expect(box?.width).toBeGreaterThan(350)
  })
  
  test('should handle touch events correctly', async ({ page }) => {
    await page.goto('/')
    
    // Tap launcher
    await page.getByTestId('widget-launcher-button').tap()
    
    // Chat should open
    await expect(page.getByTestId('chat-window')).toBeVisible()
    
    // Type message
    await page.getByTestId('message-input').fill('Mobile test')
    await page.getByTestId('send-button').tap()
    
    // Message should be sent
    await expect(page.getByTestId('message-list')).toContainText('Mobile test')
  })
  
  test('should work in landscape orientation', async ({ page }) => {
    await page.setViewportSize({ width: 844, height: 390 })
    await page.goto('/')
    
    await page.getByTestId('widget-launcher-button').click()
    await expect(page.getByTestId('chat-window')).toBeVisible()
  })
})
```

---

### Phase 6: Unit Tests (Days 15-18)

#### Step 6.1: Component Unit Tests

**File:** `widget/src/__tests__/components/MessageBubble.test.tsx`

```typescript
import { render, screen } from '@testing-library/react'
import { MessageBubble } from '../../components/MessageBubble'
import { ChatMessage } from '../factories'

describe('MessageBubble', () => {
  test('renders user message correctly', () => {
    const message: ChatMessage = {
      id: 'msg-1',
      content: 'Hello, I need help',
      role: 'user',
      timestamp: new Date(),
    }
    
    render(<MessageBubble message={message} />)
    
    expect(screen.getByTestId('message-user')).toBeInTheDocument()
    expect(screen.getByText('Hello, I need help')).toBeInTheDocument()
    expect(screen.getByTestId('message-user')).toHaveClass('message-user')
  })
  
  test('renders assistant message correctly', () => {
    const message: ChatMessage = {
      id: 'msg-2',
      content: 'How can I assist you?',
      role: 'assistant',
      timestamp: new Date(),
    }
    
    render(<MessageBubble message={message} />)
    
    expect(screen.getByTestId('message-assistant')).toBeInTheDocument()
    expect(screen.getByText('How can I assist you?')).toBeInTheDocument()
    expect(screen.getByTestId('message-assistant')).toHaveClass('message-assistant')
  })
  
  test('formats timestamp correctly', () => {
    const message: ChatMessage = {
      id: 'msg-3',
      content: 'Test message',
      role: 'user',
      timestamp: new Date('2024-01-15T10:30:00Z'),
    }
    
    render(<MessageBubble message={message} />)
    
    expect(screen.getByTestId('message-timestamp')).toHaveTextContent('10:30')
  })
  
  test('applies custom className', () => {
    const message: ChatMessage = {
      id: 'msg-4',
      content: 'Custom message',
      role: 'user',
      timestamp: new Date(),
    }
    
    render(<MessageBubble message={message} className="custom-class" />)
    
    expect(screen.getByTestId('message-user')).toHaveClass('custom-class')
  })
})
```

#### Step 6.2: Hook Unit Tests

**File:** `widget/src/__tests__/hooks/useChat.test.ts`

```typescript
import { renderHook, waitFor } from '@testing-library/react'
import { useChat } from '../../hooks/useChat'
import { ChatMessage } from '../factories'

describe('useChat', () => {
  test('initial state is empty', () => {
    const { result } = renderHook(() => useChat())
    
    expect(result.current.messages).toEqual([])
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })
  
  test('sendMessage adds user message', async () => {
    const { result } = renderHook(() => useChat())
    
    await result.current.sendMessage('Hello')
    
    expect(result.current.messages).toHaveLength(2) // User + assistant
    expect(result.current.messages[0].content).toBe('Hello')
    expect(result.current.messages[0].role).toBe('user')
  })
  
  test('clearMessages removes all messages', async () => {
    const { result } = renderHook(() => useChat())
    
    await result.current.sendMessage('Test')
    expect(result.current.messages).toHaveLength(2)
    
    result.current.clearMessages()
    
    expect(result.current.messages).toEqual([])
  })
  
  test('sets error on API failure', async () => {
    const { result } = renderHook(() => useChat())
    
    // Mock API failure
    await result.current.sendMessage('Test')
    
    expect(result.current.error).toBeNull()
  })
})
```

#### Step 6.3: Integration Tests

**File:** `widget/src/__tests__/App.test.tsx`

```typescript
import { render, screen, waitFor } from '@testing-library/react'
import { App } from '../App'

// Mock API calls
global.fetch = jest.fn().mockResolvedValue({
  json: async () => ({
    chatbotId: 'test-bot',
    welcomeMessage: 'Hello!',
    theme: 'light',
  }),
})

describe('App', () => {
  test('renders launcher button', () => {
    render(<App />)
    expect(screen.getByTestId('widget-launcher-button')).toBeInTheDocument()
  })
  
  test('opens chat window on launcher click', async () => {
    render(<App />)
    
    await screen.getByTestId('widget-launcher-button').click()
    
    await waitFor(() => {
      expect(screen.getByTestId('chat-window')).toBeVisible()
    })
  })
  
  test('loads configuration from API', async () => {
    render(<App />)
    
    await screen.getByTestId('widget-launcher-button').click()
    
    await waitFor(() => {
      expect(screen.getByTestId('message-list')).toContainText('Hello!')
    })
  })
})
```

---

### Phase 7: Verification (Days 19-21)

#### Step 7.1: Run All Tests
```bash
# Run E2E tests
npm run e2e

# Run unit tests
npm run test

# Check coverage
npm run test:coverage
```

#### Step 7.2: Coverage Report
```bash
# Generate coverage report
npm run test:coverage -- --coverage-reporters=html

# Check coverage thresholds
cat coverage/coverage-summary.json
```

---

## Progress Tracking

### Daily Checklist

- [ ] Create N new test files
- [ ] Add N new test cases
- [ ] Run tests to verify
- [ ] Update documentation
- [ ] Report progress

### Milestone Reviews

| Milestone | Target Date | Coverage Target | Status |
|-----------|-------------|-----------------|--------|
| Phase 1 Complete | Day 3 | Infrastructure ready | ⏳ |
| Phase 2 Complete | Day 7 | Chat tests 100% | ⏳ |
| Phase 3 Complete | Day 10 | Config tests 100% | ⏳ |
| Phase 4 Complete | Day 12 | Auth tests 100% | ⏳ |
| Phase 5 Complete | Day 14 | Mobile tests 100% | ⏳ |
| Phase 6 Complete | Day 18 | Unit tests 80%+ | ⏳ |
| Phase 7 Complete | Day 21 | Verification complete | ⏳ |

---

## Success Criteria

### Functional Requirements
- [ ] Chat interaction tests complete
- [ ] Widget configuration tests complete
- [ ] Theming tests complete
- [ ] Secure embedding tests complete
- [ ] Mobile responsiveness tests complete
- [ ] Unit tests for all components
- [ ] Unit tests for all hooks

### Coverage Targets
| Metric | Target | Current |
|--------|--------|---------|
| E2E Tests | 50+ | 0 |
| Unit Tests | 30+ | 0 |
| E2E Coverage | 80%+ | 0% |
| Unit Coverage | 80%+ | 0% |

---

## Dependencies

- `@playwright/test` - E2E testing framework
- `vitest` - Unit testing framework
- `@testing-library/react` - React component testing
- `@faker-js/faker` - Test data generation

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Sisyphus | Initial plan |

---

*This plan is part of the comprehensive test improvement initiative. For questions or clarifications, refer to the project documentation or consult with the team lead.*
