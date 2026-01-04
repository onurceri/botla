# Task: Implement Sitemap Source Tests

> **Task ID**: 16-sources-sitemap  
> **Source**: TEST_PATHS.md Section 5.4  
> **Priority**: High (Data Sources)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 13-sources-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for adding sitemap sources. This task covers sitemap URL input, analysis, crawling configuration, and progress tracking.

### Context

The Sitemap Source functionality allows users to crawl entire websites using sitemap.xml. Testing this functionality ensures:
- Sitemap URL validation works
- Sitemap analysis displays discovered URLs
- Crawling configuration options work
- Progress tracking shows during crawling
- URL approval workflow functions

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 5.4:

```
Add Sitemap Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   ├── Click: tab-sitemap
│   └── Assert: Sitemap input visible
│
├── Sitemap URL input
│   ├── Click: input-sitemap-url
│   ├── Type: "https://example.com/sitemap.xml"
│   ├── Click: btn-analyze
│   ├── Assert: Loading state
│   ├── Assert: Sitemap parsed
│   └── Show: URL list preview
│
├── Configuration options
    ├── Max URLs to crawl (input-number)
    │   ├── Default: 100
    │   ├── Min: 1
    │   └── Max: 1000
    │
    ├── Priority patterns (input)
    │   └── Type: "/products/*, /pricing/*"
    │
    └── Exclude patterns (input)
        └── Type: "/admin/*, /private/*"
│
├── Start crawling
│   ├── Click: btn-start-crawling
│   ├── Assert: Source created (processing)
│   ├── Assert: Job queued
│   └── Assert: Toast "Crawling started"
│
├── Crawling progress
    ├── Show: URLs processed count
    ├── Show: URLs pending count
    ├── Show: Errors count
    └── Show: Progress bar
│
├── Approve pending URLs
│   ├── Click: tab-pending
│   ├── Show: Discovered URLs list
│   ├── Click: btn-approve-all
│   ├── Assert: All URLs approved
│   └── Assert: Processing continues
│
└── Validation
    ├── Invalid sitemap → Error
    ├── Empty sitemap → Error
    └── Too many URLs → Warning
```

### Implementation Requirements

