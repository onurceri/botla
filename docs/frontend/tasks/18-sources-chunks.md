# Task: Implement Chunk Viewer Tests

> **Task ID**: 18-sources-chunks  
> **Source**: TEST_PATHS.md Section 5.6  
> **Priority**: High (Data Sources)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 13-sources-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Chunk Viewer Modal. This task covers chunk list display, detail view, search, pagination, and export functionality.

### Context

The Chunk Viewer Modal allows users to view and manage the text chunks extracted from sources. Testing this functionality ensures:
- Chunk list displays correctly
- Chunk detail view works
- Search filters chunks
- Pagination functions properly
- Export works for various formats

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 5.6:

```
Chunk Viewer Modal Flow
├── Open chunk viewer
│   ├── Click: btn-view-chunks (on source card)
│   ├── Assert: `modal-chunk-viewer` opens
│   ├── Show: Source title
│   └── Show: Chunk list
│
├── Chunk list
│   ├── Show: All chunks
│   ├── Each chunk shows:
│   │   ├── Chunk number
│   │   ├── Token count
│   │   └── Preview text
│   │
│   ├── Click: Chunk item
│   │   ├── Show: Chunk detail
│   │   ├── Show: Full text
│   │   └── Show: Metadata
│   │
│   └── Hover: Chunk item
│       └── Highlight background
│
├── Search chunks
│   ├── Type: Search term
│   ├── Assert: Filtered results
│   └── Click: Clear search
│
├── Pagination
    ├── Navigate: Pages
    └── Change: Items per page
│
└── Export chunks
    ├── Click: btn-export
    ├── Options:
    │   ├── JSON
    │   ├── CSV
    │   └── Plain text
    ├── Select: Format
    └── Download: File
```

### Implementation Requirements

