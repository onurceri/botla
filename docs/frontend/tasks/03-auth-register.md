# Task: Implement Registration Page Tests

> **Task ID**: 03-auth-register  
> **Source**: TEST_PATHS.md Section 2.2  
> **Priority**: Highest (Authentication)  
> **Estimated Effort**: 8-10 hours  
> **Prerequisite**: 01-test-naming-conventions.md (must be completed first)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Registration Page. This task covers all registration functionality including form interactions, real-time password validation, terms acceptance, and error handling.

### Context

The Registration Page is the account creation entry point for new users. Testing this page thoroughly ensures:
- Users can successfully create accounts
- Password strength requirements are enforced
- Terms and conditions are properly accepted
- Duplicate email prevention works correctly
- Proper organization and workspace are created

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 2.2:

#### 2.2.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `input-fullname` | text | Full name input |
| `input-email` | text | Email input |
| `input-password` | password | Password input |
| `input-confirm-password` | password | Confirm password |
| `checkbox-terms` | checkbox | Accept terms |
| `checkbox-privacy` | checkbox | Accept privacy policy |
| `btn-register` | submit | Register button |
| `link-login` | link | Already have account |
| `text-password-requirements` | text | Password rules display |

#### 2.2.2 Password Requirements Display

```
Password Requirements (Real-time Validation)
├── Character count ≥ 8 (checked on input)
├── Uppercase letter (checked on input)
├── Lowercase letter (checked on input)
├── Digit (checked on input)
└── Special character @$!%*?& (checked on input)
```

#### 2.2.3 Registration Flow

```
Register Flow
├── Load register page
│   ├── All inputs empty
│   ├── Password requirements visible (gray)
│   └── btn-register disabled
│
├── Fill form - Full Name
│   ├── Type: "John Doe"
│   └── Assert: Value = "John Doe"
│
├── Fill form - Email
│   ├── Type: "john@example.com"
│   ├── Blur: Trigger format validation
│   └── Assert: No error if valid
│
├── Fill form - Password
│   ├── Type: "Weak123"
│   ├── Assert: Character count requirement (check)
│   ├── Assert: Uppercase requirement (check)
│   ├── Assert: Digit requirement (check)
│   ├── Assert: Special char requirement (x)
│   ├── Type: "Weak123@" (complete)
│   └── Assert: All requirements (check)
│
├── Fill form - Confirm Password
│   ├── Type: "Weak123@"
│   ├── Assert: Matches password
│   └── Assert: No error
│
├── Submit without accepting terms
│   ├── Click: btn-register
│   ├── Assert: `toast-error` - "Accept terms required"
│   └── Assert: checkbox-terms has error class
│
├── Submit with mismatched passwords
│   ├── Change confirm password to "Different123@"
│   ├── Click: btn-register
│   ├── Assert: `input-confirm-password` error
│   └── Assert: `text-error` - "Passwords do not match"
│
├── Submit with weak password
│   ├── Change password to "weak"
│   ├── Click: btn-register
│   ├── Assert: `input-password` error
│   └── Assert: `text-error` - "Password too weak"
│
├── Successful registration
│   ├── Check: checkbox-terms
│   ├── Check: checkbox-privacy
│   ├── Click: btn-register
│   ├── Assert: Loading state
│   ├── Wait: API response
│   ├── Assert: User created in database
│   ├── Assert: Default org created
│   ├── Assert: Default workspace created
│   ├── Assert: Tokens stored
│   └── Redirect: /dashboard
│
└── Email already exists
    ├── Type: existing email
    ├── Click: btn-register
    ├── Assert: `toast-error` - "Email already exists"
    └── Assert: `input-email` error
```

#### 2.2.4 Validation States

| Field | Valid State | Invalid State |
|-------|-------------|---------------|
| Full Name | Non-empty | Empty |
| Email | RFC 5322 format | Invalid format |
| Password | All 5 requirements met | Any missing |
| Confirm Password | Matches password | Mismatch |
| Terms | Checked | Unchecked |

### Implementation Requirements

