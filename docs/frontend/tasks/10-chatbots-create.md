# Task: Implement Chatbot Creation Tests

> **Task ID**: 10-chatbots-create  
> **Source**: TEST_PATHS.md Section 4.2  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 8-10 hours  
> **Prerequisite**: 09-chatbots-list.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Chatbot Creation Flow. This task covers the create chatbot modal, form validation, and successful creation.

### Context

The Chatbot Creation Flow allows users to create new chatbots with various configuration options. Testing this flow ensures:
- Users can successfully create chatbots
- Form validation works correctly
- Default values are set properly
- API integration works as expected
- Error handling is proper

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 4.2:

#### 4.2.1 Create Dialog Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `modal-create-chatbot` | modal | Create modal |
| `input-name` | text | Chatbot name |
| `input-description` | textarea | Description |
| `select-language` | select | Default language |
| `select-model` | select | AI model |
| `slider-temperature` | slider | Temperature (0-2) |
| `input-max-tokens` | number | Max tokens |
| `btn-create` | submit | Create button |
| `btn-cancel` | button | Cancel button |

#### 4.2.2 Create Flow

```
Create Chatbot Flow
├── Open create modal
│   ├── Click: btn-create-chatbot
│   ├── Assert: `modal-create-chatbot` visible
│   └── Assert: Focus on `input-name`
│
├── Fill form
│   ├── Type: Name (required)
│   │   ├── Min: 1 character
│   │   ├── Max: 100 characters
│   │   └── Validation: On blur
│   │
│   ├── Type: Description (optional)
│   │   ├── Max: 500 characters
│   │   └── Validation: On blur
│   │
│   ├── Select: Language (default: tr)
│   │   ├── tr (Türkçe)
│   │   └── en (English)
│   │
│   ├── Select: Model (default: gpt-4o-mini)
│   │   ├── gpt-4o-mini
│   │   ├── gpt-4o
│   │   └── gpt-5 (if ultra)
│   │
│   ├── Adjust: Temperature (default: 0.7)
│   │   ├── Slider: 0.0 to 2.0
│   │   ├── Show: Value label
│   │   └── Hover: Slider track (highlight)
│   │
│   └── Input: Max tokens (default: 1000)
│       ├── Min: 100
│       ├── Max: 8000
│       └── Validation: On change
│
├── Submit validation
│   ├── Click: btn-create (without name)
│   │   ├── Assert: `input-name` error
│   │   └── Assert: `toast-error` - "Name required"
│   │
│   ├── Click: btn-create (valid form)
│   │   ├── Assert: Loading state
│   │   ├── Assert: btn-create disabled
│   │   ├── API: Create chatbot
│   │   ├── Assert: Chatbot in database
│   │   ├── Assert: Source count = 0
│   │   ├── Assert: Created with defaults
│   │   ├── Close: Modal
│   │   ├── Assert: Toast success
│   │   └── Navigate: /dashboard/chatbots/{new-id}
│   │
│   └── Click: btn-cancel
│       ├── Close: Modal
│       └── Assert: No chatbot created
│
└── Keyboard shortcuts (in modal)
    ├── Escape: Close modal
    ├── Enter: Submit (if form valid)
    └── Tab: Navigate form fields
```

### Implementation Requirements

