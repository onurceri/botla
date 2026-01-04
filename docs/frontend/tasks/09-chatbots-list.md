# Task: Implement Chatbots List Tests

> **Task ID**: 09-chatbots-list  
> **Source**: TEST_PATHS.md Section 4.1  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 10-12 hours  
> **Prerequisite**: 06-dashboard-layout.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Chatbots List Page. This task covers chatbot listing, card interactions, search, filter, pagination, and empty states.

### Context

The Chatbots List Page displays all chatbots in a grid format with various interactions. Testing this page ensures:
- Users can view all their chatbots
- Search and filtering work correctly
- Pagination handles large datasets
- Card interactions (click, hover, actions) work
- Empty states are handled gracefully

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 4.1:

#### 4.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `btn-create-chatbot` | button | Create new chatbot |
| `input-search` | text | Search chatbots |
| `select-sort` | select | Sort order |
| `select-filter` | select | Filter by status |
| `card-chatbot` | card | Chatbot card |
| `card-chatbot-name` | text | Chatbot name |
| `card-chatbot-model` | badge | Model badge |
| `card-chatbot-status` | badge | Status indicator |
| `card-chatbot-actions` | menu | Actions dropdown |
| `pagination` | pagination | Pagination controls |
| `empty-state` | component | No chatbots state |

#### 4.1.2 Chatbot Card Interactions

```
Chatbot Card Flow
├── Card hover state
│   ├── Hover: Card
│   │   ├── Shadow increase
│   │   ├── Scale 1.02
│   │   └── Cursor pointer
│   │
│   └── Hover: Card actions button
│       └── Show: Tooltip "Actions"
│
├── Click: Card body (not actions)
│   ├── Navigate: /dashboard/chatbots/{id}
│   └── Open: Chatbot detail page
│
├── Click: Actions menu
│   ├── Open: `dropdown-chatbot-actions`
│   ├── Options:
│   │   ├── Edit
│   │   ├── Duplicate
│   │   ├── Share
│   │   ├── Settings
│   │   └── Delete
│   │
│   ├── Click: Edit
│   │   └── Navigate: /dashboard/chatbots/{id}/settings
│   │
│   ├── Click: Duplicate
│   │   ├── Open: `modal-duplicate`
│   │   ├── Show: New name input
│   │   ├── Click: btn-duplicate
│   │   └── Assert: New chatbot created
│   │
│   ├── Click: Share
│   │   ├── Open: `modal-share`
│   │   ├── Show: Share link
│   │   └── Click: btn-copy-link
│   │
│   ├── Click: Delete
│   │   ├── Open: `modal-delete-confirm`
│   │   ├── Show: "Delete chatbot?" warning
│   │   ├── Type: chatbot name to confirm
│   │   ├── Click: btn-delete
│   │   └── Assert: Chatbot deleted
│   │
│   └── Hover: Menu item
        └── Highlight background
```

#### 4.1.3 Search and Filter

```
Search Chatbots
├── Type: "support"
│   ├── Filter: Chatbots matching "support"
│   ├── Update: Card list
│   └── Show: Match count
│
├── Clear search
│   ├── Click: btn-clear
│   └── Reset: Full list
│
└── Sort options
    ├── Select: Name (A-Z)
    │   └── Sort: Alphabetical
    │
    ├── Select: Name (Z-A)
    │   └── Sort: Reverse alphabetical
    │
    ├── Select: Recently updated
    │   └── Sort: UpdatedAt DESC
    │
    └── Select: Oldest
        └── Sort: CreatedAt ASC

Filter by Status
├── Select: All
│   └── Show: All chatbots
│
├── Select: Active
│   └── Show: Only active chatbots
│
├── Select: Training
│   └── Show: Chatbots with sources training
│
└── Select: Error
    └── Show: Chatbots with errors
```

#### 4.1.4 Pagination

