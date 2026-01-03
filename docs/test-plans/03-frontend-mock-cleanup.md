# Test Plan: Frontend Mock Cleanup and Selector Stabilization

**Plan ID:** TP-FRONTEND-MOCK-001  
**Priority:** HIGH  
**Estimated Duration:** 2-3 weeks  
**Target:** Stable, maintainable E2E tests with proper selectors and reliable assertions  
**Status:** Draft  

---

## Executive Summary

This plan addresses critical issues in the frontend E2E and unit test infrastructure:

1. **Fragile Selectors**: Tests use magic strings and text content that breaks with UI changes
2. **Mock Tight Coupling**: Mocks are implementation-specific and brittle
3. **Anti-Pattern `waitForTimeout`**: Unreliable timing-based waiting
4. **Missing `data-testid` Attributes**: No stable element identifiers
5. **Mixed Testing Concerns**: Tests mix UI and business logic testing

The goal is to create a robust, maintainable test suite that can withstand UI refactoring.

---

## Sisyphus Agent Prompt

```
You are Sisyphus, a senior frontend engineer with expertise in Playwright, React Testing Library, and test architecture.

### Task Context
The Botla frontend test suite has several critical issues that cause test fragility and maintenance burden:

1. FRAGILE SELECTORS: Tests use magic strings like "Hoş Geldiniz" and "Giriş Yap" which break when:
   - Turkish text changes
   - Design updates modify copy
   - Localization changes

2. UNRELIABLE TIMING: Tests use `page.waitForTimeout(2000)` which:
   - Makes tests slow when waits are too long
   - Makes tests flaky when waits are too short
   - Doesn't adapt to system load

3. NO TEST IDS: Components lack data-testid attributes for stable element selection

4. MOCK TIGHT COUPLING: E2E mocks in e2e/helpers.ts are implementation-specific and break when API changes

### Your Mission
Phase 1: Add data-testid to all testable components
- Read all React components in frontend/src/
- Add data-testid attributes to interactive elements
- Follow naming convention: data-testid="component-name-element-type"

Phase 2: Refactor E2E selectors
- Replace all getByText() with getByTestId()
- Replace all magic strings with semantic selectors
- Create page object pattern for common interactions

Phase 3: Replace waitForTimeout with explicit waits
- Find all page.waitForTimeout() calls
- Replace with explicit waitForSelector or waitForLoadState
- Add custom wait helpers for common patterns

Phase 4: Improve mock infrastructure
- Create mock factories for reusable test data
- Add mock validation to catch API changes
- Document mock patterns and usage

### Critical Rules
- NEVER break existing functionality
- Add data-testid without changing component behavior
- Keep refactoring changes minimal and focused
- Run tests after each batch of changes

### Deliverables
1. All interactive components have data-testid attributes
2. All E2E tests use getByTestId() selectors
3. No waitForTimeout in test code
4. Mock factories for common test data
5. Documentation of test patterns

Begin by analyzing the current component structure and test files.
```

---

## Current State Analysis

### Fragile Selector Inventory

#### E2E Test Files with Problematic Selectors

| File | Magic Strings | Text Selectors | WaitForTimeout |
|------|---------------|----------------|----------------|
| `auth.spec.ts` | "Hoş Geldiniz", "Giriş Yap" | 15+ | 5+ |
| `chatbot.spec.ts` | "Yeni Chatbot", "Kaydet" | 20+ | 3+ |
| `smoke.spec.ts` | "Hoş Geldiniz", "Gönder" | 10+ | 2+ |
| `widget-*.spec.ts` | Various Turkish text | 15+ | 4+ |
| `mobile-responsiveness.spec.ts` | "Menü", "Kapat" | 8+ | 2+ |
| `chunk-inspector.spec.ts` | "İncele", "Detaylar" | 12+ | 3+ |

#### Problematic Pattern Examples

```typescript
// ❌ PROBLEMATIC - Magic strings in Turkish
await page.getByText("Hoş Geldiniz").click()
await page.getByText("Giriş Yap").click()
await page.getByText("Şifremi Unuttum").click()

// ❌ PROBLEMATIC - waitForTimeout
await page.waitForTimeout(2000)
await page.waitForTimeout(5000)

// ❌ PROBLEMATIC - Text-based assertions
await expect(page.getByText("Başarılı")).toBeVisible()
await expect(page.getByText("Hata oluştu")).toBeVisible()
```

