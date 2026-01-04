# Task: Implement Chatbot Detail Page Tests

> **Task ID**: 11-chatbots-detail  
> **Source**: TEST_PATHS.md Section 4.3  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 8-10 hours  
> **Prerequisite**: 09-chatbots-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Chatbot Detail Page. This task covers tab navigation, overview content, and quick actions.

### Context

The Chatbot Detail Page displays detailed information about a single chatbot with various tabs for different functionalities. Testing this page ensures:
- Tab navigation works correctly
- Overview content displays properly
- Quick actions are accessible
- Breadcrumb navigation is correct
- Status indicators are accurate

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 4.3:

#### 4.3.1 Tab Navigation Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `tab-overview` | tab | Overview tab |
| `tab-settings` | tab | Settings tab |
| `tab-sources` | tab | Sources tab |
| `tab-actions` | tab | Actions tab |
| `tab-playground` | tab | Chat playground |
| `tab-deploy` | tab | Deployment tab |
| `tab-insights` | tab | Analytics tab |

#### 4.3.2 Tab Navigation Flow

```
Chatbot Detail Tabs
├── Load chatbot detail
│   ├── Assert: Active tab = Overview
│   ├── Assert: Sidebar highlights chatbot
│   └── Assert: Breadcrumb correct
│
├── Switch to Settings
│   ├── Click: tab-settings
│   ├── Assert: URL = /dashboard/chatbots/{id}/settings
│   ├── Assert: Tab active = Settings
│   └── Assert: Settings panel loads
│
├── Switch to Sources
│   ├── Click: tab-sources
│   ├── Assert: URL = /dashboard/chatbots/{id}/sources
│   ├── Assert: Tab active = Sources
│   └── Assert: Sources list loads
│
├── Switch to Actions
│   ├── Click: tab-actions
│   ├── Assert: URL = /dashboard/chatbots/{id}/actions
│   ├── Assert: Tab active = Actions
│   └── Assert: Actions list loads
│
├── Switch to Playground
│   ├── Click: tab-playground
│   ├── Assert: URL = /dashboard/chatbots/{id}/playground
│   ├── Assert: Tab active = Playground
│   └── Assert: Chat interface loads
│
├── Switch to Deploy
│   ├── Click: tab-deploy
│   ├── Assert: URL = /dashboard/chatbots/{id}/deploy
│   ├── Assert: Tab active = Deploy
│   └── Assert: Embed code panel loads
│
├── Switch to Insights
│   ├── Click: tab-insights
│   ├── Assert: URL = /dashboard/chatbots/{id}/insights
│   ├── Assert: Tab active = Insights
│   └── Assert: Analytics dashboard loads
│
└── Keyboard navigation
    ├── Arrow Left: Previous tab
    ├── Arrow Right: Next tab
    └── Enter: Activate focused tab
```

#### 4.3.3 Overview Tab

```
Overview Tab Flow
├── Overview content
│   ├── Show: Chatbot name
│   ├── Show: Model badge
│   ├── Show: Status indicator
│   ├── Show: Description
│   ├── Show: Created/Updated dates
│   └── Show: Quick stats (sources, messages)
│
├── Quick actions
│   ├── Click: btn-edit-settings
│   │   └── Navigate: Settings tab
│   │
│   ├── Click: btn-add-sources
│   │   └── Navigate: Sources tab
│   │
│   └── Click: btn-open-playground
│       └── Navigate: Playground tab
│
└── Status indicators
    ├── Green: Ready (sources > 0, no errors)
    ├── Yellow: Training (sources processing)
    └── Red: Error (check sources)
```

### Implementation Requirements