```
Pagination Flow
├── Assert: Pagination visible (if > items per page)
│
├── Items per page selector
│   ├── Select: 12
│   │   └── Update: itemsPerPage = 12
│   │
│   ├── Select: 24
│   │   └── Update: itemsPerPage = 24
│   │
│   └── Select: 48
│       └── Update: itemsPerPage = 48
│
├── Page navigation
│   ├── Click: btn-previous (when on page > 1)
│   │   └── Navigate: Previous page
│   │
│   ├── Click: btn-next (when more pages)
│   │   └── Navigate: Next page
│   │
│   ├── Click: Page number
│   │   └── Navigate: Specific page
│   │
│   └── Click: Ellipsis (...)
│       └── Show: Page range selector
│
└── Empty state
    ├── Assert: When no chatbots match filter
    ├── Show: Empty illustration
    ├── Show: "No chatbots found" text
    └── Show: btn-create-chatbot
```

### Implementation Requirements

1. **Create Chatbots List Test File** (`frontend/e2e/chatbots-list.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Chatbots List Page Object** (`frontend/e2e/pages/chatbots-list.page.ts`)
   - Encapsulate list page interactions
   - Card interaction methods
   - Search/filter/pagination methods

3. **Create Chatbot Card Page Object** (`frontend/e2e/pages/chatbot-card.page.ts`)
   - Individual card interactions
   - Actions menu methods
   - Status verification

4. **Create Chatbots Mocks** (`frontend/e2e/mocks/chatbots.mocks.ts`)
   - Mock chatbot list endpoint
   - Mock chatbot data
   - Mock actions (delete, duplicate, etc.)

### Expected Deliverables

1. `frontend/e2e/chatbots-list.spec.ts` - Comprehensive list tests
2. `frontend/e2e/pages/chatbots-list.page.ts` - List page object
3. `frontend/e2e/pages/chatbot-card.page.ts` - Card page object
4. `frontend/e2e/mocks/chatbots.mocks.ts` - Chatbot API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Objects

- [ ] Create `frontend/e2e/pages/chatbots-list.page.ts`:
  - Create button locator
  - Search input locator
  - Sort select locator
  - Filter select locator
  - Chatbot cards container
  - Pagination controls
  - Empty state locator
- [ ] Create `frontend/e2e/pages/chatbot-card.page.ts`:
  - Card container locator
  - Card name locator
  - Card model badge locator
  - Card status badge locator
  - Actions button locator
  - Actions dropdown locator
- [ ] Create `frontend/e2e/mocks/chatbots.mocks.ts`:
  - Mock get chatbots endpoint
  - Mock chatbot data (various statuses)
  - Mock delete chatbot
  - Mock duplicate chatbot

### Phase 2: Page Load Tests

- [ ] Test: Page loads successfully
- [ ] Test: URL is correct
- [ ] Test: Sidebar highlights chatbots
- [ ] Test: Create button visible
- [ ] Test: Search input visible
- [ ] Test: Sort select visible
- [ ] Test: Filter select visible
- [ ] Test: Chatbot cards visible (with data)
- [ ] Test: Pagination visible (with data)

### Phase 3: Chatbot Card Tests

- [ ] Test: Card displays chatbot name
- [ ] Test: Card displays model badge
- [ ] Test: Card displays status badge
- [ ] Test: Card hover state (shadow)
- [ ] Test: Card hover state (scale)
- [ ] Test: Card hover state (cursor)
- [ ] Test: Click card navigates to detail
- [ ] Test: Actions button visible
- [ ] Test: Hover actions button shows tooltip

### Phase 4: Card Actions Menu Tests

- [ ] Test: Click actions opens dropdown
- [ ] Test: Dropdown contains Edit option
- [ ] Test: Dropdown contains Duplicate option
- [ ] Test: Dropdown contains Share option
- [ ] Test: Dropdown contains Settings option
- [ ] Test: Dropdown contains Delete option
- [ ] Test: Hover menu item highlights
- [ ] Test: Click Edit navigates to settings
- [ ] Test: Click Duplicate opens modal

### Phase 5: Duplicate Chatbot Tests

- [ ] Test: Duplicate modal opens
- [ ] Test: Modal shows new name input
- [ ] Test: Default name is copy of original
- [ ] Test: Click duplicate creates new chatbot
- [ ] Test: Toast success appears
- [ ] Test: Click cancel closes modal
- [ ] Test: Modal closes on overlay click
- [ ] Test: Escape key closes modal

### Phase 6: Share Chatbot Tests

- [ ] Test: Share modal opens
- [ ] Test: Share link displayed
- [ ] Test: Copy link button works
- [ ] Test: Toast "Copied" appears
- [ ] Test: Modal closes properly

### Phase 7: Delete Chatbot Tests

- [ ] Test: Delete modal opens
- [ ] Test: Warning message displayed
- [ ] Test: Requires name confirmation
- [ ] Test: Wrong name shows error
- [ ] Test: Correct name confirms delete
- [ ] Test: API call made to delete
- [ ] Test: Toast success appears
- [ ] Test: Card removed from list

### Phase 8: Search Tests

- [ ] Test: Type filters chatbot list
- [ ] Test: Search is case-insensitive
- [ ] Test: Clear button visible when typing
- [ ] Test: Click clear resets list
- [ ] Test: No results shown when no match
- [ ] Test: Match count displayed (if implemented)
- [ ] Test: Debounce on typing
- [ ] Test: Special characters in search

### Phase 9: Sort Tests

- [ ] Test: Sort by Name (A-Z)
- [ ] Test: Sort by Name (Z-A)
- [ ] Test: Sort by Recently updated
- [ ] Test: Sort by Oldest
- [ ] Test: Sort indicator visible
- [ ] Test: Sort persists on refresh

### Phase 10: Filter Tests

- [ ] Test: Filter by All
- [ ] Test: Filter by Active
- [ ] Test: Filter by Training
- [ ] Test: Filter by Error
- [ ] Test: Filter changes results
- [ ] Test: Filter persists on refresh

### Phase 11: Pagination Tests

- [ ] Test: Pagination visible when needed
- [ ] Test: Items per page selector
- [ ] Test: Change items per page
- [ ] Test: Next button enabled on page 1
- [ ] Test: Previous button disabled on page 1
- [ ] Test: Click next goes to page 2
- [ ] Test: Click previous goes to page 1
- [ ] Test: Click page number
- [ ] Test: Ellipsis for large page counts
- [ ] Test: Current page indicator

### Phase 12: Empty State Tests

- [ ] Test: Empty state when no chatbots
- [ ] Test: Empty illustration visible
- [ ] Test: "No chatbots found" text
- [ ] Test: Create button visible
- [ ] Test: Click create opens modal
- [ ] Test: Empty with search filter
- [ ] Test: Empty with status filter

---

## Technical Notes

### Chatbots List Page Object

```typescript
// frontend/e2e/pages/chatbots-list.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ChatbotsListPage {
  readonly page: Page;
  readonly createButton: Locator;
  readonly searchInput: Locator;
  readonly sortSelect: Locator;
  readonly filterSelect: Locator;
  readonly cardsContainer: Locator;
  readonly chatbotCards: Locator;
  readonly pagination: Locator;
  readonly emptyState: Locator;
  readonly paginationInfo: Locator;

  constructor(page: Page) {
    this.page = page;
    this.createButton = page.locator('[data-testid="btn-create-chatbot"]');
    this.searchInput = page.locator('[data-testid="input-search"]');
    this.sortSelect = page.locator('[data-testid="select-sort"]');
    this.filterSelect = page.locator('[data-testid="select-filter"]');
    this.cardsContainer = page.locator('[data-testid="chatbots-cards-container"]');
    this.chatbotCards = page.locator('[data-testid="card-chatbot"]');
    this.pagination = page.locator('[data-testid="pagination"]');
    this.emptyState = page.locator('[data-testid="empty-state"]');
    this.paginationInfo = page.locator('[data-testid="pagination-info"]');
  }

  async goto() {
    await this.page.goto('/dashboard/chatbots');
  }

  async expectVisible() {
    await expect(this.createButton).toBeVisible();
    await expect(this.searchInput).toBeVisible();
  }

  async expectChatbotCount(count: number) {
    await expect(this.chatbotCards).toHaveCount(count);
  }

  async expectEmptyState() {
    await expect(this.emptyState).toBeVisible();
    await expect(this.chatbotCards).toHaveCount(0);
  }

  async search(query: string) {
    await this.searchInput.fill(query);
    await this.page.waitForTimeout(350); // Debounce
  }

  async clearSearch() {
    await this.searchInput.clear();
    await this.page.waitForTimeout(350);
  }

  async selectSort(option: string) {
    await this.sortSelect.selectOption(option);
  }

  async selectFilter(option: string) {
    await this.filterSelect.selectOption(option);
  }

  async clickCreateButton() {
    await this.createButton.click();
  }

  async clickChatbotCard(index: number) {
    await this.chatbotCards.nth(index).click();
  }

  async clickPaginationNext() {
    await this.page.locator('[data-testid="btn-pagination-next"]').click();
  }

  async clickPaginationPrevious() {
    await this.page.locator('[data-testid="btn-pagination-previous"]').click();
  }

  async clickPageNumber(pageNum: number) {
    await this.page.locator(`[data-testid="btn-page-${pageNum}"]`).click();
  }

  async expectPaginationVisible() {
    await expect(this.pagination).toBeVisible();
  }

  async expectPaginationHidden() {
    await expect(this.pagination).toBeHidden();
  }

  async getCardName(index: number): Promise<string> {
    return this.chatbotCards.nth(index).locator('[data-testid="card-chatbot-name"]').textContent();
  }

  async getCardStatus(index: number): Promise<string> {
    return this.chatbotCards.nth(index).locator('[data-testid="card-chatbot-status"]').textContent();
  }
}
```

### Chatbot Card Page Object

```typescript
// frontend/e2e/pages/chatbot-card.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ChatbotCardPage {
  readonly page: Page;
  readonly card: Locator;
  readonly name: Locator;
  readonly modelBadge: Locator;
  readonly statusBadge: Locator;
  readonly actionsButton: Locator;
  readonly actionsMenu: Locator;
  readonly menuItemEdit: Locator;
  readonly menuItemDuplicate: Locator;
  readonly menuItemShare: Locator;
  readonly menuItemSettings: Locator;
  readonly menuItemDelete: Locator;

  constructor(page: Page, cardSelector: string) {
    this.page = page;
    this.card = page.locator(cardSelector);
    this.name = page.locator('[data-testid="card-chatbot-name"]');
    this.modelBadge = page.locator('[data-testid="card-chatbot-model"]');
    this.statusBadge = page.locator('[data-testid="card-chatbot-status"]');
    this.actionsButton = page.locator('[data-testid="card-chatbot-actions"]');
    this.actionsMenu = page.locator('[data-testid="dropdown-chatbot-actions"]');
    this.menuItemEdit = page.locator('[data-testid="menu-item-edit"]');
    this.menuItemDuplicate = page.locator('[data-testid="menu-item-duplicate"]');
    this.menuItemShare = page.locator('[data-testid="menu-item-share"]');
    this.menuItemSettings = page.locator('[data-testid="menu-item-settings"]');
    this.menuItemDelete = page.locator('[data-testid="menu-item-delete"]');
  }

  async hover() {
    await this.card.hover();
  }

  async expectHoverState() {
    await expect(this.card).toHaveClass(/hover/);
  }

  async clickCardBody() {
    await this.card.locator('[data-testid="card-body"]').click();
  }

  async clickActions() {
    await this.actionsButton.click();
    await expect(this.actionsMenu).toBeVisible();
  }

  async clickEdit() {
    await this.clickActions();
    await this.menuItemEdit.click();
  }

  async clickDuplicate() {
    await this.clickActions();
    await this.menuItemDuplicate.click();
  }

  async clickShare() {
    await this.clickActions();
    await this.menuItemShare.click();
  }

  async clickSettings() {
    await this.clickActions();
    await this.menuItemSettings.click();
  }

  async clickDelete() {
    await this.clickActions();
    await this.menuItemDelete.click();
  }

  async expectStatus(status: string) {
    await expect(this.statusBadge).toHaveText(status);
  }

  async expectModel(model: string) {
    await expect(this.modelBadge).toHaveText(model);
  }
}
```

### Chatbots Mocks

```typescript
// frontend/e2e/mocks/chatbots.mocks.ts
import { APIRequestContext } from '@playwright/test';

