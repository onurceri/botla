# Task: Implement Chatbot Settings Tests

> **Task ID**: 12-chatbots-settings  
> **Source**: TEST_PATHS.md Section 4.4  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 12-16 hours  
> **Prerequisite**: 11-chatbots-detail.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Chatbot Settings Page. This task covers all 9 settings sections: Identity, Instructions, Language & Model, Appearance, Suggestions, Branding, Guardrails, Handoff, and Security.

### Context

The Chatbot Settings Page allows users to configure all aspects of their chatbot. Testing this page ensures:
- All settings sections load correctly
- Form validation works for all fields
- Save functionality persists changes
- Visual customization works properly
- Security settings are properly implemented

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 4.4:

#### 4.4.1 Settings Sections

```
Settings Page Sections
├── 1. Identity Section
│   ├── input-name (edit)
│   ├── input-description (edit)
│   ├── input-bot-display-name (edit)
│   ├── input-bot-icon (upload)
│   └── btn-save-identity
│
├── 2. Instructions Section
│   ├── textarea-custom-instruction (wysiwyg)
│   └── btn-save-instructions
│
├── 3. Language & Model Section
│   ├── select-language
│   ├── select-model
│   ├── slider-temperature
│   ├── input-max-tokens
│   └── btn-save-params
│
├── 4. Appearance Section
│   ├── color-theme (color picker)
│   ├── input-welcome-message (textarea)
│   ├── select-position
│   ├── color-bot-message
│   ├── color-user-message
│   ├── input-font-family
│   └── btn-save-appearance
│
├── 5. Suggestions Section
│   ├── toggle-suggestions-enabled
│   ├── textarea-suggested-questions
│   └── btn-save-suggestions
│
├── 6. Branding Section
│   ├── toggle-hide-branding
│   ├── input-logo-url
│   ├── input-brand-text
│   ├── input-brand-link
│   └── btn-save-branding
│
├── 7. Guardrails Section
│   ├── slider-confidence-threshold
│   ├── textarea-no-info-message
│   ├── textarea-error-message
│   ├── input-allowed-topics
│   ├── input-blocked-topics
│   └── btn-save-guardrails
│
├── 8. Handoff Section
│   ├── toggle-handoff-enabled
│   ├── select-handoff-type
│   ├── textarea-handoff-message
│   └── btn-save-handoff
│
└── 9. Security Section
    ├── toggle-secure-embed
    ├── textarea-allowed-domains
    ├── btn-regenerate-secret
    └── btn-save-security
```

#### 4.4.2 Identity Section Tests

```
Identity Section Flow
├── Edit name
│   ├── Click: input-name
│   ├── Clear: Existing name
│   ├── Type: New name
│   ├── Click: btn-save-identity
│   ├── Assert: Loading state
│   ├── Assert: Toast success
│   └── Assert: Name updated in DB
│
├── Edit description
│   ├── Click: input-description
│   ├── Type: New description
│   ├── Click: btn-save-identity
│   └── Assert: Description updated
│
├── Upload bot icon
│   ├── Click: input-bot-icon (file input)
│   ├── Select: Image file
│   ├── Assert: Preview shows image
│   ├── Assert: File size validation
│   ├── Assert: File type validation
│   ├── Click: btn-save-identity
│   └── Assert: Icon URL saved
│
└── Validation
    ├── Empty name → Error
    ├── Name > 100 chars → Error
    └── Invalid URL → Error
```

#### 4.4.3 Appearance Section Tests

```
Appearance Section Flow
├── Change theme color
│   ├── Click: color-theme (color picker)
│   ├── Assert: Color picker dropdown opens
│   ├── Select: Color from palette
│   │   ├── Assert: Color preview updates
│   │   └── Click: Outside picker to close
│   ├── Type: Hex color directly
│   │   ├── Assert: Valid hex format
│   │   └── Assert: Color updates
│   └── Click: btn-save-appearance
│       └── Assert: Theme saved
│
├── Change position
│   ├── Click: select-position
│   ├── Options:
│   │   ├── bottom-right
│   │   └── bottom-left
│   ├── Select: bottom-left
│   └── Click: btn-save-appearance
│       └── Assert: Position saved
│
├── Change message colors
│   ├── Click: color-bot-message
│   ├── Select: Bot message color
│   ├── Click: color-user-message
│   ├── Select: User message color
│   └── Click: btn-save-appearance
│       └── Assert: Colors saved
│
├── Change font family
│   ├── Click: select-font-family
│   ├── Options:
│   │   ├── System default
│   │   ├── Inter
│   │   ├── Roboto
│   │   └── Custom (input)
│   ├── Select: Inter
│   └── Click: btn-save-appearance
│       └── Assert: Font saved
│
└── Welcome message
    ├── Click: input-welcome-message
    ├── Type: Custom welcome
    ├── Click: btn-save-appearance
    └── Assert: Welcome message saved
```

