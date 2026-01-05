import { Page } from '@playwright/test'

/**
 * JWT Token structure for testing purposes
 * Matches the expected JWT payload structure from the backend
 */
export interface JWTPayload {
  sub: string // User ID
  email: string
  name: string
  iat: number // Issued at timestamp
  exp: number // Expiration timestamp
  role: string
  plan?: string
}

/**
 * Session tokens interface
 */
export interface SessionTokens {
  accessToken: string
  refreshToken: string
  user: UserData
}

/**
 * User data stored in localStorage
 */
export interface UserData {
  id: string
  email: string
  name: string
  plan?: string
}

/**
 * Token generation options
 */
export interface TokenOptions {
  expiresInSeconds?: number
  userId?: string
  email?: string
  name?: string
  role?: string
  plan?: string
}

/**
 * Default token expiration times (in seconds)
 */
export const TOKEN_EXPIRY = {
  ACCESS_TOKEN: 3600, // 1 hour
  REFRESH_TOKEN: 604800, // 7 days
  SHORT: 300, // 5 minutes (for testing refresh flow)
  EXPIRED: -3600, // Expired 1 hour ago
}

/**
 * Storage key constants (consistent with application)
 */
export const STORAGE_KEYS = {
  ACCESS_TOKEN: 'botla_token',
  REFRESH_TOKEN: 'botla_refresh_token',
  USER: 'botla_user',
  REMEMBER_ME: 'remember_me',
}

/**
 * Generate a mock JWT token with custom expiration
 * This creates a token that follows JWT structure but is not cryptographically signed
 * For testing purposes only
 */
export function generateMockToken(options: TokenOptions = {}): string {
  const {
    expiresInSeconds = TOKEN_EXPIRY.ACCESS_TOKEN,
    userId = 'test-user-' + Date.now(),
    email = 'test@example.com',
    name = 'Test User',
    role = 'user',
    plan = 'pro',
  } = options

  const now = Math.floor(Date.now() / 1000)
  const header = { alg: 'HS256', typ: 'JWT' }

  const payload: JWTPayload = {
    sub: userId,
    email,
    name,
    iat: now,
    exp: now + expiresInSeconds,
    role,
    plan,
  }

  const encodedHeader = btoa(JSON.stringify(header))
  const encodedPayload = btoa(JSON.stringify(payload))

  // Mock signature for testing
  const signature = 'mock-signature-' + Date.now()

  return `${encodedHeader}.${encodedPayload}.${signature}`
}

/**
 * Generate an expired access token (for testing token expiry handling)
 */
export function generateExpiredToken(options: Omit<TokenOptions, 'expiresInSeconds'> = {}): string {
  return generateMockToken({
    ...options,
    expiresInSeconds: TOKEN_EXPIRY.EXPIRED,
  })
}

/**
 * Generate a token that expires in a short time (for testing refresh flow)
 */
export function generateExpiringSoonToken(options: Omit<TokenOptions, 'expiresInSeconds'> = {}): string {
  return generateMockToken({
    ...options,
    expiresInSeconds: TOKEN_EXPIRY.SHORT,
  })
}

/**
 * Generate a valid token with custom expiration
 */
export function generateValidToken(expiresInSeconds: number = TOKEN_EXPIRY.ACCESS_TOKEN): string {
  return generateMockToken({ expiresInSeconds })
}

/**
 * Parse JWT payload from a token
 */
export function parseTokenPayload(token: string): JWTPayload | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return null
    }

    const payload = JSON.parse(atob(parts[1]))
    return payload as JWTPayload
  } catch {
    return null
  }
}

/**
 * Check if a token is expired
 */
export function isTokenExpired(token: string): boolean {
  const payload = parseTokenPayload(token)
  if (!payload || !payload.exp) {
    return false
  }

  const now = Math.floor(Date.now() / 1000)
  return payload.exp < now
}

/**
 * Get token expiration time in milliseconds
 */
export function getTokenExpiryTime(token: string): number | null {
  const payload = parseTokenPayload(token)
  if (!payload || !payload.exp) {
    return null
  }

  return payload.exp * 1000 // Convert to milliseconds
}

/**
 * Calculate seconds until token expiration
 */
