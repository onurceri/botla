# Task: Implement Dashboard Search Tests

> **Task ID**: 07-dashboard-search  
> **Source**: TEST_PATHS.md Section 3.2  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 06-dashboard-layout.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Dashboard Search functionality. This task covers global search, search results display, keyboard navigation, and search result interactions.

### Context

The Dashboard Search provides global search capability across the application. Testing this functionality ensures:
- Users can search for chatbots, sources, and other resources
- Search results are displayed with relevant information
- Keyboard navigation works correctly
- Debouncing prevents excessive API calls
- Clear/no results are handled properly

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 3.2:

#### Search Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `search-input` | text | Global search input |
| `search-results` | dropdown | Search results dropdown |
| `search-result-item` | list-item | Individual result |
| `btn-search-clear` | button | Clear search |
| `btn-search-submit` | button | Submit search |

#### Search Flow

```
Search Flow
├── Click: search-input
│   ├── Assert: Focus state
│   └── Type: "my chatbot"
│
├── While typing
│   ├── Debounce: 300ms
│   ├── Show: `search-results` dropdown
│   ├── Show: Loading spinner
│   ├── Hide: Results if < 2 chars
│   └── Show: No results if 0 matches
│
├── Search results displayed
│   ├── Show: Up to 5 results
│   ├── Each result shows:
│   │   ├── Icon (chatbot, source, etc.)
│   │   ├── Title
│   │   └── Description
│   │
│   ├── Hover: Result item (highlight)
│   │   ├── Background color change
│   │   └── Cursor pointer
│   │
│   ├── Click: Result item
│   │   ├── Navigate: Result URL
│   │   └── Close: Search dropdown
│   │
│   └── Click: View all results
│       └── Navigate: Search results page
│
├── Clear search
│   ├── Click: btn-search-clear
│   ├── Assert: Input cleared
│   ├── Assert: Dropdown closed
│   └── Assert: Placeholder visible
│
└── Keyboard navigation
    ├── Arrow Down: Navigate results
    ├── Arrow Up: Navigate results
    ├── Enter: Open selected result
    └── Escape: Close dropdown
```

### Implementation Requirements