### Component Analysis

#### Components Requiring data-testid

**Authentication Components:**
```
src/pages/LoginPage.tsx
src/pages/RegisterPage.tsx
src/pages/ForgotPasswordPage.tsx
```

**Chat Components:**
```
src/components/ChatWindow.tsx
src/components/MessageBubble.tsx
src/components/InputArea.tsx
src/components/ChatHeader.tsx
```

**Chatbot Management:**
```
src/pages/ChatbotList.tsx
src/pages/ChatbotDetail.tsx
src/components/ChatbotCard.tsx
src/components/SourceList.tsx
```

**Common Components:**
```
src/components/Button.tsx
src/components/Input.tsx
src/components/Select.tsx
src/components/Modal.tsx
src/components/Dropdown.tsx
```

---

## Step-by-Step Implementation Plan

### Phase 1: data-testid Attribute Addition (Days 1-5)

#### Step 1.1: Create data-testid Convention

**Naming Convention:**
```
component-name-element-type
```

**Examples:**
```
login-page-email-input
login-page-password-input
login-page-submit-button
login-page-error-message
chat-window-message-bubble
chat-window-input-area
chat-window-send-button
chatbot-list-create-button
chatbot-detail-header
chatbot-detail-sources-tab
```

#### Step 1.2: Button Components

**Pattern:**
```tsx
// BEFORE
<button className="submit-btn">Gönder</button>

// AFTER
<button 
  className="submit-btn" 
  data-testid="login-page-submit-button"
>
  Gönder
</button>
```

**Files to Update:**
- `frontend/src/components/Button.tsx` (base component)
- All button usages across components

#### Step 1.3: Input Components

**Pattern:**
```tsx
// BEFORE
<input type="email" placeholder="E-posta" />

// AFTER
<input 
  type="email" 
  placeholder="E-posta"
  data-testid="login-page-email-input"
/>
```

**Files to Update:**
- `frontend/src/components/Input.tsx` (base component)
- All input usages

#### Step 1.4: Page Components

**Pattern for Pages:**
```tsx
// BEFORE
export const LoginPage = () => {
  return (
    <div className="login-page">
      <h1>Hoş Geldiniz</h1>
      ...
    </div>
  )
}

// AFTER
export const LoginPage = () => {
  return (
    <div className="login-page" data-testid="login-page">
      <h1 data-testid="login-page-title">Hoş Geldiniz</h1>
      ...
    </div>
  )
}
```

#### Step 1.5: Interactive Elements Checklist

| Element Type | data-testid Pattern | Example |
|--------------|---------------------|---------|
| Button | `{page}-{action}-button` | `login-page-submit-button` |
| Input | `{page}-{field}-input` | `login-page-email-input` |
| Select | `{page}-{field}-select` | `settings-language-select` |
| Checkbox | `{page}-{label}-checkbox` | `settings-notifications-checkbox` |
| Link | `{page}-{label}-link` | `login-page-forgot-password-link` |
| Tab | `{page}-{name}-tab` | `chatbot-detail-sources-tab` |
| Modal | `{name}-modal` | `confirm-delete-modal` |
| Error Message | `{page}-{context}-error` | `login-page-email-error` |
| Success Message | `{page}-{context}-success` | `settings-save-success` |
| Loading Spinner | `{page}-{context}-loading` | `chat-loading` |

---

### Phase 2: E2E Selector Refactoring (Days 6-10)

#### Step 2.1: Create Page Object Pattern

**Before (scattered selectors):**
```typescript
// auth.spec.ts
test('successful login', async ({ page }) => {
  await page.goto('/login')
  await page.getByText("E-posta").fill('test@example.com')
  await page.getByText("Şifre").fill('password')
  await page.getByText("Giriş Yap").click()
  await expect(page.getByText("Hoş Geldiniz")).toBeVisible()
})
```

