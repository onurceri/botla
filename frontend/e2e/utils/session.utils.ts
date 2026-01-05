import { Page, expect } from '@playwright/test'

// Storage key constants (consistent with auth.mocks.ts)
export const TOKEN_KEY = 'botla_token'
export const REFRESH_TOKEN_KEY = 'botla_refresh_token'
export const USER_KEY = 'botla_user'

/**
 * Session data interface
 */
export interface SessionData {
  accessToken: string | null
  refreshToken: string | null
  user: Record<string, unknown> | null
}

/**
 * Get access token from localStorage
 */
export async function getAccessToken(page: Page): Promise<string | null> {
  return await page.evaluate(() => localStorage.getItem('botla_token'))
}

/**
 * Get refresh token from localStorage
 */
export async function getRefreshToken(page: Page): Promise<string | null> {
  return await page.evaluate(() => localStorage.getItem('botla_refresh_token'))
}

/**
 * Get user data from localStorage
 */
export async function getUserData(page: Page): Promise<Record<string, unknown> | null> {
  return await page.evaluate(() => {
    const userStr = localStorage.getItem('botla_user')
    return userStr ? JSON.parse(userStr) : null
  })
}

/**
 * Get complete session data from localStorage
 */
export async function getSessionData(page: Page): Promise<SessionData> {
  return await page.evaluate(() => ({
    accessToken: localStorage.getItem('botla_token'),
    refreshToken: localStorage.getItem('botla_refresh_token'),
    user: (() => {
      const userStr = localStorage.getItem('botla_user')
      return userStr ? JSON.parse(userStr) : null
    })(),
  }))
}

/**
 * Clear all session-related storage
 */
export async function clearSessionStorage(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    localStorage.removeItem('botla_user')
    sessionStorage.clear()
  })
}

/**
 * Set valid session tokens in localStorage
 */
export async function setValidSession(
  page: Page,
  accessToken: string = 'mock-access-token-' + Date.now(),
  refreshToken: string = 'mock-refresh-token-' + Date.now()
): Promise<void> {
  await page.evaluate(
    ({ accessToken, refreshToken }) => {
      localStorage.setItem('botla_token', accessToken)
      localStorage.setItem('botla_refresh_token', refreshToken)
      localStorage.setItem(
        'botla_user',
        JSON.stringify({
          id: 'user-123',
          email: 'test@example.com',
          name: 'Test User',
          plan: 'pro',
        })
      )
    },
    { accessToken, refreshToken }
  )
}

/**
 * Set an expired access token for testing session expiry
 */
export async function setExpiredToken(page: Page): Promise<void> {
  // Create a JWT that's expired (exp timestamp is in the past)
  const expiredToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.' +
    'eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMiwicm9sZSI6InVzZXIiLCJleHAiOjE1MTYyMzkwMjJ9.expired_signature'

  await page.evaluate((token) => {
    localStorage.setItem('botla_token', token)
  }, expiredToken)
}

/**
 * Set a valid but expiring soon token (for testing refresh flow)
 */
export async function setExpiringSoonToken(page: Page): Promise<void> {
  // Token that expires in 5 minutes
  const expiringToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.' +
    'eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMiwicm9sZSI6InVzZXIiLCJleHAiOjo1fQ.expiring_signature'

  await page.evaluate((token) => {
    localStorage.setItem('botla_token', token)
  }, expiringToken)
}

/**
 * Expect that all session tokens have been cleared
 */
export async function expectTokensCleared(page: Page): Promise<void> {
  const accessToken = await getAccessToken(page)
  const refreshToken = await getRefreshToken(page)

  if (accessToken !== null) {
    throw new Error(`Access token was not cleared. Found: ${accessToken.substring(0, 20)}...`)
  }
  if (refreshToken !== null) {
    throw new Error(`Refresh token was not cleared. Found: ${refreshToken.substring(0, 20)}...`)
  }
}

/**
 * Expect that user data has been cleared from storage
 */
export async function expectUserDataCleared(page: Page): Promise<void> {
  const userData = await getUserData(page)

  if (userData !== null) {
    throw new Error(`User data was not cleared. Found: ${JSON.stringify(userData)}`)
  }
}

/**
 * Expect complete session clearance (tokens + user data)
 */
export async function expectSessionCleared(page: Page): Promise<void> {
  await expectTokensCleared(page)
  await expectUserDataCleared(page)
}

/**
 * Expect session to be valid (tokens present)
 */
export async function expectValidSession(page: Page): Promise<void> {
  const sessionData = await getSessionData(page)

  expect(sessionData.accessToken).not.toBeNull()
  expect(sessionData.refreshToken).not.toBeNull()
  expect(sessionData.user).not.toBeNull()
}

/**
 * Check if user is currently authenticated (has valid tokens)
 */
export async function isAuthenticated(page: Page): Promise<boolean> {
  const accessToken = await getAccessToken(page)
  const refreshToken = await getRefreshToken(page)
  return accessToken !== null && refreshToken !== null
}

/**
 * Wait for logout API response to complete
 */
export async function waitForLogoutResponse(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForResponse(
    (response) =>
      response.url().includes('/api/v1/auth/logout') &&
      response.request().method() === 'POST',
    { timeout }
  )
}

/**
 * Wait for session to be cleared (tokens removed from storage)
 */
export async function waitForSessionCleared(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForFunction(
    () => {
      const token = localStorage.getItem('botla_token')
      const refreshToken = localStorage.getItem('botla_refresh_token')
      return token === null && refreshToken === null
    },
    { timeout }
  )
}

/**
 * Simulate BroadcastChannel message for multi-tab sync
 */
export async function simulateBroadcastMessage(
  page: Page,
  channel: string = 'auth_channel',
  message: string = 'session_terminated'
): Promise<void> {
  await page.evaluate(({ channel, message }) => {
    const bc = new BroadcastChannel(channel)
    bc.postMessage(message)
  }, { channel, message })
}

/**
 * Set up BroadcastChannel listener for session termination
 */
export async function setupBroadcastListener(
  page: Page,
  channel: string = 'auth_channel',
  redirectUrl: string = '/login'
): Promise<void> {
  await page.evaluate(({ channel, redirectUrl }) => {
    const bc = new BroadcastChannel(channel)
    bc.onmessage = (event) => {
      if (event.data === 'session_terminated') {
        window.location.href = redirectUrl
      }
    }
  }, { channel, redirectUrl })
}

/**
 * Get session expiration time from token (mock implementation)
 * Returns null if token is not a valid JWT or doesn't contain exp claim
 */
export async function getTokenExpiration(page: Page): Promise<number | null> {
  const token = await getAccessToken(page)
  if (!token) return null

  try {
    // Basic JWT parsing (header.payload.signature)
    const parts = token.split('.')
    if (parts.length !== 3) return null

    const payload = JSON.parse(atob(parts[1]))
    return payload.exp ? payload.exp * 1000 : null // Convert to milliseconds
  } catch {
    return null
  }
}

/**
 * Check if token is expired
 */
export async function isTokenExpired(page: Page): Promise<boolean> {
  const expiration = await getTokenExpiration(page)
  if (expiration === null) return false // Can't determine, assume not expired
  return Date.now() > expiration
}
