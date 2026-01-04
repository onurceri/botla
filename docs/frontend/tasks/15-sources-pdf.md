# Task: Implement PDF Source Tests

> **Task ID**: 15-sources-pdf  
> **Source**: TEST_PATHS.md Section 5.3  
> **Priority**: High (Data Sources)  
> **Estimated Effort**: 8-10 hours  
> **Prerequisite**: 13-sources-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for adding PDF sources. This task covers file upload, drag-and-drop, validation, progress tracking, and OCR options.

### Context

The PDF Source functionality allows users to upload PDF documents as knowledge sources. Testing this functionality ensures:
- File upload works correctly
- Drag-and-drop functionality works
- File validation catches errors
- Progress tracking shows during processing
- OCR option works for scanned documents

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 5.3:

```
Add PDF Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   Click: tab-pdf
│   └── Assert: File upload area visible
│
├── File upload (drag & drop)
│   ├── Drag: PDF file to drop zone
│   ├── Assert: File preview shows
│   ├── Assert: File name displayed
│   ├── Assert: File size displayed
│   ├── Click: btn-upload
│   ├── Assert: Upload progress
│   ├── Assert: Source created
│   └── Assert: Toast "PDF uploaded"
│
├── File upload (click)
│   ├── Click: drop zone
│   ├── Select: PDF file from dialog
│   └── Same as drag & drop
│
├── Multiple files
│   ├── Drag: Multiple PDFs
│   ├── Assert: File list shows all
│   ├── Remove: One file from list
│   ├── Click: btn-upload
│   └── Assert: All files uploaded
│
├── File validation
    ├── Wrong format → Error "PDF only"
    ├── File too large → Error "Max 50MB"
    ├── Corrupted PDF → Error "Invalid PDF"
    └── Encrypted PDF → Error "Password protected"
│
├── Progress indicator
    ├── Show: Upload progress %
    ├── Show: Processing stages
    │   ├── Fetching
    │   ├── Parsing
    │   ├── Chunking
    │   └── Embedding
    └── Show: Completed chunks count
│
└── OCR option (Pro+ plans)
    ├── Toggle: checkbox-enable-ocr
    ├── Click: btn-upload
    ├── Assert: OCR processing
    └── Assert: Better text extraction
```

### Implementation Requirements

