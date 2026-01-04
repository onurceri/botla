# Task: Implement Logout Flow Tests

> **Task ID**: 04-auth-logout  
> **Source**: TEST_PATHS.md Section 2.3  
> **Priority**: Highest (Authentication)  
> **Estimated Effort**: 4-6 hours  
> **Prerequisite**: 02-auth-login.md (recommended) or completed login tests

---

## Detailed Prompt

Implement comprehensive E2E tests for the Logout Flow. This task covers user logout functionality, session cleanup, and multi-tab synchronization.

### Context

The Logout Flow is critical for security and user experience. Testing this functionality ensures:
- Users can successfully log out from any page
- Session data is properly cleaned up
- Multiple tabs are synchronized
- User is redirected to login page
- No sensitive data remains in the browser

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 2.3:

```
Logout Flow
├── While logged in (any page)
│   ├── Open user menu
│   │   ├── Click: avatar or dropdown toggle
│   │   └── Assert: `dropdown-user-menu` visible
│   │
│   ├── Click: menu item "Logout"
│   │   ├── Assert: Loading state
│   │   ├── Assert: Tokens removed from storage
│   │   ├── Assert: Session cleared
│   │   └── Redirect: /login
│   │
│   └── On login page
│       └── Assert: Previous session not restored
│
├── Session expired (auto-logout)
│   ├── Wait: Access token expiry (1 hour)
│   ├── Attempt: Any API call
│   ├── Assert: 401 Unauthorized
│   ├── Assert: `modal-session-expired` visible
│   ├── Click: btn-relogin
│   └── Redirect: /login
│
└── Multiple tabs (sync logout)
    ├── User logs out in Tab A
    ├── Event: BroadcastChannel message
    ├── Tab B receives: session_terminated
    └── Tab B redirects: /login
```

### Implementation Requirements

