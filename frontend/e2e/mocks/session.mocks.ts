import { Page } from '@playwright/test'

// Response type interfaces
export interface LogoutSuccessResponse {
  success: boolean
  message: string
}

export interface LogoutErrorResponse {
  error: string
  message: string
  code?: string
}

export interface SessionExpiredResponse {
  error: string
  message: string
  code: string
  requiresRelogin: boolean
}

export interface UnauthorizedResponse {
  error: string
  message: string
  code: string
}

// Default response values
const defaultSuccessResponse: LogoutSuccessResponse = {
  success: true,
  message: 'Logged out successfully',
}

const defaultUnauthorizedResponse: UnauthorizedResponse = {
  error: 'UNAUTHORIZED',
  message: 'Access token expired',
  code: 'TOKEN_EXPIRED',
}

const defaultSessionExpiredResponse: SessionExpiredResponse = {
  error: 'TOKEN_EXPIRED',
  message: 'Session has expired',
  code: 'TOKEN_EXPIRED',
  requiresRelogin: true,
}

/**
 * Mock successful logout response
 */
export async function mockSuccessfulLogout(
  page: Page,
  response?: Partial<LogoutSuccessResponse>
): Promise<void> {
  const mergedResponse = { ...defaultSuccessResponse, ...response }

  await page.route('**/api/v1/auth/logout', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })
}

/**
 * Mock logout with server error
 */
export async function mockLogoutServerError(
  page: Page,
  statusCode: number = 500,
  message: string = 'Internal Server Error'
): Promise<void> {
  await page.route('**/api/v1/auth/logout', async (route) => {
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
 * Mock logout with network error
 */
export async function mockLogoutNetworkError(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/logout', async (route) => {
    if (route.request().method() === 'POST') {
      await route.abort('failed')
    }
  })
}

/**
 * Mock unauthorized response (for expired tokens)
 */
export async function mockUnauthorized(
  page: Page,
  response?: Partial<UnauthorizedResponse>
): Promise<void> {
  const mergedResponse = { ...defaultUnauthorizedResponse, ...response }

  await page.route('**/api/v1/**', async (route) => {
    // Only intercept API calls, skip static assets
    const url = route.request().url()
    if (url.includes('/api/v1/') && !url.includes('/auth/login') && !url.includes('/auth/register')) {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })
}

/**
 * Mock session expired response (triggers modal)
 */
export async function mockSessionExpired(
  page: Page,
  response?: Partial<SessionExpiredResponse>
): Promise<void> {
  const mergedResponse = { ...defaultSessionExpiredResponse, ...response }

  await page.route('**/api/v1/**', async (route) => {
    const url = route.request().url()
    // Skip auth endpoints
    if (url.includes('/api/v1/') && !url.includes('/auth/')) {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(mergedResponse),
      })
    }
  })
}

/**
 * Mock specific endpoint to return 401
 */
export async function mockEndpointUnauthorized(
  page: Page,
  endpointPattern: string,
  response?: Partial<UnauthorizedResponse>
): Promise<void> {
  const mergedResponse = { ...defaultUnauthorizedResponse, ...response }

  await page.route(endpointPattern, async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify(mergedResponse),
    })
  })
}

/**
 * Mock refresh token endpoint
 */
export async function mockRefreshToken(
  page: Page,
  newAccessToken?: string,
  newRefreshToken?: string
): Promise<void> {
  const accessToken = newAccessToken || 'mock-refreshed-token-' + Date.now()
  const refreshToken = newRefreshToken || 'mock-refreshed-refresh-' + Date.now()

  await page.route('**/api/v1/auth/refresh', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        access_token: accessToken,
        refresh_token: refreshToken,
      }),
    })
  })
}

/**
 * Mock refresh token failure (forces logout)
 */
export async function mockRefreshTokenFailure(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({
        error: 'TOKEN_EXPIRED',
        message: 'Refresh token expired',
        code: 'REFRESH_TOKEN_EXPIRED',
      }),
    })
  })
}

/**
 * Mock session validation endpoint
 */
export async function mockSessionValid(page: Page): Promise<void> {
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
}

/**
 * Mock session validation to return unauthorized
 */
export async function mockSessionInvalid(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({
        error: 'UNAUTHORIZED',
        message: 'Session expired',
        code: 'SESSION_EXPIRED',
      }),
    })
  })
}

/**
 * Mock rate limit error
 */
export async function mockRateLimitError(page: Page): Promise<void> {
  await page.route('**/api/v1/**', async (route) => {
    const url = route.request().url()
    if (url.includes('/api/v1/')) {
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
 * Setup all common session-related mocks
 */
export async function setupSessionMocks(page: Page): Promise<void> {
  await mockSuccessfulLogout(page)
  await mockSessionValid(page)
  await mockRefreshToken(page)
}

/**
 * Abort all API requests (for testing offline behavior)
 */
export async function abortAllRequests(page: Page): Promise<void> {
  await page.route('**', async (route) => {
    const url = route.request().url()
    // Allow localhost for dev server, abort everything else
    if (!url.includes('localhost') && !url.includes('127.0.0.1')) {
      await route.abort('failed')
    }
  })
}
