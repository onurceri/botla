/**
 * Cookie-Based Authentication Utilities for E2E Tests
 * 
 * This module provides utilities for managing authentication via HTTP-only cookies,
 * which is the authentication mechanism used by the Botla backend.
 * 
 * ## Architecture Overview
 * 
 * The Botla application uses **HttpOnly cookie-based authentication**:
 * - Backend sets `botla_token` and `botla_refresh_token` as HttpOnly, Secure cookies
 * - Frontend uses `withCredentials: true` in axios to automatically include cookies
 * - Cookies cannot be accessed/modified by JavaScript (security feature)
 * 
 * ## E2E Test Strategy
 * 
 * Since HttpOnly cookies cannot be set via `page.evaluate()`, we use two approaches:
 * 
 * 1. **Route Mocking**: Mock login/refresh API responses with proper `Set-Cookie` headers
 * 2. **Cookie Injection**: Use Playwright's `browserContext.addCookies()` for direct cookie setting
 * 
 * @module cookie-auth
 */

import { Page, BrowserContext, Cookie } from '@playwright/test'

// Cookie names (consistent with backend)
export const COOKIE_NAMES = {
  ACCESS_TOKEN: 'botla_token',
  REFRESH_TOKEN: 'botla_refresh_token',
} as const

// Default cookie options matching backend
export const DEFAULT_COOKIE_OPTIONS = {
  path: '/',
  httpOnly: true,
  secure: false, // false for localhost in tests
  sameSite: 'Strict' as const,
}

/**
 * User data stored in localStorage (for UI display purposes)
 */
export interface UserData {
  id: string
  email: string
  name: string
  full_name?: string
  plan?: string
  is_platform_admin?: boolean
}

/**
 * Options for setting up an authenticated session
 */
export interface AuthSessionOptions {
  accessToken?: string
  refreshToken?: string
  user?: Partial<UserData>
  secure?: boolean
}

/**
 * Default test user data
 */
export const DEFAULT_TEST_USER: UserData = {
  id: 'user-test-123',
  email: 'test@example.com',
  name: 'Test User',
  full_name: 'Test User',
  plan: 'pro',
  is_platform_admin: false,
}

/**
 * Default test organization
 */
export const DEFAULT_TEST_ORG = {
  id: 'org-test-123',
  name: 'Test Organization',
  slug: 'test-org',
  owner_id: DEFAULT_TEST_USER.id,
  plan_id: 'pro',
}

/**
 * Generate a mock JWT token for testing
 * Note: This is for mock purposes only, not cryptographically valid
 */
export function generateMockToken(options: {
  userId?: string
  email?: string
  name?: string
  expiresInSeconds?: number
  isPlatformAdmin?: boolean
  tokenType?: 'access' | 'refresh'
} = {}): string {
  const {
    userId = 'user-test-123',
    email = 'test@example.com',
    name = 'Test User',
    expiresInSeconds = 3600,
    isPlatformAdmin = false,
    tokenType = 'access',
  } = options

  const now = Math.floor(Date.now() / 1000)
  const header = { alg: 'HS256', typ: 'JWT' }
  const payload = {
    sub: userId,
    user_id: userId,
    email,
    name,
    iat: now,
    exp: now + expiresInSeconds,
    is_platform_admin: isPlatformAdmin,
    token_type: tokenType,
  }

  // base64url encode (not cryptographically valid, just for testing)
  const encodeBase64Url = (obj: object) => {
    const json = JSON.stringify(obj)
    const base64 = Buffer.from(json).toString('base64')
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
  }

  const encodedHeader = encodeBase64Url(header)
  const encodedPayload = encodeBase64Url(payload)
  const signature = 'mock-signature-' + Date.now()

  return `${encodedHeader}.${encodedPayload}.${signature}`
}

/**
 * Set authentication cookies directly in the browser context
 * This is the recommended way to set up authenticated tests
 * 
 * @example
 * ```typescript
 * test('authenticated user can access dashboard', async ({ page, context }) => {
 *   await setAuthCookies(context)
 *   await setupAuthenticatedMocks(page)
 *   await page.goto('/dashboard')
 *   // User is now authenticated
 * })
 * ```
 */