1. **Create Chunk Viewer Test File** (`frontend/e2e/sources-chunks.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Chunk Viewer Page Object** (`frontend/e2e/pages/chunk-viewer.page.ts`)
   - Encapsulate modal interactions
   - Chunk list methods
   - Search and pagination methods
   - Export methods

3. **Create Chunk Viewer Mocks** (`frontend/e2e/mocks/chunks.mocks.ts`)
   - Mock chunks list endpoint
   - Mock chunk detail endpoint
   - Mock export endpoint

### Expected Deliverables

1. `frontend/e2e/sources-chunks.spec.ts` - Comprehensive chunk viewer tests
2. `frontend/e2e/pages/chunk-viewer.page.ts` - Chunk viewer page object
3. `frontend/e2e/mocks/chunks.mocks.ts` - Chunks API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/chunk-viewer.page.ts`:
  - Modal container locator
  - Source title locator
  - Chunk list container locator
  - Chunk items locators
  - Search input locator
  - Pagination controls locator
  - Export button locator
  - Export format options locator

### Phase 2: Modal Open Tests

- [ ] Test: Click view chunks opens modal
- [ ] Test: Modal is visible
- [ ] Test: Source title displayed
- [ ] Test: Chunk list visible
- [ ] Test: Total chunk count displayed
- [ ] Test: Close button visible
- [ ] Test: Click overlay closes modal
- [ ] Test: Escape key closes modal

### Phase 3: Chunk List Tests

- [ ] Test: All chunks displayed
- [ ] Test: Chunk number shown
- [ ] Test: Token count shown
- [ ] Test: Preview text shown
- [ ] Test: Chunk order correct
- [ ] Test: Empty chunks state
- [ ] Test: Hover highlights item
- [ ] Test: Scroll if many chunks

### Phase 4: Chunk Detail Tests

- [ ] Test: Click chunk opens detail
- [ ] Test: Full text displayed
- [ ] Test: Metadata displayed
- [ ] Test: Token count in detail
- [ ] Test: Chunk number in detail
- [ ] Test: Back to list button
- [ ] Test: Close detail, back to list
- [ ] Test: Multiple chunk details

### Phase 5: Search Tests

- [ ] Test: Search input visible
- [ ] Test: Type search term
- [ ] Test: Results filtered
- [ ] Test: No results message
- [ ] Test: Clear search button
- [ ] Test: Click clears search
- [ ] Test: Search highlighting
- [ ] Test: Case insensitive search

### Phase 6: Pagination Tests

- [ ] Test: Pagination visible
- [ ] Test: Page numbers displayed
- [ ] Test: Click next page
- [ ] Test: Click previous page
- [ ] Test: Click specific page
- [ ] Test: Items per page selector
- [ ] Test: Change items per page
- [ ] Test: First/last page buttons

### Phase 7: Export Tests

- [ ] Test: Export button visible
- [ ] Test: Click export opens menu
- [ ] Test: JSON format option
- [ ] Test: CSV format option
- [ ] Test: Plain text format option
- [ ] Test: Select format triggers download
- [ ] Test: Download file has correct name
- [ ] Test: Download has correct content

---

## Technical Notes

### Chunk Viewer Page Object

```typescript
// frontend/e2e/pages/chunk-viewer.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ChunkViewerPage {
  readonly page: Page;
  readonly modal: Locator;
  readonly sourceTitle: Locator;
  readonly chunkList: Locator;
  readonly chunkItems: Locator;
  readonly searchInput: Locator;
  readonly clearSearchButton: Locator;
  readonly pagination: Locator;
  readonly exportButton: Locator;
  readonly exportMenu: Locator;
  readonly closeButton: Locator;
  readonly chunkDetail: Locator;
  readonly chunkDetailText: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modal = page.locator('[data-testid="modal-chunk-viewer"]');
    this.sourceTitle = page.locator('[data-testid="chunk-source-title"]');
    this.chunkList = page.locator('[data-testid="chunk-list"]');
    this.chunkItems = page.locator('[data-testid="chunk-item"]');
    this.searchInput = page.locator('[data-testid="input-chunk-search"]');
    this.clearSearchButton = page.locator('[data-testid="btn-clear-search"]');
    this.pagination = page.locator('[data-testid="chunk-pagination"]');
    this.exportButton = page.locator('[data-testid="btn-export"]');
    this.exportMenu = page.locator('[data-testid="export-menu"]');
    this.closeButton = page.locator('[data-testid="btn-close-modal"]');
    this.chunkDetail = page.locator('[data-testid="chunk-detail"]');
    this.chunkDetailText = page.locator('[data-testid="chunk-full-text"]');
  }

  async expectVisible() {
    await expect(this.modal).toBeVisible();
  }

  async expectHidden() {
    await expect(this.modal).toBeHidden();
  }

  async expectChunkCount(count: number) {
    await expect(this.chunkItems).toHaveCount(count);
  }

  async clickChunk(index: number) {
    await this.chunkItems.nth(index).click();
  }

  async expectChunkDetailVisible() {
    await expect(this.chunkDetail).toBeVisible();
  }

  async expectChunkDetailHidden() {
    await expect(this.chunkDetail).toBeHidden();
  }

  async getChunkText(index: number): Promise<string> {
    return this.chunkItems.nth(index).locator('[data-testid="chunk-preview"]').textContent();
  }

  async getChunkTokenCount(index: number): Promise<string> {
    return this.chunkItems.nth(index).locator('[data-testid="chunk-token-count"]').textContent();
  }

  async search(query: string) {
    await this.searchInput.fill(query);
    await this.page.waitForTimeout(300); // Debounce
  }

  async clearSearch() {
    await this.clearSearchButton.click();
  }

  async expectNoResults() {
    await expect(this.page.locator('[data-testid="no-chunks-found"]')).toBeVisible();
  }

  async clickExport() {
    await this.exportButton.click();
  }

  async selectExportFormat(format: 'json' | 'csv' | 'text') {
    await this.page.locator(`[data-testid="export-format-${format}"]`).click();
  }

  async clickClose() {
    await this.closeButton.click();
  }

  async clickBackToList() {
    await this.page.locator('[data-testid="btn-back-to-list"]').click();
  }

  async clickPage(pageNum: number) {
    await this.page.locator(`[data-testid="btn-page-${pageNum}"]`).click();
  }

  async clickNextPage() {
    await this.page.locator('[data-testid="btn-next-page"]').click();
  }

  async clickPreviousPage() {
    await this.page.locator('[data-testid="btn-previous-page"]').click();
  }
}
```

### Chunk Viewer Mocks

```typescript
// frontend/e2e/mocks/chunks.mocks.ts
import { APIRequestContext } from '@playwright/test';

export const mockChunksData = {
  chunks: [
    {
      id: 'chunk-1',
      chunkNumber: 1,
      content: 'This is the first chunk of text extracted from the source document.',
      tokenCount: 150,
      metadata: { page: 1, position: 1 },
    },
    {
      id: 'chunk-2',
      chunkNumber: 2,
      content: 'This is the second chunk with more content about the topic.',
      tokenCount: 180,
      metadata: { page: 1, position: 2 },
    },
    {
      id: 'chunk-3',
      chunkNumber: 3,
      content: 'The third chunk continues with additional information and details.',
      tokenCount: 165,
      metadata: { page: 2, position: 1 },
    },
  ],
  total: 45,
  page: 1,
  perPage: 10,
};

export async function mockGetChunks(request: APIRequestContext, sourceId: string) {
  await request.get(`/api/v1/sources/${sourceId}/chunks`, {
    status: 200,
    body: {
      data: mockChunksData.chunks,
      total: mockChunksData.total,
      page: mockChunksData.page,
      perPage: mockChunksData.perPage,
    },
  });
}

export async function mockGetChunkDetail(request: APIRequestContext, sourceId: string, chunkId: string) {
  const chunk = mockChunksData.chunks.find(c => c.id === chunkId) || mockChunksData.chunks[0];
  
  await request.get(`/api/v1/sources/${sourceId}/chunks/${chunkId}`, {
    status: 200,
    body: {
      ...chunk,
      fullContent: chunk.content + ' Extended content for detail view.',
      metadata: {
        ...chunk.metadata,
        source: 'document.pdf',
        createdAt: '2024-01-15T10:00:00Z',
      },
    },
  });
}

export async function mockSearchChunks(request: APIRequestContext, sourceId: string, query: string) {
  const filtered = mockChunksData.chunks.filter(c =>
    c.content.toLowerCase().includes(query.toLowerCase())
  );
  
  await request.get(`/api/v1/sources/${sourceId}/chunks`, {
    status: 200,
    body: {
      data: filtered,
      total: filtered.length,
      page: 1,
      perPage: 10,
    },
  });
}

export async function mockExportChunks(request: APIRequestContext, sourceId: string, format: string) {
  const content = format === 'json'
    ? JSON.stringify(mockChunksData.chunks, null, 2)
    : format === 'csv'
    ? 'id,content,tokenCount\nchunk-1,Sample text,150'
    : mockChunksData.chunks.map(c => c.content).join('\n\n');
  
  await request.post(`/api/v1/sources/${sourceId}/export`, {
    status: 200,
    body: content,
    headers: {
      'Content-Disposition': `attachment; filename="chunks.${format}"`,
      'Content-Type': format === 'json' ? 'application/json' : 'text/plain',
    },
  });
}
```

### Running Specific Tests

```bash
# Run all chunk viewer tests
cd frontend && npx playwright test sources-chunks.spec.ts

# Run chunk list tests
cd frontend && npx playwright test sources-chunks.spec.ts -g "chunk list"

# Run export tests
cd frontend && npx playwright test sources-chunks.spec.ts -g "export"

# Run in headed mode
cd frontend && npx playwright test sources-chunks.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] Modal open/close tested
- [ ] Chunk list displayed correctly
- [ ] Chunk detail view tested
- [ ] Search filtering tested
- [ ] Pagination tested
- [ ] Export tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Clear chunk display
- [ ] Intuitive navigation
- [ ] Helpful search
- [ ] Smooth export flow

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Chunk Display** - Test with varying chunk counts
2. **Search** - Test search filtering thoroughly
3. **Export** - Handle file download verification
4. **Pagination** - Test multi-page scenarios

### Common Issues to Avoid

1. **Skipping detail view** - Test complete detail flow
2. **Race conditions** - Wait for chunk list to load
3. **Hardcoded data** - Use mock fixtures
4. **Download handling** - Use Playwright download API

---

## Dependencies

- **Prerequisites**: 13-sources-list.md (view chunks button)
- **Environment**: Backend API with chunks endpoints
- **Test Data**: Various chunk data

---

## Related Tasks

- 13-sources-list.md - List page with view chunks action
- 14-sources-url.md - URL sources create chunks
- 15-sources-pdf.md - PDF sources create chunks
- 16-sources-sitemap.md - Sitemap sources create chunks
- 17-sources-text.md - Text sources create chunks

---

*Task created from: docs/frontend/TEST_PATHS.md Section 5.6*