**After (page objects):**
```typescript
// pages/LoginPage.ts
export class LoginPage {
  constructor(private page: Page) {}

  get emailInput() {
    return this.page.getByTestId('login-page-email-input')
  }

  get passwordInput() {
    return this.page.getByTestId('login-page-password-input')
  }

  get submitButton() {
    return this.page.getByTestId('login-page-submit-button')
  }

  get errorMessage() {
    return this.page.getByTestId('login-page-error-message')
  }

  async fillEmail(email: string) {
    await this.emailInput.fill(email)
  }

  async fillPassword(password: string) {
    await this.passwordInput.fill(password)
  }

  async clickSubmit() {
    await this.submitButton.click()
  }

  async login(email: string, password: string) {
    await this.fillEmail(email)
    await this.fillPassword(password)
    await this.clickSubmit()
  }
}
```

#### Step 2.2: Create Page Object Factory

```typescript
// pages/index.ts
import { Page } from '@playwright/test'
import { LoginPage } from './LoginPage'
import { DashboardPage } from './DashboardPage'
import { ChatbotListPage } from './ChatbotListPage'
import { ChatbotDetailPage } from './ChatbotDetailPage'

export class PageFactory {
  constructor(private page: Page) {}

  get loginPage() {
    return new LoginPage(this.page)
  }

  get dashboardPage() {
    return new DashboardPage(this.page)
  }

  get chatbotListPage() {
    return new ChatbotListPage(this.page)
  }

  get chatbotDetailPage() {
    return new ChatbotDetailPage(this.page)
  }
}

// Helper function for tests
export const createPageFactory = (page: Page) => new PageFactory(page)
```

#### Step 2.3: Refactor auth.spec.ts

**Before:**
```typescript
test('successful login', async ({ page }) => {
  await page.goto('/login')
  await page.getByText("E-posta").fill('test@example.com')
  await page.getByText("Şifre").fill('password')
  await page.getByText("Giriş Yap").click()
  await expect(page.getByText("Hoş Geldiniz")).toBeVisible()
})
```

**After:**
```typescript
import { createPageFactory } from '../pages'

test('successful login', async ({ page }) => {
  const pages = createPageFactory(page)
  
  await pages.loginPage.goto()
  await pages.loginPage.login('test@example.com', 'password')
  
  await expect(pages.dashboardPage.welcomeMessage).toBeVisible()
})
```

#### Step 2.4: Refactor All E2E Files

| File | Page Objects to Create | Priority |
|------|------------------------|----------|
| `auth.spec.ts` | LoginPage, RegisterPage | HIGH |
| `chatbot.spec.ts` | ChatbotListPage, ChatbotDetailPage | HIGH |
| `smoke.spec.ts` | All pages | HIGH |
| `mobile-responsiveness.spec.ts` | Navigation, ResponsiveLayout | MEDIUM |
| `widget-embed.spec.ts` | WidgetEmbedPage | MEDIUM |
| `widget-branding.spec.ts` | WidgetBrandingPage | MEDIUM |
| `chunk-inspector.spec.ts` | ChunkInspectorPage | LOW |

---

### Phase 3: Remove waitForTimeout (Days 11-13)

#### Step 3.1: Create Wait Helpers

```typescript
// helpers/waits.ts
import { Page, Locator } from '@playwright/test'

/**
 * Wait for element to be visible with timeout
 */
export async function waitForVisible(
  page: Page,
  locator: Locator,
  timeout: number = 10000
): Promise<void> {
  await locator.waitFor({ state: 'visible', timeout })
}

/**
 * Wait for element to be hidden
 */
export async function waitForHidden(
  page: Page,
  locator: Locator,
  timeout: number = 10000
): Promise<void> {
  await locator.waitFor({ state: 'hidden', timeout })
}

/**
 * Wait for network to be idle
 */
export async function waitForNetworkIdle(
  page: Page,
  timeout: number = 5000
): Promise<void> {
  await page.waitForLoadState('networkidle', { timeout })
}

/**
 * Wait for API call to complete
 */
export async function waitForAPIResponse(
  page: Page,
  urlPattern: string,
  timeout: number = 10000
): Promise<void> {
  await page.waitForResponse(urlPattern, { timeout })
}

/**
 * Custom wait for element containing specific text
 */
export async function waitForText(
  page: Page,
  text: string,
  timeout: number = 10000
): Promise<void> {
  await page.waitForFunction(
    (expectedText) => 
      document.body.textContent?.includes(expectedText),
    { timeout }
  )
}
```

