# Task: Implement Text Source Tests

> **Task ID**: 17-sources-text  
> **Source**: TEST_PATHS.md Section 5.5  
> **Priority**: High (Data Sources)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 13-sources-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for adding text sources. This task covers text input, character/word counting, file import, and title input.

### Context

The Text Source functionality allows users to add plain text as knowledge sources. Testing this functionality ensures:
- Text input works correctly
- Character and word counts update in real-time
- File import works for supported formats
- Title input is validated
- Empty/long text is handled properly

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 5.5:

```
Add Text Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   ├── Click: tab-text
│   └── Assert: Text input area visible
│
├── Text input
│   ├── Click: textarea-text
│   ├── Type/Paste: Content
│   ├── Assert: Character count
│   ├── Assert: Word count
│   └── Click: btn-add
│       ├── Assert: Source created
│       └── Assert: Toast "Text added"
│
├── Import from file
│   ├── Click: btn-import-file
│   ├── Select: .txt, .md, .html file
│   ├── Assert: Content imported
│   └── Assert: Source created
│
├── Title input
│   ├── Click: input-title
│   ├── Type: Source title
│   └── Assert: Title saved with source
│
└── Validation
    ├── Empty text → Error
    ├── Text too long → Error "Max 100K chars"
    └── Invalid encoding → Error
```

### Implementation Requirements