1. **Create Chatbot Create Test File** (`frontend/e2e/chatbot-create.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Chatbot Create Page Object** (`frontend/e2e/pages/chatbot-create.page.ts`)
   - Encapsulate modal interactions
   - Form field methods
   - Validation assertion methods

3. **Create Chatbot Create Mocks** (`frontend/e2e/mocks/chatbot-create.mocks.ts`)
   - Mock create chatbot API
   - Mock validation errors
   - Handle form submission

### Expected Deliverables

1. `frontend/e2e/chatbot-create.spec.ts` - Comprehensive create tests
2. `frontend/e2e/pages/chatbot-create.page.ts` - Create modal page object
3. `frontend/e2e/mocks/chatbot-create.mocks.ts` - Create API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/chatbot-create.page.ts`:
  - Modal container locator
  - Name input locator
  - Description input locator
  - Language select locator
  - Model select locator
  - Temperature slider locator
  - Max tokens input locator
  - Create button locator
  - Cancel button locator
  - Error message locators

### Phase 2: Modal Open/Close Tests

- [ ] Test: Click create button opens modal
- [ ] Test: Modal is visible
- [ ] Test: Focus is on name input
- [ ] Test: All form fields visible
- [ ] Test: Click overlay closes modal
- [ ] Test: Click cancel closes modal
- [ ] Test: Escape key closes modal
- [ ] Test: Modal has correct title
- [ ] Test: Modal has proper ARIA attributes

### Phase 3: Form Field Tests

- [ ] Test: Name input accepts text
- [ ] Test: Name input has correct placeholder
- [ ] Test: Description input accepts text
- [ ] Test: Description has character limit display
- [ ] Test: Language select has options (tr, en)
- [ ] Test: Language defaults to tr
- [ ] Test: Model select has options
- [ ] Test: Model defaults to gpt-4o-mini
- [ ] Test: Temperature slider visible
- [ ] Test: Temperature shows value label
- [ ] Test: Max tokens input accepts numbers
- [ ] Test: Max tokens defaults to 1000

### Phase 4: Temperature Slider Tests

- [ ] Test: Slider can be dragged
- [ ] Test: Slider value updates on drag
- [ ] Test: Temperature range 0.0 to 2.0
- [ ] Test: Value label updates in real-time
- [ ] Test: Hover on slider shows highlight
- [ ] Test: Click on slider track sets value
- [ ] Test: Keyboard adjustment (arrow keys)
- [ ] Test: Step increment (0.1)

### Phase 5: Form Validation Tests

- [ ] Test: Submit with empty name shows error
- [ ] Test: Name error message displayed
- [ ] Test: Name input has error class
- [ ] Test: Name validation on blur
- [ ] Test: Name max 100 characters
- [ ] Test: Description max 500 characters
- [ ] Test: Max tokens min 100
- [ ] Test: Max tokens max 8000
- [ ] Test: Validation on change
- [ ] Test: Validation on blur

### Phase 6: Successful Creation Tests

- [ ] Test: Fill form with valid data
- [ ] Test: Click create button
- [ ] Test: Loading state visible
- [ ] Test: Create button disabled
- [ ] Test: API call made to create chatbot
- [ ] Test: New chatbot in database
- [ ] Test: Source count = 0
- [ ] Test: Toast success appears
- [ ] Test: Modal closes
- [ ] Test: Navigate to new chatbot detail
- [ ] Test: New chatbot appears in list (if redirected back)

### Phase 7: Cancel Creation Tests

- [ ] Test: Click cancel with no changes
- [ ] Test: Click cancel with changes shows confirm
- [ ] Test: Confirm discard closes modal
- [ ] Test: Click keep editing
- [ ] Test: No API call on cancel
- [ ] Test: No chatbot created

### Phase 8: Keyboard Navigation Tests

- [ ] Test: Tab navigates through fields
- [ ] Test: Shift+Tab navigates backward
- [ ] Test: Enter submits when form valid
- [ ] Test: Enter doesn't submit when invalid
- [ ] Test: Escape closes modal
- [ ] Test: Focus trapped in modal
- [ ] Test: Focus returns to trigger on close

### Phase 9: Default Values Tests

- [ ] Test: Language defaults to tr
- [ ] Test: Model defaults to gpt-4o-mini
- [ ] Test: Temperature defaults to 0.7
- [ ] Test: Max tokens defaults to 1000
- [ ] Test: Description is optional
- [ ] Test: All defaults can be changed

---

## Technical Notes

### Chatbot Create Page Object

```typescript
// frontend/e2e/pages/chatbot-create.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ChatbotCreatePage {
  readonly page: Page;
  readonly modal: Locator;
  readonly nameInput: Locator;
  readonly descriptionInput: Locator;
  readonly languageSelect: Locator;
  readonly modelSelect: Locator;
  readonly temperatureSlider: Locator;
  readonly temperatureValue: Locator;
  readonly maxTokensInput: Locator;
  readonly createButton: Locator;
  readonly cancelButton: Locator;
  readonly nameError: Locator;
  readonly modalOverlay: Locator;

  constructor(page: Page) {
    this.page = page;
    this.modal = page.locator('[data-testid="modal-create-chatbot"]');
    this.nameInput = page.locator('[data-testid="input-name"]');
    this.descriptionInput = page.locator('[data-testid="input-description"]');
    this.languageSelect = page.locator('[data-testid="select-language"]');
    this.modelSelect = page.locator('[data-testid="select-model"]');
    this.temperatureSlider = page.locator('[data-testid="slider-temperature"]');
    this.temperatureValue = page.locator('[data-testid="temperature-value"]');
    this.maxTokensInput = page.locator('[data-testid="input-max-tokens"]');
    this.createButton = page.locator('[data-testid="btn-create"]');
    this.cancelButton = page.locator('[data-testid="btn-cancel"]');
    this.nameError = page.locator('[data-testid="error-name"]');
    this.modalOverlay = page.locator('[data-testid="modal-overlay"]');
  }

  async expectVisible() {
    await expect(this.modal).toBeVisible();
  }

  async expectHidden() {
    await expect(this.modal).toBeHidden();
  }

  async expectNameFocused() {
    await expect(this.nameInput).toBeFocused();
  }

  async fillName(name: string) {
    await this.nameInput.fill(name);
  }

  async fillDescription(description: string) {
    await this.descriptionInput.fill(description);
  }

  async selectLanguage(language: 'tr' | 'en') {
    await this.languageSelect.selectOption(language);
  }

  async selectModel(model: string) {
    await this.modelSelect.selectOption(model);
  }

  async setTemperature(value: number) {
    // Click on slider track to set value
    const sliderBox = await this.temperatureSlider.boundingBox();
    if (sliderBox) {
      const percentage = (value - 0) / (2 - 0);
      const x = sliderBox.x + (percentage * sliderBox.width);
      await this.page.mouse.click(x, sliderBox.y + sliderBox.height / 2);
    }
  }

  async fillMaxTokens(tokens: number) {
    await this.maxTokensInput.fill(tokens.toString());
  }

  async clickCreate() {
    await this.createButton.click();
  }

  async clickCancel() {
    await this.cancelButton.click();
  }

  async expectLoading() {
    await expect(this.createButton).toBeDisabled();
    await expect(this.createButton).toHaveText(/Creating.../);
  }

  async expectNameError(message: string) {
    await expect(this.nameError).toHaveText(message);
    await expect(this.nameInput).toHaveClass(/error/);
  }

  async expectTemperatureValue(value: string) {
    await expect(this.temperatureValue).toHaveText(value);
  }

  async expectLanguageValue(value: string) {
    await expect(this.languageSelect).toHaveValue(value);
  }

  async expectModelValue(value: string) {
    await expect(this.modelSelect).toHaveValue(value);
  }

  async expectDefaultValues() {
    await this.expectLanguageValue('tr');
    await this.expectModelValue('gpt-4o-mini');
    await this.expectTemperatureValue('0.7');
    await expect(this.maxTokensInput).toHaveValue('1000');
  }

  async createChatbot(data: {
    name: string;
    description?: string;
    language?: string;
    model?: string;
    temperature?: number;
    maxTokens?: number;
  }) {
    await this.fillName(data.name);
    if (data.description) {
      await this.fillDescription(data.description);
    }
    if (data.language) {
      await this.selectLanguage(data.language as 'tr' | 'en');
    }
    if (data.model) {
      await this.selectModel(data.model);
    }
    if (data.temperature !== undefined) {
      await this.setTemperature(data.temperature);
    }
    if (data.maxTokens) {
      await this.fillMaxTokens(data.maxTokens);
    }
    await this.clickCreate();
  }
}
```

### Create API Mocks

```typescript
// frontend/e2e/mocks/chatbot-create.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockSuccessfulCreate(request: APIRequestContext) {
  await request.post('/api/v1/chatbots', {
    status: 201,
    body: {
      id: 'chatbot-new-' + Date.now(),
      name: 'New Chatbot',
      description: 'Test description',
      language: 'tr',
      model: 'gpt-4o-mini',
      temperature: 0.7,
      maxTokens: 1000,
      status: 'active',
      sourceCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
  });
}

export async function mockValidationError(request: APIRequestContext, field: string, message: string) {
  await request.post('/api/v1/chatbots', {
    status: 400,
    body: {
      error: 'VALIDATION_ERROR',
      field,
      message,
    },
  });
}

export async function mockDuplicateNameError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots', {
    status: 409,
    body: {
      error: 'CONFLICT',
      field: 'name',
      message: 'Chatbot with this name already exists',
    },
  });
}

export async function mockServerError(request: APIRequestContext) {
  await request.post('/api/v1/chatbots', {
    status: 500,
    body: {
      error: 'INTERNAL_ERROR',
      message: 'An error occurred while creating chatbot',
    },
  });
}
```

### Form Test Data

```typescript
// frontend/e2e/fixtures/chatbot.fixture.ts
export const validChatbotData = {
  name: 'Test Chatbot',
  description: 'A test chatbot for E2E testing',
  language: 'en' as const,
  model: 'gpt-4o-mini',
  temperature: 0.7,
  maxTokens: 1000,
};

export const invalidChatbotData = {
  emptyName: { name: '' },
  longName: { name: 'a'.repeat(101) },
  longDescription: { description: 'b'.repeat(501) },
  lowTokens: { maxTokens: 50 },
  highTokens: { maxTokens: 9000 },
};

export const modelOptions = [
  'gpt-4o-mini',
  'gpt-4o',
  'gpt-5',
];

export const languageOptions = [
  { value: 'tr', label: 'Türkçe' },
  { value: 'en', label: 'English' },
];
```

### Running Specific Tests

```bash
# Run all create tests
cd frontend && npx playwright test chatbot-create.spec.ts

# Run modal tests
cd frontend && npx playwright test chatbot-create.spec.ts -g "modal"

# Run validation tests
cd frontend && npx playwright test chatbot-create.spec.ts -g "validation"

# Run successful creation tests
cd frontend && npx playwright test chatbot-create.spec.ts -g "successful"

# Run in headed mode
cd frontend && npx playwright test chatbot-create.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All modal open/close tested
- [ ] All form fields tested
- [ ] All validations tested
- [ ] Success flow tested
- [ ] Cancel flow tested
- [ ] Keyboard navigation tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Focus management correct
- [ ] Loading states visible
- [ ] Error messages clear
- [ ] Default values reasonable

### 4. Accessibility Verification
- [ ] Keyboard navigation works
- [ ] ARIA labels present
- [ ] Focus trap in modal
- [ ] Screen reader compatible

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Modal State** - Ensure modal is properly opened/closed between tests
2. **API Mocking** - Mock create endpoint for success/error scenarios
3. **Focus Testing** - Verify focus management in modal
4. **Slider Testing** - Be careful with slider interaction tests

### Common Issues to Avoid

1. **Race conditions** - Wait for modal to appear
2. **Not waiting for API** - Wait for create response
3. **Hardcoded values** - Use fixture data
4. **Skipping keyboard tests** - Test keyboard navigation

### Test Setup

```typescript
// Open modal before tests
test.beforeEach(async ({ page }) => {
  await page.goto('/dashboard/chatbots');
  await page.click('[data-testid="btn-create-chatbot"]');
  await expect(page.locator('[data-testid="modal-create-chatbot"]')).toBeVisible();
});
```

---

## Dependencies

- **Prerequisites**: 09-chatbots-list.md (for navigation and create button)
- **Environment**: Backend API with chatbot creation endpoint
- **Test Data**: Various chatbot configurations

---

## Related Tasks

- 09-chatbots-list.md - List page with create button
- 11-chatbots-detail.md - Navigate to detail after creation
- 12-chatbots-settings.md - Settings accessible from actions menu

---

*Task created from: docs/frontend/TEST_PATHS.md Section 4.2*