1. **Create Chatbot Detail Test File** (`frontend/e2e/chatbot-detail.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Chatbot Detail Page Object** (`frontend/e2e/pages/chatbot-detail.page.ts`)
   - Encapsulate tab navigation
   - Overview content assertions
   - Quick action methods

3. **Create Chatbot Detail Mocks** (`frontend/e2e/mocks/chatbot-detail.mocks.ts`)
   - Mock chatbot detail endpoint
   - Mock tab content endpoints

### Expected Deliverables

1. `frontend/e2e/chatbot-detail.spec.ts` - Comprehensive detail tests
2. `frontend/e2e/pages/chatbot-detail.page.ts` - Detail page object
3. `frontend/e2e/mocks/chatbot-detail.mocks.ts` - Detail API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/chatbot-detail.page.ts`:
  - Tab container locator
  - Tab button locators
  - Overview content locators
  - Quick action button locators
  - Status indicator locators
  - Stats display locators

### Phase 2: Page Load Tests

- [ ] Test: Page loads successfully
- [ ] Test: URL matches chatbot ID
- [ ] Test: Sidebar highlights chatbot
- [ ] Test: Breadcrumb shows correct path
- [ ] Test: Overview tab is active by default
- [ ] Test: All tabs visible

### Phase 3: Tab Navigation Tests

- [ ] Test: Click Settings tab activates
- [ ] Test: URL changes to settings path
- [ ] Test: Click Sources tab activates
- [ ] Test: URL changes to sources path
- [ ] Test: Click Actions tab activates
- [ ] Test: URL changes to actions path
- [ ] Test: Click Playground tab activates
- [ ] Test: URL changes to playground path
- [ ] Test: Click Deploy tab activates
- [ ] Test: URL changes to deploy path
- [ ] Test: Click Insights tab activates
- [ ] Test: URL changes to insights path

### Phase 4: Keyboard Navigation Tests

- [ ] Test: Arrow Left moves to previous tab
- [ ] Test: Arrow Right moves to next tab
- [ ] Test: First tab -> Arrow Left goes to last
- [ ] Test: Last tab -> Arrow Right goes to first
- [ ] Test: Enter activates focused tab
- [ ] Test: Focus indicator visible
- [ ] Test: Tab order is logical

### Phase 5: Overview Content Tests

- [ ] Test: Chatbot name displayed
- [ ] Test: Model badge displayed
- [ ] Test: Status indicator displayed
- [ ] Test: Description displayed
- [ ] Test: Created date displayed
- [ ] Test: Updated date displayed
- [ ] Test: Source count displayed
- [ ] Test: Message count displayed

### Phase 6: Status Indicator Tests

- [ ] Test: Green status for ready chatbot
- [ ] Test: Yellow status for training chatbot
- [ ] Test: Red status for error chatbot
- [ ] Test: Status tooltip on hover
- [ ] Test: Status changes on data refresh

### Phase 7: Quick Actions Tests

- [ ] Test: Edit settings button visible
- [ ] Test: Click edit settings goes to settings
- [ ] Test: Add sources button visible
- [ ] Test: Click add sources goes to sources
- [ ] Test: Open playground button visible
- [ ] Test: Click open playground goes to playground
- [ ] Test: Button hover states
- [ ] Test: Button disabled states (if applicable)

### Phase 8: Tab Content Tests (Overview)

- [ ] Test: Overview has title
- [ ] Test: Overview has description section
- [ ] Test: Overview has stats grid
- [ ] Test: Overview has actions section
- [ ] Test: Overview loads without error
- [ ] Test: Overview data is correct

---

## Technical Notes

### Chatbot Detail Page Object

```typescript
// frontend/e2e/pages/chatbot-detail.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ChatbotDetailPage {
  readonly page: Page;
  readonly tabContainer: Locator;
  readonly tabOverview: Locator;
  readonly tabSettings: Locator;
  readonly tabSources: Locator;
  readonly tabActions: Locator;
  readonly tabPlayground: Locator;
  readonly tabDeploy: Locator;
  readonly tabInsights: Locator;
  
  // Overview content
  readonly chatbotName: Locator;
  readonly modelBadge: Locator;
  readonly statusIndicator: Locator;
  readonly description: Locator;
  readonly createdDate: Locator;
  readonly updatedDate: Locator;
  readonly sourceCount: Locator;
  readonly messageCount: Locator;
  
  // Quick actions
  readonly btnEditSettings: Locator;
  readonly btnAddSources: Locator;
  readonly btnOpenPlayground: Locator;

  constructor(page: Page) {
    this.page = page;
    this.tabContainer = page.locator('[data-testid="chatbot-tabs"]');
    this.tabOverview = page.locator('[data-testid="tab-overview"]');
    this.tabSettings = page.locator('[data-testid="tab-settings"]');
    this.tabSources = page.locator('[data-testid="tab-sources"]');
    this.tabActions = page.locator('[data-testid="tab-actions"]');
    this.tabPlayground = page.locator('[data-testid="tab-playground"]');
    this.tabDeploy = page.locator('[data-testid="tab-deploy"]');
    this.tabInsights = page.locator('[data-testid="tab-insights"]');
    
    this.chatbotName = page.locator('[data-testid="chatbot-name"]');
    this.modelBadge = page.locator('[data-testid="model-badge"]');
    this.statusIndicator = page.locator('[data-testid="status-indicator"]');
    this.description = page.locator('[data-testid="chatbot-description"]');
    this.createdDate = page.locator('[data-testid="created-date"]');
    this.updatedDate = page.locator('[data-testid="updated-date"]');
    this.sourceCount = page.locator('[data-testid="source-count"]');
    this.messageCount = page.locator('[data-testid="message-count"]');
    
    this.btnEditSettings = page.locator('[data-testid="btn-edit-settings"]');
    this.btnAddSources = page.locator('[data-testid="btn-add-sources"]');
    this.btnOpenPlayground = page.locator('[data-testid="btn-open-playground"]');
  }

  async goto(chatbotId: string) {
    await this.page.goto(`/dashboard/chatbots/${chatbotId}`);
  }

  async expectVisible() {
    await expect(this.tabContainer).toBeVisible();
  }

  async expectActiveTab(tabName: string) {
    await expect(this.page.locator(`[data-testid="tab-${tabName}"]`)).toHaveClass(/active/);
  }

  async clickTab(tabName: string) {
    await this.page.locator(`[data-testid="tab-${tabName}"]`).click();
  }

  async expectUrlContains(path: string) {
    await expect(this.page).toHaveURL(new RegExp(path));
  }

  async clickEditSettings() {
    await this.btnEditSettings.click();
  }

  async clickAddSources() {
    await this.btnAddSources.click();
  }

  async clickOpenPlayground() {
    await this.btnOpenPlayground.click();
  }

  async expectChatbotName(name: string) {
    await expect(this.chatbotName).toHaveText(name);
  }

  async expectModel(model: string) {
    await expect(this.modelBadge).toHaveText(model);
  }

  async expectStatus(status: 'ready' | 'training' | 'error') {
    await expect(this.statusIndicator).toHaveClass(new RegExp(status));
  }

  async expectSourceCount(count: number) {
    await expect(this.sourceCount).toHaveText(count.toString());
  }

  async expectMessageCount(count: number) {
    await expect(this.messageCount).toHaveText(count.toString());
  }

  async expectDescription(description: string) {
    await expect(this.description).toHaveText(description);
  }

  // Keyboard navigation
  async pressTabRight() {
    await this.page.keyboard.press('ArrowRight');
  }

  async pressTabLeft() {
    await this.page.keyboard.press('ArrowLeft');
  }

  async pressEnter() {
    await this.page.keyboard.press('Enter');
  }
}
```

### Detail API Mocks

```typescript
// frontend/e2e/mocks/chatbot-detail.mocks.ts
import { APIRequestContext } from '@playwright/test';

export const mockChatbotDetail = {
  id: 'chatbot-123',
  name: 'Customer Support Bot',
  description: 'A chatbot for handling customer inquiries',
  model: 'gpt-4o-mini',
  status: 'ready',
  language: 'tr',
  temperature: 0.7,
  maxTokens: 1000,
  sourceCount: 5,
  messageCount: 150,
  createdAt: '2024-01-15T10:00:00Z',
  updatedAt: '2024-01-20T14:30:00Z',
};

export async function mockGetChatbotDetail(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}`, {
    status: 200,
    body: mockChatbotDetail,
  });
}

