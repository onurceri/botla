import { test as base } from '@playwright/test'
import { RegisterPage } from '../pages/register.page'

// Test data interfaces
export interface ValidRegistration {
  fullname: string
  email: string
  password: string
  confirmPassword?: string
  acceptTerms?: boolean
  acceptPrivacy?: boolean
}

export interface InvalidRegistration {
  fullname: string
  email: string
  password: string
  confirmPassword?: string
  acceptTerms?: boolean
  acceptPrivacy?: boolean
  expectedError: string | RegExp
}

export interface PasswordTestCase {
  password: string
  description: string
  validCharCount: boolean
  validUppercase: boolean
  validLowercase: boolean
  validDigit: boolean
  validSpecialChar: boolean
}

export interface EmailTestCase {
  email: string
  description: string
  valid: boolean
}

export interface ValidationTestData {
  emptyFullname: string
  emptyEmail: string
  emptyPassword: string
  emptyConfirmPassword: string
  invalidEmail: string
  weakPassword: string
  mismatchedPasswords: boolean
}

export interface EdgeCaseData {
  fullnameWithSpecialChars: string
  emailWithSpaces: string
  emailWithSpecialChars: string
  veryLongFullname: string
  veryLongEmail: string
  veryLongPassword: string
  passwordWithOnlySpaces: string
}

// Test fixtures interface
interface RegisterFixtures {
  // Page objects
  registerPage: RegisterPage

  // Valid test data
  validRegistration: ValidRegistration
  validFullname: string
  validEmail: string
  validPassword: string

  // Invalid test data
  invalidEmail: string
  weakPassword: string
  invalidRegistration: InvalidRegistration

  // Password test cases
  passwordTestCases: PasswordTestCase[]
  validPasswords: string[]
  invalidPasswords: Array<{ password: string; reason: string }>

  // Email test cases
  emailTestCases: EmailTestCase[]
  validEmails: string[]
  invalidEmails: Array<{ email: string; reason: string }>

  // Validation test data
  validationData: ValidationTestData

  // Edge case data
  edgeCaseData: EdgeCaseData

  // Test state
  testEmail: string
  testUser: {
    fullname: string
    email: string
    password: string
  }
}

