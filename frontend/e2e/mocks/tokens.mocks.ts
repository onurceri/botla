import { Page } from '@playwright/test'
import { COOKIE_NAMES } from '../utils/cookie-auth'

/**
 * Token refresh success response interface
 */
export interface TokenRefreshSuccessResponse {
  access_token: string
  refresh_token?: string
  expires_in: number
  token_type: string
}

/**
 * Token refresh error response interface
 */
export interface TokenRefreshErrorResponse {
  error: string
  message: string
  code?: string
}

/**
 * Session status response interface
 */
export interface SessionStatusResponse {
  active: boolean
  user?: {
    id: string
    email: string
    name: string
    plan?: string
  }
  expires_at?: string
  reason?: string
}

/**
 * Token validation response interface
 */
export interface TokenValidationResponse {
  valid: boolean
  user?: {
    id: string
    email: string
    name: string
    plan?: string
  }
  error?: string
}

/**
 * Default mock response values
 */

const defaultUnauthorizedResponse = {
  error: 'UNAUTHORIZED',
  message: 'Access token expired or invalid',
  code: 'TOKEN_EXPIRED',
}

const defaultSessionExpiredResponse = {
  error: 'TOKEN_EXPIRED',
  message: 'Session has expired',
  code: 'TOKEN_EXPIRED',
  requiresRelogin: true,
}

const defaultRefreshTokenExpiredResponse = {
  error: 'REFRESH_TOKEN_EXPIRED',
  message: 'Refresh token has expired. Please login again.',
  code: 'REFRESH_TOKEN_EXPIRED',
}

const defaultInvalidRefreshTokenResponse = {
  error: 'INVALID_REFRESH_TOKEN',
  message: 'Invalid refresh token',
  code: 'INVALID_REFRESH_TOKEN',
}

const defaultRevokedTokenResponse = {
  error: 'TOKEN_REVOKED',
  message: 'Refresh token has been revoked',
  code: 'TOKEN_REVOKED',
}

/**
 * Mock successful token refresh response
 */
export async function mockSuccessfulTokenRefresh(
  page: Page,
  newAccessToken?: string,
  newRefreshToken?: string,
  expiresIn: number = 3600
): Promise<void> {
  const accessToken = newAccessToken || 'mock-refreshed-token-' + Date.now()
  const refreshToken = newRefreshToken || 'mock-refreshed-refresh-' + Date.now()

  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        headers: {
          'Set-Cookie': [
            `${COOKIE_NAMES.ACCESS_TOKEN}=${accessToken}; Path=/; HttpOnly; SameSite=Strict`,
            `${COOKIE_NAMES.REFRESH_TOKEN}=${refreshToken}; Path=/; HttpOnly; SameSite=Strict`,
          ].join(', '),
        },
        body: JSON.stringify({
          access_token: accessToken,
          refresh_token: refreshToken,
          expires_in: expiresIn,
          token_type: 'Bearer',
        }),
      })
    }
  })
}

/**
 * Mock token refresh that returns new tokens only (no refresh token update)
 */
export async function mockTokenRefreshAccessTokenOnly(
  page: Page,
  newAccessToken?: string
): Promise<void> {
  const accessToken = newAccessToken || 'mock-refreshed-token-' + Date.now()

  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        headers: {
          'Set-Cookie': `${COOKIE_NAMES.ACCESS_TOKEN}=${accessToken}; Path=/; HttpOnly; SameSite=Strict`,
        },
        body: JSON.stringify({
          access_token: accessToken,
          expires_in: 3600,
          token_type: 'Bearer',
        }),
      })
    }
  })
}

/**
 * Mock expired refresh token response (forces re-login)
 */
export async function mockExpiredRefreshToken(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(defaultRefreshTokenExpiredResponse),
      })
    }
  })
}

/**
 * Mock invalid refresh token response
 */
export async function mockInvalidRefreshToken(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(defaultInvalidRefreshTokenResponse),
      })
    }
  })
}

/**
 * Mock revoked refresh token response
 */
export async function mockRevokedRefreshToken(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(defaultRevokedTokenResponse),
      })
    }
  })
}

/**
 * Mock refresh token server error
 */
export async function mockRefreshTokenServerError(
  page: Page,
  statusCode: number = 500,
  message: string = 'Internal Server Error'
): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
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
 * Mock refresh token network error
 */
export async function mockRefreshTokenNetworkError(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.abort('failed')
    }
  })
}

/**
 * Mock session status endpoint (active session)
 */
export async function mockSessionStatusActive(
  page: Page,
  userData?: { id?: string; email?: string; name?: string; plan?: string }
): Promise<void> {
  await page.route('**/api/v1/auth/session', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        active: true,
        user: {
          id: userData?.id || 'user-123',
          email: userData?.email || 'test@example.com',
          name: userData?.name || 'Test User',
          plan: userData?.plan || 'pro',
        },
        expires_at: new Date(Date.now() + 3600000).toISOString(),
      }),
    })
  })
}

/**
 * Mock session status endpoint (expired session)
 */
export async function mockSessionStatusExpired(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/session', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({
        active: false,
        reason: 'Session expired',
      }),
    })
  })
}

/**
 * Mock token validation endpoint (valid token)
 */
export async function mockTokenValidationValid(
  page: Page,
  userData?: { id?: string; email?: string; name?: string; plan?: string }
): Promise<void> {
  await page.route('**/api/v1/auth/validate', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        valid: true,
        user: {
          id: userData?.id || 'user-123',
          email: userData?.email || 'test@example.com',
          name: userData?.name || 'Test User',
          plan: userData?.plan || 'pro',
        },
      }),
    })
  })
}