export async function setAuthCookies(
  context: BrowserContext,
  options: AuthSessionOptions = {}
): Promise<void> {
  const {
    accessToken = generateMockToken({ tokenType: 'access' }),
    refreshToken = generateMockToken({ tokenType: 'refresh', expiresInSeconds: 604800 }),
    secure = false,
  } = options

  // Get base URL from environment or default
  const baseUrl = process.env.VITE_API_BASE_URL || 'http://localhost:5173'
  const url = new URL(baseUrl)
  // When using 'url', we generally don't need 'domain'. 'path' is inferred from url usually, but can be explicit.
  // The error "Cookie should have either url or path" implies we might be missing something valid or providing a conflict.
  // Actually, strictly: url OR (domain AND path).
  // If we use url, we should drop domain.
  
  // Let's create specific cookie objects.
  const accessTokenCookie: Cookie = {
    name: COOKIE_NAMES.ACCESS_TOKEN,
    value: accessToken,
    url: baseUrl, 
    // when url is present, domain is not needed. path defaults to / if not specified or implied.
    httpOnly: true,
    secure,
    sameSite: 'Strict',
    expires: Math.floor(Date.now() / 1000) + 3600, // 1 hour
  } as unknown as Cookie

  const refreshTokenCookie: Cookie = {
    name: COOKIE_NAMES.REFRESH_TOKEN,
    value: refreshToken,
    url: baseUrl,
    httpOnly: true,
    secure,
    sameSite: 'Strict',
    expires: Math.floor(Date.now() / 1000) + 604800, // 7 days
  } as unknown as Cookie

  await context.addCookies([accessTokenCookie, refreshTokenCookie])
}

/**
 * Clear authentication cookies
 */
export async function clearAuthCookies(context: BrowserContext): Promise<void> {
  await context.clearCookies()
}

/**
 * Set user data in localStorage (for UI display)
 * Note: Authentication is done via cookies, this is just for UI purposes
 */
export async function setUserDataInStorage(
  page: Page,
  user: Partial<UserData> = {}
): Promise<void> {
  const userData = { ...DEFAULT_TEST_USER, ...user }
  
  await page.evaluate((data) => {
    localStorage.setItem('botla_user', JSON.stringify(data))
  }, userData)
}

/**
 * Set organization data in localStorage
 */
export async function setOrgDataInStorage(
  page: Page,
  orgId: string = DEFAULT_TEST_ORG.id
): Promise<void> {
  await page.evaluate((id) => {
    localStorage.setItem('botla_last_org_id', id)
  }, orgId)
}

/**
 * Setup all required mocks for authenticated routes
 * This should be called before navigating to authenticated pages
 */
export async function setupAuthenticatedMocks(
  page: Page,
  options: {
    user?: Partial<UserData>
    includeOrg?: boolean
    includeChatbots?: boolean
    includeOnboarding?: boolean
  } = {}
): Promise<void> {
  const {
    user = {},
    includeOrg = true,
    includeChatbots = true,
    includeOnboarding = true,
  } = options

  const userData = { ...DEFAULT_TEST_USER, ...user }

  // Mock user profile endpoint
  await page.route('**/api/v1/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(userData),
    })
  })

  // Mock auth/me endpoint
  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(userData),
    })
  })

  // Mock onboarding status
  if (includeOnboarding) {
    await page.route('**/api/v1/me/onboarding', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ completed: true, skipped: false }),
      })
    })
  }

  // Mock organizations
  if (includeOrg) {
    await page.route('**/api/v1/organizations', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([{
            ...DEFAULT_TEST_ORG,
            owner_id: userData.id,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          }]),
        })
      } else {
        await route.continue()
      }
    })

    await page.route('**/api/v1/organizations/*/workspaces', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{
          id: 'ws-test-123',
          organization_id: DEFAULT_TEST_ORG.id,
          name: 'Default Workspace',
          slug: 'default',
          created_at: new Date().toISOString(),
        }]),
      })
    })
  }

  // Mock chatbots list
  if (includeChatbots) {
    await page.route('**/api/v1/chatbots', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([]),
        })
      } else {
        await route.continue()
      }
    })
  }

  // Mock refresh token endpoint
  await page.route('**/api/v1/auth/refresh', async (route) => {
    const newAccessToken = generateMockToken({ tokenType: 'access' })
    const newRefreshToken = generateMockToken({ tokenType: 'refresh', expiresInSeconds: 604800 })
    
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      headers: {
        'Set-Cookie': [
          `${COOKIE_NAMES.ACCESS_TOKEN}=${newAccessToken}; Path=/; HttpOnly; SameSite=Strict`,
          `${COOKIE_NAMES.REFRESH_TOKEN}=${newRefreshToken}; Path=/; HttpOnly; SameSite=Strict`,
        ].join(', '),
      },
      body: JSON.stringify({
        token: newAccessToken,
        refresh_token: newRefreshToken,
      }),
    })
  })
}