export async function mockGetChatbotDetailTraining(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}`, {
    status: 200,
    body: {
      ...mockChatbotDetail,
      status: 'training',
      sourceCount: 2,
    },
  });
}

export async function mockGetChatbotDetailError(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}`, {
    status: 200,
    body: {
      ...mockChatbotDetail,
      status: 'error',
    },
  });
}

export async function mockChatbotNotFound(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}`, {
    status: 404,
    body: {
      error: 'NOT_FOUND',
      message: 'Chatbot not found',
    },
  });
}
```

### Tab Test Data

```typescript
export const tabNavigationTests = [
  { tab: 'overview', path: /\/dashboard\/chatbots\/[^\/]+$/ },
  { tab: 'settings', path: /\/settings$/ },
  { tab: 'sources', path: /\/sources$/ },
  { tab: 'actions', path: /\/actions$/ },
  { tab: 'playground', path: /\/playground$/ },
  { tab: 'deploy', path: /\/deploy$/ },
  { tab: 'insights', path: /\/insights$/ },
];

export const statusIndicators = [
  { status: 'ready', class: 'status-ready', color: 'green' },
  { status: 'training', class: 'status-training', color: 'yellow' },
  { status: 'error', class: 'status-error', color: 'red' },
];
```

### Running Specific Tests