/**
 * Mock token validation endpoint (invalid token)
 */
export async function mockTokenValidationInvalid(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/validate', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({
        valid: false,
        error: 'Invalid token',
      }),
    })
  })
}

/**
 * Mock 401 response for any API call (triggers token refresh flow)
 */
export async function mockUnauthorizedForApi(
  page: Page,
  response?: Partial<typeof defaultUnauthorizedResponse>
): Promise<void> {
  const mergedResponse = { ...defaultUnauthorizedResponse, ...response }

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
 * Mock 401 with session expired modal trigger
 */
export async function mockSessionExpired(page: Page): Promise<void> {
  await page.route('**/api/v1/**', async (route) => {
    const url = route.request().url()
    // Skip auth endpoints
    if (url.includes('/api/v1/') && !url.includes('/auth/')) {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify(defaultSessionExpiredResponse),
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
  response?: Partial<typeof defaultUnauthorizedResponse>
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
 * Mock rate limit error for token refresh
 */
export async function mockRefreshRateLimit(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 429,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Too Many Requests',
          message: 'Token refresh rate limit exceeded. Please wait before trying again.',
          retryAfter: 60,
        }),
      })
    }
  })
}

/**
 * Mock concurrent refresh requests (testing deduplication)
 * This mock returns different tokens for each request to verify only one actual refresh occurs
 */
export async function mockConcurrentRefreshRequests(page: Page): Promise<void> {
  let requestCount = 0

  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      requestCount++

      const accessToken = `mock-refreshed-token-${requestCount}-${Date.now()}`
      const refreshToken = `mock-refreshed-refresh-${requestCount}-${Date.now()}`

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        headers: {
          'Set-Cookie': [
            `${COOKIE_NAMES.ACCESS_TOKEN}=${accessToken}; Path=/; HttpOnly; SameSite=Strict`,
            `${COOKIE_NAMES.REFRESH_TOKEN}=${refreshToken}; Path=/; HttpOnly; SameSite=Strict`,
          ].join(', '),
        },
        body: JSON.stringify({
          access_token: accessToken,
          refresh_token: refreshToken,
          expires_in: 3600,
          token_type: 'Bearer',
        }),
      })
    }
  })
}

/**
 * Mock token refresh with delayed response (for testing loading states)
 */
export async function mockDelayedTokenRefresh(
  page: Page,
  delayMs: number = 1000
): Promise<void> {
  await page.route('**/api/v1/auth/refresh', async (route) => {
    if (route.request().method() === 'POST') {
      // Delay the response
      await new Promise(resolve => setTimeout(resolve, delayMs))

      const accessToken = 'mock-delayed-refresh-token-' + Date.now()
      const refreshToken = 'mock-delayed-refresh-refresh-' + Date.now()
      
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        headers: {
          'Set-Cookie': [
            `${COOKIE_NAMES.ACCESS_TOKEN}=${accessToken}; Path=/; HttpOnly; SameSite=Strict`,
            `${COOKIE_NAMES.REFRESH_TOKEN}=${refreshToken}; Path=/; HttpOnly; SameSite=Strict`,
          ].join(', '),
        },
        body: JSON.stringify({
          access_token: accessToken,
          refresh_token: refreshToken,
          expires_in: 3600,
          token_type: 'Bearer',
        }),
      })
    }
  })
}

/**
 * Mock user info endpoint
 */
export async function mockUserInfo(
  page: Page,
  userData?: { id?: string; email?: string; name?: string; plan?: string }
): Promise<void> {
  const response = {
    id: userData?.id || 'user-123',
    email: userData?.email || 'test@example.com',
    name: userData?.name || 'Test User',
    plan: userData?.plan || 'pro',
  }

  // Mock both endpoints to be safe
  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })

  await page.route('**/api/v1/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })
}

/**
 * Mock user info endpoint with 401
 */
export async function mockUserInfoUnauthorized(page: Page): Promise<void> {
  const response = {
    error: 'UNAUTHORIZED',
    message: 'Session expired',
    code: 'TOKEN_EXPIRED',
  }

  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })

  await page.route('**/api/v1/me', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })
}

/**
 * Setup comprehensive session mocks for authenticated tests
 */
export async function setupSessionMocks(page: Page): Promise<void> {
  await mockSuccessfulTokenRefresh(page)
  await mockUserInfo(page)
  await mockSessionStatusActive(page)
}

/**
 * Setup session mocks with expired token handling
 */
export async function setupSessionMocksWithExpiry(page: Page): Promise<void> {
  await mockSuccessfulTokenRefresh(page)
  await mockUserInfo(page)
  await mockSessionExpired(page)
}

/**
 * Setup all token-related mocks for full session testing
 */
export async function setupAllTokenMocks(page: Page): Promise<void> {
  await mockSuccessfulTokenRefresh(page)
  await mockExpiredRefreshToken(page)
  await mockInvalidRefreshToken(page)
  await mockRevokedRefreshToken(page)
  await mockUserInfo(page)
  await mockSessionStatusActive(page)
  await mockTokenValidationValid(page)
}

/**
 * Clear all mock routes (reset to default network behavior)
 */
export async function clearMockRoutes(page: Page): Promise<void> {
  await page.unroute('**/api/v1/**')
  await page.unroute('**/api/v1/auth/refresh')
}