1. **Create Logout Test File** (`frontend/e2e/logout.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create User Menu Page Object** (`frontend/e2e/pages/user-menu.page.ts`)
   - Encapsulate user menu interactions
   - Provide logout method
   - Handle menu toggle states

3. **Create Session Management Utilities** (`frontend/e2e/utils/session.utils.ts`)
   - Token verification helpers
   - Storage cleanup verification
   - BroadcastChannel mocking

4. **Create Session Mock Handlers** (`frontend/e2e/mocks/session.mocks.ts`)
   - Mock session expiration
   - Mock logout API responses
   - Handle unauthorized responses

### Expected Deliverables

1. `frontend/e2e/logout.spec.ts` - Comprehensive logout tests
2. `frontend/e2e/pages/user-menu.page.ts` - User menu page object
3. `frontend/e2e/utils/session.utils.ts` - Session utilities
4. `frontend/e2e/mocks/session.mocks.ts` - Session mock handlers

---

## Implementation Plan

### Phase 1: Setup and Utilities

- [ ] Create `frontend/e2e/utils/session.utils.ts`:
  - Token verification helpers
  - Storage cleanup helpers
  - Session state checkers
- [ ] Create `frontend/e2e/mocks/session.mocks.ts`:
  - Mock logout API
  - Mock 401 responses
  - Mock session validation
- [ ] Create `frontend/e2e/pages/user-menu.page.ts`:
  - User avatar locator
  - Dropdown toggle
  - Logout menu item
  - Menu close functionality

### Phase 2: Manual Logout Tests

- [ ] Test: Open user menu from dashboard
- [ ] Test: Open user menu from any authenticated page
- [ ] Test: Menu dropdown visibility
- [ ] Test: Click logout menu item
- [ ] Test: Loading state during logout
- [ ] Test: Access token removed from storage
- [ ] Test: Refresh token removed from storage
- [ ] Test: Session data cleared
- [ ] Test: Redirect to login page
- [ ] Test: Previous session not restored on login page

### Phase 3: Session Expiration Tests

- [ ] Test: Access token expiry detection
- [ ] Test: 401 response triggers modal
- [ ] Test: Session expired modal visibility
- [ ] Test: Modal contains relogin button
- [ ] Test: Clicking relogin redirects to login
- [ ] Test: Multiple 401 responses handled
- [ ] Test: No infinite redirect loops

### Phase 4: Multi-tab Synchronization Tests

- [ ] Test: BroadcastChannel event emission
- [ ] Test: Other tabs receive session_terminated
- [ ] Test: Other tabs redirect to login
- [ ] Test: All tabs synchronized logout
- [ ] Test: No duplicate logout API calls

### Phase 5: Edge Cases Tests

- [ ] Test: Logout during API call
- [ ] Test: Logout with pending uploads
- [ ] Test: Network error during logout
- [ ] Test: Quick logout then relogin
- [ ] Test: Logout from deep link page

---

## Technical Notes

### User Menu Page Object

```typescript
// frontend/e2e/pages/user-menu.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class UserMenu {
  readonly page: Page;
  readonly avatarButton: Locator;
  readonly dropdownMenu: Locator;
  readonly menuItemProfile: Locator;
  readonly menuItemSettings: Locator;
  readonly menuItemHelp: Locator;
  readonly menuItemLogout: Locator;

  constructor(page: Page) {
    this.page = page;
    this.avatarButton = page.locator('[data-testid="user-avatar"]');
    this.dropdownMenu = page.locator('[data-testid="dropdown-user-menu"]');
    this.menuItemProfile = page.locator('[data-testid="menu-item-profile"]');
    this.menuItemSettings = page.locator('[data-testid="menu-item-settings"]');
    this.menuItemHelp = page.locator('[data-testid="menu-item-help"]');
    this.menuItemLogout = page.locator('[data-testid="menu-item-logout"]');
  }

  async open() {
    await this.avatarButton.click();
    await expect(this.dropdownMenu).toBeVisible();
  }

  async clickLogout() {
    await this.menuItemLogout.click();
  }

  async clickProfile() {
    await this.menuItemProfile.click();
  }

  async clickSettings() {
    await this.menuItemSettings.click();
  }

  async clickHelp() {
    await this.menuItemHelp.click();
  }

  async expectMenuVisible() {
    await expect(this.dropdownMenu).toBeVisible();
  }

  async expectMenuHidden() {
    await expect(this.dropdownMenu).toBeHidden();
  }
}
```

### Session Utilities

```typescript
// frontend/e2e/utils/session.utils.ts
import { Page } from '@playwright/test';

export async function getAccessToken(page: Page): Promise<string | null> {
  return page.evaluate(() => localStorage.getItem('access_token'));
}

export async function getRefreshToken(page: Page): Promise<string | null> {
  return page.evaluate(() => localStorage.getItem('refresh_token'));
}

export async function clearAllStorage(page: Page) {
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
}

export async function expectTokensCleared(page: Page) {
  const accessToken = await getAccessToken(page);
  const refreshToken = await getRefreshToken(page);
  
  if (accessToken !== null) {
    throw new Error('Access token was not cleared');
  }
  if (refreshToken !== null) {
    throw new Error('Refresh token was not cleared');
  }
}

export async function setExpiredToken(page: Page) {
  // Set an expired JWT token for testing
  const expiredToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMiwicm9sZSI6InVzZXIiLCJleHAiOjE1MTYyMzkwMjJ9.invalid';
  await page.evaluate((token) => {
    localStorage.setItem('access_token', token);
  }, expiredToken);
}
```

### Multi-tab Testing

Use Playwright's context isolation for multi-tab tests:

```typescript
// frontend/e2e/utils/tab.utils.ts
import { BrowserContext, Page } from '@playwright/test';

export async function createSecondTab(context: BrowserContext): Promise<Page> {
  const [newPage] = await Promise.all([
    context.waitForEvent('page'),
    context.pages()[0].click('a[target="_blank"]').catch(() => {}), // Optional trigger
  ]);
  return newPage;
}

