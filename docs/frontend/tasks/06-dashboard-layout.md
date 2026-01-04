# Task: Implement Dashboard Layout Tests

> **Task ID**: 06-dashboard-layout  
> **Source**: TEST_PATHS.md Section 3.1  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 8-10 hours  
> **Prerequisite**: 05-auth-sessions.md (recommended)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Dashboard Layout. This task covers sidebar navigation, top bar, breadcrumb navigation, and overall layout structure.

### Context

The Dashboard Layout is the main shell of the authenticated application. Testing this layout ensures:
- Consistent navigation across all pages
- Proper layout structure and responsiveness
- Active state management for navigation items
- User menu and organization switcher functionality
- Breadcrumb navigation for deep links

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 3.1:

#### 3.1.1 Layout Structure

```
Dashboard Layout
├── Sidebar (Left)
│   ├── Logo/Brand
│   ├── Navigation Menu
│   │   ├── Dashboard (Home)
│   │   ├── Chatbots
│   │   ├── Settings
│   │   └── Admin (if admin)
│   ├── Organization Switcher
│   └── User Menu
│       ├── Profile
│       ├── Settings
│       ├── Help
│       └── Logout
│
├── Top Bar
│   ├── Breadcrumb Navigation
│   ├── Search Bar
│   └── Action Buttons
│
└── Main Content Area
    └── Dynamic Content
```

#### 3.1.2 Sidebar Navigation Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `sidebar-logo` | image | Brand logo |
| `nav-item-dashboard` | nav-item | Dashboard link |
| `nav-item-chatbots` | nav-item | Chatbots list |
| `nav-item-settings` | nav-item | Settings |
| `nav-item-admin` | nav-item | Admin panel (admin only) |
| `org-switcher` | dropdown | Organization selector |
| `btn-sidebar-toggle` | button | Collapse/expand sidebar |

#### 3.1.3 Navigation Paths

```
Dashboard Navigation Flow
├── Load dashboard (authenticated)
│   ├── Assert: Sidebar visible
│   ├── Assert: Active nav item = Dashboard
│   └── Assert: Main content = Dashboard stats
│
├── Navigate to Chatbots
│   ├── Click: nav-item-chatbots
│   ├── Assert: URL changes to /dashboard/chatbots
│   ├── Assert: Active nav item = Chatbots
│   └── Assert: Content = Chatbots list page
│
├── Navigate to Settings
│   ├── Click: nav-item-settings
│   ├── Assert: URL changes to /settings
│   ├── Assert: Active nav item = Settings
│   └── Assert: Content = Settings page
│
├── Toggle Sidebar Collapse
│   ├── Click: btn-sidebar-toggle
│   ├── Assert: Sidebar collapses
│   ├── Assert: Icons only (no text)
│   ├── Click: btn-sidebar-toggle (expand)
│   └── Assert: Sidebar expands (full width)
│
├── Switch Organization
│   ├── Click: org-switcher
│   ├── Assert: `dropdown-org-list` visible
│   ├── Hover: Organization item (highlight)
│   ├── Click: Organization item
│   ├── Assert: Context switched
│   ├── Assert: Data refreshed for new org
│   └── Assert: org-switcher updated
│
└── Open User Menu
    ├── Click: user-avatar
    ├── Assert: `dropdown-user-menu` visible
    ├── Hover: Menu items (highlight)
    ├── Click: Profile
    │   └── Navigate: /settings/profile
    ├── Click: Settings
    │   └── Navigate: /settings
    ├── Click: Help
    │   └── Navigate: /help
    └── Click: Logout
        └── Execute: Logout flow
```

#### 3.1.4 Breadcrumb Navigation

```
Breadcrumb Flow (on /dashboard/chatbots/chatbot-id/settings)
├── Assert: Breadcrumb visible
│   ├── Home > Chatbots > [Chatbot Name] > Settings
│   │
│   ├── Click: Home (/)
│   │   └── Navigate: /dashboard
│   │
│   ├── Click: Chatbots (/)
│   │   └── Navigate: /dashboard/chatbots
│   │
│   ├── Click: [Chatbot Name] (/)
│   │   └── Navigate: /dashboard/chatbots/chatbot-id
│   │
│   └── Current: Settings (active, no click)
│
└── Hover: Breadcrumb item
    └── Assert: Tooltip if truncated
```

### Implementation Requirements