1. **Create/Update Register Test File** (`frontend/e2e/register.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Register Page Object** (`frontend/e2e/pages/register.page.ts`)
   - Encapsulate registration page interactions
   - Provide reusable methods
   - Improve test maintainability

3. **Create Test Data Fixtures** (`frontend/e2e/fixtures/register.fixture.ts`)
   - Valid registration data
   - Invalid registration data
   - Password strength test cases

4. **Create Mock API Handlers** (`frontend/e2e/mocks/register.mocks.ts`)
   - Mock successful registration responses
   - Mock error responses (email exists, validation errors)
   - Handle registration endpoints

### Expected Deliverables

1. `frontend/e2e/register.spec.ts` - Comprehensive registration tests
2. `frontend/e2e/pages/register.page.ts` - Page object model
3. `frontend/e2e/fixtures/register.fixture.ts` - Test data fixtures
4. `frontend/e2e/mocks/register.mocks.ts` - API mock handlers

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/register.page.ts` with:
  - Locators for all page elements
  - Methods for all interactions
  - Password requirement indicator helpers
- [ ] Create `frontend/e2e/fixtures/register.fixture.ts` with test data
- [ ] Create `frontend/e2e/mocks/register.mocks.ts` with API handlers

### Phase 2: Page Load Tests

- [ ] Test: Page loads successfully
- [ ] Test: All inputs are empty
- [ ] Test: Password requirements are visible (gray/unchecked)
- [ ] Test: Register button is disabled initially
- [ ] Test: Navigation links are present

### Phase 3: Form Field Tests

- [ ] Test: Full name input works correctly
- [ ] Test: Email input with format validation
- [ ] Test: Password input character counting
- [ ] Test: Confirm password matching
- [ ] Test: Terms checkbox functionality
- [ ] Test: Privacy checkbox functionality

### Phase 4: Password Validation Tests

- [ ] Test: Character count ≥ 8 (real-time)
- [ ] Test: Uppercase letter requirement (real-time)
- [ ] Test: Lowercase letter requirement (real-time)
- [ ] Test: Digit requirement (real-time)
- [ ] Test: Special character requirement (real-time)
- [ ] Test: All requirements met state
- [ ] Test: Individual requirement toggle states

### Phase 5: Form Submission Validation Tests

- [ ] Test: Submit without full name
- [ ] Test: Submit with invalid email format
- [ ] Test: Submit with weak password
- [ ] Test: Submit with mismatched passwords
- [ ] Test: Submit without accepting terms
- [ ] Test: Submit without accepting privacy policy
- [ ] Test: Error messages display correctly

### Phase 6: Successful Registration Tests

- [ ] Test: Successful registration with valid data
- [ ] Test: Loading state during registration
- [ ] Test: Button disabled during submission
- [ ] Test: User created in database (API mock)
- [ ] Test: Default organization created
- [ ] Test: Default workspace created
- [ ] Test: Access token stored
- [ ] Test: Refresh token stored
- [ ] Test: Redirect to dashboard

### Phase 7: Error Scenarios Tests

- [ ] Test: Email already exists error
- [ ] Test: Network error handling
- [ ] Test: Server error handling
- [ ] Test: Multiple registration attempts

### Phase 8: Navigation Tests

- [ ] Test: Navigate to login page via link
- [ ] Test: Forgot password link works
- [ ] Test: Browser back button behavior

---

## Technical Notes

### Page Object Pattern

Create a reusable page object for registration:

```typescript
// frontend/e2e/pages/register.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class RegisterPage {
  readonly page: Page;
  readonly fullnameInput: Locator;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly termsCheckbox: Locator;
  readonly privacyCheckbox: Locator;
  readonly registerButton: Locator;
  readonly loginLink: Locator;
  readonly errorMessage: Locator;
  
  // Password requirements indicators
  readonly reqCharCount: Locator;
  readonly reqUppercase: Locator;
  readonly reqLowercase: Locator;
  readonly reqDigit: Locator;
  readonly reqSpecialChar: Locator;

  constructor(page: Page) {
    this.page = page;
    this.fullnameInput = page.locator('[data-testid="input-fullname"]');
    this.emailInput = page.locator('[data-testid="input-email"]');
    this.passwordInput = page.locator('[data-testid="input-password"]');
    this.confirmPasswordInput = page.locator('[data-testid="input-confirm-password"]');
    this.termsCheckbox = page.locator('[data-testid="checkbox-terms"]');
    this.privacyCheckbox = page.locator('[data-testid="checkbox-privacy"]');
    this.registerButton = page.locator('[data-testid="btn-register"]');
    this.loginLink = page.locator('[data-testid="link-login"]');
    this.errorMessage = page.locator('[data-testid="text-error"]');
    
    // Password requirement indicators
    this.reqCharCount = page.locator('[data-testid="req-char-count"]');
    this.reqUppercase = page.locator('[data-testid="req-uppercase"]');
    this.reqLowercase = page.locator('[data-testid="req-lowercase"]');
    this.reqDigit = page.locator('[data-testid="req-digit"]');
    this.reqSpecialChar = page.locator('[data-testid="req-special"]');
  }

  async goto() {
    await this.page.goto('/register');
  }

  async fillFullname(name: string) {
    await this.fullnameInput.fill(name);
  }

  async fillEmail(email: string) {
    await this.emailInput.fill(email);
    await this.emailInput.blur();
  }

  async fillPassword(password: string) {
    await this.passwordInput.fill(password);
  }

  async fillConfirmPassword(password: string) {
    await this.confirmPasswordInput.fill(password);
  }

  async checkTerms() {
    await this.termsCheckbox.check();
  }

  async checkPrivacy() {
    await this.privacyCheckbox.check();
  }

  async clickRegister() {
    await this.registerButton.click();
  }

  async register(data: {
    fullname: string;
    email: string;
    password: string;
    confirmPassword?: string;
    acceptTerms?: boolean;
    acceptPrivacy?: boolean;
  }) {
    await this.fillFullname(data.fullname);
    await this.fillEmail(data.email);
    await this.fillPassword(data.password);
    if (data.confirmPassword) {
      await this.fillConfirmPassword(data.confirmPassword);
    }
    if (data.acceptTerms) {
      await this.checkTerms();
    }
    if (data.acceptPrivacy) {
      await this.checkPrivacy();
    }
    await this.clickRegister();
  }

  // Password requirement checkers
  async expectCharCountValid() {
    await expect(this.reqCharCount).toHaveClass(/valid/);
  }

  async expectUppercaseValid() {
    await expect(this.reqUppercase).toHaveClass(/valid/);
  }

  async expectDigitValid() {
    await expect(this.reqDigit).toHaveClass(/valid/);
  }

  async expectAllRequirementsValid() {
    await this.expectCharCountValid();
    await this.expectUppercaseValid();
    await expect(this.reqLowercase).toHaveClass(/valid/);
    await this.expectDigitValid();
    await expect(this.reqSpecialChar).toHaveClass(/valid/);
  }
}
```

### Password Strength Test Data

Create comprehensive test cases for password validation:

```typescript
// frontend/e2e/fixtures/register.fixture.ts
export const testPasswords = {
  valid: 'SecurePass123!',
  tooShort: 'Short1!',
  noUppercase: 'lowercase123!',
  noLowercase: 'UPPERCASE123!',
  noDigit: 'NoDigits!@',
  noSpecial: 'NoSpecial123',
  empty: '',
  onlySpaces: '     ',
};

export const testEmails = {
  valid: 'newuser@example.com',
  invalid: 'not-an-email',
  empty: '',
  noDomain: 'user@',
  noAt: 'userexample.com',
};
```

### API Mocking

Mock registration API with proper responses:

```typescript
// frontend/e2e/mocks/register.mocks.ts
export async function mockSuccessfulRegistration(request: APIRequestContext) {
  await request.post('/api/v1/auth/register', {
    status: 201,
    body: {
      access_token: 'mock-access-token',
      refresh_token: 'mock-refresh-token',
      user: {
        id: 'user-new-123',
        email: 'newuser@example.com',
        name: 'New User',
      },
      organization: {
        id: 'org-123',
        name: 'My Organization',
      },
      workspace: {
        id: 'workspace-123',
        name: 'Default Workspace',
      },
    },
  });
}

export async function mockEmailExists(request: APIRequestContext) {
  await request.post('/api/v1/auth/register', {
    status: 409,
    body: {
      error: 'CONFLICT',
      message: 'Email already registered',
      field: 'email',
    },
  });
}

export async function mockValidationError(request: APIRequestContext, field: string) {
  await request.post('/api/v1/auth/register', {
    status: 400,
    body: {
      error: 'VALIDATION_ERROR',
      message: `Invalid ${field}`,
      field,
    },
  });
}
```

### Running Specific Tests

```bash
# Run all registration tests
cd frontend && npx playwright test register.spec.ts

# Run password validation tests
cd frontend && npx playwright test register.spec.ts -g "password"

# Run success tests
cd frontend && npx playwright test register.spec.ts -g "successful"

# Run in headed mode
cd frontend && npx playwright test register.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All form fields tested
- [ ] All password requirements tested
- [ ] All validation scenarios covered
- [ ] Success flow tested
- [ ] All error scenarios tested
- [ ] Navigation tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] No flaky tests
- [ ] Proper timeout handling
- [ ] Clean test isolation

### 3. Code Quality Verification
- [ ] Page object pattern used
- [ ] Selectors follow naming conventions
- [ ] Tests are maintainable
- [ ] No hardcoded values

### 4. User Flow Verification
- [ ] Real-time password validation works
- [ ] Error messages are clear
- [ ] Success flow is smooth
- [ ] Loading states are visible

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Real-time validation** - Password requirements should update as user types
2. **Button state** - Register button should be disabled until all requirements met
3. **Terms required** - Both terms and privacy must be checked
4. **Data cleanup** - Consider using unique emails for testing to avoid conflicts
5. **Mock responses** - Verify organization/workspace creation in mocks

### Common Issues to Avoid

1. **Skipping blur events** - Email validation often triggers on blur
2. **Not waiting for requirements** - Check classes update after typing
3. **Race conditions** - Wait for API response before checking redirect
4. **Hardcoded passwords** - Use fixture data instead

### Test Order

1. Page load tests first
2. Individual field tests
3. Password validation tests (comprehensive)
4. Form submission validation
5. Success flow tests
6. Error handling tests

---

## Dependencies

- **Prerequisite**: 01-test-naming-conventions.md (must be completed)
- **Environment**: Backend API must be running
- **Test Data**: Need unique email addresses for testing

---

## Related Tasks

- 01-test-naming-conventions.md - Foundation for naming conventions
- 02-auth-login.md - Login page tests (uses similar patterns)
- 04-auth-logout.md - Logout flow tests
- 05-auth-sessions.md - Session management tests

---

*Task created from: docs/frontend/TEST_PATHS.md Section 2.2*
