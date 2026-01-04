import { Page, expect } from '@playwright/test'

// Registration response types
export interface RegisterSuccessResponse {
  access_token: string
  refresh_token: string
  user: {
    id: string
    email: string
    name: string
    plan?: string
  }
}

export interface RegisterErrorResponse {
  error: string
  message: string
  field?: string
  code?: string
}

// Default successful registration response
const defaultSuccessResponse: RegisterSuccessResponse = {
  access_token: 'mock-access-token-' + Date.now(),
  refresh_token: 'mock-refresh-token-' + Date.now(),
  user: {
    id: 'user-new-' + Date.now(),
    email: 'newuser@example.com',
    name: 'New User',
    plan: 'free',
  },
}

// Default error responses
const defaultEmailExistsResponse: RegisterErrorResponse = {
  error: 'CONFLICT',
  message: 'Email already registered',
  field: 'email',
  code: 'AUTH_002',
}

/**
 * Mock successful registration flow
 * This mocks both the register and subsequent login API calls
 */
export async function mockSuccessfulRegistration(page: Page): Promise<void> {
  const successResponse = defaultSuccessResponse

  // Mock registration endpoint
  await page.route('**/api/v1/auth/register', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ success: true }),
      })
    }
  })

  // Mock login endpoint (called after successful registration)
  await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          token: successResponse.access_token,
          refresh_token: successResponse.refresh_token,
        }),
      })
    }
  })

  // Mock onboarding check endpoint
  await page.route('**/api/v1/me/onboarding', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ completed: true, skipped: false }),
    })
  })
}

/**
 * Mock email already exists error
 */
export async function mockEmailExists(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/register', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 409,
        contentType: 'application/json',
        body: JSON.stringify(defaultEmailExistsResponse),
      })
    }
  })
}

/**
 * Mock validation error on registration
 */
export async function mockValidationError(
  page: Page,
  statusCode: number = 400,
  message: string = 'Validation failed'
): Promise<void> {
  await page.route('**/api/v1/auth/register', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: statusCode,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'VALIDATION_ERROR',
          message,
        }),
      })
    }
  })
}

/**
 * Mock network error on registration attempt
 */
export async function mockNetworkError(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/register', async (route) => {
    await route.abort('failed')
  })
}

/**
 * Mock server error on registration
 */
export async function mockServerError(
  page: Page,
  statusCode: number = 500,
  message: string = 'Internal Server Error'
): Promise<void> {
  await page.route('**/api/v1/auth/register', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: statusCode,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Server Error',
          message,
        }),
      })
    }
  })
}

/**
 * Mock too many requests error (rate limiting)
 */
export async function mockRateLimitError(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/register', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 429,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Too Many Requests',
          message: 'Please wait before trying again',
          retryAfter: 60,
        }),
      })
    }
  })
}

/**
 * Clear all auth-related localStorage and sessionStorage
 */
export async function clearAuthStorage(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    localStorage.removeItem('botla_user')
    sessionStorage.clear()
  })
}

/**
 * Set auth tokens in localStorage manually
 */
export async function setAuthTokens(
  page: Page,
  accessToken: string,
  refreshToken: string
): Promise<void> {
  await page.evaluate(
    ({ accessToken, refreshToken }) => {
      localStorage.setItem('botla_token', accessToken)
      localStorage.setItem('botla_refresh_token', refreshToken)
    },
    { accessToken, refreshToken }
  )
}

/**
 * Get auth tokens from localStorage
 */
export async function getAuthTokens(page: Page): Promise<{
  accessToken: string | null
  refreshToken: string | null
}> {
  return await page.evaluate(() => ({
    accessToken: localStorage.getItem('botla_token'),
    refreshToken: localStorage.getItem('botla_refresh_token'),
  }))
}

/**
 * Wait for registration API response to complete
 */
export async function waitForRegistrationResponse(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForResponse(
    (response) =>
      response.url().includes('/api/v1/auth/register') &&
      response.request().method() === 'POST',
    { timeout }
  )
}

/**
 * Wait for loading spinner to disappear after registration
 */
export async function waitForLoadingToComplete(
  page: Page,
  timeout: number = 15000
): Promise<void> {
  const spinner = page.locator('[data-testid="loading-spinner"]')
  if (await spinner.isVisible({ timeout: 1000 })) {
    await spinner.waitFor({ state: 'hidden', timeout })
  }
}

/**
 * Verify tokens are stored in localStorage after registration
 */
export async function verifyTokensStored(
  page: Page,
  expectTokens: boolean = true
): Promise<{
  accessToken: string | null
  refreshToken: string | null
}> {
  const tokens = await getAuthTokens(page)

  if (expectTokens) {
    expect(tokens.accessToken).toBeTruthy()
    expect(tokens.refreshToken).toBeTruthy()
  }

  return tokens
}