1. **Create PDF Source Test File** (`frontend/e2e/sources-pdf.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create PDF Source Page Object** (`frontend/e2e/pages/add-pdf-source.page.ts`)
   - Encapsulate PDF upload interactions
   - File handling methods
   - Progress tracking methods

3. **Create PDF Source Mocks** (`frontend/e2e/mocks/sources-pdf.mocks.ts`)
   - Mock PDF upload endpoint
   - Mock processing stages
   - Mock validation errors

### Expected Deliverables

1. `frontend/e2e/sources-pdf.spec.ts` - Comprehensive PDF source tests
2. `frontend/e2e/pages/add-pdf-source.page.ts` - PDF source page object
3. `frontend/e2e/mocks/sources-pdf.mocks.ts` - PDF source API mock handlers
4. Test PDF fixtures for upload testing

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/add-pdf-source.page.ts`:
  - Modal container locator
  - PDF tab locator
  - Drop zone locator
  - File input locator
  - File list locator
  - Upload button locator
  - OCR checkbox locator
  - Progress bar locators

### Phase 2: Modal Open Tests

- [ ] Test: Click add source opens modal
- [ ] Test: Click PDF tab shows drop zone
- [ ] Test: Drop zone visible
- [ ] Test: Drop zone has instructions
- [ ] Test: Click to upload button visible
- [ ] Test: Supported formats listed

### Phase 3: Drag and Drop Tests

- [ ] Test: Drag PDF file to drop zone
- [ ] Test: Drop zone highlights on drag over
- [ ] Test: File preview appears
- [ ] Test: File name displayed
- [ ] Test: File size displayed
- [ ] Test: File type icon shows
- [ ] Test: Remove button appears
- [ ] Test: Multiple files can be added

### Phase 4: Click Upload Tests

- [ ] Test: Click drop zone opens file dialog
- [ ] Test: Select PDF file
- [ ] Test: File added to list
- [ ] Test: Same as drag & drop behavior

### Phase 5: File Validation Tests

- [ ] Test: Non-PDF file rejected
- [ ] Test: Wrong format error message
- [ ] Test: File > 50MB rejected
- [ ] Test: File size error message
- [ ] Test: Corrupted PDF rejected
- [ ] Test: Invalid PDF error message
- [ ] Test: Encrypted PDF rejected
- [ ] Test: Password protected error message

### Phase 6: Upload Process Tests

- [ ] Test: Click upload button
- [ ] Test: Upload progress visible
- [ ] Test: Progress percentage updates
- [ ] Test: Processing stages shown
- [ ] Test: Fetching stage
- [ ] Test: Parsing stage
- [ ] Test: Chunking stage
- [ ] Test: Embedding stage
- [ ] Test: Completed chunks count
- [ ] Test: Upload button disabled during

### Phase 7: Multiple Files Tests

- [ ] Test: Drag multiple PDFs
- [ ] Test: File list shows all files
- [ ] Test: Remove one file
- [ ] Test: Remaining files stay
- [ ] Test: Upload all files
- [ ] Test: Progress per file
- [ ] Test: Success for all files

### Phase 8: OCR Option Tests

- [ ] Test: OCR checkbox visible
- [ ] Test: OCR toggle works
- [ ] Test: OCR disabled by default
- [ ] Test: OCR enabled for upload
- [ ] Test: OCR processing indicator
- [ ] Test: OCR only on Pro+ plans (check UI)

### Phase 9: Remove File Tests

- [ ] Test: Remove button on file
- [ ] Test: Click removes file
- [ ] Test: File removed from list
- [ ] Test: All files can be removed
- [ ] Test: Upload button disabled when empty

---

## Technical Notes

### PDF Source Page Object

```typescript
// frontend/e2e/pages/add-pdf-source.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class AddPdfSourcePage {
  readonly page: Page;
  readonly modal: Locator;
  readonly pdfTab: Locator;
  readonly dropZone: Locator;
  readonly fileInput: Locator;
  readonly fileList: Locator;
  readonly uploadButton: Locator;
  readonly ocrCheckbox: Locator;
  readonly progressContainer: Locator;
  readonly progressBar: Locator;
  readonly progressPercent: Locator;
  readonly processingStage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modal = page.locator('[data-testid="modal-add-source"]');
    this.pdfTab = page.locator('[data-testid="tab-pdf"]');
    this.dropZone = page.locator('[data-testid="drop-zone-pdf"]');
    this.fileInput = page.locator('[data-testid="input-file"]');
    this.fileList = page.locator('[data-testid="file-list"]');
    this.uploadButton = page.locator('[data-testid="btn-upload"]');
    this.ocrCheckbox = page.locator('[data-testid="checkbox-enable-ocr"]');
    this.progressContainer = page.locator('[data-testid="progress-container"]');
    this.progressBar = page.locator('[data-testid="progress-bar"]');
    this.progressPercent = page.locator('[data-testid="progress-percent"]');
    this.processingStage = page.locator('[data-testid="processing-stage"]');
  }

  async expectVisible() {
    await expect(this.modal).toBeVisible();
  }

  async clickPdfTab() {
    await this.pdfTab.click();
  }

  async expectDropZoneVisible() {
    await expect(this.dropZone).toBeVisible();
  }

  async uploadFile(filePath: string) {
    await this.fileInput.setInputFiles(filePath);
  }

  async uploadMultipleFiles(filePaths: string[]) {
    await this.fileInput.setInputFiles(filePaths);
  }

  async expectFileInList(fileName: string) {
    await expect(
      this.fileList.locator(`[data-testid="file-item"]:has-text("${fileName}")`)
    ).toBeVisible();
  }

  async expectFileCount(count: number) {
    await expect(this.fileList.locator('[data-testid="file-item"]')).toHaveCount(count);
  }

  async removeFile(fileName: string) {
    await this.fileList
      .locator(`[data-testid="file-item"]:has-text("${fileName}")`)
      .locator('[data-testid="btn-remove"]')
      .click();
  }

  async toggleOcr(enabled: boolean) {
    const isChecked = await this.ocrCheckbox.isChecked();
    if (isChecked !== enabled) {
      await this.ocrCheckbox.click();
    }
  }

  async clickUpload() {
    await this.uploadButton.click();
  }

  async expectUploadButtonDisabled() {
    await expect(this.uploadButton).toBeDisabled();
  }

  async expectUploadButtonEnabled() {
    await expect(this.uploadButton).toBeEnabled();
  }

  async expectProgressVisible() {
    await expect(this.progressContainer).toBeVisible();
  }

  async expectProgressPercent(percent: number) {
    await expect(this.progressPercent).toHaveText(`${percent}%`);
  }

  async expectProcessingStage(stage: string) {
    await expect(this.processingStage).toHaveText(new RegExp(stage, 'i'));
  }

  async expectErrorMessage(message: string) {
    await expect(this.page.locator('[data-testid="error-message"]')).toContainText(message);
  }
}
```

### PDF Source Mocks

```typescript
// frontend/e2e/mocks/sources-pdf.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockUploadPdf(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 201,
    body: {
      id: 'source-pdf-' + Date.now(),
      name: 'document.pdf',
      type: 'pdf',
      status: 'processing',
      chunkCount: 0,
      fileSize: 1024000,
      createdAt: new Date().toISOString(),
    },
  });
}

export async function mockUploadPdfComplete(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/sources`, {
    status: 201,
    body: {
      id: 'source-pdf-' + Date.now(),
      name: 'document.pdf',
      type: 'pdf',
      status: 'completed',
      chunkCount: 45,
      fileSize: 1024000,
      createdAt: new Date().toISOString(),
    },
  });
}