#### 4.4.4 Suggestions Section Tests

```
Suggestions Section Flow
├── Toggle suggestions
│   ├── Click: toggle-suggestions-enabled
│   ├── Assert: Toggle state changes
│   └── Assert: Suggestions input enabled/disabled
│
├── Add suggested questions
│   ├── Click: textarea-suggested-questions
│   ├── Type: Question 1
│   ├── Press: Enter
│   ├── Type: Question 2
│   ├── Press: Enter
│   ├── Type: Question 3
│   ├── Click: btn-save-suggestions
│   ├── Assert: Toast success
│   └── Assert: Questions saved
│
├── Edit suggested question
│   ├── Hover: Question item
│   ├── Click: Edit icon
│   ├── Modify: Question text
│   ├── Click: Save
│   └── Assert: Question updated
│
├── Delete suggested question
│   ├── Hover: Question item
│   ├── Click: Delete icon
│   └── Assert: Question removed
│
└── Reorder questions
    ├── Drag: Question item
    ├── Drop: New position
    └── Assert: Order saved
```

#### 4.4.5 Guardrails Section Tests

```
Guardrails Section Flow
├── Adjust confidence threshold
│   ├── Click: slider-confidence-threshold
│   ├── Drag: To 0.6
│   ├── Assert: Value label = 0.6
│   └── Click: btn-save-guardrails
│       └── Assert: Threshold saved
│
├── Configure fallback messages
│   ├── Click: textarea-no-info-message
│   ├── Type: "I couldn't find information..."
│   ├── Click: textarea-error-message
│   ├── Type: "Something went wrong..."
│   └── Click: btn-save-guardrails
│       └── Assert: Messages saved
│
├── Set topic restrictions
│   ├── Click: input-allowed-topics
│   ├── Type: "product, pricing, features"
│   ├── Click: input-blocked-topics
│   ├── Type: "politics, religion"
│   └── Click: btn-save-guardrails
│       └── Assert: Topics saved
│
└── Toggle threshold warnings
    ├── Click: toggle-show-warnings
    └── Assert: Toggle state saved
```

#### 4.4.6 Security Section Tests

```
Security Section Flow
├── Toggle secure embed
│   ├── Click: toggle-secure-embed
│   ├── Assert: Toggle enabled
│   ├── Click: btn-save-security
│   └── Assert: Secure embed enabled
│
├── Set allowed domains
│   ├── Click: textarea-allowed-domains
│   ├── Type: "example.com, www.example.com"
│   ├── Click: btn-save-security
│   └── Assert: Domains saved
│
├── Regenerate embed secret
│   ├── Click: btn-regenerate-secret
│   ├── Assert: `modal-confirm-regenerate` opens
│   ├── Click: btn-confirm
│   ├── Assert: New secret generated
│   ├── Assert: Toast success
│   └── Assert: Old secret invalidated
│
└── View embed secret
    ├── Click: btn-show-secret (eye icon)
    ├── Assert: Secret visible
    ├── Click: btn-copy-secret
    └── Assert: Toast "Copied to clipboard"
```

### Implementation Requirements