// Custom test implementation
export const test = base.extend<RegisterFixtures>({
  // Register page fixture - creates and provides RegisterPage instance
  registerPage: async ({ page }, use) => {
    const registerPage = new RegisterPage(page)
    await registerPage.goto()
    await use(registerPage)
  },

  // Valid registration data
  validRegistration: async () => ({
    fullname: 'John Doe',
    email: 'newuser@example.com',
    password: 'SecurePass123!',
    confirmPassword: 'SecurePass123!',
    acceptTerms: true,
    acceptPrivacy: true,
  }),

  validFullname: async () => 'John Doe',

  validEmail: async () => 'newuser@example.com',

  validPassword: async () => 'SecurePass123!',

  // Invalid email formats
  invalidEmail: async () => 'not-an-email',

  // Weak password for testing
  weakPassword: async () => 'weak',

  // Invalid registration with expected error
  invalidRegistration: async () => ({
    fullname: 'Test User',
    email: 'invalid-email',
    password: 'weak',
    expectedError: /validation|invalid|required/i,
  }),

  // Password test cases with validation results
  passwordTestCases: async () => [
    {
      password: 'SecurePass123!',
      description: 'Valid password - all requirements met',
      validCharCount: true,
      validUppercase: true,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: true,
    },
    {
      password: 'Short1!',
      description: 'Invalid - too short (less than 8 characters)',
      validCharCount: false,
      validUppercase: false,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: true,
    },
    {
      password: 'lowercase123!',
      description: 'Invalid - no uppercase letter',
      validCharCount: true,
      validUppercase: false,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: true,
    },
    {
      password: 'UPPERCASE123!',
      description: 'Invalid - no lowercase letter',
      validCharCount: true,
      validUppercase: true,
      validLowercase: false,
      validDigit: true,
      validSpecialChar: true,
    },
    {
      password: 'NoDigits!@',
      description: 'Invalid - no digit',
      validCharCount: true,
      validUppercase: true,
      validLowercase: true,
      validDigit: false,
      validSpecialChar: true,
    },
    {
      password: 'NoSpecial123',
      description: 'Invalid - no special character',
      validCharCount: true,
      validUppercase: true,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: false,
    },
    {
      password: 'ABCdef123',
      description: 'Invalid - no special character',
      validCharCount: true,
      validUppercase: true,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: false,
    },
    {
      password: 'abc123!@#',
      description: 'Invalid - no uppercase letter',
      validCharCount: true,
      validUppercase: false,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: true,
    },
    {
      password: 'Short!',
      description: 'Invalid - too short and missing requirements',
      validCharCount: false,
      validUppercase: false,
      validLowercase: true,
      validDigit: false,
      validSpecialChar: true,
    },
    {
      password: 'PerfectPass123!',
      description: 'Valid - all requirements met with longer password',
      validCharCount: true,
      validUppercase: true,
      validLowercase: true,
      validDigit: true,
      validSpecialChar: true,
    },
  ],

  // Valid passwords for testing
  validPasswords: async () => [
    'SecurePass123!',
    'MyP@ssw0rd',
    'C0mplex!@#',
    'Str0ng#Pass',
    'Valid@123Pass',
    'Test!23456',
    'Password123!',
    'A1b2c3d4!',
    'AlphaNumeric!1',
    'M!nimum8Ch@rs',
  ],

  // Invalid passwords with reasons
  invalidPasswords: async () => [
    { password: 'Short1!', reason: 'Less than 8 characters' },
    { password: 'NoNumbers!', reason: 'Missing digit' },
    { password: 'nouppercase123!', reason: 'No uppercase letter' },
    { password: 'NOLOWER CASE123!', reason: 'No lowercase letter' },
    { password: 'NoSpecial123', reason: 'Missing special character' },
    { password: '', reason: 'Empty password' },
    { password: '     ', reason: 'Only spaces' },
    { password: 'a'.repeat(7) + '1!', reason: '7 characters (needs 8)' },
    { password: '12345678', reason: 'Only digits, no letters' },
    { password: 'abcdefgh!', reason: 'No uppercase or digit' },
  ],

  // Email test cases with validation results
  emailTestCases: async () => [
    {
      email: 'user@example.com',
      description: 'Valid standard email',
      valid: true,
    },
    {
      email: 'user.name@example.com',
      description: 'Valid email with dot in local part',
      valid: true,
    },
    {
      email: 'user+tag@example.com',
      description: 'Valid email with plus addressing',
      valid: true,
    },
    {
      email: 'user@subdomain.example.com',
      description: 'Valid email with subdomain',
      valid: true,
    },
    {
      email: 'user@example.co.uk',
      description: 'Valid email with multi-level domain',
      valid: true,
    },
    {
      email: 'not-an-email',
      description: 'Invalid - missing @ symbol',
      valid: false,
    },
    {
      email: 'user@',
      description: 'Invalid - missing domain',
      valid: false,
    },
    {
      email: '@example.com',
      description: 'Invalid - missing local part',
      valid: false,
    },
    {
      email: 'user@example',
      description: 'Invalid - missing TLD',
      valid: false,
    },
    {
      email: '',
      description: 'Invalid - empty email',
      valid: false,
    },
    {
      email: 'user name@example.com',
      description: 'Invalid - space in email',
      valid: false,
    },
    {
      email: 'user@exam ple.com',
      description: 'Invalid - space in domain',
      valid: false,
    },
  ],

  // Valid emails for testing
  validEmails: async () => [
    'newuser@example.com',
    'test.user@example.com',
    'user+tag@example.com',
    'user@sub.example.com',
    'user123@example.co.uk',
    'first.last@domain.org',
    'user_name@example.net',
    'user-name@example.io',
  ],

  // Invalid emails with reasons
  invalidEmails: async () => [
    { email: 'not-an-email', reason: 'Missing @ symbol' },
    { email: 'user@', reason: 'Missing domain' },
    { email: '@example.com', reason: 'Missing local part' },
    { email: 'user@example', reason: 'Missing TLD' },
    { email: '', reason: 'Empty email' },
    { email: 'user name@example.com', reason: 'Space in email' },
    { email: 'user@exam ple.com', reason: 'Space in domain' },
  ],

  // Validation test data
  validationData: async () => ({
    emptyFullname: '',
    emptyEmail: '',
    emptyPassword: '',
    emptyConfirmPassword: '',
    invalidEmail: 'invalid-email',
    weakPassword: 'weak',
    mismatchedPasswords: true,
  }),

  // Edge case test data
  edgeCaseData: async () => ({
    fullnameWithSpecialChars: "O'Brien-Smith Jr.",
    emailWithSpaces: '  test@example.com  ',
    emailWithSpecialChars: 'test+tag@example.com',
    veryLongFullname: 'A'.repeat(100),
    veryLongEmail: 'a'.repeat(100) + '@example.com',
    veryLongPassword: 'b'.repeat(200),
    passwordWithOnlySpaces: '     ',
  }),

  // Dynamic test email for unique test runs
  testEmail: async () => `test-${Date.now()}@example.com`,

  // Test user object
  testUser: async () => ({
    fullname: `Test User ${Date.now()}`,
    email: `test-${Date.now()}@example.com`,
    password: 'SecurePass123!',
  }),
})

// Re-export for convenience
export { expect } from '@playwright/test'