export function getSecondsUntilExpiry(token: string): number | null {
  const expiryTime = getTokenExpiryTime(token)
  if (expiryTime === null) {
    return null
  }

  const now = Date.now()
  const secondsRemaining = Math.floor((expiryTime - now) / 1000)

  return Math.max(0, secondsRemaining)
}

/**
 * Set session tokens in localStorage using addInitScript
 * This must be called before page navigation to avoid redirect issues
 */
export async function setSessionStorage(
  page: Page,
  tokens: SessionTokens
): Promise<void> {
  await page.addInitScript(
    ({ accessToken, refreshToken, user }) => {
      localStorage.setItem('botla_token', accessToken)
      localStorage.setItem('botla_refresh_token', refreshToken)
      localStorage.setItem('botla_user', JSON.stringify(user))
    },
    {
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
      user: tokens.user,
    }
  )
}

/**
 * Set session with Remember Me flag enabled
 */
export async function setRememberMeSession(
  page: Page,
  options: TokenOptions = {}
): Promise<void> {
  const accessToken = generateMockToken(options)
  const refreshToken = generateMockToken({ ...options, expiresInSeconds: TOKEN_EXPIRY.REFRESH_TOKEN })

  await page.addInitScript(
    ({ accessToken, refreshToken }) => {
      localStorage.setItem('botla_token', accessToken)
      localStorage.setItem('botla_refresh_token', refreshToken)
      localStorage.setItem('botla_user', JSON.stringify({
        id: 'test-user-' + Date.now(),
        email: 'test@example.com',
        name: 'Test User',
        plan: 'pro',
      }))
      localStorage.setItem('remember_me', 'true')
    },
    { accessToken, refreshToken }
  )
}

/**
 * Set session with expired access token (for testing refresh flow)
 */
export async function setExpiredSession(
  page: Page,
  validRefreshToken?: string
): Promise<void> {
  const expiredToken = generateExpiredToken()
  const refreshToken = validRefreshToken || generateMockToken({ expiresInSeconds: TOKEN_EXPIRY.REFRESH_TOKEN })

  await page.addInitScript(
    ({ accessToken, refreshToken }) => {
      localStorage.setItem('botla_token', accessToken)
      localStorage.setItem('botla_refresh_token', refreshToken)
      localStorage.setItem('botla_user', JSON.stringify({
        id: 'test-user-' + Date.now(),
        email: 'test@example.com',
        name: 'Test User',
        plan: 'pro',
      }))
    },
    { accessToken: expiredToken, refreshToken }
  )
}

/**
 * Set session with expiring soon token (for testing automatic refresh)
 */
export async function setExpiringSession(
  page: Page,
  expiresInSeconds: number = TOKEN_EXPIRY.SHORT
): Promise<void> {
  const expiringToken = generateMockToken({ expiresInSeconds })
  const refreshToken = generateMockToken({ expiresInSeconds: TOKEN_EXPIRY.REFRESH_TOKEN })

  await page.addInitScript(
    ({ accessToken, refreshToken }) => {
      localStorage.setItem('botla_token', accessToken)
      localStorage.setItem('botla_refresh_token', refreshToken)
      localStorage.setItem('botla_user', JSON.stringify({
        id: 'test-user-' + Date.now(),
        email: 'test@example.com',
        name: 'Test User',
        plan: 'pro',
      }))
    },
    { accessToken: expiringToken, refreshToken }
  )
}

/**
 * Clear all session-related storage
 */
export async function clearSessionStorage(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('botla_token')
    localStorage.removeItem('botla_refresh_token')
    localStorage.removeItem('botla_user')
    localStorage.removeItem('remember_me')
    sessionStorage.clear()
  })
}

/**
 * Clear only authentication tokens (keep user data for Remember Me)
 */
export async function clearAuthTokens(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN)
    localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN)
  })
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
export async function getUserData(page: Page): Promise<UserData | null> {
  return await page.evaluate(() => {
    const userStr = localStorage.getItem('botla_user')
    return userStr ? JSON.parse(userStr) : null
  })
}

/**
 * Get complete session data from localStorage
 */
export async function getSessionData(page: Page): Promise<SessionTokens | null> {
  return await page.evaluate(() => {
    const accessToken = localStorage.getItem('botla_token')
    const refreshToken = localStorage.getItem('botla_refresh_token')
    const userStr = localStorage.getItem('botla_user')

    if (!accessToken || !refreshToken || !userStr) {
      return null
    }

    return {
      accessToken,
      refreshToken,
      user: JSON.parse(userStr),
    }
  })
}