#### Step 3.2: Replace waitForTimeout in Each File

**Before:**
```typescript
await page.goto('/chatbot')
await page.waitForTimeout(2000)  // Wait for data to load
await expect(page.getByText("Chatbot 1")).toBeVisible()
```

**After:**
```typescript
await page.goto('/chatbot')
await page.waitForLoadState('networkidle')  // Wait for all API calls
await expect(page.getByTestId('chatbot-list')).toBeVisible()
```

**Replacement Strategy:**

| Original Wait | Replacement | Rationale |
|---------------|-------------|-----------|
| `waitForTimeout(2000)` after navigation | `waitForLoadState('networkidle')` | Waits for actual data, not arbitrary time |
| `waitForTimeout(1000)` after click | `waitForSelector('[data-testid=...]')` | Waits for specific element |
| `waitForTimeout(3000)` for animation | `waitForFunction()` with animation check | Waits for animation completion |
| `waitForTimeout(5000)` polling | `waitForResponse()` | Waits for API response |

#### Step 3.3: Files to Update

| File | waitForTimeout Count | Priority |
|------|---------------------|----------|
| `auth.spec.ts` | 5+ | HIGH |
| `chatbot.spec.ts` | 3+ | HIGH |
| `smoke.spec.ts` | 2+ | HIGH |
| `widget-*.spec.ts` | 4+ | MEDIUM |
| `mobile-responsiveness.spec.ts` | 2+ | MEDIUM |
| `chunk-inspector.spec.ts` | 3+ | MEDIUM |

---

### Phase 4: Mock Infrastructure Improvement (Days 14-18)

#### Step 4.1: Create Test Data Factories

```typescript
// factories/test-data.ts
import { faker } from '@faker-js/faker'

export interface UserData {
  email: string
  password: string
  fullName: string
  isVerified: boolean
  planCode: string
}

export interface ChatbotData {
  name: string
  description: string
  welcomeMessage: string
  systemPrompt: string
  temperature: number
  maxTokens: number
  languageCode: string
}

export interface MessageData {
  content: string
  isFromUser: boolean
  timestamp: Date
}

// User Factory
export const createUser = (overrides: Partial<UserData> = {}): UserData => ({
  email: faker.internet.email(),
  password: faker.internet.password({ length: 12 }),
  fullName: faker.person.fullName(),
  isVerified: true,
  planCode: 'free',
  ...overrides,
})

// Chatbot Factory
export const createChatbot = (overrides: Partial<ChatbotData> = {}): ChatbotData => ({
  name: faker.commerce.productName(),
  description: faker.lorem.sentence(),
  welcomeMessage: faker.hacker.phrase(),
  systemPrompt: faker.lorem.paragraph(),
  temperature: 0.7,
  maxTokens: 1000,
  languageCode: 'tr',
  ...overrides,
})

// Message Factory
export const createMessage = (overrides: Partial<MessageData> = {}): MessageData => ({
  content: faker.lorem.sentence(),
  isFromUser: faker.datatype.boolean(),
  timestamp: new Date(),
  ...overrides,
})

// Chatbot List Factory
export const createChatbotList = (count: number = 5): ChatbotData[] => {
  return Array.from({ length: count }, () => createChatbot())
}
```

#### Step 4.2: Create Mock API Handlers

