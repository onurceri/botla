# Task: Implement Sources List Tests

> **Task ID**: 13-sources-list  
> **Source**: TEST_PATHS.md Section 5.1  
> **Priority**: High (Data Sources)  
> **Estimated Effort**: 8-10 hours  
> **Prerequisite**: 11-chatbots-detail.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Sources List Page. This task covers source listing, status indicators, card interactions, and source management.

### Context

The Sources List Page displays all knowledge sources for a chatbot. Testing this page ensures:
- Users can view all sources
- Status indicators reflect source state
- Source card interactions work correctly
- Actions menu provides proper functionality
- Progress tracking works during processing

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 5.1:

#### 5.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `btn-add-source` | button | Add source button |
| `tab-source-type` | tabs | URL / PDF / Text / Sitemap |
| `card-source` | card | Source card |
| `card-source-status` | badge | Status indicator |
| `card-source-type` | badge | Type badge |
| `card-source-chunk-count` | text | Chunk count |
| `card-source-actions` | menu | Actions dropdown |
| `progress-bar` | progress | Training progress |
| `input-url` | text | URL input |
| `input-file` | file | File upload |
| `textarea-text` | textarea | Text content |

#### 5.1.2 Source Card Interactions

```
Source Card Flow
├── Card hover state
│   ├── Hover: Card
│   │   ├── Shadow increase
│   │   └── Scale 1.01
│   │
│   └── Hover: Actions button
│       └── Show: Tooltip
│
├── Status states
    ├── pending (yellow) → Show spinner
    ├── processing (blue) → Show progress bar
    ├── completed (green) → Show chunk count
    └── failed (red) → Show error message
│
├── Click: Source card
│   ├── Open: Source detail panel
│   ├── Show: Source info
│   ├── Show: Sample chunks
│   └── Show: Actions
│
└── Source actions menu
    ├── Click: btn-refresh
    │   ├── Open: `modal-refresh-confirm`
    │   ├── Click: btn-confirm
    │   └── Assert: Source re-processing
    │
    ├── Click: btn-view-chunks
    │   ├── Open: `modal-chunk-viewer`
    │   ├── Show: All chunks
    │   ├── Search: Chunk content
    │   └── Export: Chunk list
    │
    ├── Click: btn-download
    │   └── Download: Source content
    │
    └── Click: btn-delete
        ├── Open: `modal-delete-source`
        ├── Type: DELETE to confirm
        ├── Click: btn-delete
        └── Assert: Source deleted
```

### Implementation Requirements