1. **Create Dashboard Layout Test File** (`frontend/e2e/dashboard.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Sidebar Page Object** (`frontend/e2e/pages/sidebar.page.ts`)
   - Encapsulate sidebar interactions
   - Navigation methods
   - Collapse/expand functionality

3. **Create User Menu Page Object** (`frontend/e2e/pages/user-menu.page.ts`)
   - Already partially created in task 04, expand it

4. **Create Organization Switcher Page Object** (`frontend/e2e/pages/org-switcher.page.ts`)
   - Organization selection methods
   - Multi-org handling

5. **Create Breadcrumb Page Object** (`frontend/e2e/pages/breadcrumb.page.ts`)
   - Breadcrumb navigation methods
   - Path verification

### Expected Deliverables

1. `frontend/e2e/dashboard.spec.ts` - Comprehensive dashboard layout tests
2. `frontend/e2e/pages/sidebar.page.ts` - Sidebar page object
3. `frontend/e2e/pages/org-switcher.page.ts` - Organization switcher page object
4. `frontend/e2e/pages/breadcrumb.page.ts` - Breadcrumb page object
5. Updated `frontend/e2e/pages/user-menu.page.ts` with dashboard-specific methods

---

## Implementation Plan

### Phase 1: Setup and Page Objects

- [ ] Create `frontend/e2e/pages/sidebar.page.ts`:
  - Sidebar container locator
  - Navigation item locators
  - Logo locator
  - Toggle button locator
  - Collapse/expand methods
- [ ] Create `frontend/e2e/pages/org-switcher.page.ts`:
  - Organization switcher locator
  - Dropdown list locator
  - Organization item locators
  - Selection methods
- [ ] Create `frontend/e2e/pages/breadcrumb.page.ts`:
  - Breadcrumb container locator
  - Breadcrumb item locators
  - Navigation methods
  - Tooltip handling

### Phase 2: Sidebar Navigation Tests

- [ ] Test: Sidebar visible on dashboard
- [ ] Test: All nav items present
- [ ] Test: Dashboard nav item active on home
- [ ] Test: Click Chatbots navigates to chatbots page
- [ ] Test: Click Settings navigates to settings
- [ ] Test: Admin nav visible only for admins
- [ ] Test: Logo links to dashboard
- [ ] Test: Navigation active states update correctly

### Phase 3: Sidebar Collapse Tests

- [ ] Test: Click toggle collapses sidebar
- [ ] Test: Collapsed sidebar shows icons only
- [ ] Test: Collapsed sidebar hides text labels
- [ ] Test: Click toggle expands sidebar
- [ ] Test: Expanded sidebar shows full width
- [ ] Test: Navigation works when collapsed
- [ ] Test: Collapse state persists on refresh
- [ ] Test: Responsive collapse on mobile

### Phase 4: Organization Switcher Tests

- [ ] Test: Organization switcher visible
- [ ] Test: Current org displayed
- [ ] Test: Click opens dropdown
- [ ] Test: Organization list visible
- [ ] Test: Hover highlights org item
- [ ] Test: Click org switches context
- [ ] Test: Data refreshes for new org
- [ ] Test: Switcher updates with new org name
- [ ] Test: Single org hides switcher
- [ ] Test: Multi-org shows switcher

### Phase 5: User Menu Tests

- [ ] Test: User avatar visible
- [ ] Test: Click opens user menu
- [ ] Test: Menu contains Profile
- [ ] Test: Menu contains Settings
- [ ] Test: Menu contains Help
- [ ] Test: Menu contains Logout
- [ ] Test: Hover highlights menu items
- [ ] Test: Click Profile navigates to profile
- [ ] Test: Click Settings navigates to settings
- [ ] Test: Click Help navigates to help
- [ ] Test: Menu closes when clicking outside

### Phase 6: Breadcrumb Navigation Tests

- [ ] Test: Breadcrumb visible on inner pages
- [ ] Test: Correct breadcrumb path
- [ ] Test: Home link navigates to dashboard
- [ ] Test: Intermediate links work
- [ ] Test: Current page not clickable
- [ ] Test: Hover shows tooltip for truncated
- [ ] Test: Dynamic content in breadcrumbs
- [ ] Test: Breadcrumb updates on navigation

### Phase 7: Top Bar Tests

- [ ] Test: Search bar visible
- [ ] Test: Action buttons present
- [ ] Test: User info displayed
- [ ] Test: Notification indicator (if present)

### Phase 8: Responsive Tests

- [ ] Test: Sidebar behavior on mobile
- [ ] Test: Hamburger menu on mobile
- [ ] Test: Collapsed sidebar on tablet
- [ ] Test: Full sidebar on desktop

---

## Technical Notes

### Sidebar Page Object

```typescript
// frontend/e2e/pages/sidebar.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class Sidebar {
  readonly page: Page;
  readonly container: Locator;
  readonly logo: Locator;
  readonly navDashboard: Locator;
  readonly navChatbots: Locator;
  readonly navSettings: Locator;
  readonly navAdmin: Locator;
  readonly toggleButton: Locator;
  readonly orgSwitcher: Locator;
  readonly userAvatar: Locator;

  constructor(page: Page) {
    this.page = page;
    this.container = page.locator('[data-testid="sidebar"]');
    this.logo = page.locator('[data-testid="sidebar-logo"]');
    this.navDashboard = page.locator('[data-testid="nav-item-dashboard"]');
    this.navChatbots = page.locator('[data-testid="nav-item-chatbots"]');
    this.navSettings = page.locator('[data-testid="nav-item-settings"]');
    this.navAdmin = page.locator('[data-testid="nav-item-admin"]');
    this.toggleButton = page.locator('[data-testid="btn-sidebar-toggle"]');
    this.orgSwitcher = page.locator('[data-testid="org-switcher"]');
    this.userAvatar = page.locator('[data-testid="user-avatar"]');
  }

  async expectVisible() {
    await expect(this.container).toBeVisible();
  }

  async expectCollapsed() {
    await expect(this.container).toHaveClass(/collapsed/);
    await expect(this.logo).toBeHidden();
    await expect(this.navDashboard).not.toHaveText(/Dashboard/); // Icon only
  }

  async expectExpanded() {
    await expect(this.container).not.toHaveClass(/collapsed/);
    await expect(this.logo).toBeVisible();
  }

  async clickToggle() {
    await this.toggleButton.click();
  }

  async navigateToDashboard() {
    await this.navDashboard.click();
    await expect(this.page).toHaveURL(/\/dashboard$/);
  }

  async navigateToChatbots() {
    await this.navChatbots.click();
    await expect(this.page).toHaveURL(/\/dashboard\/chatbots/);
  }

  async navigateToSettings() {
    await this.navSettings.click();
    await expect(this.page).toHaveURL(/\/settings/);
  }

  async clickUserAvatar() {
    await this.userAvatar.click();
  }

  async expectUserMenuVisible() {
    await expect(this.page.locator('[data-testid="dropdown-user-menu"]')).toBeVisible();
  }

  async isAdminNavVisible(): Promise<boolean> {
    return this.navAdmin.isVisible();
  }
}
```

### Organization Switcher Page Object

```typescript
// frontend/e2e/pages/org-switcher.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class OrgSwitcher {
  readonly page: Page;
  readonly trigger: Locator;
  readonly dropdown: Locator;
  readonly currentOrg: Locator;
  readonly orgList: Locator;
  readonly orgItems: Locator;

  constructor(page: Page) {
    this.page = page;
    this.trigger = page.locator('[data-testid="org-switcher"]');
    this.dropdown = page.locator('[data-testid="dropdown-org-list"]');
    this.currentOrg = page.locator('[data-testid="org-current-name"]');
    this.orgList = page.locator('[data-testid="org-list"]');
    this.orgItems = page.locator('[data-testid="org-item"]');
  }

  async expectVisible() {
    await expect(this.trigger).toBeVisible();
  }

  async expectCurrentOrg(name: string) {
    await expect(this.currentOrg).toHaveText(name);
  }

  async click() {
    await this.trigger.click();
  }

  async expectDropdownVisible() {
    await expect(this.dropdown).toBeVisible();
  }

  async expectDropdownHidden() {
    await expect(this.dropdown).toBeHidden();
  }

  async selectOrg(orgName: string) {
    await this.click();
    await this.expectDropdownVisible();
    await this.page.locator(`[data-testid="org-item"]:has-text("${orgName}")`).click();
  }

  async expectOrgCount(count: number) {
    await expect(this.orgItems).toHaveCount(count);
  }
}
```

### Breadcrumb Page Object

```typescript
// frontend/e2e/pages/breadcrumb.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class Breadcrumb {
  readonly page: Page;
  readonly container: Locator;
  readonly homeLink: Locator;
  readonly items: Locator;

  constructor(page: Page) {
    this.page = page;
    this.container = page.locator('[data-testid="breadcrumb"]');
    this.homeLink = page.locator('[data-testid="breadcrumb-home"]');
    this.items = page.locator('[data-testid^="breadcrumb-item-"]');
  }

  async expectVisible() {
    await expect(this.container).toBeVisible();
  }

  async expectHidden() {
    await expect(this.container).toBeHidden();
  }

  async expectItemCount(count: number) {
    await expect(this.items).toHaveCount(count);
  }

  async clickHome() {
    await this.homeLink.click();
    await expect(this.page).toHaveURL(/\/dashboard/);
  }

  async clickItem(index: number) {
    const item = this.items.nth(index);
    await item.locator('a').click();
  }

  async expectPath(path: string[]) {
    const items = this.items.all();
    for (let i = 0; i < path.length; i++) {
      await expect(items[i]).toHaveText(path[i]);
    }
  }

  async hoverItem(index: number) {
    const item = this.items.nth(index);
    await item.hover();
  }

  async expectTooltipOnHover(index: number) {
    const item = this.items.nth(index);
    const tooltip = this.page.locator('[data-testid="tooltip"]');
    await this.hoverItem(index);
    await expect(tooltip).toBeVisible();
  }
}
```

### Running Specific Tests

```bash
# Run all dashboard layout tests
cd frontend && npx playwright test dashboard.spec.ts