/**
 * Check if Remember Me is enabled
 */
export async function isRememberMeEnabled(page: Page): Promise<boolean> {
  return await page.evaluate(() => {
    return localStorage.getItem('remember_me') === 'true'
  })
}

/**
 * Set Remember Me flag
 */
export async function setRememberMe(page: Page, enabled: boolean = true): Promise<void> {
  await page.evaluate((enabled) => {
    localStorage.setItem('remember_me', enabled ? 'true' : 'false')
  }, enabled)
}

/**
 * Check if user is currently authenticated
 */
export async function isAuthenticated(page: Page): Promise<boolean> {
  const accessToken = await getAccessToken(page)
  const refreshToken = await getRefreshToken(page)

  if (!accessToken || !refreshToken) {
    return false
  }

  // Check if access token is expired
  if (isTokenExpired(accessToken)) {
    return false
  }

  return true
}

/**
 * Get token expiration time from storage
 */
export async function getTokenExpiry(page: Page): Promise<number | null> {
  const token = await getAccessToken(page)
  if (!token) return null

  return getTokenExpiryTime(token)
}

/**
 * Check if token in storage is expired
 */
export async function isTokenExpiredInStorage(page: Page): Promise<boolean> {
  const token = await getAccessToken(page)
  if (!token) return true

  return isTokenExpired(token)
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
 * Wait for session to be valid (tokens present and not expired)
 */
export async function waitForSessionValid(
  page: Page,
  timeout: number = 10000
): Promise<void> {
  await page.waitForFunction(
    () => {
      const token = localStorage.getItem('botla_token')
      const refreshToken = localStorage.getItem('botla_refresh_token')

      if (!token || !refreshToken) {
        return false
      }

      // Check token expiration
      try {
        const payload = JSON.parse(atob(token.split('.')[1]))
        if (payload.exp) {
          const now = Math.floor(Date.now() / 1000)
          if (payload.exp < now) {
            return false
          }
        }
      } catch {
        return false
      }

      return true
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
 * Create a complete session with all tokens and user data
 */
export function createSession(
  tokenOptions: TokenOptions = {},
  userData?: Partial<UserData>
): SessionTokens {
  const timestamp = Date.now()
  return {
    accessToken: generateMockToken(tokenOptions),
    refreshToken: generateMockToken({
      ...tokenOptions,
      expiresInSeconds: TOKEN_EXPIRY.REFRESH_TOKEN,
    }),
    user: {
      id: userData?.id || 'test-user-' + timestamp,
      email: userData?.email || 'test@example.com',
      name: userData?.name || 'Test User',
      plan: userData?.plan || 'pro',
    },
  }
}

/**
 * Create a session with expired access token
 */
export function createExpiredSession(refreshTokenOptions: TokenOptions = {}): SessionTokens {
  return {
    accessToken: generateExpiredToken(),
    refreshToken: generateMockToken({
      ...refreshTokenOptions,
      expiresInSeconds: TOKEN_EXPIRY.REFRESH_TOKEN,
    }),
    user: {
      id: 'test-user-' + Date.now(),
      email: 'test@example.com',
      name: 'Test User',
      plan: 'pro',
    },
  }
}

/**
 * Create a session with expiring access token
 */
export function createExpiringSession(expiresInSeconds: number = TOKEN_EXPIRY.SHORT): SessionTokens {
  return {
    accessToken: generateMockToken({ expiresInSeconds }),
    refreshToken: generateMockToken({ expiresInSeconds: TOKEN_EXPIRY.REFRESH_TOKEN }),
    user: {
      id: 'test-user-' + Date.now(),
      email: 'test@example.com',
      name: 'Test User',
      plan: 'pro',
    },
  }
}

/**
 * Extract token claims for verification
 */
export async function getTokenClaims(page: Page): Promise<{
  accessToken: JWTPayload | null
  refreshToken: JWTPayload | null
}> {
  const accessToken = await getAccessToken(page)
  const refreshToken = await getRefreshToken(page)

  return {
    accessToken: accessToken ? parseTokenPayload(accessToken) : null,
    refreshToken: refreshToken ? parseTokenPayload(refreshToken) : null,
  }
}
