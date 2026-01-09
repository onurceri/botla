/**
 * E2E Test Utilities - Main Export File
 *
 * This file provides a clean single entry point for all E2E test utilities.
 * Import utilities from here to keep test files clean and maintainable.
 *
 * @example
 * import { setAuthCookies, mockUserInfo, TEST_IDS } from './utils'
 */

// ============================================================================
// Authentication & Session Management
// ============================================================================
export {
  // Cookie management
  COOKIE_NAMES,
  DEFAULT_COOKIE_OPTIONS,
  setAuthCookies,
  clearAuthCookies,
  getAuthCookies,
  isAuthenticated,
  // User and org data
  DEFAULT_TEST_USER,
  DEFAULT_TEST_ORG,
  setUserDataInStorage,
  setOrgDataInStorage,
  // Session setup (recommended one-liners)
  setupAuthenticatedSession,
  setupAuthenticatedMocks,
  setupLoginMock,
  setupLogoutMock,
  // Token generation
  generateMockToken,
  // Types
  type UserData,
  type AuthSessionOptions,
} from './cookie-auth'

// Re-export session manager utilities (for backward compatibility)
export {
  TOKEN_EXPIRY,
  STORAGE_KEYS,
  // Token utilities
  generateExpiredToken,
  generateExpiringSoonToken,
  generateValidToken,
  parseTokenPayload,
  isTokenExpired,
  getTokenExpiryTime,
  getSecondsUntilExpiry,
  // Session storage
  setSessionStorage,
  clearSessionStorage,
  getAccessToken,
  getRefreshToken,
  getUserData,
  getSessionData,
  // Remember Me
  isRememberMeEnabled,
  setRememberMe,
  setRememberMeSession,
  setExpiredSession,
  setExpiringSession,
  // Session wait helpers
  waitForSessionCleared,
  waitForSessionValid,
  // Session factory functions
  createSession,
  createExpiredSession,
  createExpiringSession,
  // Types
  type JWTPayload,
  type SessionTokens,
  type TokenOptions,
} from './session-manager'

// ============================================================================
// Test Constants
// ============================================================================
export { TEST_IDS, TURKISH, ENGLISH, PAGE_URLS } from '../test-constants'