1. **Create Sources List Test File** (`frontend/e2e/sources-list.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Sources List Page Object** (`frontend/e2e/pages/sources-list.page.ts`)
   - Encapsulate list page interactions
   - Source card interaction methods
   - Status verification methods

3. **Create Sources Mocks** (`frontend/e2e/mocks/sources.mocks.ts`)
   - Mock sources list endpoint
   - Mock source data with various statuses
   - Mock source actions (delete, refresh)

### Expected Deliverables

1. `frontend/e2e/sources-list.spec.ts` - Comprehensive sources list tests
2. `frontend/e2e/pages/sources-list.page.ts` - Sources list page object
3. `frontend/e2e/mocks/sources.mocks.ts` - Sources API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/sources-list.page.ts`:
  - Add source button locator
  - Source type tabs locator
  - Source cards container locator
  - Individual card locators
  - Actions button locators
  - Progress bar locators
  - Empty state locator

### Phase 2: Page Load Tests

- [ ] Test: Page loads successfully
- [ ] Test: URL matches chatbot sources path
- [ ] Test: Add source button visible
- [ ] Test: Source type tabs visible
- [ ] Test: Sources list visible (with data)
- [ ] Test: Empty state when no sources

### Phase 3: Source Card Tests

- [ ] Test: Card displays source name
- [ ] Test: Card displays source type badge
- [ ] Test: Card displays status badge
- [ ] Test: Card hover state (shadow)
- [ ] Test: Card hover state (scale)
- [ ] Test: Actions button visible
- [ ] Test: Hover actions button shows tooltip
- [ ] Test: Click card opens detail panel

### Phase 4: Status Indicator Tests

- [ ] Test: Pending status (yellow with spinner)
- [ ] Test: Processing status (blue with progress bar)
- [ ] Test: Completed status (green with chunk count)
- [ ] Test: Failed status (red with error message)
- [ ] Test: Progress bar shows percentage
- [ ] Test: Status updates on refresh
- [ ] Test: Status tooltip on hover

### Phase 5: Source Type Tabs Tests

- [ ] Test: URL tab visible
- [ ] Test: PDF tab visible
- [ ] Test: Text tab visible
- [ ] Test: Sitemap tab visible
- [ ] Test: Click URL tab filters sources
- [ ] Test: Click PDF tab filters sources
- [ ] Test: Click All shows all sources

### Phase 6: Actions Menu Tests

- [ ] Test: Click actions opens dropdown
- [ ] Test: Dropdown contains Refresh option
- [ ] Test: Dropdown contains View Chunks option
- [ ] Test: Dropdown contains Download option
- [ ] Test: Dropdown contains Delete option
- [ ] Test: Hover menu item highlights

### Phase 7: Refresh Source Tests

- [ ] Test: Click refresh opens confirm modal
- [ ] Test: Modal has confirmation message
- [ ] Test: Click confirm re-processes source
- [ ] Test: Status changes to processing
- [ ] Test: Toast success appears
- [ ] Test: Click cancel closes modal

### Phase 8: View Chunks Tests

- [ ] Test: Click view chunks opens modal
- [ ] Test: Modal shows chunk list
- [ ] Test: Each chunk shows preview text
- [ ] Test: Chunk search functionality
- [ ] Test: Chunk pagination
- [ ] Test: Export chunks option
- [ ] Test: Modal close functionality

### Phase 9: Download Source Tests

- [ ] Test: Click download triggers download
- [ ] Test: Download file has correct name
- [ ] Test: Download format is correct

### Phase 10: Delete Source Tests

- [ ] Test: Click delete opens confirm modal
- [ ] Test: Warning message displayed
- [ ] Test: Requires DELETE confirmation
- [ ] Test: Wrong confirmation shows error
- [ ] Test: Correct confirmation deletes source
- [ ] Test: Source removed from list
- [ ] Test: Toast success appears

---

## Technical Notes

### Sources List Page Object

```typescript
// frontend/e2e/pages/sources-list.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class SourcesListPage {
  readonly page: Page;
  readonly addSourceButton: Locator;
  readonly sourceTypeTabs: Locator;
  readonly sourcesContainer: Locator;
  readonly sourceCards: Locator;
  readonly emptyState: Locator;
  readonly progressBar: Locator;
  readonly actionsButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.addSourceButton = page.locator('[data-testid="btn-add-source"]');
    this.sourceTypeTabs = page.locator('[data-testid="tab-source-type"]');
    this.sourcesContainer = page.locator('[data-testid="sources-container"]');
    this.sourceCards = page.locator('[data-testid="card-source"]');
    this.emptyState = page.locator('[data-testid="empty-state"]');
    this.progressBar = page.locator('[data-testid="progress-bar"]');
    this.actionsButton = page.locator('[data-testid="card-source-actions"]');
  }

  async goto(chatbotId: string) {
    await this.page.goto(`/dashboard/chatbots/${chatbotId}/sources`);
  }

  async expectVisible() {
    await expect(this.addSourceButton).toBeVisible();
    await expect(this.sourceTypeTabs).toBeVisible();
  }

  async expectSourceCount(count: number) {
    await expect(this.sourceCards).toHaveCount(count);
  }

  async expectEmptyState() {
    await expect(this.emptyState).toBeVisible();
    await expect(this.sourceCards).toHaveCount(0);
  }

  async clickAddSource() {
    await this.addSourceButton.click();
  }

  async clickSourceTypeTab(type: string) {
    await this.page.locator(`[data-testid="tab-${type}"]`).click();
  }

  async clickSourceCard(index: number) {
    await this.sourceCards.nth(index).click();
  }

  async clickActionsOnCard(index: number) {
    await this.sourceCards.nth(index).locator('[data-testid="card-source-actions"]').click();
  }

  async getSourceName(index: number): Promise<string> {
    return this.sourceCards.nth(index).locator('[data-testid="card-source-name"]').textContent();
  }

  async getSourceStatus(index: number): Promise<string> {
    return this.sourceCards.nth(index).locator('[data-testid="card-source-status"]').textContent();
  }

  async getSourceType(index: number): Promise<string> {
    return this.sourceCards.nth(index).locator('[data-testid="card-source-type"]').textContent();
  }

  async expectStatus(index: number, status: string) {
    await expect(
      this.sourceCards.nth(index).locator(`[data-testid="card-source-status"]`)
    ).toHaveClass(new RegExp(status));
  }

  async expectProcessingWithProgress(index: number, progress: number) {
    const card = this.sourceCards.nth(index);
    await expect(card.locator('[data-testid="progress-bar"]')).toBeVisible();
    await expect(card.locator('[data-testid="progress-value"]')).toHaveText(`${progress}%`);
  }
}
```

### Sources Mocks

```typescript
// frontend/e2e/mocks/sources.mocks.ts
import { APIRequestContext } from '@playwright/test';

export const mockSourcesData = [
  {
    id: 'source-1',
    name: 'Product Documentation',
    type: 'url',
    status: 'completed',
    chunkCount: 42,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-20T14:30:00Z',
  },
  {
    id: 'source-2',
    name: 'User Guide PDF',
    type: 'pdf',
    status: 'processing',
    progress: 65,
    chunkCount: 0,
    createdAt: '2024-01-18T08:00:00Z',
    updatedAt: '2024-01-19T16:00:00Z',
  },
  {
    id: 'source-3',
    name: 'FAQ Text',
    type: 'text',
    status: 'pending',
    chunkCount: 0,
    createdAt: '2024-01-19T12:00:00Z',
    updatedAt: '2024-01-19T12:00:00Z',
  },
  {
    id: 'source-4',
    name: 'Old Documentation',
    type: 'url',
    status: 'error',
    errorMessage: 'Failed to fetch URL',
    chunkCount: 0,
    createdAt: '2024-01-10T09:00:00Z',
    updatedAt: '2024-01-17T11:00:00Z',
  },
];

export async function mockGetSources(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 200,
    body: {
      data: mockSourcesData,
      total: mockSourcesData.length,
    },
  });
}

export async function mockGetSourcesEmpty(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 200,
    body: {
      data: [],
      total: 0,
    },
  });
}

export async function mockRefreshSource(request: APIRequestContext, sourceId: string) {
  await request.post(`/api/v1/sources/${sourceId}/refresh`, {
    status: 200,
    body: {
      success: true,
      message: 'Source refresh started',
    },
  });
}

export async function mockDeleteSource(request: APIRequestContext, sourceId: string) {
  await request.delete(`/api/v1/sources/${sourceId}`, {
    status: 200,
    body: {
      success: true,
      message: 'Source deleted successfully',
    },
  });
}

export async function mockGetChunks(request: APIRequestContext, sourceId: string) {
  await request.get(`/api/v1/sources/${sourceId}/chunks`, {
    status: 200,
    body: {
      data: [
        { id: 'chunk-1', content: 'First chunk content...', tokenCount: 150 },
        { id: 'chunk-2', content: 'Second chunk content...', tokenCount: 180 },
        { id: 'chunk-3', content: 'Third chunk content...', tokenCount: 165 },
      ],
      total: 3,
    },
  });
}
```

### Running Specific Tests

```bash
# Run all sources list tests
cd frontend && npx playwright test sources-list.spec.ts

# Run status indicator tests
cd frontend && npx playwright test sources-list.spec.ts -g "status"

# Run actions menu tests
cd frontend && npx playwright test sources-list.spec.ts -g "actions"

# Run in headed mode
cd frontend && npx playwright test sources-list.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All page elements tested
- [ ] All status types tested
- [ ] All actions tested
- [ ] Empty states tested
- [ ] Tab filtering tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Clear status indicators
- [ ] Smooth hover effects
- [ ] Loading states visible
- [ ] Error messages clear

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Status States** - Test all four status types (pending, processing, completed, error)
2. **Progress Bars** - Verify progress updates during processing
3. **File Downloads** - Handle download verification carefully
4. **Modal Flows** - Test complete modal interactions

### Common Issues to Avoid

1. **Skipping status tests** - Each status needs verification
2. **Race conditions** - Wait for status updates
3. **Hardcoded data** - Use mock fixtures
4. **Not testing modals** - Test all modal interactions

---

## Dependencies

- **Prerequisites**: 11-chatbots-detail.md (navigation to sources)
- **Environment**: Backend API with sources endpoints
- **Test Data**: Sources with various statuses

---

## Related Tasks

- 11-chatbots-detail.md - Tab navigation to sources
- 14-sources-url.md - URL source creation
- 15-sources-pdf.md - PDF source creation
- 16-sources-sitemap.md - Sitemap source creation
- 17-sources-text.md - Text source creation
- 18-sources-chunks.md - Chunk viewer

---

*Task created from: docs/frontend/TEST_PATHS.md Section 5.1*