```typescript
// mocks/api-handlers.ts
import { UserData, ChatbotData, MessageData } from '../factories/test-data'

export interface MockConfig {
  user?: UserData
  chatbots?: ChatbotData[]
  messages?: MessageData[]
  delay?: number
  shouldFail?: boolean
  errorMessage?: string
}

// Default mock configuration
const defaultConfig: MockConfig = {
  user: {
    email: 'test@example.com',
    password: 'password123',
    fullName: 'Test User',
    isVerified: true,
    planCode: 'free',
  },
  chatbots: [],
  messages: [],
  delay: 100,
  shouldFail: false,
  errorMessage: 'An error occurred',
}

// Login Handler
export const createLoginHandler = (config: MockConfig = defaultConfig) => ({
  method: 'POST',
  path: '/api/v1/auth/login',
  handler: async (request, response, context) => {
    const { email, password } = await request.body()
    
    if (config.shouldFail) {
      return response.status(401).send({
        error: 'Unauthorized',
        message: config.errorMessage,
      })
    }
    
    return response.send({
      token: 'mock-jwt-token',
      user: config.user,
    })
  },
})

// Get Chatbots Handler
export const createGetChatbotsHandler = (config: MockConfig = defaultConfig) => ({
  method: 'GET',
  path: '/api/v1/chatbots',
  handler: async (request, response, context) => {
    if (config.delay) {
      await new Promise(resolve => setTimeout(resolve, config.delay))
    }
    
    return response.send({
      data: config.chatbots || [],
      total: config.chatbots?.length || 0,
    })
  },
})

// Create Chatbot Handler
export const createCreateChatbotHandler = (config: MockConfig = defaultConfig) => ({
  method: 'POST',
  path: '/api/v1/chatbots',
  handler: async (request, response, context) => {
    const body = await request.body()
    const newChatbot = {
      id: `bot-${Date.now()}`,
      ...body,
      createdAt: new Date().toISOString(),
    }
    
    return response.status(201).send(newChatbot)
  },
})

// Export all handlers
export const createAllHandlers = (config: MockConfig = defaultConfig) => [
  createLoginHandler(config),
  createGetChatbotsHandler(config),
  createCreateChatbotHandler(config),
]
```

#### Step 4.3: Create Mock Setup Utilities

```typescript
// e2e/mock-setup.ts
import { Page, Request, Response } from '@playwright/test'
import { MockConfig, createAllHandlers } from '../mocks/api-handlers'
import { UserData, ChatbotData } from '../factories/test-data'

export interface MockSetup {
  setupUser: (user: UserData) => void
  setupChatbots: (chatbots: ChatbotData[]) => void
  setupDelay: (ms: number) => void
  setupFailure: (endpoint: string, error: string) => void
  resetMocks: () => void
}

export function createMockSetup(page: Page): MockSetup {
  const handlers: any[] = []
  
  return {
    setupUser: (user: UserData) => {
      handlers.push({
        method: 'POST',
        path: '/api/v1/auth/login',
        handler: (request, response) => {
          response.send({
            token: 'mock-jwt-token',
            user: user,
          })
        },
      })
    },
    
    setupChatbots: (chatbots: ChatbotData[]) => {
      handlers.push({
        method: 'GET',
        path: '/api/v1/chatbots',
        handler: (request, response) => {
          response.send({
            data: chatbots,
            total: chatbots.length,
          })
        },
      })
    },
    
    setupDelay: (ms: number) => {
      // Add delay handler for all requests
      page.route('**', (route) => {
        setTimeout(() => route.continue(), ms)
      })
    },
    
    setupFailure: (endpoint: string, error: string) => {
      page.route(endpoint, (route) => {
        route.fulfill({
          status: 500,
          body: JSON.stringify({ error, message: error }),
        })
      })
    },
    
    resetMocks: () => {
      handlers.length = 0
      page.unrouteAll()
    },
  }
}

// Convenience function for full mock setup
export async function setupAllMocks(
  page: Page,
  config?: MockConfig
): Promise<void> {
  const mocks = createMockSetup(page)
  
  // Setup default mocks
  if (config?.user) {
    mocks.setupUser(config.user)
  }
  
  if (config?.chatbots) {
    mocks.setupChatbots(config.chatbots)
  }
  
  if (config?.delay) {
    mocks.setupDelay(config.delay)
  }
}
```

---

### Phase 5: Verification (Days 19-21)

#### Step 5.1: Run All Tests
```bash
# Run E2E tests
npm run e2e

# Run unit tests
npm run test

# Check for any waitForTimeout remaining
grep -rn "waitForTimeout" --include="*.ts" e2e/
```