export const mockChatbotsData = [
  {
    id: 'chatbot-1',
    name: 'Customer Support Bot',
    model: 'gpt-4o-mini',
    status: 'active',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-20T14:30:00Z',
    sourceCount: 5,
    messageCount: 150,
  },
  {
    id: 'chatbot-2',
    name: 'Sales Assistant',
    model: 'gpt-4o',
    status: 'training',
    createdAt: '2024-01-10T08:00:00Z',
    updatedAt: '2024-01-19T16:00:00Z',
    sourceCount: 3,
    messageCount: 0,
  },
  {
    id: 'chatbot-3',
    name: 'HR Helper',
    model: 'gpt-4o-mini',
    status: 'error',
    createdAt: '2024-01-05T12:00:00Z',
    updatedAt: '2024-01-18T09:00:00Z',
    sourceCount: 2,
    messageCount: 25,
  },
];

export async function mockGetChatbots(request: APIRequestContext) {
  await request.get('/api/v1/chatbots', {
    status: 200,
    body: {
      data: mockChatbotsData,
      total: mockChatbotsData.length,
      page: 1,
      perPage: 12,
    },
  });
}

export async function mockGetChatbotsEmpty(request: APIRequestContext) {
  await request.get('/api/v1/chatbots', {
    status: 200,
    body: {
      data: [],
      total: 0,
      page: 1,
      perPage: 12,
    },
  });
}