1. **Create Search Test File** (`frontend/e2e/search.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Search Page Object** (`frontend/e2e/pages/search.page.ts`)
   - Encapsulate search interactions
   - Results handling methods
   - Keyboard navigation support

3. **Create Search Mocks** (`frontend/e2e/mocks/search.mocks.ts`)
   - Mock search API responses
   - Mock search result data
   - Handle debouncing behavior

### Expected Deliverables

1. `frontend/e2e/search.spec.ts` - Comprehensive search tests
2. `frontend/e2e/pages/search.page.ts` - Search page object
3. `frontend/e2e/mocks/search.mocks.ts` - Search API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/search.page.ts`:
  - Search input locator
  - Results dropdown locator
  - Result item locators
  - Clear button locator
  - Loading spinner locator
- [ ] Create `frontend/e2e/mocks/search.mocks.ts`:
  - Mock search endpoint
  - Mock various result types
  - Mock empty results

### Phase 2: Basic Search Tests

- [ ] Test: Search input visible
- [ ] Test: Click search input shows focus state
- [ ] Test: Type in search input
- [ ] Test: Input accepts text
- [ ] Test: Input clears correctly
- [ ] Test: Placeholder text visible

### Phase 3: Search Results Tests

- [ ] Test: Results dropdown appears after typing
- [ ] Test: Results show after debounce (300ms)
- [ ] Test: Loading spinner visible during search
- [ ] Test: Results hidden when < 2 characters
- [ ] Test: "No results" message when 0 matches
- [ ] Test: Results display up to 5 items
- [ ] Test: Each result shows icon
- [ ] Test: Each result shows title
- [ ] Test: Each result shows description

### Phase 4: Result Interaction Tests

- [ ] Test: Hover highlights result item
- [ ] Test: Cursor changes to pointer
- [ ] Test: Click result navigates to result URL
- [ ] Test: Dropdown closes after click
- [ ] Test: "View all results" link present
- [ ] Test: Click "View all results" navigates to search page
- [ ] Test: Multiple result types (chatbots, sources, etc.)

### Phase 5: Clear Search Tests

- [ ] Test: Clear button visible when text entered
- [ ] Test: Click clear button clears input
- [ ] Test: Click clear button closes dropdown
- [ ] Test: Clear button hidden when empty
- [ ] Test: Escape key clears and closes

### Phase 6: Keyboard Navigation Tests

- [ ] Test: Arrow Down navigates to first result
- [ ] Test: Arrow Down from result 1 goes to result 2
- [ ] Test: Arrow Up navigates backward
- [ ] Test: Arrow Up from first wraps to last
- [ ] Test: Enter opens selected result
- [ ] Test: Escape closes dropdown
- [ ] Test: Tab moves focus out
- [ ] Test: Focus indicator visible

### Phase 7: Debouncing Tests

- [ ] Test: Search not triggered immediately on type
- [ ] Test: Search triggered after 300ms
- [ ] Test: Rapid typing debounces correctly
- [ ] Test: Only one search request for rapid input
- [ ] Test: New search cancels pending request

### Phase 8: Edge Cases Tests

- [ ] Test: Special characters in search
- [ ] Test: Very long search query
- [ ] Test: Search with emoji
- [ ] Test: Network error during search
- [ ] Test: Empty results state
- [ ] Test: Maximum results displayed

---

## Technical Notes

### Search Page Object

```typescript
// frontend/e2e/pages/search.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class SearchPage {
  readonly page: Page;
  readonly input: Locator;
  readonly resultsDropdown: Locator;
  readonly resultItems: Locator;
  readonly clearButton: Locator;
  readonly loadingSpinner: Locator;
  readonly noResultsMessage: Locator;
  readonly viewAllLink: Locator;

  constructor(page: Page) {
    this.page = page;
    this.input = page.locator('[data-testid="search-input"]');
    this.resultsDropdown = page.locator('[data-testid="search-results"]');
    this.resultItems = page.locator('[data-testid="search-result-item"]');
    this.clearButton = page.locator('[data-testid="btn-search-clear"]');
    this.loadingSpinner = page.locator('[data-testid="search-loading"]');
    this.noResultsMessage = page.locator('[data-testid="search-no-results"]');
    this.viewAllLink = page.locator('[data-testid="search-view-all"]');
  }

  async focus() {
    await this.input.click();
  }

  async expectFocused() {
    await expect(this.input).toBeFocused();
  }

  async type(query: string) {
    await this.input.fill(query);
  }

  async clear() {
    await this.clearButton.click();
  }

  async pressKey(key: 'ArrowDown' | 'ArrowUp' | 'Enter' | 'Escape' | 'Tab') {
    await this.input.press(key);
  }

  async expectResultsVisible() {
    await expect(this.resultsDropdown).toBeVisible();
  }

  async expectResultsHidden() {
    await expect(this.resultsDropdown).toBeHidden();
  }

  async expectResultCount(count: number) {
    await expect(this.resultItems).toHaveCount(count);
  }

  async expectNoResults() {
    await expect(this.noResultsMessage).toBeVisible();
    await expect(this.resultItems).toHaveCount(0);
  }

  async expectLoading() {
    await expect(this.loadingSpinner).toBeVisible();
  }

  async expectNotLoading() {
    await expect(this.loadingSpinner).toBeHidden();
  }

  async clickResult(index: number) {
    await this.resultItems.nth(index).click();
  }

  async clickViewAll() {
    await this.viewAllLink.click();
  }

  async expectClearButtonVisible() {
    await expect(this.clearButton).toBeVisible();
  }

  async expectClearButtonHidden() {
    await expect(this.clearButton).toBeHidden();
  }

  // Keyboard navigation helpers
  async navigateResults(direction: 'down' | 'up', count: number) {
    for (let i = 0; i < count; i++) {
      await this.pressKey(direction === 'down' ? 'ArrowDown' : 'ArrowUp');
    }
  }
}
```

### Search Mocks

```typescript
// frontend/e2e/mocks/search.mocks.ts
import { APIRequestContext } from '@playwright/test';

// Mock search results data
export const mockSearchResults = {
  chatbots: [
    {
      id: 'chatbot-1',
      type: 'chatbot',
      title: 'Customer Support Bot',
      description: 'Handles customer inquiries and support tickets',
      icon: 'bot-icon',
      url: '/dashboard/chatbots/chatbot-1',
    },
    {
      id: 'chatbot-2',
      type: 'chatbot',
      title: 'Sales Assistant',
      description: 'Helps with product inquiries and pricing',
      icon: 'bot-icon',
      url: '/dashboard/chatbots/chatbot-2',
    },
  ],
  sources: [
    {
      id: 'source-1',
      type: 'source',
      title: 'Documentation',
      description: 'Product documentation and guides',
      icon: 'document-icon',
      url: '/dashboard/chatbots/chatbot-1/sources/source-1',
    },
  ],
};

export async function mockSuccessfulSearch(request: APIRequestContext, query: string = '') {
  const results = query.toLowerCase().includes('support')
    ? mockSearchResults.chatbots
    : mockSearchResults.chatbots.filter(c => 
        c.title.toLowerCase().includes(query.toLowerCase())
      );

  await request.get('/api/v1/search', {
    status: 200,
    body: {
      query,
      results,
      total: results.length,
      limit: 5,
    },
  });
}

export async function mockEmptySearch(request: APIRequestContext) {
  await request.get('/api/v1/search', {
    status: 200,
    body: {
      query: '',
      results: [],
      total: 0,
      limit: 5,
    },
  });
}

export async function mockSearchError(request: APIRequestContext) {
  await request.get('/api/v1/search', {
    status: 500,
    body: {
      error: 'SEARCH_FAILED',
      message: 'Search service temporarily unavailable',
    },
  });
}
```

### Debounce Testing

```typescript
// Helper for testing debounce behavior
export async function waitForDebounce(timeout: number = 350) {
  await new Promise(resolve => setTimeout(resolve, timeout));
}

export async function typeWithDebounce(page: Page, input: Locator, text: string, debounceMs: number = 300) {
  await input.fill(text);
  await page.waitForTimeout(debounceMs + 50); // Wait for debounce + buffer
}
```

### Running Specific Tests

```bash
# Run all search tests
cd frontend && npx playwright test search.spec.ts

# Run basic search tests
cd frontend && npx playwright test search.spec.ts -g "basic"

# Run keyboard navigation tests
cd frontend && npx playwright test search.spec.ts -g "keyboard"

# Run debounce tests
cd frontend && npx playwright test search.spec.ts -g "debounce"

# Run in headed mode
cd frontend && npx playwright test search.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All search interactions tested
- [ ] All keyboard navigation tested
- [ ] Debouncing behavior tested
- [ ] All result types tested
- [ ] Edge cases covered
- [ ] Error handling tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with API mocking
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Loading states visible
- [ ] Clear feedback for no results
- [ ] Smooth dropdown animation
- [ ] Clear hover states

### 4. Accessibility Verification
- [ ] Keyboard navigation works
- [ ] Focus states visible
- [ ] ARIA labels present
- [ ] Screen reader compatible

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Debouncing** - Use proper timing to test debounce behavior
2. **API Mocking** - Mock search endpoint for consistent results
3. **Timing** - Be careful with race conditions in async tests
4. **Keyboard** - Test all keyboard interactions thoroughly

### Common Issues to Avoid

1. **Not waiting for debounce** - Always wait for debounce in tests
2. **Race conditions** - Wait for dropdown before interacting
3. **Hardcoded results** - Use mock data for consistency
4. **Missing focus states** - Test focus indicators

### Test Data Setup

```typescript
// Mock search results fixture
const mockChatbotResults = [
  { id: '1', title: 'Support Bot', type: 'chatbot' },
  { id: '2', title: 'Sales Bot', type: 'chatbot' },
];

// Use in tests
test('search shows chatbot results', async ({ page }) => {
  await page.route('/api/v1/search', route => {
    route.fulfill({ body: { results: mockChatbotResults } });
  });
  // Test implementation
});
```

---

## Dependencies

- **Prerequisites**: 06-dashboard-layout.md (for search input location)
- **Environment**: Backend search API
- **Test Data**: Various search terms and results

---

## Related Tasks

- 06-dashboard-layout.md - Dashboard layout (search input location)
- 08-dashboard-toast.md - Toast notifications (error display)
- 09-chatbots-list.md - Chatbots list (search functionality)

---

*Task created from: docs/frontend/TEST_PATHS.md Section 3.2*