# Run sidebar tests
cd frontend && npx playwright test dashboard.spec.ts -g "sidebar"

# Run navigation tests
cd frontend && npx playwright test dashboard.spec.ts -g "navigation"

# Run organization switcher tests
cd frontend && npx playwright test dashboard.spec.ts -g "org"

# Run breadcrumb tests
cd frontend && npx playwright test dashboard.spec.ts -g "breadcrumb"

# Run in headed mode
cd frontend && npx playwright test dashboard.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All sidebar elements tested
- [ ] All navigation paths tested
- [ ] Collapse/expand tested
- [ ] Organization switcher tested
- [ ] User menu tested
- [ ] Breadcrumbs tested
- [ ] Responsive behavior tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] Tests work with authenticated state
- [ ] No flaky navigation tests
- [ ] Proper timeout handling

### 3. Layout Verification
- [ ] Sidebar layout consistent
- [ ] Active states clear
- [ ] Hover states visible
- [ ] Responsive behavior correct

### 4. UX Verification
- [ ] Navigation intuitive
- [ ] Clear visual feedback
- [ ] Loading states handled
- [ ] Error states graceful

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Authentication State** - Tests need authenticated page context
2. **Multiple Pages** - Navigation tests need multiple pages to test
3. **Admin Tests** - Admin nav tests need admin user context
4. **Responsive Tests** - Use viewport configuration for responsive tests
5. **Wait for Navigation** - Always wait for URL change after clicking nav