/**
 * Setup login mock that properly sets cookies via headers
 */
export async function setupLoginMock(
  page: Page,
  options: {
    user?: Partial<UserData>
    shouldSucceed?: boolean
    errorMessage?: string
  } = {}
): Promise<void> {
  const { user = {}, shouldSucceed = true, errorMessage = 'Invalid credentials' } = options
  const userData = { ...DEFAULT_TEST_USER, ...user }

  await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() !== 'POST') {
      await route.continue()
      return
    }

    if (!shouldSucceed) {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Unauthorized',
          message: errorMessage,
          code: 'AUTH_001',
        }),
      })
      return
    }

    const accessToken = generateMockToken({ 
      userId: userData.id,
      email: userData.email,
      name: userData.name,
      tokenType: 'access',
    })
    const refreshToken = generateMockToken({ 
      userId: userData.id,
      tokenType: 'refresh',
      expiresInSeconds: 604800,
    })

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
        token: accessToken,
        refresh_token: refreshToken,
      }),
    })
  })
}

/**
 * Setup logout mock that clears cookies
 */
export async function setupLogoutMock(page: Page): Promise<void> {
  await page.route('**/api/v1/auth/logout', async (route) => {
    if (route.request().method() !== 'POST') {
      await route.continue()
      return
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      headers: {
        'Set-Cookie': [
          `${COOKIE_NAMES.ACCESS_TOKEN}=; Path=/; HttpOnly; Max-Age=0`,
          `${COOKIE_NAMES.REFRESH_TOKEN}=; Path=/; HttpOnly; Max-Age=0`,
        ].join(', '),
      },
      body: JSON.stringify({ success: true }),
    })
  })
}

/**
 * Full authenticated session setup - combines cookies, storage, and mocks
 * This is the recommended one-liner for setting up authenticated tests
 * 
 * @example
 * ```typescript
 * test.beforeEach(async ({ page, context }) => {
 *   await setupAuthenticatedSession(page, context)
 * })
 * 
 * test('user can access dashboard', async ({ page }) => {
 *   await page.goto('/dashboard')
 *   await expect(page.getByTestId('dashboard')).toBeVisible()
 * })
 * ```
 */
export async function setupAuthenticatedSession(
  page: Page,
  context: BrowserContext,
  options: AuthSessionOptions & {
    user?: Partial<UserData>
  } = {}
): Promise<void> {
  const { user = {}, ...authOptions } = options
  
  // Set auth cookies
  await setAuthCookies(context, authOptions)
  
  // Setup all mocks
  await setupAuthenticatedMocks(page, { user })
  await setupLogoutMock(page)
}

/**
 * Check if user is authenticated by checking cookies
 */
export async function isAuthenticated(context: BrowserContext): Promise<boolean> {
  const cookies = await context.cookies()
  const hasAccessToken = cookies.some(c => c.name === COOKIE_NAMES.ACCESS_TOKEN)
  const hasRefreshToken = cookies.some(c => c.name === COOKIE_NAMES.REFRESH_TOKEN)
  return hasAccessToken && hasRefreshToken
}

/**
 * Get current auth cookies
 */
export async function getAuthCookies(context: BrowserContext): Promise<{
  accessToken: string | null
  refreshToken: string | null
}> {
  const cookies = await context.cookies()
  const accessToken = cookies.find(c => c.name === COOKIE_NAMES.ACCESS_TOKEN)?.value || null
  const refreshToken = cookies.find(c => c.name === COOKIE_NAMES.REFRESH_TOKEN)?.value || null
  return { accessToken, refreshToken }
}