1. **Create Text Source Test File** (`frontend/e2e/sources-text.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Text Source Page Object** (`frontend/e2e/pages/add-text-source.page.ts`)
   - Encapsulate text source modal interactions
   - Text input and counting methods
   - File import methods

3. **Create Text Source Mocks** (`frontend/e2e/mocks/sources-text.mocks.ts`)
   - Mock text source creation endpoint
   - Mock validation errors
   - Mock file import

### Expected Deliverables

1. `frontend/e2e/sources-text.spec.ts` - Comprehensive text source tests
2. `frontend/e2e/pages/add-text-source.page.ts` - Text source page object
3. `frontend/e2e/mocks/sources-text.mocks.ts` - Text source API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/add-text-source.page.ts`:
  - Modal container locator
  - Text tab locator
  - Title input locator
  - Text textarea locator
  - Character count locator
  - Word count locator
  - Import button locator
  - Add button locator

### Phase 2: Modal Open Tests

- [ ] Test: Click add source opens modal
- [ ] Test: Click text tab shows input
- [ ] Test: Title input visible
- [ ] Test: Text textarea visible
- [ ] Test: Character count visible
- [ ] Test: Word count visible
- [ ] Test: Import button visible
- [ ] Test: Add button visible

### Phase 3: Text Input Tests

- [ ] Test: Textarea accepts text
- [ ] Test: Textarea accepts paste
- [ ] Test: Character count updates
- [ ] Test: Word count updates
- [ ] Test: Real-time counting
- [ ] Test: Empty text shows 0/0
- [ ] Test: Long text shows count

### Phase 4: Title Input Tests

- [ ] Test: Title input accepts text
- [ ] Test: Title is optional?
- [ ] Test: Title with special characters
- [ ] Test: Title max length
- [ ] Test: Title saved with source

### Phase 5: Validation Tests

- [ ] Test: Empty text shows error
- [ ] Test: Text > 100K chars shows error
- [ ] Test: Max characters message
- [ ] Test: Error prevents submission
- [ ] Test: Invalid encoding error

### Phase 6: Add Text Tests

- [ ] Test: Enter valid text
- [ ] Test: Enter title (optional)
- [ ] Test: Click add button
- [ ] Test: Loading state visible
- [ ] Test: Source created
- [ ] Test: Toast success appears
- [ ] Test: Modal closes
- [ ] Test: Source in list

### Phase 7: Import File Tests

- [ ] Test: Click import button
- [ ] Test: File dialog opens
- [ ] Test: Select .txt file
- [ ] Test: Content imported
- [ ] Test: Character count updates
- [ ] Test: Select .md file
- [ ] Test: Select .html file
- [ ] Test: Non-text file rejected
- [ ] Test: Import then add

---

## Technical Notes

### Text Source Page Object

```typescript
// frontend/e2e/pages/add-text-source.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class AddTextSourcePage {
  readonly page: Page;
  readonly modal: Locator;
  readonly textTab: Locator;
  readonly titleInput: Locator;
  readonly textTextarea: Locator;
  readonly charCount: Locator;
  readonly wordCount: Locator;
  readonly importButton: Locator;
  readonly addButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modal = page.locator('[data-testid="modal-add-source"]');
    this.textTab = page.locator('[data-testid="tab-text"]');
    this.titleInput = page.locator('[data-testid="input-title"]');
    this.textTextarea = page.locator('[data-testid="textarea-text"]');
    this.charCount = page.locator('[data-testid="char-count"]');
    this.wordCount = page.locator('[data-testid="word-count"]');
    this.importButton = page.locator('[data-testid="btn-import-file"]');
    this.addButton = page.locator('[data-testid="btn-add"]');
    this.errorMessage = page.locator('[data-testid="error-message"]');
  }

  async expectVisible() {
    await expect(this.modal).toBeVisible();
  }

  async clickTextTab() {
    await this.textTab.click();
  }

  async fillTitle(title: string) {
    await this.titleInput.fill(title);
  }

  async fillText(text: string) {
    await this.textTextarea.fill(text);
  }

  async pasteText(text: string) {
    await this.textTextarea.focus();
    await this.page.keyboard.insertText(text);
  }

  async getCharCount(): Promise<string> {
    return this.charCount.textContent();
  }

  async getWordCount(): Promise<string> {
    return this.wordCount.textContent();
  }

  async clickImport() {
    await this.importButton.click();
  }

  async clickAdd() {
    await this.addButton.click();
  }

  async expectCharCount(expected: string) {
    await expect(this.charCount).toHaveText(expected);
  }

  async expectWordCount(expected: string) {
    await expect(this.wordCount).toHaveText(expected);
  }

  async expectError(message: string) {
    await expect(this.errorMessage).toContainText(message);
  }

  async expectNoError() {
    await expect(this.errorMessage).toBeHidden();
  }

  async expectLoading() {
    await expect(this.addButton).toBeDisabled();
    await expect(this.addButton).toHaveText(/Adding.../);
  }

  async addTextSource(title: string, text: string) {
    if (title) await this.fillTitle(title);
    await this.fillText(text);
    await this.clickAdd();
  }
}
```

### Text Source Mocks

```typescript
// frontend/e2e/mocks/sources-text.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockCreateTextSource(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 201,
    body: {
      id: 'source-text-' + Date.now(),
      name: 'Text Source',
      type: 'text',
      content: 'Sample text content...',
      charCount: 23,
      wordCount: 4,
      status: 'completed',
      chunkCount: 1,
      createdAt: new Date().toISOString(),
    },
  });
}

export async function mockEmptyTextError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'VALIDATION_ERROR',
      field: 'content',
      message: 'Text content is required',
    },
  });
}

export async function mockTextTooLongError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'TEXT_TOO_LONG',
      message: 'Text exceeds maximum of 100,000 characters',
      maxLength: 100000,
      currentLength: 100001,
    },
  });
}

export async function mockInvalidEncodingError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'INVALID_ENCODING',
      message: 'File encoding is not supported',
    },
  });
}

export async function mockImportFile(request: APIRequestContext) {
  await request.post('/api/v1/sources/import-text', {
    status: 200,
    body: {
      content: 'Imported text content...',
      charCount: 28,
      wordCount: 5,
      encoding: 'utf-8',
    },
  });
}

export async function mockInvalidFileType(request: APIRequestContext) {
  await request.post('/api/v1/sources/import-text', {
    status: 400,
    body: {
      error: 'INVALID_FILE_TYPE',
      message: 'Only .txt, .md, and .html files are supported',
    },
  });
}
```

### Test Data

```typescript
export const testTexts = {
  short: 'This is a short text.',
  medium: 'This is a medium length text that contains several words for testing purposes.',
  long: 'a'.repeat(1000), // 1000 characters
  veryLong: 'b'.repeat(100000), // 100K characters
  empty: '',
  withSpecialChars: 'Special chars: @#$%^&*()_+-=[]{}|;\':",./<>?',
  withEmojis: 'Text with emojis: 🚀 🎉 💻 🔧',
  multiline: `Line 1
Line 2
Line 3
Line 4`,
};

export const importFiles = {
  validTxt: 'e2e/fixtures/imports/valid.txt',
  validMd: 'e2e/fixtures/imports/document.md',
  validHtml: 'e2e/fixtures/imports/document.html',
  invalidPdf: 'e2e/fixtures/imports/document.pdf',
  invalidJson: 'e2e/fixtures/imports/data.json',
};
```

### Running Specific Tests

```bash
# Run all text source tests
cd frontend && npx playwright test sources-text.spec.ts

# Run validation tests
cd frontend && npx playwright test sources-text.spec.ts -g "validation"

# Run import tests
cd frontend && npx playwright test sources-text.spec.ts -g "import"

# Run in headed mode
cd frontend && npx playwright test sources-text.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] Text input tested
- [ ] Character/word counting tested
- [ ] Validation tested
- [ ] Import functionality tested
- [ ] Success flow tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Counts update in real-time
- [ ] Clear error messages
- [ ] Import flow intuitive
- [ ] Character limit clear

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Counting** - Test character and word counting accuracy
2. **Pasting** - Test paste functionality
3. **Import** - Test file import with various formats
4. **Limits** - Test max length validation

### Common Issues to Avoid

1. **Skipping validation** - Test empty and long text
2. **Not testing import** - Test file import thoroughly
3. **Race conditions** - Wait for counts to update
4. **Hardcoded values** - Use test fixtures

---

## Dependencies

- **Prerequisites**: 13-sources-list.md (add source button)
- **Environment**: Backend API with text source endpoint
- **Test Data**: Various text content and import files

---

## Related Tasks

- 13-sources-list.md - List page with add source button
- 14-sources-url.md - URL source creation
- 15-sources-pdf.md - PDF source creation
- 16-sources-sitemap.md - Sitemap source creation

---

*Task created from: docs/frontend/TEST_PATHS.md Section 5.5*
