import { test as base, Page, Locator } from '@playwright/test'
import { LoginPage } from '../pages/login.page'

// Test data interfaces
export interface ValidUser {
  email: string
  password: string
  name?: string
}

export interface InvalidCredentials {
  email: string
  password: string
  errorMessage: string
}

export interface ValidationTestData {
  emptyEmail: string
  emptyPassword: string
  invalidEmail: string
  invalidPassword: string
}

export interface EdgeCaseData {
  emailWithSpaces: string
  emailWithSpecialChars: string
  veryLongEmail: string
  veryLongPassword: string
  passwordWithOnlySpaces: string
}

// Test fixtures interface
interface AuthFixtures {
  // Page objects
  loginPage: LoginPage

  // Valid test data
  validUser: ValidUser
  validEmail: string
  validPassword: string

  // Invalid test data
  invalidEmail: string
  invalidPassword: string
  invalidCredentials: InvalidCredentials

  // Validation test data
  validationData: ValidationTestData

  // Edge case data
  edgeCaseData: EdgeCaseData

  // Test state
  testEmail: string
}

// Custom test implementation
export const test = base.extend<AuthFixtures>({
  // Login page fixture - creates and provides LoginPage instance
  loginPage: async ({ page }, use) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await use(loginPage)
  },

  // Valid user credentials
  validUser: async () => ({
    email: 'test@example.com',
    password: 'SecurePass123!',
    name: 'Test User',
  }),

  validEmail: async () => 'test@example.com',

  validPassword: async () => 'SecurePass123!',

  // Invalid email formats
  invalidEmail: async () => 'not-an-email',

  // Invalid password
  invalidPassword: async () => 'wrongpassword',

  // Invalid credentials with expected error
  invalidCredentials: async () => ({
    email: 'wrong@example.com',
    password: 'wrongpassword',
    errorMessage: /invalid|incorrect|failed/i,
  }),

  // Validation test data
  validationData: async () => ({
    emptyEmail: '',
    emptyPassword: '',
    invalidEmail: 'invalid-email',
    invalidPassword: '123', // Too short
  }),

  // Edge case test data
  edgeCaseData: async () => ({
    emailWithSpaces: '  test@example.com  ',
    emailWithSpecialChars: 'test+tag@example.com',
    veryLongEmail: 'a'.repeat(100) + '@example.com',
    veryLongPassword: 'b'.repeat(200),
    passwordWithOnlySpaces: '     ',
  }),

  // Dynamic test email for unique test runs
  testEmail: async () => `test-${Date.now()}@example.com`,
})

// Re-export for convenience
export { expect } from '@playwright/test'