export async function waitForBroadcastMessage(page: Page, expectedMessage: string): Promise<void> {
  // Set up BroadcastChannel listener
  await page.evaluate((message) => {
    const bc = new BroadcastChannel('auth_channel');
    bc.onmessage = (event) => {
      if (event.data === message) {
        window.location.href = '/login';
      }
    };
  }, expectedMessage);
}
```

### Mocking Session Expiration

```typescript
// frontend/e2e/mocks/session.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockLogoutSuccess(request: APIRequestContext) {
  await request.post('/api/v1/auth/logout', {
    status: 200,
    body: {
      success: true,
      message: 'Logged out successfully',
    },
  });
}

export async function mockUnauthorized(request: APIRequestContext) {
  await request.anyEndpoint().match(() => true).abort(); // Abort all requests
  
  // Or intercept specific endpoints
  await request.get('/api/v1/user/me', {
    status: 401,
    body: {
      error: 'UNAUTHORIZED',
      message: 'Access token expired',
      code: 'TOKEN_EXPIRED',
    },
  });
}

export async function mockSessionExpiredModal(page: Page) {
  // Mock the modal to appear
  await page.route('**/api/v1/auth/refresh', async (route) => {
    await route.fulfill({
      status: 401,
      body: {
        error: 'TOKEN_EXPIRED',
        message: 'Session has expired',
        requiresRelogin: true,
      },
    });
  });
}
```

### Running Specific Tests

```bash
# Run all logout tests
cd frontend && npx playwright test logout.spec.ts

# Run manual logout tests
cd frontend && npx playwright test logout.spec.ts -g "manual logout"

# Run session expiration tests
cd frontend && npx playwright test logout.spec.ts -g "session expired"

# Run multi-tab tests
cd frontend && npx playwright test logout.spec.ts -g "multi-tab"

# Run in headed mode
cd frontend && npx playwright test logout.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] Manual logout tested from various pages
- [ ] Session cleanup verified
- [ ] Token removal verified
- [ ] Redirect to login verified
- [ ] Session expiration handled
- [ ] Multi-tab sync tested
- [ ] Error scenarios tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work in parallel
- [ ] No race conditions
- [ ] Clean test isolation

### 3. Security Verification
- [ ] Tokens are properly cleared
- [ ] No sensitive data in storage
- [ ] 401 responses handled securely
- [ ] No infinite redirect loops

### 4. UX Verification
- [ ] Loading states visible
- [ ] Modal appears for expired sessions
- [ ] Clear error messages
- [ ] Smooth logout experience

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Storage State** - Use Playwright's storageState for authenticated state
2. **Multi-tab Testing** - Requires proper context management
3. **BroadcastChannel** - Native browser API for tab sync
4. **Timing** - Session expiration tests need careful timing
5. **Mocking** - Use API mocking to simulate token expiry

### Common Issues to Avoid

1. **Not waiting for logout API** - Wait for response before checking redirect
2. **Skipping storage checks** - Verify tokens are actually removed
3. **Multi-tab race conditions** - Use proper synchronization
4. **Hardcoded token values** - Use dynamic test data

### Test Data Setup

```typescript
// Use authenticated storage state
test.use({
  storageState: 'e2e/.auth/authenticated.json',
});
```

### Authentication Helper

```typescript
// Helper to set up authenticated state
async function setupAuthenticatedState(page: Page) {
  await page.evaluate(() => {
    localStorage.setItem('access_token', 'valid-test-token');
    localStorage.setItem('refresh_token', 'valid-refresh-token');
    localStorage.setItem('user', JSON.stringify({
      id: 'test-user',
      email: 'test@example.com',
      name: 'Test User',
    }));
  });
}
```

---

## Dependencies

- **Prerequisite**: 02-auth-login.md (recommended for understanding auth flow)
- **Environment**: Backend API must be running
- **Test Data**: Authenticated user session

---

## Related Tasks

- 02-auth-login.md - Login page tests
- 03-auth-register.md - Registration tests
- 05-auth-sessions.md - Session management (comprehensive)
- 06-dashboard-layout.md - Dashboard tests (uses user menu)

---

*Task created from: docs/frontend/TEST_PATHS.md Section 2.3*
