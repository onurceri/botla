/**
 * E2E Mock Utilities - Main Export File
 *
 * This file provides a clean single entry point for all API mocks.
 * Import mocks from here to keep test files clean and maintainable.
 *
 * @example
 * import { mockSuccessfulLogin, mockUserInfo } from './mocks'
 */

// ============================================================================
// Auth Mocks
// ============================================================================
export {
  mockSuccessfulLogin,
  mockFailedLogin,
  mockNetworkError as mockAuthNetworkError,
  mockServerError as mockAuthServerError,
  mockRateLimitError as mockAuthRateLimitError,
  mockValidationError as mockAuthValidationError,
  mockRememberMeLogin,
  mockLogout,
  mockRefreshToken as mockAuthRefreshToken,
  mockForgotPassword,
  mockResetPassword,
  setupAuthMocks,
  clearAuthStorage,
  setAuthTokens,
  getAuthTokens,
  waitForLoginResponse,
  waitForLoadingToComplete,
  type LoginSuccessResponse,
  type LoginErrorResponse,
} from './auth.mocks'

// ============================================================================
// Token & Session Mocks
// ============================================================================
export {
  // Token refresh
  mockSuccessfulTokenRefresh,
  mockTokenRefreshAccessTokenOnly,
  mockExpiredRefreshToken,
  mockInvalidRefreshToken,
  mockRevokedRefreshToken,
  mockRefreshTokenServerError,
  mockRefreshTokenNetworkError,
  mockConcurrentRefreshRequests,
  mockDelayedTokenRefresh,
  mockRefreshRateLimit,
  // Session status
  mockSessionStatusActive,
  mockSessionStatusExpired,
  mockSessionExpired as mockTokenSessionExpired,
  // Token validation
  mockTokenValidationValid,
  mockTokenValidationInvalid,
  // User info
  mockUserInfo,
  mockUserInfoUnauthorized,
  // 401 handling
  mockUnauthorizedForApi,
  mockEndpointUnauthorized as mockTokenEndpointUnauthorized,
  // Setup helpers
  setupSessionMocks as setupTokenSessionMocks,
  setupSessionMocksWithExpiry,
  setupAllTokenMocks,
  clearMockRoutes,
  // Types
  type TokenRefreshSuccessResponse,
  type TokenRefreshErrorResponse,
  type SessionStatusResponse,
  type TokenValidationResponse,
} from './tokens.mocks'

// ============================================================================
// Register Mocks
// ============================================================================
export {
  mockSuccessfulRegistration,
  mockEmailExists,
  mockValidationError as mockRegisterValidationError,
  mockNetworkError as mockRegisterNetworkError,
  mockServerError as mockRegisterServerError,
  mockRateLimitError as mockRegisterRateLimitError,
  clearAuthStorage as clearRegisterAuthStorage,
  setAuthTokens as setRegisterAuthTokens,
  getAuthTokens as getRegisterAuthTokens,
  waitForRegistrationResponse,
  waitForLoadingToComplete as waitForRegisterLoadingToComplete,
  verifyTokensStored,
  type RegisterSuccessResponse,
  type RegisterErrorResponse,
} from './register.mocks'

// ============================================================================
// Session/Logout Mocks
// ============================================================================
export {
  mockSuccessfulLogout,
  mockLogoutServerError,
  mockLogoutNetworkError,
  mockUnauthorized,
  mockSessionExpired,
  mockEndpointUnauthorized,
  mockRefreshToken,
  mockRefreshTokenFailure,
  mockSessionValid,
  mockSessionInvalid,
  mockRateLimitError as mockSessionRateLimitError,
  setupSessionMocks,
  abortAllRequests,
  type LogoutSuccessResponse,
  type LogoutErrorResponse,
  type SessionExpiredResponse,
  type UnauthorizedResponse,
} from './session.mocks'
