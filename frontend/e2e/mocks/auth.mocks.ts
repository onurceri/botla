import { Page } from '@playwright/test'

// Login response types
export interface LoginSuccessResponse {
  access_token: string
  refresh_token: string
  user: {
    id: string
    email: string
    name: string
    plan?: string
  }
}

export interface LoginErrorResponse {
  error: string
  message: string
  code?: string
}

// Default successful login response
const defaultSuccessResponse: LoginSuccessResponse = {
  access_token: 'mock-access-token-' + Date.now(),
  refresh_token: 'mock-refresh-token-' + Date.now(),
  user: {
    id: 'user-123',
    email: 'test@example.com',
    name: 'Test User',
    plan: 'pro',
  },
}

// Default error responses
const defaultUnauthorizedResponse: LoginErrorResponse = {
  error: 'Unauthorized',
  message: 'Invalid email or password',
  code: 'AUTH_001',
}

const defaultValidationErrorResponse: LoginErrorResponse = {
  error: 'Validation Error',
  message: 'Email format is invalid',
  code: 'VALIDATION_001',
}

/**
 * Mock successful login response
 */
export async function mockSuccessfulLogin(
  page: Page,
  response?: Partial<LoginSuccessResponse>
): Promise<void> {
  const mergedResponse = { ...defaultSuccessResponse, ...response }

  await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })

  // Also mock the /me endpoint for session validation
  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(mergedResponse.user),
    })
  })

  // Mock onboarding check
  await page.route('**/api/v1/me/onboarding', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ completed: true, skipped: false }),
    })
  })
}

/**
 * Mock failed login with unauthorized response
 */
export async function mockFailedLogin(
  page: Page,
  response?: Partial<LoginErrorResponse>
): Promise<void> {
  const mergedResponse = { ...defaultUnauthorizedResponse, ...response }

  await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })
}

/**
 * Mock validation error on login
 */
export async function mockValidationError(
  page: Page,
  response?: Partial<LoginErrorResponse>
): Promise<void> {
  const mergedResponse = { ...defaultValidationErrorResponse, ...response }

  await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })
}

/**
 * Mock network error on login attempt
 */
export async function mockNetworkError(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/login', async (route) => {
    // Simulate network error by aborting
    await route.abort('failed')
  })
}

/**
 * Mock server error on login
 */
export async function mockServerError(
  page: Page,
  statusCode: number = 500,
  message: string = 'Internal Server Error'
): Promise<void> {
  await page.route('**/api/v1/auth/login', async (route) => {
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
  await page.route('**/api/v1/auth/login', async (route) => {
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
 * Mock remember me functionality - sets tokens in localStorage
 */
export async function mockRememberMeLogin(
  page: Page,
  response?: Partial<LoginSuccessResponse>
): Promise<void> {
  const mergedResponse = { ...defaultSuccessResponse, ...response }

  await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })

  // After login, verify localStorage has the tokens
  await page.evaluate((response) => {
    localStorage.setItem('botla_token', response.access_token)
    localStorage.setItem('botla_refresh_token', response.refresh_token)
  }, mergedResponse)
}

/**
 * Mock logout endpoint
 */
export async function mockLogout(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/logout', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ success: true }),
    })
  })
}

/**
 * Mock refresh token endpoint
 */
export async function mockRefreshToken(
  page: Page,
  newAccessToken?: string
): Promise<void> {
  const token = newAccessToken || 'mock-refreshed-token-' + Date.now()

  await page.route('**/api/v1/auth/refresh', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        access_token: token,
        refresh_token: token,
      }),
    })
  })
}

/**
 * Mock forgot password endpoint
 */
export async function mockForgotPassword(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/forgot-password', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Password reset link sent to your email',
        }),
      })
    }
  })
}

/**
 * Mock reset password endpoint
 */
export async function mockResetPassword(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/reset-password', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Password reset successfully',
        }),
      })
    }
  })
}

/**
 * Setup all common authentication mocks
 * Includes login, logout, refresh, and session validation
 */
export async function setupAuthMocks(page: Page): Promise<void> {
  await mockSuccessfulLogin(page)

  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        plan: 'pro',
      }),
    })
  })

  await page.route('**/api/v1/me/onboarding', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ completed: true, skipped: false }),
    })
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
 * Wait for login API response to complete
 */
export async function waitForLoginResponse(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForResponse(
    (response) =>
      response.url().includes('/api/v1/auth/login') &&
      response.request().method() === 'POST',
    { timeout }
  )
}

/**
 * Wait for loading spinner to disappear after login
 */
export async function waitForLoadingToComplete(
  page: Page,
  timeout: number = 15000
): Promise<void> {
  const spinner = page.getByTestId('loading-spinner')
  if (await spinner.isVisible({ timeout: 1000 })) {
    await spinner.waitFor({ state: 'hidden', timeout })
  }
}
