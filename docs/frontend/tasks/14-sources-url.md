# Task: Implement URL Source Tests

> **Task ID**: 14-sources-url  
> **Source**: TEST_PATHS.md Section 5.2  
> **Priority**: High (Data Sources)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 13-sources-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for adding URL sources. This task covers URL input, validation, discovery options, and path filters.

### Context

The URL Source functionality allows users to add web pages as knowledge sources. Testing this functionality ensures:
- URL validation works correctly
- Discovery options function properly
- Path filters work as expected
- Source creation succeeds with valid URLs
- Error handling handles invalid URLs

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 5.2:

```
Add URL Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   ├── Click: tab-url
│   └── Assert: URL input visible
│
├── Single URL input
│   ├── Click: input-url
│   ├── Type: "https://example.com/page"
│   ├── Assert: URL validation
│   ├── Click: btn-add
│   ├── Assert: Loading state
│   ├── Assert: Source created (pending)
│   └── Assert: Toast "Source added"
│
├── URL with discovery
│   ├── Click: input-url
│   ├── Type: "https://example.com"
│   ├── Toggle: checkbox-discover-pages
│   ├── Click: btn-add
│   ├── Assert: Source created
│   ├── Assert: Discovery started
│   └── Assert: Pending URLs will be discovered
│
├── Path filters
│   ├── Click: input-include-paths
│   ├── Type: "/docs/, /guide/"
│   ├── Click: input-exclude-paths
│   ├── Type: "/admin/, /private/"
│   └── Click: btn-add
│       └── Assert: Filters saved
│
├── Validation
    ├── Empty URL → Error
    ├── Invalid URL → Error
    ├── Blocked domain → Error
    └── Private IP → Error (SSRF protection)
│
└── Cancel
    ├── Click: btn-cancel
    └── Assert: Modal closed, no source created
```

### Implementation Requirements

