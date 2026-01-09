# Task: Implement Session Management Tests

> **Task ID**: 05-auth-sessions  
> **Source**: TEST_PATHS.md (comprehensive session testing)  
> **Priority**: Highest (Authentication)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 02-auth-login.md, 04-auth-logout.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for Session Management. This task covers token refresh, session persistence, Remember Me functionality, and session security features.

### Context

Session Management is critical for authentication security and user experience. Testing this functionality ensures:
- Token refresh works seamlessly
- Sessions persist across browser restarts
- Remember Me functionality works correctly
- Session security is maintained
- Token expiration is handled gracefully

### Reference Specifications

From various TEST_PATHS.md sections and authentication patterns:

#### Token Management Flow

```
Token Lifecycle
├── Login with credentials
│   ├── Receive: Access token (1 hour expiry)
│   ├── Receive: Refresh token (7 day expiry)
│   └── Store: In secure storage
│
├── Access token expires
│   ├── Intercept: 401 response
│   ├── Trigger: Token refresh
│   ├── Request: New access token
│   ├── Receive: Fresh tokens
│   └── Continue: Original request
│
├── Refresh token expires
│   ├── Refresh: Fails with 401
│   ├── Trigger: Force re-login
│   ├── Show: Session expired modal
│   └── Redirect: /login
│
└── Logout
    ├── Revoke: Refresh token
    ├── Clear: All tokens
    └── Redirect: /login
```

#### Remember Me Flow

```
Remember Me Feature
├── Check: Remember me checkbox on login
│   ├── Store: Refresh token in localStorage
│   └── Store: User data in localStorage
│
├── Session expires
│   ├── Check: Remember me enabled
│   ├── Use: Refresh token to get new access token
│   └── Continue: User session without login
│
└── Browser restart
    ├── Check: Stored refresh token
    ├── Validate: Token not expired
    ├── Refresh: Access token automatically
    └── Restore: User session
```

### Implementation Requirements

1. **Create Session Test File** (`frontend/e2e/sessions.spec.ts`)
   - Token refresh tests
   - Session persistence tests
   - Remember Me tests
   - Session security tests

2. **Create Session Utilities** (`frontend/e2e/utils/session-manager.ts`)
   - Token generation helpers
   - Token expiration helpers
   - Storage management helpers

3. **Create Session Mocks** (`frontend/e2e/mocks/tokens.mocks.ts`)
   - Token refresh endpoint
   - Token validation endpoint
   - Session status endpoint

4. **Update Authentication Tests** to include session scenarios

### Expected Deliverables

1. `frontend/e2e/sessions.spec.ts` - Comprehensive session tests
2. `frontend/e2e/utils/session-manager.ts` - Session utilities
3. `frontend/e2e/mocks/tokens.mocks.ts` - Token mock handlers
4. Updated auth.spec.ts with Remember Me tests

---

## Implementation Plan

### Phase 1: Setup and Utilities

- [x] Create `frontend/e2e/utils/session-manager.ts`:
  - [x] Token generation helpers
  - [x] Token parsing utilities
  - [x] Expiration simulation
  - [x] Storage manipulation
- [x] Create `frontend/e2e/mocks/tokens.mocks.ts`:
  - [x] Mock token refresh endpoint
  - [x] Mock token validation
  - [x] Mock session status
- [x] Create `frontend/e2e/pages/session.page.ts` (if needed)

### Phase 2: Token Refresh Tests

- [x] Test: Automatic token refresh on 401
- [x] Test: Token refresh with valid refresh token
- [x] Test: Token refresh with expired refresh token
- [x] Test: Token refresh API error handling
- [x] Test: Multiple concurrent refresh requests
- [x] Test: Refresh request deduplication
- [x] Test: Token refresh during active user session

### Phase 3: Remember Me Tests

- [x] Test: Login with Remember Me checked
- [x] Test: Token persistence in localStorage
- [x] Test: Session restoration after browser restart
- [x] Test: Remember Me with expired tokens
- [x] Test: Remember Me with valid tokens
- [x] Test: Remember Me checkbox persistence
- [x] Test: Unchecking Remember Me clears storage

### Phase 4: Session Persistence Tests

- [x] Test: Session persists after page refresh
- [x] Test: Session persists after browser close
- [x] Test: Session persists after browser restart
- [x] Test: Multiple browser sessions (different devices)
- [x] Test: Session timeout handling
- [x] Test: Inactivity timeout

### Phase 5: Session Security Tests

- [x] Test: Invalid token rejected
- [x] Test: Tampered token rejected
- [x] Test: Token used after logout rejected
- [x] Test: Multiple tab session consistency
- [x] Test: Session hijacking prevention
- [x] Test: XSS token protection
- [x] Test: CSRF token handling

### Phase 6: Edge Cases Tests