### Common Issues to Avoid

1. **Race conditions** - Wait for navigation to complete
2. **Hardcoded URLs** - Use URL patterns for flexibility
3. **Missing auth state** - Ensure page is authenticated
4. **Not testing mobile** - Always test responsive behavior

### Viewport Configuration

```typescript
// Test with different viewports
test.use({
  viewport: { width: 1280, height: 720 }, // Desktop
});

test('mobile sidebar', async ({ page }) => {
  test.use({
    viewport: { width: 375, height: 667 }, // Mobile
  });
  // Mobile specific tests
});
```

### Admin User Setup

```typescript
// Test that requires admin privileges
test('admin nav visible for admin user', async ({ page }) => {
  test.use({
    storageState: 'e2e/.auth/admin.json',
  });
  // Admin specific tests
});
```

---

## Dependencies

- **Prerequisites**: 05-auth-sessions.md (for authenticated state)
- **Environment**: Multiple pages available for navigation tests
- **Test Data**: Admin user for admin nav tests

---

## Related Tasks

- 05-auth-sessions.md - Session management (authenticated state)
- 07-dashboard-search.md - Search bar tests
- 08-dashboard-toast.md - Toast notification tests
- 09-chatbots-list.md - Chatbots page tests
- All tests that navigate from dashboard

---

*Task created from: docs/frontend/TEST_PATHS.md Section 3.1*