1. **Create URL Source Test File** (`frontend/e2e/sources-url.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create URL Source Page Object** (`frontend/e2e/pages/add-url-source.page.ts`)
   - Encapsulate URL source modal interactions
   - URL input and validation methods
   - Discovery and filter methods

3. **Create URL Source Mocks** (`frontend/e2e/mocks/sources-url.mocks.ts`)
   - Mock URL source creation endpoint
   - Mock validation errors
   - Mock discovery responses

### Expected Deliverables

1. `frontend/e2e/sources-url.spec.ts` - Comprehensive URL source tests
2. `frontend/e2e/pages/add-url-source.page.ts` - URL source page object
3. `frontend/e2e/mocks/sources-url.mocks.ts` - URL source API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/add-url-source.page.ts`:
  - Modal container locator
  - URL tab locator
  - URL input locator
  - Discovery checkbox locator
  - Include paths input locator
  - Exclude paths input locator
  - Add button locator
  - Cancel button locator

### Phase 2: Modal Open Tests

- [ ] Test: Click add source opens modal
- [ ] Test: Click URL tab shows URL input
- [ ] Test: URL input is focused
- [ ] Test: All fields visible
- [ ] Test: Add button visible
- [ ] Test: Cancel button visible
- [ ] Test: Modal has correct title

### Phase 3: URL Input Tests

- [ ] Test: URL input accepts text
- [ ] Test: Placeholder text visible
- [ ] Test: Input validation on blur
- [ ] Test: Input validation on submit
- [ ] Test: Real-time validation feedback

### Phase 4: URL Validation Tests

- [ ] Test: Empty URL shows error
- [ ] Test: Invalid URL format shows error
- [ ] Test: Missing protocol shows error
- [ ] Test: Blocked domain shows error
- [ ] Test: Private IP shows SSRF error
- [ ] Test: Valid URL passes validation
- [ ] Test: Error message displayed
- [ ] Test: Input has error class

### Phase 5: Single URL Addition Tests

- [ ] Test: Enter valid URL
- [ ] Test: Click add button
- [ ] Test: Loading state visible
- [ ] Test: Button disabled
- [ ] Test: API call made
- [ ] Test: Source created (pending status)
- [ ] Test: Toast success appears
- [ ] Test: Modal closes
- [ ] Test: Source appears in list

### Phase 6: Discovery Option Tests

- [ ] Test: Discovery checkbox visible
- [ ] Test: Checkbox can be toggled
- [ ] Test: Discovery enabled by default?
- [ ] Test: Disable discovery
- [ ] Test: Enable discovery
- [ ] Test: Discovery with valid URL
- [ ] Test: Discovery starts processing
- [ ] Test: Pending URLs added to queue

### Phase 7: Path Filter Tests

- [ ] Test: Include paths input visible
- [ ] Test: Exclude paths input visible
- [ ] Test: Enter include patterns
- [ ] Test: Enter exclude patterns
- [ ] Test: Multiple patterns separated
- [ ] Test: Patterns saved with source
- [ ] Test: Filter validation (if any)

### Phase 8: Cancel Tests

- [ ] Test: Click cancel closes modal
- [ ] Test: No source created
- [ ] Test: Click overlay closes modal
- [ ] Test: Escape key closes modal
- [ ] Test: Changes not saved

---

## Technical Notes

### URL Source Page Object

```typescript
// frontend/e2e/pages/add-url-source.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class AddUrlSourcePage {
  readonly page: Page;
  readonly modal: Locator;
  readonly urlTab: Locator;
  readonly urlInput: Locator;
  readonly discoveryCheckbox: Locator;
  readonly includePathsInput: Locator;
  readonly excludePathsInput: Locator;
  readonly addButton: Locator;
  readonly cancelButton: Locator;
  readonly urlError: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modal = page.locator('[data-testid="modal-add-source"]');
    this.urlTab = page.locator('[data-testid="tab-url"]');
    this.urlInput = page.locator('[data-testid="input-url"]');
    this.discoveryCheckbox = page.locator('[data-testid="checkbox-discover-pages"]');
    this.includePathsInput = page.locator('[data-testid="input-include-paths"]');
    this.excludePathsInput = page.locator('[data-testid="input-exclude-paths"]');
    this.addButton = page.locator('[data-testid="btn-add"]');
    this.cancelButton = page.locator('[data-testid="btn-cancel"]');
    this.urlError = page.locator('[data-testid="error-url"]');
  }

  async expectVisible() {
    await expect(this.modal).toBeVisible();
  }

  async expectHidden() {
    await expect(this.modal).toBeHidden();
  }

  async clickUrlTab() {
    await this.urlTab.click();
  }

  async fillUrl(url: string) {
    await this.urlInput.fill(url);
  }

  async clearUrl() {
    await this.urlInput.clear();
  }

  async toggleDiscovery(enabled: boolean) {
    const isChecked = await this.discoveryCheckbox.isChecked();
    if (isChecked !== enabled) {
      await this.discoveryCheckbox.click();
    }
  }

  async fillIncludePaths(paths: string) {
    await this.includePathsInput.fill(paths);
  }

  async fillExcludePaths(paths: string) {
    await this.excludePathsInput.fill(paths);
  }

  async clickAdd() {
    await this.addButton.click();
  }

  async clickCancel() {
    await this.cancelButton.click();
  }

  async expectUrlError(message: string) {
    await expect(this.urlError).toHaveText(message);
  }

  async expectNoUrlError() {
    await expect(this.urlError).toBeHidden();
  }

  async expectLoading() {
    await expect(this.addButton).toBeDisabled();
    await expect(this.addButton).toHaveText(/Adding.../);
  }

  async addUrl(url: string, options?: {
    discover?: boolean;
    includePaths?: string;
    excludePaths?: string;
  }) {
    await this.fillUrl(url);
    
    if (options?.discover !== undefined) {
      await this.toggleDiscovery(options.discover);
    }
    
    if (options?.includePaths) {
      await this.fillIncludePaths(options.includePaths);
    }
    
    if (options?.excludePaths) {
      await this.fillExcludePaths(options.excludePaths);
    }
    
    await this.clickAdd();
  }
}
```

### URL Source Mocks

```typescript
// frontend/e2e/mocks/sources-url.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockCreateUrlSource(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 201,
    body: {
      id: 'source-new-' + Date.now(),
      name: 'https://example.com/page',
      type: 'url',
      url: 'https://example.com/page',
      status: 'pending',
      chunkCount: 0,
      createdAt: new Date().toISOString(),
    },
  });
}

export async function mockCreateUrlSourceWithDiscovery(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 201,
    body: {
      id: 'source-new-' + Date.now(),
      name: 'https://example.com',
      type: 'url',
      url: 'https://example.com',
      status: 'processing',
      discoveryEnabled: true,
      pendingUrls: 5,
      chunkCount: 0,
      createdAt: new Date().toISOString(),
    },
  });
}

export async function mockInvalidUrlError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'VALIDATION_ERROR',
      field: 'url',
      message: 'Invalid URL format',
    },
  });
}

export async function mockBlockedDomainError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 403,
    body: {
      error: 'DOMAIN_BLOCKED',
      message: 'This domain is not allowed',
    },
  });
}

export async function mockSsrfError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 403,
    body: {
      error: 'SSRF_DETECTED',
      message: 'Private IP addresses are not allowed',
    },
  });
}
```

### Test Data

```typescript
export const validUrls = [
  'https://example.com',
  'https://example.com/page',
  'https://docs.example.com/getting-started',
  'http://localhost:3000',
];

export const invalidUrls = [
  '',
  'not-a-url',
  'example.com',
  'ftp://example.com',
  'https://',
];

export const blockedDomains = [
  'https://blocked-domain.com',
  'https://malicious.com',
];

export const privateIps = [
  'http://192.168.1.1',
  'http://10.0.0.1',
  'http://localhost',
  'http://127.0.0.1',
];

export const pathPatterns = {
  valid: {
    include: '/docs/, /guide/, /api/',
    exclude: '/admin/, /private/, /internal/',
  },
};
```

### Running Specific Tests

```bash
# Run all URL source tests
cd frontend && npx playwright test sources-url.spec.ts

# Run validation tests
cd frontend && npx playwright test sources-url.spec.ts -g "validation"

# Run discovery tests
cd frontend && npx playwright test sources-url.spec.ts -g "discovery"

# Run in headed mode
cd frontend && npx playwright test sources-url.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All URL validation tested
- [ ] All error cases tested
- [ ] Discovery option tested
- [ ] Path filters tested
- [ ] Success flow tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. Security Verification
- [ ] SSRF protection tested
- [ ] Blocked domains tested
- [ ] Invalid URLs rejected

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Security Testing** - Test SSRF and blocked domain scenarios
2. **URL Parsing** - Test various URL formats
3. **Discovery** - Test discovery functionality separately
4. **Modal State** - Ensure clean modal state between tests

### Common Issues to Avoid

1. **Skipping validation** - Test all validation rules
2. **Security gaps** - Don't skip SSRF tests
3. **Race conditions** - Wait for API responses
4. **Hardcoded URLs** - Use test fixtures

---

## Dependencies

- **Prerequisites**: 13-sources-list.md (add source button)
- **Environment**: Backend API with URL validation
- **Test Data**: Various URLs for testing

---

## Related Tasks

- 13-sources-list.md - List page with add source button
- 15-sources-pdf.md - PDF source creation
- 16-sources-sitemap.md - Sitemap source creation
- 17-sources-text.md - Text source creation

---

*Task created from: docs/frontend/TEST_PATHS.md Section 5.2*