export async function mockDeleteChatbot(request: APIRequestContext, id: string) {
  await request.delete(`/api/v1/chatbots/${id}`, {
    status: 200,
    body: {
      success: true,
      message: 'Chatbot deleted successfully',
    },
  });
}

export async function mockDuplicateChatbot(request: APIRequestContext, originalId: string) {
  await request.post(`/api/v1/chatbots/${originalId}/duplicate`, {
    status: 201,
    body: {
      id: 'chatbot-new',
      name: `${originalId} (Copy)`,
      model: 'gpt-4o-mini',
      status: 'active',
    },
  });
}

export async function mockSearchChatbots(request: APIRequestContext, query: string) {
  const filtered = mockChatbotsData.filter(c => 
    c.name.toLowerCase().includes(query.toLowerCase())
  );
  
  await request.get('/api/v1/chatbots', {
    status: 200,
    body: {
      data: filtered,
      total: filtered.length,
      page: 1,
      perPage: 12,
    },
  });
}
```

### Running Specific Tests

```bash
# Run all chatbots list tests
cd frontend && npx playwright test chatbots-list.spec.ts

# Run card tests
cd frontend && npx playwright test chatbots-list.spec.ts -g "card"

# Run search tests
cd frontend && npx playwright test chatbots-list.spec.ts -g "search"