1. **Create Sitemap Source Test File** (`frontend/e2e/sources-sitemap.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Sitemap Source Page Object** (`frontend/e2e/pages/add-sitemap-source.page.ts`)
   - Encapsulate sitemap modal interactions
   - URL input and analysis methods
   - Configuration and crawling methods

3. **Create Sitemap Source Mocks** (`frontend/e2e/mocks/sources-sitemap.mocks.ts`)
   - Mock sitemap analysis endpoint
   - Mock crawling progress
   - Mock URL discovery

### Expected Deliverables

1. `frontend/e2e/sources-sitemap.spec.ts` - Comprehensive sitemap tests
2. `frontend/e2e/pages/add-sitemap-source.page.ts` - Sitemap page object
3. `frontend/e2e/mocks/sources-sitemap.mocks.ts` - Sitemap API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/add-sitemap-source.page.ts`:
  - Modal container locator
  - Sitemap tab locator
  - Sitemap URL input locator
  - Analyze button locator
  - Max URLs input locator
  - Priority patterns input locator
  - Exclude patterns input locator
  - Start crawling button locator
  - URL list container locator
  - Pending URLs tab locator

### Phase 2: Modal Open Tests

- [ ] Test: Click add source opens modal
- [ ] Test: Click sitemap tab shows input
- [ ] Test: Sitemap URL input visible
- [ ] Test: Analyze button visible
- [ ] Test: Configuration options visible
- [ ] Test: Default values set

### Phase 3: Sitemap URL Tests

- [ ] Test: URL input accepts text
- [ ] Test: Placeholder text visible
- [ ] Test: Input validation
- [ ] Test: Click analyze
- [ ] Test: Loading state visible
- [ ] Test: Sitemap parsed successfully
- [ ] Test: URL list preview appears
- [ ] Test: Total URLs count displayed

### Phase 4: Sitemap Validation Tests

- [ ] Test: Empty URL shows error
- [ ] Test: Invalid URL format shows error
- [ ] Test: Invalid sitemap shows error
- [ ] Test: Empty sitemap shows error
- [ ] Test: Too many URLs shows warning
- [ ] Test: Error message displayed
- [ ] Test: Retry after error

### Phase 5: Configuration Tests

- [ ] Test: Max URLs input visible
- [ ] Test: Default value 100
- [ ] Test: Min value 1 validation
- [ ] Test: Max value 1000 validation
- [ ] Test: Priority patterns input
- [ ] Test: Exclude patterns input
- [ ] Test: Multiple patterns separated
- [ ] Test: Patterns saved

### Phase 6: Crawling Tests

- [ ] Test: Click start crawling
- [ ] Test: Source created (processing)
- [ ] Test: Job queued
- [ ] Test: Toast success appears
- [ ] Test: Modal closes
- [ ] Test: Source appears in list

### Phase 7: Progress Tracking Tests

- [ ] Test: URLs processed count
- [ ] Test: URLs pending count
- [ ] Test: Errors count
- [ ] Test: Progress bar visible
- [ ] Test: Progress percentage
- [ ] Test: Real-time updates

### Phase 8: URL Approval Tests

- [ ] Test: Click pending URLs tab
- [ ] Test: Discovered URLs list visible
- [ ] Test: Each URL shows checkbox
- [ ] Test: Select single URL
- [ ] Test: Select multiple URLs
- [ ] Test: Click approve selected
- [ ] Test: Click approve all
- [ ] Test: URLs approved successfully
- [ ] Test: Processing continues

---

## Technical Notes

### Sitemap Source Page Object

```typescript
// frontend/e2e/pages/add-sitemap-source.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class AddSitemapSourcePage {
  readonly page: Page;
  readonly modal: Locator;
  readonly sitemapTab: Locator;
  readonly sitemapUrlInput: Locator;
  readonly analyzeButton: Locator;
  readonly maxUrlsInput: Locator;
  readonly priorityPatternsInput: Locator;
  readonly excludePatternsInput: Locator;
  readonly startCrawlingButton: Locator;
  readonly urlListContainer: Locator;
  readonly pendingUrlsTab: Locator;
  readonly urlCount: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modal = page.locator('[data-testid="modal-add-source"]');
    this.sitemapTab = page.locator('[data-testid="tab-sitemap"]');
    this.sitemapUrlInput = page.locator('[data-testid="input-sitemap-url"]');
    this.analyzeButton = page.locator('[data-testid="btn-analyze"]');
    this.maxUrlsInput = page.locator('[data-testid="input-max-urls"]');
    this.priorityPatternsInput = page.locator('[data-testid="input-priority-patterns"]');
    this.excludePatternsInput = page.locator('[data-testid="input-exclude-patterns"]');
    this.startCrawlingButton = page.locator('[data-testid="btn-start-crawling"]');
    this.urlListContainer = page.locator('[data-testid="url-list-container"]');
    this.pendingUrlsTab = page.locator('[data-testid="tab-pending"]');
    this.urlCount = page.locator('[data-testid="url-count"]');
  }

  async expectVisible() {
    await expect(this.modal).toBeVisible();
  }

  async clickSitemapTab() {
    await this.sitemapTab.click();
  }

  async fillSitemapUrl(url: string) {
    await this.sitemapUrlInput.fill(url);
  }

  async clickAnalyze() {
    await this.analyzeButton.click();
  }

  async expectAnalyzing() {
    await expect(this.analyzeButton).toBeDisabled();
    await expect(this.analyzeButton).toHaveText(/Analyzing.../);
  }

  async expectUrlListVisible() {
    await expect(this.urlListContainer).toBeVisible();
  }

  async expectUrlCount(count: number) {
    await expect(this.urlCount).toHaveText(`${count} URLs found`);
  }

  async setMaxUrls(count: number) {
    await this.maxUrlsInput.clear();
    await this.maxUrlsInput.fill(count.toString());
  }

  async fillPriorityPatterns(patterns: string) {
    await this.priorityPatternsInput.fill(patterns);
  }

  async fillExcludePatterns(patterns: string) {
    await this.excludePatternsInput.fill(patterns);
  }

  async clickStartCrawling() {
    await this.startCrawlingButton.click();
  }

  async clickPendingUrlsTab() {
    await this.pendingUrlsTab.click();
  }

  async expectError(message: string) {
    await expect(this.page.locator('[data-testid="error-message"]')).toContainText(message);
  }

  async expectWarning(message: string) {
    await expect(this.page.locator('[data-testid="warning-message"]')).toContainText(message);
  }
}
```

### Sitemap Source Mocks

```typescript
// frontend/e2e/mocks/sources-sitemap.mocks.ts
import { APIRequestContext } from '@playwright/test';

export const mockSitemapData = {
  urls: [
    { url: 'https://example.com/', priority: 1.0, changefreq: 'daily' },
    { url: 'https://example.com/products', priority: 0.8, changefreq: 'weekly' },
    { url: 'https://example.com/about', priority: 0.5, changefreq: 'monthly' },
  ],
  totalCount: 3,
};

export async function mockAnalyzeSitemap(request: APIRequestContext) {
  await request.post('/api/v1/sources/analyze-sitemap', {
    status: 200,
    body: {
      urls: mockSitemapData.urls,
      totalCount: mockSitemapData.totalCount,
      isValid: true,
    },
  });
}

export async function mockInvalidSitemap(request: APIRequestContext) {
  await request.post('/api/v1/sources/analyze-sitemap', {
    status: 400,
    body: {
      error: 'INVALID_SITEMAP',
      message: 'Could not parse sitemap XML',
    },
  });
}

export async function mockEmptySitemap(request: APIRequestContext) {
  await request.post('/api/v1/sources/analyze-sitemap', {
    status: 200,
    body: {
      urls: [],
      totalCount: 0,
      isValid: true,
      warning: 'Sitemap is empty',
    },
  });
}

export async function mockTooManyUrls(request: APIRequestContext) {
  await request.post('/api/v1/sources/analyze-sitemap', {
    status: 200,
    body: {
      urls: Array(1500).fill({ url: 'https://example.com/page' }),
      totalCount: 1500,
      isValid: true,
      warning: 'Sitemap contains 1500 URLs. Maximum allowed is 1000.',
    },
  });
}

export async function mockStartCrawling(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 201,
    body: {
      id: 'source-sitemap-' + Date.now(),
      name: 'https://example.com/sitemap.xml',
      type: 'sitemap',
      status: 'processing',
      urlsDiscovered: 3,
      urlsProcessed: 0,
      urlsPending: 3,
      urlsError: 0,
      createdAt: new Date().toISOString(),
    },
  });
}

export async function mockCrawlingProgress(request: APIRequestContext, sourceId: string) {
  await request.get(`/api/v1/sources/${sourceId}/progress`, {
    status: 200,
    body: {
      urlsDiscovered: 3,
      urlsProcessed: 2,
      urlsPending: 1,
      urlsError: 0,
      progress: 66,
    },
  });
}

export async function mockApproveUrls(request: APIRequestContext, sourceId: string) {
  await request.post(`/api/v1/sources/${sourceId}/approve-urls`, {
    status: 200,
    body: {
      success: true,
      approvedCount: 3,
    },
  });
}
```

### Running Specific Tests

```bash
# Run all sitemap source tests
cd frontend && npx playwright test sources-sitemap.spec.ts

# Run analysis tests
cd frontend && npx playwright test sources-sitemap.spec.ts -g "analyze"

# Run crawling tests
cd frontend && npx playwright test sources-sitemap.spec.ts -g "crawling"

# Run in headed mode
cd frontend && npx playwright test sources-sitemap.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] URL input and analysis tested
- [ ] Validation tested
- [ ] Configuration options tested
- [ ] Crawling flow tested
- [ ] Progress tracking tested
- [ ] URL approval tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Clear progress indicators
- [ ] Helpful error messages
- [ ] Intuitive URL approval flow
- [ ] Configuration options clear

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Sitemap XML** - Mock XML parsing scenarios
2. **Progress Updates** - Test progress polling
3. **URL Approval** - Test selection and approval flow
4. **Large Sitemaps** - Test warning for too many URLs

### Common Issues to Avoid

1. **Skipping validation** - Test all sitemap errors
2. **Race conditions** - Wait for analysis to complete
3. **Hardcoded URLs** - Use mock data
4. **Not testing approval** - Test complete workflow

---

## Dependencies

- **Prerequisites**: 13-sources-list.md (add source button)
- **Environment**: Backend API with sitemap analysis
- **Test Data**: Various sitemap scenarios

---

## Related Tasks

- 13-sources-list.md - List page with add source button
- 14-sources-url.md - URL source creation
- 15-sources-pdf.md - PDF source creation
- 17-sources-text.md - Text source creation

---

*Task created from: docs/frontend/TEST_PATHS.md Section 5.4*