export async function mockInvalidFileType(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'INVALID_FILE_TYPE',
      message: 'Only PDF files are allowed',
      acceptedTypes: ['application/pdf'],
    },
  });
}

export async function mockFileTooLarge(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'FILE_TOO_LARGE',
      message: 'File exceeds 50MB limit',
      maxSize: 50 * 1024 * 1024,
    },
  });
}

export async function mockCorruptedPdf(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'INVALID_PDF',
      message: 'Could not parse PDF file',
    },
  });
}

export async function mockEncryptedPdf(request: APIRequestContext) {
  await request.post('/api/v1/chatbots/chatbot-123/sources', {
    status: 400,
    body: {
      error: 'ENCRYPTED_PDF',
      message: 'Password protected PDFs are not supported',
    },
  });
}
```

### Test PDF Fixtures

```typescript
// e2e/fixtures/test-pdfs/
export const testPdfs = {
  valid: 'e2e/fixtures/test-pdfs/valid-document.pdf',
  small: 'e2e/fixtures/test-pdfs/small-document.pdf',
  multiPage: 'e2e/fixtures/test-pdfs/multi-page-document.pdf',
  corrupted: 'e2e/fixtures/test-pdfs/corrupted.pdf',
  encrypted: 'e2e/fixtures/test-pdfs/encrypted.pdf',
  tooLarge: 'e2e/fixtures/test-pdfs/large-file.pdf', // > 50MB
  notPdf: 'e2e/fixtures/test-pdfs/document.txt',
};
```

### Running Specific Tests

```bash
# Run all PDF source tests
cd frontend && npx playwright test sources-pdf.spec.ts

# Run upload tests
cd frontend && npx playwright test sources-pdf.spec.ts -g "upload"

# Run validation tests
cd frontend && npx playwright test sources-pdf.spec.ts -g "validation"

# Run in headed mode
cd frontend && npx playwright test sources-pdf.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All upload methods tested
- [ ] All validations tested
- [ ] Progress tracking tested
- [ ] Multiple files tested
- [ ] OCR option tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with file fixtures
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Drop zone responsive
- [ ] Progress clear
- [ ] Error messages helpful
- [ ] File list intuitive

---

## Execution Notes for Developer Agent

### Key Considerations

1. **File Fixtures** - Need actual PDF files for testing
2. **Upload Timing** - Progress tests need careful timing
3. **File Size** - Test with various file sizes
4. **Multiple Files** - Test queue behavior

### Common Issues to Avoid

1. **Missing fixtures** - Create test PDF files
2. **Skipping validation** - Test all error cases
3. **Race conditions** - Wait for uploads to complete
4. **Large files** - Don't use actual large files

---

## Dependencies

- **Prerequisites**: 13-sources-list.md (add source button)
- **Environment**: Backend API with file upload
- **Test Data**: PDF test files of various types

---

## Related Tasks

- 13-sources-list.md - List page with add source button
- 14-sources-url.md - URL source creation
- 16-sources-sitemap.md - Sitemap source creation
- 17-sources-text.md - Text source creation

---

*Task created from: docs/frontend/TEST_PATHS.md Section 5.3*