#### Step 5.2: Verify Selector Stability
```bash
# Check all testids are present
grep -rn "getByTestId" --include="*.ts" e2e/ | wc -l

# Verify no text-based selectors remain
grep -rn "getByText" --include="*.ts" e2e/ | wc -l
```

#### Step 5.3: Performance Comparison
```bash
# Time E2E tests before and after
time npm run e2e
```

---

## Progress Tracking

### Daily Checklist

- [ ] Add data-testid to N components
- [ ] Refactor N selectors to use getByTestId
- [ ] Replace N waitForTimeout calls
- [ ] Run tests to verify no regression
- [ ] Update documentation

### Milestone Reviews

| Milestone | Target Date | Deliverables | Status |
|-----------|-------------|--------------|--------|
| Phase 1 Complete | Day 5 | All components have data-testid | ⏳ |
| Phase 2 Complete | Day 10 | Page objects for all pages | ⏳ |
| Phase 3 Complete | Day 13 | No waitForTimeout | ⏳ |
| Phase 4 Complete | Day 18 | Mock factories complete | ⏳ |
| Phase 5 Complete | Day 21 | Verification complete | ⏳ |

---

## Success Criteria

### Functional Requirements
- [ ] All interactive elements have data-testid
- [ ] All E2E tests use getByTestId() selectors
- [ ] Zero waitForTimeout in test code
- [ ] All tests pass without regression
- [ ] Test execution time improved

### Non-Functional Requirements
- [ ] Tests are maintainable
- [ ] Selectors are stable across refactoring
- [ ] Mocks are reusable and documented
- [ ] Test code is readable and follows patterns

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| data-testid adds visual clutter | User experience | data-testid is invisible (data attribute) |
| Some components hard to identify | Coverage gap | Use semantic naming, document exceptions |
| Performance impact of selectors | Test speed | Use unique testids, avoid over-specification |
| Mock changes break tests | Test failure | Version mocks, add breaking change detection |

---

## File Templates

### data-testid Template for Components

```tsx
// ComponentTemplate.tsx
import React from 'react'

interface ComponentProps {
  // props
}

export const Component: React.FC<ComponentProps> = (props) => {
  return (
    <div 
      className="component-container" 
      data-testid="component-container"
    >
      <header 
        className="component-header"
        data-testid="component-header"
      >
        <h1 data-testid="component-title">Title</h1>
      </header>
      
      <main data-testid="component-content">
        <button
          data-testid="component-action-button"
          onClick={props.onAction}
        >
          Action
        </button>
      </main>
      
      {props.error && (
        <div 
          className="component-error"
          data-testid="component-error-message"
        >
          {props.error}
        </div>
      )}
    </div>
  )
}
```

### Page Object Template

```typescript
// pages/PageName.ts
import { Page, Locator, expect } from '@playwright/test'

export class PageName {
  private page: Page
  
  // Selectors
  private readonly title = this.page.getByTestId('page-title')
  private readonly submitButton = this.page.getByTestId('page-submit-button')
  private readonly errorMessage = this.page.getByTestId('page-error-message')
  
  constructor(page: Page) {
    this.page = page
  }
  
  // Actions
  async goto(): Promise<void> {
    await this.page.goto('/page-url')
  }
  
  async fillForm(data: FormData): Promise<void> {
    // Implementation
  }
  
  async submit(): Promise<void> {
    await this.submitButton.click()
  }
  
  // Assertions
  async expectTitleVisible(): Promise<void> {
    await expect(this.title).toBeVisible()
  }
  
  async expectError(message: string): Promise<void> {
    await expect(this.errorMessage).toContainText(message)
  }
}
```

---

## Dependencies

- `@playwright/test` - E2E testing framework
- `@testing-library/react` - React component testing
- `faker-js/faker` - Test data generation
- Frontend component files
- E2E test files

---

## References

- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Testing Library Guiding Principles](https://testing-library.com/docs/guiding-principles)
- [Page Object Pattern](https://playwright.dev/docs/pom)

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Sisyphus | Initial plan |

---

*This plan is part of the comprehensive test improvement initiative. For questions or clarifications, refer to the project documentation or consult with the team lead.*