```bash
# Run all chatbot detail tests
cd frontend && npx playwright test chatbot-detail.spec.ts

# Run tab navigation tests
cd frontend && npx playwright test chatbot-detail.spec.ts -g "tab"

# Run overview tests
cd frontend && npx playwright test chatbot-detail.spec.ts -g "overview"

# Run status indicator tests
cd frontend && npx playwright test chatbot-detail.spec.ts -g "status"

# Run in headed mode
cd frontend && npx playwright test chatbot-detail.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All tabs tested
- [ ] All tab navigation paths tested
- [ ] All overview content tested
- [ ] All status indicators tested
- [ ] All quick actions tested
- [ ] Keyboard navigation tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Active tab clearly visible
- [ ] Smooth transitions
- [ ] Loading states present
- [ ] Clear status indicators

### 4. Navigation Verification
- [ ] URL updates correctly
- [ ] Breadcrumb correct
- [ ] Sidebar highlights correctly

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Tab State** - Ensure tabs maintain proper active state
2. **URL Updates** - Verify URL changes match tab clicks
3. **Breadcrumb** - Check breadcrumb updates on navigation
4. **Mock Data** - Use different chatbot states for testing

### Common Issues to Avoid

1. **Race conditions** - Wait for tab content to load
2. **URL patterns** - Use regex for URL matching
3. **Hardcoded IDs** - Use dynamic chatbot IDs
4. **Skipping keyboard tests** - Test keyboard navigation

### Test Setup

```typescript
// Use authenticated state
test.use({
  storageState: 'e2e/.auth/user.json',
});

// Test with different chatbot statuses
test('shows correct status indicator', async ({ page }) => {
  await page.route('/api/v1/chatbots/chatbot-123', route => {
    route.fulfill({ body: JSON.stringify(mockChatbotDetailTraining) });
  });
  
  await page.goto('/dashboard/chatbots/chatbot-123');
  await expect(page.locator('[data-testid="status-indicator"]')).toHaveClass(/training/);
});
```

---

## Dependencies

- **Prerequisites**: 09-chatbots-list.md (navigation to detail)
- **Environment**: Backend API with chatbot detail endpoint
- **Test Data**: Chatbot with various statuses

---

## Related Tasks

- 09-chatbots-list.md - Click card to navigate to detail
- 10-chatbots-create.md - Creates new chatbot
- 12-chatbots-settings.md - Settings tab content
- 13-sources-list.md - Sources tab content
- 19-playground.md - Playground tab content

---

*Task created from: docs/frontend/TEST_PATHS.md Section 4.3*