# Run pagination tests
cd frontend && npx playwright test chatbots-list.spec.ts -g "pagination"

# Run in headed mode
cd frontend && npx playwright test chatbots-list.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All page elements tested
- [ ] All card interactions tested
- [ ] All actions tested
- [ ] All search/filter/sort tested
- [ ] All pagination tested
- [ ] Empty states tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No flaky tests
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Clear card hover states
- [ ] Smooth actions menu
- [ ] Loading states visible
- [ ] Error messages clear

### 4. Performance Verification
- [ ] Pagination loads quickly
- [ ] Search debounced correctly
- [ ] Large datasets handled

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Data Mocking** - Mock API responses for consistent testing
2. **Multiple Cards** - Test with multiple chatbots for pagination
3. **Status States** - Test all status types (active, training, error)
4. **Actions Flow** - Test complete action flows (open modal, confirm, verify)

### Common Issues to Avoid

1. **Hardcoded indices** - Be careful with card indices
2. **Skipping empty states** - Test empty state thoroughly
3. **Not waiting for API** - Wait for list to load
4. **Race conditions** - Wait for dropdown to appear

### Test Data Setup

```typescript
// Use mocked API for consistent tests
test.use({
  baseURL: 'http://localhost:3000',
});

test('displays chatbots correctly', async ({ page }) => {
  await page.route('/api/v1/chatbots', route => {
    route.fulfill({ body: JSON.stringify({ data: mockChatbotsData }) });
  });
  
  await page.goto('/dashboard/chatbots');
  await expect(page.locator('[data-testid="card-chatbot"]')).toHaveCount(3);
});
```

---

## Dependencies

- **Prerequisites**: 06-dashboard-layout.md (navigation to page)
- **Environment**: Backend API with chatbot endpoints
- **Test Data**: Various chatbot data with different statuses

---

## Related Tasks

- 06-dashboard-layout.md - Navigation to chatbots page
- 10-chatbots-create.md - Create chatbot modal
- 11-chatbots-detail.md - Chatbot detail page
- 12-chatbots-settings.md - Chatbot settings

---

*Task created from: docs/frontend/TEST_PATHS.md Section 4.1*