- [x] Test: Network error during token refresh
- [x] Test: Server error during refresh
- [x] Test: Rapid token refresh attempts
- [x] Test: Token refresh with missing permissions
- [x] Test: Session with revoked tokens
- [x] Test: Concurrent session limit
- [x] Test: Token refresh with delayed response
- [x] Test: Expired access token with valid refresh token

---

## Technical Notes

### Session Manager Utilities

```typescript
// frontend/e2e/utils/session-manager.ts
import { Page } from '@playwright/test';

// JWT token structure (for testing purposes)
interface JWTPayload {
  sub: string;
  email: string;
  name: string;
  iat: number;
  exp: number;
  role: string;
}

// Generate a mock JWT token with custom expiration
export function generateMockToken(payload: Partial<JWTPayload>, expiresInSeconds: number = 3600): string {
  const header = { alg: 'HS256', typ: 'JWT' };
  const now = Math.floor(Date.now() / 1000);
  
  const tokenPayload: JWTPayload = {
    sub: payload.sub || 'test-user-id',
    email: payload.email || 'test@example.com',
    name: payload.name || 'Test User',
    iat: payload.iat || now,
    exp: payload.exp || now + expiresInSeconds,
    role: payload.role || 'user',
    ...payload,
  };

  // In real implementation, this would be a proper JWT
  // For testing, we use a mock token
  const mockPayload = Buffer.from(JSON.stringify(tokenPayload)).toString('base64');
  return `mock.${mockPayload}.signature`;
}

// Generate expired token
export function generateExpiredToken(): string {
  return generateMockToken({}, -3600); // Expired 1 hour ago
}

// Generate valid token with custom expiration
export function generateValidToken(expiresInSeconds: number = 3600): string {
  return generateMockToken({}, expiresInSeconds);
}

// Set tokens in storage
export async function setTokens(page: Page, accessToken: string, refreshToken: string) {
  await page.evaluate(({ access, refresh }) => {
    localStorage.setItem('access_token', access);
    localStorage.setItem('refresh_token', refresh);
  }, { access: accessToken, refresh: refreshToken });
}

// Clear all auth tokens
export async function clearTokens(page: Page) {
  await page.evaluate(() => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    sessionStorage.clear();
  });
}

// Get token expiration time
export async function getTokenExpiry(page: Page): Promise<number | null> {
  return page.evaluate(() => {
    const token = localStorage.getItem('access_token');
    if (!token) return null;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp || null;
    } catch {
      return null;
    }
  });
}

// Check if token is expired
export async function isTokenExpired(page: Page): Promise<boolean> {
  const expiry = await getTokenExpiry(page);
  if (!expiry) return true;
  
  const now = Math.floor(Date.now() / 1000);
  return expiry < now;
}

// Set remember me flag
export async function setRememberMe(page: Page, enabled: boolean = true) {
  await page.evaluate((enabled) => {
    localStorage.setItem('remember_me', enabled ? 'true' : 'false');
  }, enabled);
}
```

### Token Refresh Mock Handler

```typescript
// frontend/e2e/mocks/tokens.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockSuccessfulTokenRefresh(request: APIRequestContext) {
  await request.post('/api/v1/auth/refresh', {
    status: 200,
    body: {
      access_token: 'new-access-token-' + Date.now(),
      refresh_token: 'new-refresh-token-' + Date.now(),
      expires_in: 3600,
    },
  });
}

export async function mockExpiredRefreshToken(request: APIRequestContext) {
  await request.post('/api/v1/auth/refresh', {
    status: 401,
    body: {
      error: 'REFRESH_TOKEN_EXPIRED',
      message: 'Refresh token has expired. Please login again.',
      code: 'TOKEN_EXPIRED',
    },
  });
}

export async function mockInvalidRefreshToken(request: APIRequestContext) {
  await request.post('/api/v1/auth/refresh', {
    status: 401,
    body: {
      error: 'INVALID_REFRESH_TOKEN',
      message: 'Invalid refresh token',
    },
  });
}

export async function mockRevokedRefreshToken(request: APIRequestContext) {
  await request.post('/api/v1/auth/refresh', {
    status: 401,
    body: {
      error: 'TOKEN_REVOKED',
      message: 'Refresh token has been revoked',
    },
  });
}

export async function mockSessionStatus(request: APIRequestContext, active: boolean = true) {
  await request.get('/api/v1/auth/session', {
    status: active ? 200 : 401,
    body: active
      ? {
          active: true,
          user: { id: 'test-user', email: 'test@example.com' },
          expires_at: new Date(Date.now() + 3600000).toISOString(),
        }
      : {
          active: false,
          reason: 'Session expired',
        },
  });
}
```

### Remember Me Page Object Extension