1. **Create Chatbot Settings Test File** (`frontend/e2e/chatbot-settings.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Chatbot Settings Page Object** (`frontend/e2e/pages/chatbot-settings.page.ts`)
   - Encapsulate all settings section interactions
   - Section navigation methods
   - Form field methods for each section

3. **Create Settings Mocks** (`frontend/e2e/mocks/chatbot-settings.mocks.ts`)
   - Mock settings API endpoints
   - Mock save operations
   - Mock validation errors

### Expected Deliverables

1. `frontend/e2e/chatbot-settings.spec.ts` - Comprehensive settings tests
2. `frontend/e2e/pages/chatbot-settings.page.ts` - Settings page object
3. `frontend/e2e/mocks/chatbot-settings.mocks.ts` - Settings API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/chatbot-settings.page.ts` with:
  - All section locators
  - All form field locators
  - Save button locators
  - Toggle locators
  - Color picker locators
  - Slider locators

### Phase 2: Identity Section Tests

- [ ] Test: Identity section loads
- [ ] Test: Name field editable
- [ ] Test: Description field editable
- [ ] Test: Bot display name field editable
- [ ] Test: Bot icon upload
- [ ] Test: File type validation
- [ ] Test: File size validation
- [ ] Test: Image preview
- [ ] Test: Save identity changes
- [ ] Test: Validation - empty name
- [ ] Test: Validation - name too long
- [ ] Test: Validation - invalid URL

### Phase 3: Instructions Section Tests

- [ ] Test: Instructions section loads
- [ ] Test: WYSIWYG editor loads
- [ ] Test: Custom instruction text
- [ ] Test: Save instructions
- [ ] Test: Instructions persist

### Phase 4: Language & Model Section Tests

- [ ] Test: Language select options
- [ ] Test: Model select options
- [ ] Test: Temperature slider
- [ ] Test: Max tokens input
- [ ] Test: Save parameters
- [ ] Test: Validation - out of range tokens

### Phase 5: Appearance Section Tests

- [ ] Test: Theme color picker
- [ ] Test: Color selection from palette
- [ ] Test: Hex color input
- [ ] Test: Position select
- [ ] Test: Bot message color
- [ ] Test: User message color
- [ ] Test: Font family select
- [ ] Test: Welcome message textarea
- [ ] Test: Save appearance
- [ ] Test: Preview updates

### Phase 6: Suggestions Section Tests

- [ ] Test: Suggestions toggle
- [ ] Test: Toggle enables/disables input
- [ ] Test: Add suggested questions
- [ ] Test: Delete question
- [ ] Test: Edit question
- [ ] Test: Reorder questions (drag and drop)
- [ ] Test: Save suggestions
- [ ] Test: Validation - empty question

### Phase 7: Branding Section Tests

- [ ] Test: Hide branding toggle
- [ ] Test: Logo URL input
- [ ] Test: Brand text input
- [ ] Test: Brand link input
- [ ] Test: Save branding
- [ ] Test: Validation - invalid URL

### Phase 8: Guardrails Section Tests

- [ ] Test: Confidence threshold slider
- [ ] Test: No info message textarea
- [ ] Test: Error message textarea
- [ ] Test: Allowed topics input
- [ ] Test: Blocked topics input
- [ ] Test: Save guardrails
- [ ] Test: Slider value display

### Phase 9: Handoff Section Tests

- [ ] Test: Handoff toggle
- [ ] Test: Handoff type select
- [ ] Test: Handoff message textarea
- [ ] Test: Save handoff settings
- [ ] Test: Toggle enables/disables options

### Phase 10: Security Section Tests

- [ ] Test: Secure embed toggle
- [ ] Test: Allowed domains textarea
- [ ] Test: Regenerate secret button
- [ ] Test: Confirm modal for regeneration
- [ ] Test: Show secret (eye icon)
- [ ] Test: Copy secret to clipboard
- [ ] Test: Save security settings
- [ ] Test: Validation - invalid domains

---

## Technical Notes

### Chatbot Settings Page Object

```typescript
// frontend/e2e/pages/chatbot-settings.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ChatbotSettingsPage {
  readonly page: Page;
  
  // Identity Section
  readonly nameInput: Locator;
  readonly descriptionInput: Locator;
  readonly botDisplayNameInput: Locator;
  readonly botIconInput: Locator;
  readonly saveIdentityButton: Locator;
  
  // Instructions Section
  readonly instructionsEditor: Locator;
  readonly saveInstructionsButton: Locator;
  
  // Language & Model Section
  readonly languageSelect: Locator;
  readonly modelSelect: Locator;
  readonly temperatureSlider: Locator;
  readonly temperatureValue: Locator;
  readonly maxTokensInput: Locator;
  readonly saveParamsButton: Locator;
  
  // Appearance Section
  readonly themeColorPicker: Locator;
  readonly welcomeMessageInput: Locator;
  readonly positionSelect: Locator;
  readonly botMessageColor: Locator;
  readonly userMessageColor: Locator;
  readonly fontFamilySelect: Locator;
  readonly saveAppearanceButton: Locator;
  
  // Suggestions Section
  readonly suggestionsToggle: Locator;
  readonly suggestedQuestionsInput: Locator;
  readonly saveSuggestionsButton: Locator;
  
  // Branding Section
  readonly hideBrandingToggle: Locator;
  readonly logoUrlInput: Locator;
  readonly brandTextInput: Locator;
  readonly brandLinkInput: Locator;
  readonly saveBrandingButton: Locator;
  
  // Guardrails Section
  readonly confidenceThresholdSlider: Locator;
  readonly confidenceValue: Locator;
  readonly noInfoMessageInput: Locator;
  readonly errorMessageInput: Locator;
  readonly allowedTopicsInput: Locator;
  readonly blockedTopicsInput: Locator;
  readonly saveGuardrailsButton: Locator;
  
  // Handoff Section
  readonly handoffToggle: Locator;
  readonly handoffTypeSelect: Locator;
  readonly handoffMessageInput: Locator;
  readonly saveHandoffButton: Locator;
  
  // Security Section
  readonly secureEmbedToggle: Locator;
  readonly allowedDomainsInput: Locator;
  readonly regenerateSecretButton: Locator;
  readonly showSecretButton: Locator;
  readonly copySecretButton: Locator;
  readonly saveSecurityButton: Locator;

  constructor(page: Page) {
    this.page = page;
    
    // Identity
    this.nameInput = page.locator('[data-testid="input-name"]');
    this.descriptionInput = page.locator('[data-testid="input-description"]');
    this.botDisplayNameInput = page.locator('[data-testid="input-bot-display-name"]');
    this.botIconInput = page.locator('[data-testid="input-bot-icon"]');
    this.saveIdentityButton = page.locator('[data-testid="btn-save-identity"]');
    
    // Instructions
    this.instructionsEditor = page.locator('[data-testid="textarea-custom-instruction"]');
    this.saveInstructionsButton = page.locator('[data-testid="btn-save-instructions"]');
    
    // Language & Model
    this.languageSelect = page.locator('[data-testid="select-language"]');
    this.modelSelect = page.locator('[data-testid="select-model"]');
    this.temperatureSlider = page.locator('[data-testid="slider-temperature"]');
    this.temperatureValue = page.locator('[data-testid="temperature-value"]');
    this.maxTokensInput = page.locator('[data-testid="input-max-tokens"]');
    this.saveParamsButton = page.locator('[data-testid="btn-save-params"]');
    
    // Appearance
    this.themeColorPicker = page.locator('[data-testid="color-theme"]');
    this.welcomeMessageInput = page.locator('[data-testid="input-welcome-message"]');
    this.positionSelect = page.locator('[data-testid="select-position"]');
    this.botMessageColor = page.locator('[data-testid="color-bot-message"]');
    this.userMessageColor = page.locator('[data-testid="color-user-message"]');
    this.fontFamilySelect = page.locator('[data-testid="select-font-family"]');
    this.saveAppearanceButton = page.locator('[data-testid="btn-save-appearance"]');
    
    // Suggestions
    this.suggestionsToggle = page.locator('[data-testid="toggle-suggestions-enabled"]');
    this.suggestedQuestionsInput = page.locator('[data-testid="textarea-suggested-questions"]');
    this.saveSuggestionsButton = page.locator('[data-testid="btn-save-suggestions"]');
    
    // Branding
    this.hideBrandingToggle = page.locator('[data-testid="toggle-hide-branding"]');
    this.logoUrlInput = page.locator('[data-testid="input-logo-url"]');
    this.brandTextInput = page.locator('[data-testid="input-brand-text"]');
    this.brandLinkInput = page.locator('[data-testid="input-brand-link"]');
    this.saveBrandingButton = page.locator('[data-testid="btn-save-branding"]');
    
    // Guardrails
    this.confidenceThresholdSlider = page.locator('[data-testid="slider-confidence-threshold"]');
    this.confidenceValue = page.locator('[data-testid="confidence-value"]');
    this.noInfoMessageInput = page.locator('[data-testid="textarea-no-info-message"]');
    this.errorMessageInput = page.locator('[data-testid="textarea-error-message"]');
    this.allowedTopicsInput = page.locator('[data-testid="input-allowed-topics"]');
    this.blockedTopicsInput = page.locator('[data-testid="input-blocked-topics"]');
    this.saveGuardrailsButton = page.locator('[data-testid="btn-save-guardrails"]');
    
    // Handoff
    this.handoffToggle = page.locator('[data-testid="toggle-handoff-enabled"]');
    this.handoffTypeSelect = page.locator('[data-testid="select-handoff-type"]');
    this.handoffMessageInput = page.locator('[data-testid="textarea-handoff-message"]');
    this.saveHandoffButton = page.locator('[data-testid="btn-save-handoff"]');
    
    // Security
    this.secureEmbedToggle = page.locator('[data-testid="toggle-secure-embed"]');
    this.allowedDomainsInput = page.locator('[data-testid="textarea-allowed-domains"]');
    this.regenerateSecretButton = page.locator('[data-testid="btn-regenerate-secret"]');
    this.showSecretButton = page.locator('[data-testid="btn-show-secret"]');
    this.copySecretButton = page.locator('[data-testid="btn-copy-secret"]');
    this.saveSecurityButton = page.locator('[data-testid="btn-save-security"]');
  }

  async goto(chatbotId: string) {
    await this.page.goto(`/dashboard/chatbots/${chatbotId}/settings`);
  }

  // Identity methods
  async updateName(name: string) {
    await this.nameInput.clear();
    await this.nameInput.fill(name);
  }

  async updateDescription(description: string) {
    await this.descriptionInput.clear();
    await this.descriptionInput.fill(description);
  }

  async uploadBotIcon(filePath: string) {
    await this.botIconInput.setInputFiles(filePath);
  }

  async saveIdentity() {
    await this.saveIdentityButton.click();
  }

  // Appearance methods
  async selectThemeColor(color: string) {
    await this.themeColorPicker.click();
    await this.page.locator(`[data-testid="color-option"]:has-text("${color}")`).click();
  }

  async selectPosition(position: 'bottom-right' | 'bottom-left') {
    await this.positionSelect.selectOption(position);
  }

  async saveAppearance() {
    await this.saveAppearanceButton.click();
  }

  // Suggestions methods
  async toggleSuggestions(enabled: boolean) {
    const isChecked = await this.suggestionsToggle.isChecked();
    if (isChecked !== enabled) {
      await this.suggestionsToggle.click();
    }
  }

  async addSuggestedQuestion(question: string) {
    await this.suggestedQuestionsInput.fill(question);
    await this.suggestedQuestionsInput.press('Enter');
  }

  async saveSuggestions() {
    await this.saveSuggestionsButton.click();
  }

  // Guardrails methods
  async setConfidenceThreshold(value: number) {
    const slider = this.confidenceThresholdSlider;
    const box = await slider.boundingBox();
    if (box) {
      const percentage = (value - 0) / (1 - 0);
      const x = box.x + (percentage * box.width);
      await this.page.mouse.click(x, box.y + box.height / 2);
    }
  }

  async saveGuardrails() {
    await this.saveGuardrailsButton.click();
  }

  // Security methods
  async toggleSecureEmbed(enabled: boolean) {
    const isChecked = await this.secureEmbedToggle.isChecked();
    if (isChecked !== enabled) {
      await this.secureEmbedToggle.click();
    }
  }

  async setAllowedDomains(domains: string) {
    await this.allowedDomainsInput.clear();
    await this.allowedDomainsInput.fill(domains);
  }

  async regenerateSecret() {
    await this.regenerateSecretButton.click();
    await this.page.locator('[data-testid="btn-confirm-regenerate"]').click();
  }

  async copySecret() {
    await this.copySecretButton.click();
  }

  async saveSecurity() {
    await this.saveSecurityButton.click();
  }

  // Toast expectations
  async expectSuccessToast(message: string) {
    await expect(this.page.locator('[data-testid="toast-success"]')).toContainText(message);
  }

  async expectErrorToast(message: string) {
    await expect(this.page.locator('[data-testid="toast-error"]')).toContainText(message);
  }
}
```

### Settings API Mocks

```typescript
// frontend/e2e/mocks/chatbot-settings.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockGetSettings(request: APIRequestContext, chatbotId: string) {
  await request.get(`/api/v1/chatbots/${chatbotId}/settings`, {
    status: 200,
    body: {
      name: 'Customer Support Bot',
      description: 'A helpful chatbot',
      botDisplayName: 'Support Bot',
      botIcon: 'https://example.com/icon.png',
      instructions: 'You are a helpful customer support agent.',
      language: 'tr',
      model: 'gpt-4o-mini',
      temperature: 0.7,
      maxTokens: 1000,
      themeColor: '#3B82F6',
      welcomeMessage: 'Hello! How can I help you?',
      position: 'bottom-right',
      botMessageColor: '#E5E7EB',
      userMessageColor: '#3B82F6',
      fontFamily: 'Inter',
      suggestionsEnabled: true,
      suggestedQuestions: ['How do I get started?', 'What features do you have?'],
      hideBranding: false,
      logoUrl: '',
      brandText: '',
      brandLink: '',
      confidenceThreshold: 0.7,
      noInfoMessage: 'I couldn\'t find information about that.',
      errorMessage: 'Something went wrong. Please try again.',
      allowedTopics: 'product, pricing',
      blockedTopics: 'politics, religion',
      handoffEnabled: false,
      handoffType: 'email',
      handoffMessage: 'Let me connect you with a human agent.',
      secureEmbed: false,
      allowedDomains: '',
      embedSecret: 'secret-abc123',
    },
  });
}

export async function mockSaveSettings(request: APIRequestContext, chatbotId: string) {
  await request.patch(`/api/v1/chatbots/${chatbotId}/settings`, {
    status: 200,
    body: {
      success: true,
      message: 'Settings saved successfully',
    },
  });
}

export async function mockRegenerateSecret(request: APIRequestContext, chatbotId: string) {
  await request.post(`/api/v1/chatbots/${chatbotId}/regenerate-secret`, {
    status: 200,
    body: {
      embedSecret: 'new-secret-' + Date.now(),
      success: true,
    },
  });
}

export async function mockValidationError(request: APIRequestContext) {
  await request.patch(`/api/v1/chatbots/chatbot-123/settings`, {
    status: 400,
    body: {
      error: 'VALIDATION_ERROR',
      field: 'name',
      message: 'Name is required',
    },
  });
}
```

### Running Specific Tests

```bash
# Run all settings tests
cd frontend && npx playwright test chatbot-settings.spec.ts

# Run identity section tests
cd frontend && npx playwright test chatbot-settings.spec.ts -g "identity"

# Run appearance section tests
cd frontend && npx playwright test chatbot-settings.spec.ts -g "appearance"

# Run security section tests
cd frontend && npx playwright test chatbot-settings.spec.ts -g "security"

# Run in headed mode
cd frontend && npx playwright test chatbot-settings.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All 9 sections tested
- [ ] All form fields tested
- [ ] All validations tested
- [ ] All save operations tested
- [ ] All toggles tested
- [ ] All color pickers tested
- [ ] All sliders tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with mocked API
- [ ] No race conditions
- [ ] Proper timeout handling

### 3. UX Verification
- [ ] Clear section organization
- [ ] Loading states visible
- [ ] Success feedback clear
- [ ] Error messages helpful

### 4. Security Verification
- [ ] Secret regeneration works
- [ ] Copy to clipboard works
- [ ] Confirmation required for sensitive actions

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Multiple Sections** - Test each section independently
2. **Complex Interactions** - Color pickers, sliders need careful testing
3. **File Uploads** - Use test fixtures for file upload testing
4. **Toggles** - Test both enabled and disabled states

### Common Issues to Avoid

1. **Not testing all sections** - Each section needs testing
2. **Skipping validation** - Test all validation rules
3. **Race conditions** - Wait for save to complete
4. **Hardcoded values** - Use test fixtures

### File Upload Testing

```typescript
// Test bot icon upload
test('uploads bot icon', async ({ page }) => {
  await page.goto('/dashboard/chatbots/chatbot-123/settings');
  
  // Set up file chooser listener
  const [fileChooser] = await Promise.all([
    page.waitForEvent('filechooser'),
    page.click('[data-testid="input-bot-icon"]'),
  ]);
  
  await fileChooser.setFiles(['e2e/fixtures/test-icon.png']);
  await expect(page.locator('[data-testid="icon-preview"]')).toBeVisible();
});
```

---

## Dependencies

- **Prerequisites**: 11-chatbots-detail.md (navigation to settings)
- **Environment**: Backend API with settings endpoints
- **Test Data**: Various settings configurations

---

## Related Tasks

- 11-chatbots-detail.md - Tab navigation to settings
- 10-chatbots-create.md - Creates chatbot with defaults
- 09-chatbots-list.md - Access settings from actions menu

---

*Task created from: docs/frontend/TEST_PATHS.md Section 4.4*