```typescript
// frontend/e2e/pages/login.page.ts (extension)
export class LoginPage {
  // ... existing code ...

  readonly rememberMeCheckbox: Locator;

  constructor(page: Page) {
    // ... existing initializers ...
    this.rememberMeCheckbox = page.locator('[data-testid="checkbox-remember"]');
  }

  async checkRememberMe() {
    if (!(await this.rememberMeCheckbox.isChecked())) {
      await this.rememberMeCheckbox.check();
    }
  }

  async uncheckRememberMe() {
    if (await this.rememberMeCheckbox.isChecked()) {
      await this.rememberMeCheckbox.uncheck();
    }
  }

  async isRememberMeChecked(): Promise<boolean> {
    return this.rememberMeCheckbox.isChecked();
  }

  async expectRememberMeChecked() {
    await expect(this.rememberMeCheckbox).toBeChecked();
  }

  async expectRememberMeUnchecked() {
    await expect(this.rememberMeCheckbox).not.toBeChecked();
  }
}
```

### Session Test Setup

```typescript
// frontend/e2e/fixtures/session.fixture.ts
import { test as base } from '@playwright/test';

export const test = base.extend({
  // Authenticated page fixture
  authenticatedPage: async ({ page }, use) => {
    // Set up authenticated state
    await page.evaluate(() => {
      const accessToken = generateMockToken({}, 3600); // 1 hour
      const refreshToken = 'valid-refresh-token';
      
      localStorage.setItem('access_token', accessToken);
      localStorage.setItem('refresh_token', refreshToken);
      localStorage.setItem('user', JSON.stringify({
        id: 'test-user',
        email: 'test@example.com',
        name: 'Test User',
      }));
    });
    
    await use(page);
  },

  // Page with expired token
  expiredTokenPage: async ({ page }, use) => {
    await page.evaluate(() => {
      const expiredToken = generateMockToken({}, -3600); // Expired 1 hour ago
      localStorage.setItem('access_token', expiredToken);
      localStorage.setItem('refresh_token', 'valid-refresh-token');
    });
    
    await use(page);
  },

  // Remember me enabled page
  rememberMePage: async ({ page }, use) => {
    await page.evaluate(() => {
      const accessToken = generateMockToken({}, 3600);
      const refreshToken = 'valid-refresh-token';
      
      localStorage.setItem('access_token', accessToken);
      localStorage.setItem('refresh_token', refreshToken);
      localStorage.setItem('remember_me', 'true');
    });
    
    await use(page);
  },
});
```

### Running Specific Tests

```bash
# Run all session tests
cd frontend && npx playwright test sessions.spec.ts

# Run token refresh tests
cd frontend && npx playwright test sessions.spec.ts -g "refresh"

# Run remember me tests
cd frontend && npx playwright test sessions.spec.ts -g "remember"

# Run session persistence tests
cd frontend && npx playwright test sessions.spec.ts -g "persistence"

# Run security tests
cd frontend && npx playwright test sessions.spec.ts -g "security"

# Run in headed mode
cd frontend && npx playwright test sessions.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [x] Token refresh scenarios tested
- [x] Remember Me functionality tested
- [x] Session persistence tested
- [x] Security scenarios tested
- [x] Edge cases covered
- [x] Error handling tested

### 2. Test Execution Verification
- [x] All tests pass locally
- [x] Tests work with mocked APIs
- [x] No race conditions in token refresh
- [x] Clean test isolation

### 3. Security Verification
- [x] Invalid tokens rejected
- [x] Expired tokens handled
- [x] Token storage secure
- [x] No token leakage

### 4. UX Verification
- [x] Seamless token refresh
- [x] No visible login prompts during refresh
- [x] Clear error messages for expired sessions
- [x] Remember Me persists correctly

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Token Mocking** - Use realistic token structures for testing
2. **Timing** - Token expiration tests need careful timing
3. **Storage** - Test both localStorage and sessionStorage
4. **Interceptors** - Use request interception for API mocking
5. **Page Reload** - Test session persistence across reloads

### Common Issues to Avoid

1. **Hardcoded tokens** - Use dynamic token generation
2. **Skipping refresh tests** - This is critical functionality
3. **Not testing error cases** - Expired refresh token is common
4. **Race conditions** - Multiple concurrent refresh requests

### Testing Strategy

1. **Unit session tests** - Test individual session functions
2. **Integration tests** - Test full auth flow with API mocks
3. **E2E tests** - Test complete user journeys

### Browser Context Setup

```typescript
// Create browser context with specific settings
const context = await browser.newContext({
  storageState: undefined, // Don't persist storage
  permissions: [],
});
```

---

## Dependencies

- **Prerequisites**: 02-auth-login.md, 04-auth-logout.md
- **Environment**: Backend API with token endpoints
- **Test Data**: Valid and expired test tokens

---

## Related Tasks

- 02-auth-login.md - Login page tests (Remember Me integration)
- 04-auth-logout.md - Logout tests (session cleanup)
- 06-dashboard-layout.md - Dashboard tests (authenticated state)
- All authenticated tests depend on proper session management

---

*Task created from: docs/frontend/TEST_PATHS.md authentication patterns*
