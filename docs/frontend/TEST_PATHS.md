# Botla-Co Comprehensive Test Paths Documentation

> **Document Version**: 1.0  
> **Last Updated**: January 2026  
> **Scope**: All E2E and Integration Test Paths for Frontend, Widget, and Admin  
> **Test Framework**: Playwright  

---

## Table of Contents

1. [Test Naming Conventions](#1-test-naming-conventions)
2. [Authentication Flows](#2-authentication-flows)
3. [Dashboard & Navigation](#3-dashboard--navigation)
4. [Chatbot Management](#4-chatbot-management)
5. [Source Management](#5-source-management)
6. [Chat & Playground](#6-chat--playground)
7. [Smart Actions](#7-smart-actions)
8. [Settings & Configuration](#8-settings--configuration)
9. [Organization & Workspace](#9-organization--workspace)
10. [Admin Panel](#10-admin-panel)
11. [Widget Integration](#11-widget-integration)
12. [Edge Cases & Error States](#12-edge-cases--error-states)
13. [Accessibility Tests](#13-accessibility-tests)
14. [Performance Tests](#14-performance-tests)

---

## 1. Test Naming Conventions

### 1.1 File Naming Pattern

```
{page-or-feature}.spec.ts
```

### 1.2 Test Naming Pattern

```typescript
test.describe('Feature Area', () => {
  test('should perform action when user does X', async () => { ... });
  test('should show error when Y condition', async () => { ... });
  test('should handle hover state on element', async () => { ... });
});
```

### 1.3 Element Naming

| Element Type | Prefix | Example |
|--------------|--------|---------|
| Button | `btn` | `btn-create-chatbot` |
| Input | `input` | `input-email` |
| Select | `select` | `select-plan` |
| Link | `link` | `link-login` |
| Tab | `tab` | `tab-settings` |
| Modal | `modal` | `modal-confirm-delete` |
| Toast | `toast` | `toast-success` |
| Dropdown | `dropdown` | `dropdown-menu` |
| Checkbox | `checkbox` | `checkbox-terms` |
| Radio | `radio` | `radio-mode` |
| Toggle | `toggle` | `toggle-visibility` |

---

## 2. Authentication Flows

### 2.1 Login Page (`auth.spec.ts`)

#### 2.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `input-email` | text | Email input field |
| `input-password` | password | Password input field |
| `btn-login` | submit | Login button |
| `link-forgot-password` | link | Forgot password link |
| `link-register` | link | Register new account link |
| `checkbox-remember` | checkbox | Remember me checkbox |
| `text-error` | text | Error message display |

#### 2.1.2 Interactions

```
Login Flow
├── Load login page
│   ├── Hover: Email input label (shows tooltip)
│   ├── Hover: Password input label (shows tooltip)
│   ├── Click: Email input → focus state
│   ├── Click: Password input → focus state
│   ├── Tab: Navigate through fields
│   ├── Enter: Email input → focus password
│   ├── Enter: Password input → submit form
│   └── Type: Email field (validation on blur)
│
├── Submit with empty fields
│   ├── Click: btn-login
│   ├── Assert: `toast-error` - "Email is required"
│   └── Assert: `input-email` has error class
│
├── Submit with invalid email
│   ├── Type: "invalid-email"
│   ├── Blur: Email input
│   ├── Assert: `input-email` has error class
│   └── Assert: `text-error` - "Invalid email format"
│
├── Submit with valid credentials
│   ├── Type: Valid email
│   ├── Type: Valid password
│   ├── Click: btn-login
│   ├── Assert: Loading spinner visible
│   ├── Assert: btn-login disabled
│   ├── Wait: API response
│   └── Redirect: /dashboard
│
├── Remember me checkbox
│   ├── Check: checkbox-remember
│   ├── Login successfully
│   └── Assert: Refresh token stored in localStorage
│
└── Forgot password flow
    ├── Click: link-forgot-password
    ├── Assert: URL contains /forgot-password
    ├── Type: email
    ├── Click: btn-send-reset
    └── Assert: `toast-success` - "Reset link sent"
```

#### 2.1.3 Hover States

| Element | Expected Hover Behavior |
|---------|------------------------|
| `btn-login` | Darken background, scale 1.02 |
| `link-forgot-password` | Underline, color change |
| `link-register` | Underline, color change |
| Input labels | Slight color change |

#### 2.1.4 Keyboard Navigation

| Key | Action |
|-----|--------|
| `Tab` | Navigate forward through inputs |
| `Shift+Tab` | Navigate backward |
| `Enter` | Submit form (when focused on submit) |
| `Escape` | Close any open dropdowns/modals |

### 2.2 Register Page (`register.spec.ts`)

#### 2.2.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `input-fullname` | text | Full name input |
| `input-email` | text | Email input |
| `input-password` | password | Password input |
| `input-confirm-password` | password | Confirm password |
| `checkbox-terms` | checkbox | Accept terms |
| `checkbox-privacy` | checkbox | Accept privacy policy |
| `btn-register` | submit | Register button |
| `link-login` | link | Already have account |
| `text-password-requirements` | text | Password rules display |

#### 2.2.2 Password Requirements Display

```
Password Requirements (Real-time Validation)
├── Character count ≥ 8 (checked on input)
├── Uppercase letter (checked on input)
├── Lowercase letter (checked on input)
├── Digit (checked on input)
└── Special character @$!%*?& (checked on input)
```

#### 2.2.3 Registration Flow

```
Register Flow
├── Load register page
│   ├── All inputs empty
│   ├── Password requirements visible (gray)
│   └── btn-register disabled
│
├── Fill form - Full Name
│   ├── Type: "John Doe"
│   └── Assert: Value = "John Doe"
│
├── Fill form - Email
│   ├── Type: "john@example.com"
│   ├── Blur: Trigger format validation
│   └── Assert: No error if valid
│
├── Fill form - Password
│   ├── Type: "Weak123"
│   ├── Assert: Character count requirement (check)
│   ├── Assert: Uppercase requirement (check)
│   ├── Assert: Digit requirement (check)
│   ├── Assert: Special char requirement (x)
│   ├── Type: "Weak123@" (complete)
│   └── Assert: All requirements (check)
│
├── Fill form - Confirm Password
│   ├── Type: "Weak123@"
│   ├── Assert: Matches password
│   └── Assert: No error
│
├── Submit without accepting terms
│   ├── Click: btn-register
│   ├── Assert: `toast-error` - "Accept terms required"
│   └── Assert: checkbox-terms has error class
│
├── Submit with mismatched passwords
│   ├── Change confirm password to "Different123@"
│   ├── Click: btn-register
│   ├── Assert: `input-confirm-password` error
│   └── Assert: `text-error` - "Passwords do not match"
│
├── Submit with weak password
│   ├── Change password to "weak"
│   ├── Click: btn-register
│   ├── Assert: `input-password` error
│   └── Assert: `text-error` - "Password too weak"
│
├── Successful registration
│   ├── Check: checkbox-terms
│   ├── Check: checkbox-privacy
│   ├── Click: btn-register
│   ├── Assert: Loading state
│   ├── Wait: API response
│   ├── Assert: User created in database
│   ├── Assert: Default org created
│   ├── Assert: Default workspace created
│   ├── Assert: Tokens stored
│   └── Redirect: /dashboard
│
└── Email already exists
    ├── Type: existing email
    ├── Click: btn-register
    ├── Assert: `toast-error` - "Email already exists"
    └── Assert: `input-email` error
```

#### 2.2.4 Validation States

| Field | Valid State | Invalid State |
|-------|-------------|---------------|
| Full Name | Non-empty | Empty |
| Email | RFC 5322 format | Invalid format |
| Password | All 5 requirements met | Any missing |
| Confirm Password | Matches password | Mismatch |
| Terms | Checked | Unchecked |

### 2.3 Logout Flow (`logout.spec.ts`)

```
Logout Flow
├── While logged in (any page)
│   ├── Open user menu
│   │   ├── Click: avatar or dropdown toggle
│   │   └── Assert: `dropdown-menu` visible
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

---

## 3. Dashboard & Navigation

### 3.1 Dashboard Layout (`dashboard.spec.ts`)

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

#### 3.1.2 Sidebar Navigation

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

### 3.2 Search Functionality (`search.spec.ts`)

| Element ID | Type | Description |
|------------|------|-------------|
| `search-input` | text | Global search input |
| `search-results` | dropdown | Search results dropdown |
| `search-result-item` | list-item | Individual result |
| `btn-search-clear` | button | Clear search |
| `btn-search-submit` | button | Submit search |

```
Search Flow
├── Click: search-input
│   ├── Assert: Focus state
│   └── Type: "my chatbot"
│
├── While typing
│   ├── Debounce: 300ms
│   ├── Show: `search-results` dropdown
│   ├── Show: Loading spinner
│   ├── Hide: Results if < 2 chars
│   └── Show: No results if 0 matches
│
├── Search results displayed
│   ├── Show: Up to 5 results
│   ├── Each result shows:
│   │   ├── Icon (chatbot, source, etc.)
│   │   ├── Title
│   │   └── Description
│   │
│   ├── Hover: Result item (highlight)
│   │   ├── Background color change
│   │   └── Cursor pointer
│   │
│   ├── Click: Result item
│   │   ├── Navigate: Result URL
│   │   └── Close: Search dropdown
│   │
│   └── Click: View all results
│       └── Navigate: Search results page
│
├── Clear search
│   ├── Click: btn-search-clear
│   ├── Assert: Input cleared
│   ├── Assert: Dropdown closed
│   └── Assert: Placeholder visible
│
└── Keyboard navigation
    ├── Arrow Down: Navigate results
    ├── Arrow Up: Navigate results
    ├── Enter: Open selected result
    └── Escape: Close dropdown
```

### 3.3 Toast Notifications (`toast.spec.ts`)

| Element ID | Type | Description |
|------------|------|-------------|
| `toast-container` | container | Toast container |
| `toast-item` | toast | Individual toast |
| `btn-toast-close` | button | Close toast |
| `toast-progress` | progress | Auto-dismiss progress bar |

```
Toast Notification Flow
├── Success Toast
│   ├── Trigger: Successful operation
│   ├── Show: Green toast
│   ├── Icon: Checkmark
│   ├── Message: Operation completed
│   ├── Duration: 5 seconds
│   ├── Show: Progress bar (shrinking)
│   └── Auto-dismiss: After duration
│
├── Error Toast
│   ├── Trigger: Failed operation
│   ├── Show: Red toast
│   ├── Icon: X mark
│   ├── Message: Error description
│   ├── Duration: 8 seconds (longer)
│   └── Auto-dismiss: After duration
│
├── Warning Toast
│   ├── Trigger: Warning condition
│   ├── Show: Yellow toast
│   ├── Icon: Warning triangle
│   └── Message: Warning text
│
├── Dismiss toast manually
│   ├── Hover: Toast
│   ├── Click: btn-toast-close
│   └── Assert: Toast removed from DOM
│
├── Multiple toasts
│   ├── Stack: Vertical (newest on top)
│   ├── Max: 5 visible
│   ├── Older: Dismissed when max exceeded
│   └── Animation: Slide in/out
│
└── Toast interaction
    ├── Click: Toast body (if link)
    │   └── Navigate: Related page
    └── Hover: Pause auto-dismiss timer
```

---

## 4. Chatbot Management

### 4.1 Chatbots List Page (`chatbots-list.spec.ts`)

#### 4.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `btn-create-chatbot` | button | Create new chatbot |
| `input-search` | text | Search chatbots |
| `select-sort` | select | Sort order |
| `select-filter` | select | Filter by status |
| `card-chatbot` | card | Chatbot card |
| `card-chatbot-name` | text | Chatbot name |
| `card-chatbot-model` | badge | Model badge |
| `card-chatbot-status` | badge | Status indicator |
| `card-chatbot-actions` | menu | Actions dropdown |
| `pagination` | pagination | Pagination controls |
| `empty-state` | component | No chatbots state |

#### 4.1.2 Chatbot Card Interactions

```
Chatbot Card Flow
├── Card hover state
│   ├── Hover: Card
│   │   ├── Shadow increase
│   │   ├── Scale 1.02
│   │   └── Cursor pointer
│   │
│   └── Hover: Card actions button
│       └── Show: Tooltip "Actions"
│
├── Click: Card body (not actions)
│   ├── Navigate: /dashboard/chatbots/{id}
│   └── Open: Chatbot detail page
│
├── Click: Actions menu
│   ├── Open: `dropdown-chatbot-actions`
│   ├── Options:
│   │   ├── Edit
│   │   ├── Duplicate
│   │   ├── Share
│   │   ├── Settings
│   │   └── Delete
│   │
│   ├── Click: Edit
│   │   └── Navigate: /dashboard/chatbots/{id}/settings
│   │
│   ├── Click: Duplicate
│   │   ├── Open: `modal-duplicate`
│   │   ├── Show: New name input
│   │   ├── Click: btn-duplicate
│   │   └── Assert: New chatbot created
│   │
│   ├── Click: Share
│   │   ├── Open: `modal-share`
│   │   ├── Show: Share link
│   │   └── Click: btn-copy-link
│   │
│   ├── Click: Delete
│   │   ├── Open: `modal-delete-confirm`
│   │   ├── Show: "Delete chatbot?" warning
│   │   ├── Type: chatbot name to confirm
│   │   ├── Click: btn-delete
│   │   └── Assert: Chatbot deleted
│   │
│   └── Hover: Menu item
        └── Highlight background
```

#### 4.1.3 Search and Filter

```
Search Chatbots
├── Type: "support"
│   ├── Filter: Chatbots matching "support"
│   ├── Update: Card list
│   └── Show: Match count
│
├── Clear search
│   ├── Click: btn-clear
│   └── Reset: Full list
│
└── Sort options
    ├── Select: Name (A-Z)
    │   └── Sort: Alphabetical
    │
    ├── Select: Name (Z-A)
    │   └── Sort: Reverse alphabetical
    │
    ├── Select: Recently updated
    │   └── Sort: UpdatedAt DESC
    │
    └── Select: Oldest
        └── Sort: CreatedAt ASC

Filter by Status
├── Select: All
│   └── Show: All chatbots
│
├── Select: Active
│   └── Show: Only active chatbots
│
├── Select: Training
│   └── Show: Chatbots with sources training
│
└── Select: Error
    └── Show: Chatbots with errors
```

#### 4.1.4 Pagination

```
Pagination Flow
├── Assert: Pagination visible (if > items per page)
│
├── Items per page selector
│   ├── Select: 12
│   │   └── Update: itemsPerPage = 12
│   │
│   ├── Select: 24
│   │   └── Update: itemsPerPage = 24
│   │
│   └── Select: 48
│       └── Update: itemsPerPage = 48
│
├── Page navigation
│   ├── Click: btn-previous (when on page > 1)
│   │   └── Navigate: Previous page
│   │
│   ├── Click: btn-next (when more pages)
│   │   └── Navigate: Next page
│   │
│   ├── Click: Page number
│   │   └── Navigate: Specific page
│   │
│   └── Click: Ellipsis (...)
│       └── Show: Page range selector
│
└── Empty state
    ├── Assert: When no chatbots match filter
    ├── Show: Empty illustration
    ├── Show: "No chatbots found" text
    └── Show: btn-create-chatbot
```

### 4.2 Create Chatbot Flow (`chatbot-create.spec.ts`)

#### 4.2.1 Create Dialog Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `modal-create-chatbot` | modal | Create modal |
| `input-name` | text | Chatbot name |
| `input-description` | textarea | Description |
| `select-language` | select | Default language |
| `select-model` | select | AI model |
| `slider-temperature` | slider | Temperature (0-2) |
| `input-max-tokens` | number | Max tokens |
| `btn-create` | submit | Create button |
| `btn-cancel` | button | Cancel button |

#### 4.2.2 Create Flow

```
Create Chatbot Flow
├── Open create modal
│   ├── Click: btn-create-chatbot
│   ├── Assert: `modal-create-chatbot` visible
│   └── Assert: Focus on `input-name`
│
├── Fill form
│   ├── Type: Name (required)
│   │   ├── Min: 1 character
│   │   ├── Max: 100 characters
│   │   └── Validation: On blur
│   │
│   ├── Type: Description (optional)
│   │   ├── Max: 500 characters
│   │   └── Validation: On blur
│   │
│   ├── Select: Language (default: tr)
│   │   ├── tr (Türkçe)
│   │   └── en (English)
│   │
│   ├── Select: Model (default: gpt-4o-mini)
│   │   ├── gpt-4o-mini
│   │   ├── gpt-4o
│   │   └── gpt-5 (if ultra)
│   │
│   ├── Adjust: Temperature (default: 0.7)
│   │   ├── Slider: 0.0 to 2.0
│   │   ├── Show: Value label
│   │   └── Hover: Slider track (highlight)
│   │
│   └── Input: Max tokens (default: 1000)
│       ├── Min: 100
│       ├── Max: 8000
│       └── Validation: On change
│
├── Submit validation
│   ├── Click: btn-create (without name)
│   │   ├── Assert: `input-name` error
│   │   └── Assert: `toast-error` - "Name required"
│   │
│   ├── Click: btn-create (valid form)
│   │   ├── Assert: Loading state
│   │   ├── Assert: btn-create disabled
│   │   ├── API: Create chatbot
│   │   ├── Assert: Chatbot in database
│   │   ├── Assert: Source count = 0
│   │   ├── Assert: Created with defaults
│   │   ├── Close: Modal
│   │   ├── Assert: Toast success
│   │   └── Navigate: /dashboard/chatbots/{new-id}
│   │
│   └── Click: btn-cancel
│       ├── Close: Modal
│       └── Assert: No chatbot created
│
└── Keyboard shortcuts (in modal)
    ├── Escape: Close modal
    ├── Enter: Submit (if form valid)
    └── Tab: Navigate form fields
```

### 4.3 Chatbot Detail Page (`chatbot-detail.spec.ts`)

#### 4.3.1 Tab Navigation

| Element ID | Type | Description |
|------------|------|-------------|
| `tab-overview` | tab | Overview tab |
| `tab-settings` | tab | Settings tab |
| `tab-sources` | tab | Sources tab |
| `tab-actions` | tab | Actions tab |
| `tab-playground` | tab | Chat playground |
| `tab-deploy` | tab | Deployment tab |
| `tab-insights` | tab | Analytics tab |

#### 4.3.2 Tab Navigation Flow

```
Chatbot Detail Tabs
├── Load chatbot detail
│   ├── Assert: Active tab = Overview
│   ├── Assert: Sidebar highlights chatbot
│   └── Assert: Breadcrumb correct
│
├── Switch to Settings
│   ├── Click: tab-settings
│   ├── Assert: URL = /dashboard/chatbots/{id}/settings
│   ├── Assert: Tab active = Settings
│   └── Assert: Settings panel loads
│
├── Switch to Sources
│   ├── Click: tab-sources
│   ├── Assert: URL = /dashboard/chatbots/{id}/sources
│   ├── Assert: Tab active = Sources
│   └── Assert: Sources list loads
│
├── Switch to Actions
│   ├── Click: tab-actions
│   ├── Assert: URL = /dashboard/chatbots/{id}/actions
│   ├── Assert: Tab active = Actions
│   └── Assert: Actions list loads
│
├── Switch to Playground
│   ├── Click: tab-playground
│   ├── Assert: URL = /dashboard/chatbots/{id}/playground
│   ├── Assert: Tab active = Playground
│   └── Assert: Chat interface loads
│
├── Switch to Deploy
│   ├── Click: tab-deploy
│   ├── Assert: URL = /dashboard/chatbots/{id}/deploy
│   ├── Assert: Tab active = Deploy
│   └── Assert: Embed code panel loads
│
├── Switch to Insights
│   ├── Click: tab-insights
│   ├── Assert: URL = /dashboard/chatbots/{id}/insights
│   ├── Assert: Tab active = Insights
│   └── Assert: Analytics dashboard loads
│
└── Keyboard navigation
    ├── Arrow Left: Previous tab
    ├── Arrow Right: Next tab
    └── Enter: Activate focused tab
```

#### 4.3.3 Overview Tab

```
Overview Tab Flow
├── Overview content
│   ├── Show: Chatbot name
│   ├── Show: Model badge
│   ├── Show: Status indicator
│   ├── Show: Description
│   ├── Show: Created/Updated dates
│   └── Show: Quick stats (sources, messages)
│
├── Quick actions
│   ├── Click: btn-edit-settings
│   │   └── Navigate: Settings tab
│   │
│   ├── Click: btn-add-sources
│   │   └── Navigate: Sources tab
│   │
│   └── Click: btn-open-playground
│       └── Navigate: Playground tab
│
└── Status indicators
    ├── Green: Ready (sources > 0, no errors)
    ├── Yellow: Training (sources processing)
    └── Red: Error (check sources)
```

### 4.4 Chatbot Settings (`chatbot-settings.spec.ts`)

#### 4.4.1 Settings Sections

```
Settings Page Sections
├── 1. Identity Section
│   ├── input-name (edit)
│   ├── input-description (edit)
│   ├── input-bot-display-name (edit)
│   ├── input-bot-icon (upload)
│   └── btn-save-identity
│
├── 2. Instructions Section
│   ├── textarea-custom-instruction (wysiwyg)
│   └── btn-save-instructions
│
├── 3. Language & Model Section
│   ├── select-language
│   ├── select-model
│   ├── slider-temperature
│   ├── input-max-tokens
│   └── btn-save-params
│
├── 4. Appearance Section
│   ├── color-theme (color picker)
│   ├── input-welcome-message (textarea)
│   ├── select-position
│   ├── color-bot-message
│   ├── color-user-message
│   ├── input-font-family
│   └── btn-save-appearance
│
├── 5. Suggestions Section
│   ├── toggle-suggestions-enabled
│   ├── textarea-suggested-questions
│   └── btn-save-suggestions
│
├── 6. Branding Section
│   ├── toggle-hide-branding
│   ├── input-logo-url
│   ├── input-brand-text
│   ├── input-brand-link
│   └── btn-save-branding
│
├── 7. Guardrails Section
│   ├── slider-confidence-threshold
│   ├── textarea-no-info-message
│   ├── textarea-error-message
│   ├── input-allowed-topics
│   ├── input-blocked-topics
│   └── btn-save-guardrails
│
├── 8. Handoff Section
│   ├── toggle-handoff-enabled
│   ├── select-handoff-type
│   ├── textarea-handoff-message
│   └── btn-save-handoff
│
└── 9. Security Section
    ├── toggle-secure-embed
    ├── textarea-allowed-domains
    ├── btn-regenerate-secret
    └── btn-save-security
```

#### 4.4.2 Identity Section Tests

```
Identity Section Flow
├── Edit name
│   ├── Click: input-name
│   ├── Clear: Existing name
│   ├── Type: New name
│   ├── Click: btn-save-identity
│   ├── Assert: Loading state
│   ├── Assert: Toast success
│   └── Assert: Name updated in DB
│
├── Edit description
│   ├── Click: input-description
│   ├── Type: New description
│   ├── Click: btn-save-identity
│   └── Assert: Description updated
│
├── Upload bot icon
│   ├── Click: input-bot-icon (file input)
│   ├── Select: Image file
│   ├── Assert: Preview shows image
│   ├── Assert: File size validation
│   ├── Assert: File type validation
│   ├── Click: btn-save-identity
│   └── Assert: Icon URL saved
│
└── Validation
    ├── Empty name → Error
    ├── Name > 100 chars → Error
    └── Invalid URL → Error
```

#### 4.4.3 Appearance Section Tests

```
Appearance Section Flow
├── Change theme color
│   ├── Click: color-theme (color picker)
│   ├── Assert: Color picker dropdown opens
│   ├── Select: Color from palette
│   │   ├── Assert: Color preview updates
│   │   └── Click: Outside picker to close
│   ├── Type: Hex color directly
│   │   ├── Assert: Valid hex format
│   │   └── Assert: Color updates
│   └── Click: btn-save-appearance
│       └── Assert: Theme saved
│
├── Change position
│   ├── Click: select-position
│   ├── Options:
│   │   ├── bottom-right
│   │   └── bottom-left
│   ├── Select: bottom-left
│   └── Click: btn-save-appearance
│       └── Assert: Position saved
│
├── Change message colors
│   ├── Click: color-bot-message
│   ├── Select: Bot message color
│   ├── Click: color-user-message
│   ├── Select: User message color
│   └── Click: btn-save-appearance
│       └── Assert: Colors saved
│
├── Change font family
│   ├── Click: select-font-family
│   ├── Options:
│   │   ├── System default
│   │   ├── Inter
│   │   ├── Roboto
│   │   └── Custom (input)
│   ├── Select: Inter
│   └── Click: btn-save-appearance
│       └── Assert: Font saved
│
└── Welcome message
    ├── Click: input-welcome-message
    ├── Type: Custom welcome
    ├── Click: btn-save-appearance
    └── Assert: Welcome message saved
```

#### 4.4.4 Suggestions Section Tests

```
Suggestions Section Flow
├── Toggle suggestions
│   ├── Click: toggle-suggestions-enabled
│   ├── Assert: Toggle state changes
│   └── Assert: Suggestions input enabled/disabled
│
├── Add suggested questions
│   ├── Click: textarea-suggested-questions
│   ├── Type: Question 1
│   ├── Press: Enter
│   ├── Type: Question 2
│   ├── Press: Enter
│   ├── Type: Question 3
│   ├── Click: btn-save-suggestions
│   ├── Assert: Toast success
│   └── Assert: Questions saved
│
├── Edit suggested question
│   ├── Hover: Question item
│   ├── Click: Edit icon
│   ├── Modify: Question text
│   ├── Click: Save
│   └── Assert: Question updated
│
├── Delete suggested question
│   ├── Hover: Question item
│   ├── Click: Delete icon
│   └── Assert: Question removed
│
└── Reorder questions
    ├── Drag: Question item
    ├── Drop: New position
    └── Assert: Order saved
```

#### 4.4.5 Guardrails Section Tests

```
Guardrails Section Flow
├── Adjust confidence threshold
│   ├── Click: slider-confidence-threshold
│   ├── Drag: To 0.6
│   ├── Assert: Value label = 0.6
│   └── Click: btn-save-guardrails
│       └── Assert: Threshold saved
│
├── Configure fallback messages
│   ├── Click: textarea-no-info-message
│   ├── Type: "I couldn't find information..."
│   ├── Click: textarea-error-message
│   ├── Type: "Something went wrong..."
│   └── Click: btn-save-guardrails
│       └── Assert: Messages saved
│
├── Set topic restrictions
│   ├── Click: input-allowed-topics
│   ├── Type: "product, pricing, features"
│   ├── Click: input-blocked-topics
│   ├── Type: "politics, religion"
│   └── Click: btn-save-guardrails
│       └── Assert: Topics saved
│
└── Toggle threshold warnings
    ├── Click: toggle-show-warnings
    └── Assert: Toggle state saved
```

#### 4.4.6 Security Section Tests

```
Security Section Flow
├── Toggle secure embed
│   ├── Click: toggle-secure-embed
│   ├── Assert: Toggle enabled
│   ├── Click: btn-save-security
│   └── Assert: Secure embed enabled
│
├── Set allowed domains
│   ├── Click: textarea-allowed-domains
│   ├── Type: "example.com, www.example.com"
│   ├── Click: btn-save-security
│   └── Assert: Domains saved
│
├── Regenerate embed secret
│   ├── Click: btn-regenerate-secret
│   ├── Assert: `modal-confirm-regenerate` opens
│   ├── Click: btn-confirm
│   ├── Assert: New secret generated
│   ├── Assert: Toast success
│   └── Assert: Old secret invalidated
│
└── View embed secret
    ├── Click: btn-show-secret (eye icon)
    ├── Assert: Secret visible
    ├── Click: btn-copy-secret
    └── Assert: Toast "Copied to clipboard"
```

---

## 5. Source Management

### 5.1 Sources List Page (`sources-list.spec.ts`)

#### 5.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `btn-add-source` | button | Add source button |
| `tab-source-type` | tabs | URL / PDF / Text / Sitemap |
| `card-source` | card | Source card |
| `card-source-status` | badge | Status indicator |
| `card-source-type` | badge | Type badge |
| `card-source-chunk-count` | text | Chunk count |
| `card-source-actions` | menu | Actions dropdown |
| `progress-bar` | progress | Training progress |
| `input-url` | text | URL input |
| `input-file` | file | File upload |
| `textarea-text` | textarea | Text content |

#### 5.1.2 Source Card Interactions

```
Source Card Flow
├── Card hover state
│   ├── Hover: Card
│   │   ├── Shadow increase
│   │   └── Scale 1.01
│   │
│   └── Hover: Actions button
│       └── Show: Tooltip
│
├── Status states
    ├── pending (yellow) → Show spinner
    ├── processing (blue) → Show progress bar
    ├── completed (green) → Show chunk count
    └── failed (red) → Show error message
│
├── Click: Source card
│   ├── Open: Source detail panel
│   ├── Show: Source info
│   ├── Show: Sample chunks
│   └── Show: Actions
│
└── Source actions menu
    ├── Click: btn-refresh
    │   ├── Open: `modal-refresh-confirm`
    │   ├── Click: btn-confirm
    │   └── Assert: Source re-processing
    │
    ├── Click: btn-view-chunks
    │   ├── Open: `modal-chunk-viewer`
    │   ├── Show: All chunks
    │   ├── Search: Chunk content
    │   └── Export: Chunk list
    │
    ├── Click: btn-download
    │   └── Download: Source content
    │
    └── Click: btn-delete
        ├── Open: `modal-delete-source`
        ├── Type: DELETE to confirm
        ├── Click: btn-delete
        └── Assert: Source deleted
```

### 5.2 Add URL Source (`sources-url.spec.ts`)

```
Add URL Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   ├── Click: tab-url
│   └── Assert: URL input visible
│
├── Single URL input
│   ├── Click: input-url
│   ├── Type: "https://example.com/page"
│   ├── Assert: URL validation
│   ├── Click: btn-add
│   ├── Assert: Loading state
│   ├── Assert: Source created (pending)
│   └── Assert: Toast "Source added"
│
├── URL with discovery
│   ├── Click: input-url
│   ├── Type: "https://example.com"
│   ├── Toggle: checkbox-discover-pages
│   ├── Click: btn-add
│   ├── Assert: Source created
│   ├── Assert: Discovery started
│   └── Assert: Pending URLs will be discovered
│
├── Path filters
│   ├── Click: input-include-paths
│   ├── Type: "/docs/, /guide/"
│   ├── Click: input-exclude-paths
│   ├── Type: "/admin/, /private/"
│   └── Click: btn-add
│       └── Assert: Filters saved
│
├── Validation
    ├── Empty URL → Error
    ├── Invalid URL → Error
    ├── Blocked domain → Error
    └── Private IP → Error (SSRF protection)
│
└── Cancel
    ├── Click: btn-cancel
    └── Assert: Modal closed, no source created
```

### 5.3 Add PDF Source (`sources-pdf.spec.ts`)

```
Add PDF Source Flow
├── Open add source modal
│   ├── Click: btn   ├──-add-source
│ Click: tab-pdf
│   └── Assert: File upload area visible
│
├── File upload (drag & drop)
│   ├── Drag: PDF file to drop zone
│   ├── Assert: File preview shows
│   ├── Assert: File name displayed
│   ├── Assert: File size displayed
│   ├── Click: btn-upload
│   ├── Assert: Upload progress
│   ├── Assert: Source created
│   └── Assert: Toast "PDF uploaded"
│
├── File upload (click)
│   ├── Click: drop zone
│   ├── Select: PDF file from dialog
│   └── Same as drag & drop
│
├── Multiple files
│   ├── Drag: Multiple PDFs
│   ├── Assert: File list shows all
│   ├── Remove: One file from list
│   ├── Click: btn-upload
│   └── Assert: All files uploaded
│
├── File validation
    ├── Wrong format → Error "PDF only"
    ├── File too large → Error "Max 50MB"
    ├── Corrupted PDF → Error "Invalid PDF"
    └── Encrypted PDF → Error "Password protected"
│
├── Progress indicator
    ├── Show: Upload progress %
    ├── Show: Processing stages
    │   ├── Fetching
    │   ├── Parsing
    │   ├── Chunking
    │   └── Embedding
    └── Show: Completed chunks count
│
└── OCR option (Pro+ plans)
    ├── Toggle: checkbox-enable-ocr
    ├── Click: btn-upload
    ├── Assert: OCR processing
    └── Assert: Better text extraction
```

### 5.4 Add Sitemap Source (`sources-sitemap.spec.ts`)

```
Add Sitemap Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   ├── Click: tab-sitemap
│   └── Assert: Sitemap input visible
│
├── Sitemap URL input
│   ├── Click: input-sitemap-url
│   ├── Type: "https://example.com/sitemap.xml"
│   ├── Click: btn-analyze
│   ├── Assert: Loading state
│   ├── Assert: Sitemap parsed
│   └── Show: URL list preview
│
├── Configuration options
    ├── Max URLs to crawl (input-number)
    │   ├── Default: 100
    │   ├── Min: 1
    │   └── Max: 1000
    │
    ├── Priority patterns (input)
    │   └── Type: "/products/*, /pricing/*"
    │
    └── Exclude patterns (input)
        └── Type: "/admin/*, /private/*"
│
├── Start crawling
│   ├── Click: btn-start-crawling
│   ├── Assert: Source created (processing)
│   ├── Assert: Job queued
│   └── Assert: Toast "Crawling started"
│
├── Crawling progress
    ├── Show: URLs processed count
    ├── Show: URLs pending count
    ├── Show: Errors count
    └── Show: Progress bar
│
├── Approve pending URLs
│   ├── Click: tab-pending
│   ├── Show: Discovered URLs list
│   ├── Click: btn-approve-all
│   ├── Assert: All URLs approved
│   └── Assert: Processing continues
│
└── Validation
    ├── Invalid sitemap → Error
    ├── Empty sitemap → Error
    └── Too many URLs → Warning
```

### 5.5 Add Text Source (`sources-text.spec.ts`)

```
Add Text Source Flow
├── Open add source modal
│   ├── Click: btn-add-source
│   ├── Click: tab-text
│   └── Assert: Text input area visible
│
├── Text input
│   ├── Click: textarea-text
│   ├── Type/Paste: Content
│   ├── Assert: Character count
│   ├── Assert: Word count
│   └── Click: btn-add
│       ├── Assert: Source created
│       └── Assert: Toast "Text added"
│
├── Import from file
│   ├── Click: btn-import-file
│   ├── Select: .txt, .md, .html file
│   ├── Assert: Content imported
│   └── Assert: Source created
│
├── Title input
│   ├── Click: input-title
│   ├── Type: Source title
│   └── Assert: Title saved with source
│
└── Validation
    ├── Empty text → Error
    ├── Text too long → Error "Max 100K chars"
    └── Invalid encoding → Error
```

### 5.6 Chunk Viewer (`sources-chunks.spec.ts`)

```
Chunk Viewer Modal Flow
├── Open chunk viewer
│   ├── Click: btn-view-chunks (on source card)
│   ├── Assert: `modal-chunk-viewer` opens
│   ├── Show: Source title
│   └── Show: Chunk list
│
├── Chunk list
│   ├── Show: All chunks
│   ├── Each chunk shows:
│   │   ├── Chunk number
│   │   ├── Token count
│   │   └── Preview text
│   │
│   ├── Click: Chunk item
│   │   ├── Show: Chunk detail
│   │   ├── Show: Full text
│   │   └── Show: Metadata
│   │
│   └── Hover: Chunk item
│       └── Highlight background
│
├── Search chunks
│   ├── Type: Search term
│   ├── Assert: Filtered results
│   └── Click: Clear search
│
├── Pagination
    ├── Navigate: Pages
    └── Change: Items per page
│
└── Export chunks
    ├── Click: btn-export
    ├── Options:
    │   ├── JSON
    │   ├── CSV
    │   └── Plain text
    ├── Select: Format
    └── Download: File
```

---

## 6. Chat & Playground

### 6.1 Playground Page (`playground.spec.ts`)

#### 6.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `chat-container` | container | Chat messages area |
| `message-user` | component | User message bubble |
| `message-bot` | component | Bot message bubble |
| `message-loading` | component | Loading indicator |
| `message-feedback` | component | Thumbs up/down |
| `input-message` | textarea | Message input |
| `btn-send` | button | Send button |
| `suggestions-carousel` | component | Suggested questions |
| `btn-clear-chat` | button | Clear conversation |
| `btn-download-chat` | button | Download chat history |

#### 6.1.2 Chat Interaction Flow

```
Chat Flow
├── Load playground
│   ├── Assert: Chat container empty
│   ├── Assert: Welcome message shown
│   ├── Assert: Suggestions visible (if enabled)
│   └── Assert: Input enabled
│
├── Send message
│   ├── Type: "Hello, how are you?"
│   ├── Assert: Message appears (user)
│   ├── Assert: Loading indicator
│   ├── Wait: Bot response
│   ├── Assert: Message appears (bot)
│   ├── Assert: Sources cited (if any)
│   └── Assert: Feedback buttons visible
│
├── Send empty message
│   ├── Type: ""
│   ├── Click: btn-send
│   └── Assert: No message sent
│
├── Send long message
│   ├── Type: "A" x 4000
│   ├── Assert: Character count = 4000/4000
│   ├── Type: 1 more char
│   └── Assert: Error "Max 4000 characters"
│
├── Typing indicator
│   ├── Send: Message
│   ├── Assert: Bot shows typing
│   ├── Show: Animated dots
│   └── Hide: After response
│
├── Suggestions
│   ├── Click: Suggestion chip
│   │   ├── Copy: Text to input
│   │   └── Auto-send: After delay
│   │
│   └── Hover: Suggestion chip
│       └── Highlight background
│
└── Clear chat
    ├── Click: btn-clear-chat
    ├── Assert: `modal-confirm-clear` opens
    ├── Click: btn-confirm
    ├── Assert: Chat cleared
    └── Assert: Welcome message shown
```

#### 6.1.3 Message Feedback Flow

```
Feedback Flow
├── User receives bot message
│   ├── Show: Thumbs up/down buttons
│   │
│   ├── Click: Thumbs up
│   │   ├── Assert: Button highlighted
│   │   ├── Assert: Toast "Thanks for feedback"
│   │   └── API: Send feedback
│   │
│   ├── Click: Thumbs down
│   │   ├── Assert: Button highlighted
│   │   ├── Show: Optional feedback form
│   │   ├── Type: Feedback comment
│   │   ├── Click: btn-submit-feedback
│   │   ├── Assert: Toast "Feedback sent"
│   │   └── API: Send feedback with comment
│   │
│   └── Hover: Feedback buttons
│       └── Show: Tooltip
│
└── Toggle feedback
    ├── Click: Already selected thumbs up
    │   ├── Assert: Selection cleared
    │   └── API: Remove feedback
    └── Click: Thumbs down (after thumbs up)
        ├── Assert: Thumbs up cleared
        └── Assert: Thumbs down selected
```

#### 6.1.4 Markdown Rendering

```
Markdown Rendering Flow
├── Bot sends message with markdown
│   ├── Render: **bold** text
│   ├── Render: *italic* text
│   ├── Render: ~~strikethrough~~
│   ├── Render: `inline code`
│   ├── Render: ```code blocks```
│   ├── Render: # Heading 1
│   ├── Render: ## Heading 2
│   ├── Render: - Bullet list
│   ├── Render: 1. Numbered list
│   ├── Render: [link](url)
│   ├── Render: ![image](url)
│   ├── Render: > Blockquote
│   ├── Render: Table
│   └── Render: Horizontal rule
│
├── Click: Link
│   ├── Open: New tab
│   └── Navigate: URL
│
├── Click: Image
│   ├── Open: Lightbox
│   ├── Zoom: Controls
│   └── Close: X button or Escape
│
└── Code block
    ├── Copy: Code button
    │   ├── Click: btn-copy
    │   └── Assert: Toast "Copied"
    └── Language badge
        └── Show: Language name
```

### 6.2 Chat History (`chat-history.spec.ts`)

```
Chat History Flow
├── Conversation list (sidebar)
│   ├── Show: All conversations
│   ├── Each shows:
│   │   ├── Title (first message)
│   │   ├── Date
│   │   └── Message count
│   │
│   ├── Click: Conversation
│   │   ├── Load: Messages
│   │   └── Show: Chat interface
│   │
│   ├── Hover: Conversation
│   │   ├── Show: Options menu
│   │   ├── Click: Rename
│   │   │   ├── Open: `modal-rename`
│   │   │   ├── Type: New title
│   │   │   └── Click: btn-save
│   │   │
│   │   ├── Click: Delete
│   │   │   ├── Open: `modal-confirm`
│   │   │   └── Click: btn-delete
│   │   │
│   │   └── Click: Export
│   │       └── Download: Chat export
│   │
│   └── Create: New conversation
│       ├── Click: btn-new-chat
│       └── Assert: Empty chat shown
│
├── Export chat
│   ├── Click: btn-download-chat
│   ├── Options:
│   │   ├── JSON
│   │   ├── Markdown
│   │   └── PDF
│   ├── Select: Format
│   └── Download: File
│
└── Search in chat
    ├── Type: Search term
    ├── Highlight: Matching messages
    ├── Navigate: Next/Previous match
    └── Clear: Search
```

### 6.3 Human Handoff (`handoff.spec.ts`)

```
Human Handoff Flow
├── Bot response triggers handoff
│   ├── Show: Handoff message
│   ├── Show: "Talk to human" button
│   └── Show: Email input form
│
├── Request handoff
│   ├── Click: btn-request-handoff
│   ├── Assert: Form expands
│   ├── Type: Email
│   ├── Click: btn-submit
│   ├── Assert: Toast "Request submitted"
│   ├── Assert: Confirmation message
│   └── API: Create handoff request
│
├── Handoff status
│   ├── Show: Pending status
│   ├── Show: Request ID
│   └── Show: Expected response time
│
├── After handoff submitted
│   ├── Disable: Email input
│   ├── Show: "We'll contact you at email"
│   └── Show: Status updates
│
└── Admin views handoff
    ├── Navigate: Admin > Handoff Requests
    ├── Show: All pending requests
    ├── Click: Request
    ├── Show: Conversation history
    ├── Click: btn-contact-user
    │   ├── Open: Email client
    │   └── Pre-fill: User email
    └── Click: btn-mark-resolved
        └── Assert: Request resolved
```

---

## 7. Smart Actions

### 7.1 Actions List Page (`actions-list.spec.ts`)

#### 7.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `btn-create-action` | button | Create action |
| `card-action` | card | Action card |
| `card-action-type` | badge | HTTP / Function |
| `card-action-status` | badge | Enabled / Disabled |
| `card-action-logs` | badge | Execution count |

#### 7.1.2 Actions List Flow

```
Actions List Flow
├── Load actions page
│   ├── Assert: Empty state (if no actions)
│   ├── Assert: btn-create-action visible
│   │
│   ├── If actions exist:
│   │   ├── Show: Action cards grid
│   │   ├── Each card shows:
│   │   │   ├── Name
│   │   │   ├── Description
│   │   │   ├── Type badge
│   │   │   ├── Endpoint/method
│   │   │   └── Status toggle
│   │   │
│   │   └── Hover: Action card
│   │       └── Show: Edit and Delete buttons
│   │
│   └── Click: btn-create-action
│       └── Open: `modal-create-action`
│
├── Toggle action status
│   ├── Click: toggle-enabled (on card)
│   ├── Assert: Toggle state changes
│   └── Assert: Toast "Action updated"
│
├── View action logs
│   ├── Click: btn-logs (on card)
│   ├── Open: `modal-action-logs`
│   ├── Show: Execution history
│   ├── Each log shows:
│   │   ├── Timestamp
│   │   ├── Status (success/failed)
│   │   ├── Duration
│   │   └── Request/Response preview
│   └── Filter: By date/status
│
└── Search and filter
    ├── Type: Search term
    └── Filter: By type (HTTP/Function)
```

### 7.2 Create Action Flow (`action-create.spec.ts`)

```
Create Action Flow
├── Open create modal
│   ├── Click: btn-create-action
│   ├── Assert: Modal opens
│   └── Assert: Focus on input-name
│
├── Action type selection
│   ├── Select: HTTP (default)
│   │   └── Show: HTTP configuration
│   │       ├── input-endpoint
│   │       ├── select-method
│   │       ├── textarea-headers
│   │       ├── textarea-body
│   │       └── textarea-parameters
│   │
│   └── Select: Function
│       └── Show: Function configuration
│           ├── textarea-code
│           ├── input-name
│           └── textarea-parameters
│
├── HTTP Action Configuration
│   ├── Name input
│   │   ├── Type: "Get Weather"
│   │   └── Required
│   │
│   ├── Description
│   │   ├── Type: "Fetches weather data"
│   │   └── Optional
│   │
│   ├── Method select
│   │   ├── GET
│   │   ├── POST
│   │   ├── PUT
│   │   └── DELETE
│   │
│   ├── Endpoint input
│   │   ├── Type: "https://api.weather.com"
│   │   └── Validation: Valid URL
│   │
│   ├── Headers (JSON)
│   │   ├── Type: {"Authorization": "Bearer..."}
│   │   └── Validation: Valid JSON
│   │
│   ├── Body (JSON) - for POST/PUT
│   │   ├── Type: {"param": "value"}
│   │   └── Validation: Valid JSON
│   │
│   └── Parameters (array of objects)
│       ├── Add: btn-add-parameter
│       ├── Each parameter:
│       │   ├── name (string)
│       │   ├── type (string/number/boolean)
│       │   ├── required (boolean)
│       │   └── default (optional)
│       └── Remove: btn-remove-parameter
│
├── Function Action Configuration
│   ├── Name input
│   ├── Description
│   ├── Code editor
│   │   ├── Write: JavaScript function
│   │   ├── Validate: Syntax check
│   │   └── Test: btn-test-function
│   └── Parameters (same as HTTP)
│
├── Test action
│   ├── Click: btn-test-action
│   ├── Assert: Test modal opens
│   ├── Fill: Parameter values
│   ├── Click: btn-run
│   ├── Show: Execution result
│   ├── Show: Response/output
│   └── Show: Duration
│
├── Save action
│   ├── Click: btn-save
│   ├── Assert: Validation
│   ├── Assert: Toast success
│   └── Assert: Action created
│
└── Cancel
    ├── Click: btn-cancel
    └── Assert: Modal closed, no action created
```

### 7.3 Edit Action Flow (`action-edit.spec.ts`)

```
Edit Action Flow
├── Open edit modal
│   ├── Hover: Action card
│   ├── Click: btn-edit
│   ├── Assert: Modal opens with current values
│   └── Assert: All fields pre-filled
│
├── Modify fields
│   ├── Change: Name
│   ├── Change: Endpoint
│   ├── Change: Parameters
│   └── Click: btn-save
│       └── Assert: Toast "Action updated"
│
├── Toggle enabled
│   ├── Click: toggle-enabled
│   ├── Assert: Status changes
│   └── Assert: Toast "Action disabled/enabled"
│
├── Test action
│   ├── Click: btn-test
│   ├── Run: With current config
│   └── Assert: Test result
│
└── Delete action
    ├── Click: btn-delete
    ├── Open: `modal-confirm-delete`
    ├── Click: btn-confirm
    └── Assert: Action deleted
```

### 7.4 Action Execution Logs (`action-logs.spec.ts`)

```
Action Logs Flow
├── Open logs modal
│   ├── Click: btn-logs (on action card)
│   ├── Assert: Modal opens
│   └── Show: Execution history
│
├── Log entries
│   ├── Each entry shows:
│   │   ├── Timestamp
│   │   ├── Status icon (green/red)
│   │   ├── Duration (ms)
│   │   ├── Triggered by (user message)
│   │   └── Request/Response preview
│   │
│   ├── Click: Log entry
│   │   ├── Expand: Full details
│   │   ├── Show: Request headers/body
│   │   ├── Show: Response headers/body
│   │   └── Show: Error message (if failed)
│   │
│   └── Hover: Log entry
│       └── Highlight background
│
├── Filter logs
│   ├── Select: Date range
│   ├── Select: Status (all/success/failed)
│   ├── Select: Action
│   └── Assert: Filtered results
│
├── Search logs
│   ├── Type: Search term
│   └── Assert: Filtered by term
│
├── Export logs
│   ├── Click: btn-export
│   ├── Select: Format (JSON/CSV)
│   └── Download: File
│
└── Pagination
    ├── Navigate: Pages
    └── Change: Items per page
```

---

## 8. Settings & Configuration

### 8.1 Profile Settings (`settings-profile.spec.ts`)

```
Profile Settings Flow
├── Load profile page
│   ├── Assert: Current values displayed
│   ├── Show: Avatar
│   ├── Show: Email (read-only)
│   ├── Show: Full name
│   └── Show: Language preference
│
├── Edit profile
│   ├── Click: btn-edit-profile
│   ├── Assert: Form becomes editable
│   │
│   ├── Upload avatar
│   │   ├── Click: input-avatar
│   │   ├── Select: Image file
│   │   ├── Assert: Preview
│   │   ├── Assert: Size validation
│   │   └── Click: btn-save
│   │       └── Assert: Avatar updated
│   │
│   ├── Edit full name
│   │   ├── Clear: Name field
│   │   ├── Type: New name
│   │   └── Click: btn-save
│   │       └── Assert: Name updated
│   │
│   ├── Change language
│   │   ├── Click: select-language
│   │   ├── Select: English
│   │   └── Assert: UI language changes
│   │
│   └── Cancel edit
│       ├── Click: btn-cancel
│       └── Assert: Original values restored
│
├── Change password
│   ├── Click: btn-change-password
│   ├── Assert: `modal-change-password` opens
│   ├── Type: Current password
│   ├── Type: New password
│   ├── Type: Confirm new password
│   ├── Click: btn-update
│   ├── Assert: Validation
│   ├── Assert: Toast "Password updated"
│   └── Assert: User logged out
│
└── Validation
    ├── Empty name → Error
    ├── Invalid avatar → Error
    └── Weak password → Error
```

### 8.2 Organization Settings (`settings-organization.spec.ts`)

```
Organization Settings Flow
├── Load organization page
│   ├── Assert: Current org info displayed
│   ├── Show: Organization name
│   ├── Show: Slug
│   ├── Show: Member count
│   └── Show: Created date
│
├── Edit organization
│   ├── Click: btn-edit-org
│   ├── Assert: Form editable
│   │
│   ├── Edit name
│   │   ├── Clear: Name field
│   │   ├── Type: New name
│   │   └── Click: btn-save
│   │       └── Assert: Name updated
│   │
│   └── Upload logo
│       ├── Click: input-logo
│       ├── Select: Image
│       └── Click: btn-save
│           └── Assert: Logo updated
│
├── Members management
│   ├── Show: Member list table
│   ├── Each member shows:
│   │   ├── Name
│   │   ├── Email
│   │   ├── Role (owner/admin/member)
│   │   └── Status (active/invited)
│   │
│   ├── Invite member
│   │   ├── Click: btn-invite
│   │   ├── Type: Email
│   │   ├── Select: Role
│   │   ├── Click: btn-send-invite
│   │   └── Assert: Invitation sent
│   │
│   ├── Update role
│   │   ├── Click: Role dropdown
│   │   ├── Select: New role
│   │   └── Assert: Role updated
│   │
│   ├── Remove member
│   │   ├── Click: btn-remove
│   │   ├── Open: `modal-confirm`
│   │   ├── Click: btn-confirm
│   │   └── Assert: Member removed
│   │
│   └── Resend invitation
│       ├── Hover: Invited member
│       ├── Click: btn-resend
│       └── Assert: Invitation resent
│
└── Delete organization
    ├── Click: btn-delete-org
    ├── Open: `modal-confirm-delete`
    ├── Type: Organization name to confirm
    ├── Click: btn-delete
    └── Assert: Organization deleted
```

### 8.3 Workspace Settings (`settings-workspace.spec.ts`)

```
Workspace Settings Flow
├── Load workspace page
│   ├── Assert: Current workspaces listed
│   ├── Show: Workspace name
│   ├── Show: Slug
│   └── Show: Chatbot count
│
├── Create workspace
│   ├── Click: btn-create-workspace
│   ├── Type: Name
│   ├── Type: Slug (auto-generated)
│   ├── Click: btn-create
│   └── Assert: Workspace created
│
├── Edit workspace
│   ├── Click: btn-edit (on workspace)
│   ├── Change: Name
│   ├── Click: btn-save
│   └── Assert: Workspace updated
│
├── Delete workspace
│   ├── Click: btn-delete (on workspace)
│   ├── Open: `modal-confirm-delete`
│   ├── Click: btn-confirm
│   └── Assert: Workspace deleted
│
└── Set default workspace
    ├── Click: btn-set-default
    └── Assert: Default workspace updated
```

### 8.4 Plan & Billing (`settings-plan.spec.ts`)

```
Plan & Billing Flow
├── Load plan page
│   ├── Assert: Current plan displayed
│   ├── Show: Plan features
│   ├── Show: Usage stats
│   └── Show: Billing info
│
├── Current plan details
│   ├── Show: Plan name
│   ├── Show: Monthly cost
│   ├── Show: Renewal date
│   └── Show: Payment method
│
├── Usage statistics
│   ├── Show: Tokens used / limit
│   ├── Show: Chatbots used / limit
│   ├── Show: Storage used / limit
│   ├── Show: Files uploaded / limit
│   └── Show: Progress bars for each
│
├── Available plans
│   ├── Show: Current plan (highlighted)
│   ├── Show: Upgrade options
│   ├── Each plan shows:
│   │   ├── Name
│   │   ├── Price
│   │   ├── Features list
│   │   └── btn-upgrade
│   │
│   └── Click: btn-upgrade (on higher plan)
│       ├── Open: `modal-upgrade`
│       ├── Show: Plan comparison
│       ├── Select: Billing period (monthly/yearly)
│       ├── Click: btn-confirm-upgrade
│       └── Assert: Plan upgraded
│
├── Cancel subscription
│   ├── Click: btn-cancel-subscription
│   ├── Open: `modal-confirm-cancel`
│   ├── Show: Consequences
│   ├── Type: "CANCEL" to confirm
│   └── Click: btn-confirm
│       └── Assert: Subscription cancelled
│
├── Update payment method
│   ├── Click: btn-update-payment
│   ├── Open: `modal-payment`
│   ├── Enter: Card details (Stripe Elements)
│   ├── Click: btn-save
│   └── Assert: Payment method updated
│
└── View invoices
    ├── Click: btn-invoices
    ├── Show: Invoice history
    ├── Each invoice:
    │   ├── Date
    │   ├── Amount
    │   ├── Status (paid/pending)
    │   └── Download PDF
    └── Click: Download PDF
        └── Download: Invoice file
```

### 8.5 Privacy Settings (`settings-privacy.spec.ts`)

```
Privacy Settings Flow
├── Load privacy page
│   ├── Assert: Current settings displayed
│   └── Show: Data export options
│
├── Data export
│   ├── Click: btn-export-data
│   ├── Assert: `modal-export-data` opens
│   ├── Select: Data to include
│   │   ├── Chatbots
│   │   ├── Conversations
│   │   ├── Sources
│   │   └── Actions
│   ├── Click: btn-generate-export
│   ├── Wait: Export generation
│   ├── Assert: Download ready
│   └── Download: ZIP file
│
├── Delete account
│   ├── Click: btn-delete-account
│   ├── Open: `modal-delete-account`
│   ├── Show: Consequences
│   ├── Type: Password to confirm
│   ├── Type: "DELETE" to confirm
│   ├── Click: btn-delete
│   ├── Assert: Account deleted
│   └── Assert: All data removed (KVKK compliance)
│
├── Data retention settings
│   ├── Show: Current retention period
│   ├── Change: Retention period
│   │   ├── 30 days
│   │   ├── 90 days
│   │   ├── 1 year
│   │   └── Forever
│   └── Click: btn-save
│       └── Assert: Settings saved
│
└── Activity log
    ├── Show: Recent activity
    ├── Each entry:
    │   ├── Action
    │   ├── Timestamp
    │   └── IP address
    └── Export: Activity log
```

---

## 9. Admin Panel

### 9.1 Admin Dashboard (`admin-dashboard.spec.ts`)

```
Admin Dashboard Flow
├── Load admin dashboard
│   ├── Assert: Sidebar shows Admin section
│   ├── Assert: Active nav item = Dashboard
│   └── Show: System overview
│
├── System stats cards
│   ├── Total Users
│   ├── Total Organizations
│   ├── Total Chatbots
│   ├── Active Sessions
│   ├── Queue Jobs (pending/processing/failed)
│   └── API Response Time
│
├── Recent activity
│   ├── Show: Latest signups
│   ├── Show: Latest chatbots
│   ├── Show: System errors
│   └── Each item clickable → Navigate to detail
│
├── Health indicators
│   ├── Database: Green/Red
│   ├── Redis: Green/Red
│   ├── Qdrant: Green/Red
│   ├── Storage: Green/Red
│   └── Each → Click for details
│
├── Quick actions
│   ├── btn-flush-cache
│   ├── btn-run-migrations
│   ├── btn-send-announcement
│   └── btn-export-stats
│
└── Charts
    ├── Daily active users (line chart)
    ├── Chatbot creations (bar chart)
    ├── Token usage (area chart)
    └── API requests (pie chart)
```

### 9.2 User Management (`admin-users.spec.ts`)

```
Admin Users Flow
├── Load users page
│   ├── Assert: User table visible
│   ├── Show: All users with pagination
│   └── Each user shows:
│       ├── Avatar
│       ├── Name
│       ├── Email
│       ├── Plan
│       ├── Status (active/suspended)
│       ├── Created date
│       └── Last login
│
├── Search users
│   ├── Type: Search term
│   └── Assert: Filtered results
│
├── Filter users
│   ├── By plan (free/pro/ultra)
│   ├── By status (active/suspended)
│   └── By date range
│
├── View user detail
│   ├── Click: User row
│   ├── Show: User detail panel
│   ├── Show: Chatbots created
│   ├── Show: Usage statistics
│   ├── Show: Activity log
│   └── Show: Login history
│
├── Edit user
│   ├── Click: btn-edit (on user)
│   ├── Change: Name
│   ├── Change: Plan
│   ├── Toggle: Admin status
│   └── Click: btn-save
│       └── Assert: User updated
│
├── Suspend user
│   ├── Click: btn-suspend
│   ├── Open: `modal-suspend`
│   ├── Type: Reason
│   ├── Click: btn-confirm
│   └── Assert: User suspended
│
├── Unsuspend user
│   ├── Click: btn-unsuspend (on suspended user)
│   └── Assert: User active
│
├── Delete user
│   ├── Click: btn-delete
│   ├── Open: `modal-confirm-delete`
│   ├── Click: btn-confirm
│   └── Assert: User deleted
│
└── Export users
    ├── Click: btn-export
    ├── Select: Format (CSV/JSON)
    └── Download: File
```

### 9.3 Organization Management (`admin-organizations.spec.ts`)

```
Admin Organizations Flow
├── Load organizations page
│   ├── Assert: Org table visible
│   └── Show: All organizations
│
├── Organization details
│   ├── Name
│   ├── Slug
│   ├── Owner
│   ├── Member count
│   ├── Chatbot count
│   ├── Plan
│   ├── Created date
│   └── Status (active/suspended)
│
├── View organization
│   ├── Click: Org row
│   ├── Show: Detail panel
│   ├── Show: Members list
│   ├── Show: Chatbots list
│   ├── Show: Usage stats
│   └── Show: Billing history
│
├── Edit organization
│   ├── Click: btn-edit
│   ├── Change: Name
│   ├── Change: Plan
│   └── Click: btn-save
│
├── Suspend organization
│   ├── Click: btn-suspend
│   ├── Reason: Type suspension reason
│   └── Assert: Org suspended (all users locked)
│
├── Merge organizations
│   ├── Click: btn-merge
│   ├── Select: Source org
│   ├── Select: Target org
│   ├── Click: btn-merge-confirm
│   └── Assert: Data merged
│
└── Delete organization
    ├── Click: btn-delete
    ├── Open: `modal-confirm-delete`
    ├── Click: btn-confirm
    └── Assert: Org deleted
```

### 9.4 System Health (`admin-health.spec.ts`)

```
System Health Flow
├── Load health page
│   ├── Assert: Service status grid
│   └── Show: Last check time
│
├── Service status
    ├── PostgreSQL: Connected/Disconnected
    ├── Redis: Connected/Disconnected
    ├── Qdrant: Connected/Disconnected
    ├── Storage (R2): Connected/Disconnected
    ├── OpenAI API: Connected/Disconnected
    └── Email service: Connected/Disconnected
│
├── Each service shows
    ├── Status indicator
    ├── Response time
    ├── Last successful check
    └── Error message (if failed)
│
├── Run health check
│   ├── Click: btn-run-check
│   └── Assert: All services checked
│
├── Service logs
│   ├── Click: btn-logs (on service)
│   ├── Show: Recent logs
│   ├── Filter: By level (error/warn/info)
│   └── Search: By term
│
├── Alert configuration
│   ├── Show: Alert thresholds
│   ├── Edit: CPU threshold
│   ├── Edit: Memory threshold
│   ├── Edit: Response time threshold
│   └── Edit: Email recipients
│
└── Incident history
    ├── Show: Past incidents
    ├── Each shows:
    │   ├── Service
    │   ├── Start time
    │   ├── End time
    │   ├── Severity
    │   └── Description
```

### 9.5 Queue Monitoring (`admin-queues.spec.ts`)

```
Queue Monitoring Flow
├── Load queues page
│   ├── Assert: Queue overview visible
│   └── Show: All job queues
│
├── Queue types
    ├── Source processing
    ├── Embedding generation
    ├── Refresh jobs
    ├── Retention jobs
    └── Analytics aggregation
│
├── Each queue shows
    ├── Pending jobs count
    ├── Processing jobs count
    ├── Failed jobs count
    ├── Average wait time
    └── Success rate
│
├── View queue jobs
│   ├── Click: Queue row
│   ├── Show: Job list
│   ├── Each job shows:
│   │   ├── ID
│   │   ├── Status
│   │   ├── Progress
│   │   ├── Created at
    │   └── Started at (if processing)
│   │
│   ├── Filter: By status
│   ├── Filter: By date range
│   └── Search: By ID
│
├── Job actions
│   ├── Click: btn-retry (failed job)
│   │   └── Assert: Job requeued
│   │
│   ├── Click: btn-cancel (pending job)
│   │   └── Assert: Job cancelled
│   │
│   └── Click: btn-view-log (job)
│       └── Show: Job execution log
│
├── Bulk actions
│   ├── Select: Multiple jobs
│   ├── Click: btn-retry-selected
│   └── Assert: All failed jobs retried
│
├── Pause queue
│   ├── Click: btn-pause
│   ├── Assert: Queue paused
│   └── Assert: No new jobs started
│
├── Resume queue
│   ├── Click: btn-resume
│   └── Assert: Queue processing
│
└── Clear queue
    ├── Click: btn-clear
    ├── Open: `modal-confirm`
    ├── Click: btn-confirm
    └── Assert: All pending jobs removed
```

### 9.6 Error Logs (`admin-errors.spec.ts`)

```
Error Logs Flow
├── Load errors page
│   ├── Assert: Error log table visible
│   └── Show: Recent errors
│
├── Error entry shows
    ├── Timestamp
    ├── Level (ERROR/WARN)
    ├── Service/component
    ├── Message
    ├── Request ID
    └── User ID (if applicable)
│
├── Filter errors
    ├── By level (ERROR/WARN/INFO)
    ├── By service
    ├── By date range
    └── By user
│
├── Search errors
    ├── Type: Search term
    └── Assert: Filtered by message
│
├── View error detail
│   ├── Click: Error row
│   ├── Show: Full stack trace
│   ├── Show: Request context
│   ├── Show: User context
│   ├── Show: Related logs
│   └── Show: Timeline
│
├── Error actions
    ├── Click: btn-assign
    │   ├── Select: Developer
    │   └── Assert: Error assigned
    │
    ├── Click: btn-mark-resolved
    │   └── Assert: Error marked resolved
    │
    └── Click: btn-create-ticket
        └── Assert: Ticket created in issue tracker
│
├── Error statistics
    ├── Show: Errors per hour (chart)
    ├── Show: Errors by service (pie chart)
    ├── Show: Top errors (list)
    └── Export: Error report
│
└── Alert rules
    ├── Show: Current alert rules
    ├── Add: New alert rule
    │   ├── Error threshold
    │   ├── Time window
    │   └── Notification channel
    ├── Edit: Existing rule
    └── Delete: Rule
```

---

## 10. Widget Integration

### 10.1 Embed Code Generation (`deploy-embed.spec.ts`)

```
Embed Code Flow
├── Load deploy tab
│   ├── Assert: Embed code section visible
│   └── Show: Generated embed code
│
├── Embed code options
    ├── Script tag (default)
    │   ├── Copy: btn-copy-script
    │   └── Show: Code block
    │
    ├── Iframe tag
    │   ├── Copy: btn-copy-iframe
    │   └── Show: Iframe code
    │
    └── React component
        ├── Copy: btn-copy-component
        └── Show: React component code
│
├── Configuration options
    ├── Position (bottom-right/bottom-left)
    ├── Theme color
    ├── Custom welcome message
    ├── Language
    └── Custom branding
│
├── Preview widget
│   ├── Click: btn-open-preview
│   ├── Assert: Preview modal opens
│   ├── Show: Widget in isolation
│   ├── Test: Chat interaction
│   └── Close: btn-close-preview
│
├── Copy embed code
│   ├── Click: btn-copy-code
│   ├── Assert: Toast "Copied to clipboard"
│   └── Assert: Code in clipboard
│
└── Test on site
    ├── Click: btn-test-on-site
    ├── Open: Test page with embed
    └── Assert: Widget loads correctly
```

### 10.2 Widget Configuration Preview (`deploy-preview.spec.ts`)

```
Widget Configuration Flow
├── Configuration form
│   ├── Position select
│   ├── Color picker (theme color)
│   ├── Color picker (header color)
│   ├── Color picker (bot message color)
│   ├── Color picker (user message color)
│   ├── Font family select
│   ├── Toggle: Auto open
│   ├── Toggle: Hide branding
│   ├── Input: Welcome message
│   ├── Input: Bot display name
│   └── Input: Custom CSS variables
│
├── Live preview
│   ├── Show: Widget preview
│   ├── Real-time: Updates on config change
│   ├── Click: Toggle widget open/close
│   └── Click: Send test message
│
├── Reset to defaults
│   ├── Click: btn-reset-defaults
│   └── Assert: All values reset
│
├── Export configuration
│   ├── Click: btn-export-config
│   ├── Download: JSON file
│   └── Import: JSON file
│
└── Save configuration
    ├── Click: btn-save
    ├── Assert: Toast "Configuration saved"
    └── Assert: Embed code updated
```

---

## 11. Edge Cases & Error States

### 11.1 Network Error Handling

```
Network Error Scenarios
├── API timeout (30s)
│   ├── Show: Loading spinner timeout
│   ├── Show: Toast "Request timed out"
│   ├── Option: Retry button
│   └── Click: Retry
│       └── Re-attempt API call
│
├── 401 Unauthorized
│   ├── Show: Session expired modal
│   ├── Click: btn-relogin
│   └── Redirect: /login
│
├── 403 Forbidden
│   ├── Show: Access denied message
│   └── Option: Contact admin
│
├── 404 Not Found
│   ├── Show: 404 page
│   ├── Click: btn-go-home
│   └── Navigate: /dashboard
│
├── 429 Rate Limited
│   ├── Show: Rate limit toast
│   ├── Show: Retry after timer
│   └── Block: Input until timer expires
│
├── 500 Server Error
│   ├── Show: Error page
│   ├── Show: Error ID (for support)
│   ├── Option: Report issue
│   └── Option: Retry
│
├── Network offline
│   ├── Show: Offline indicator
│   ├── Disable: All API calls
│   ├── Queue: Actions for later
│   └── When online: Auto-retry queued
│
└── WebSocket disconnect
    ├── Show: Reconnecting indicator
    ├── Auto: Retry connection
    └── After reconnect: Resume chat
```

### 11.2 Form Validation Errors

```
Form Validation Scenarios
├── Required field empty
│   ├── Blur: Field
│   ├── Show: Error message
│   ├── Show: Error icon
│   └── Prevent: Form submission
│
├── Invalid format
│   ├── Email: "not-an-email"
│   │   ├── Show: "Invalid email format"
│   │   └── Suggest: "did you mean...?"
│   │
│   ├── URL: "not-a-url"
│   │   ├── Show: "Invalid URL"
│   │   └── Hint: "https://..."
│   │
│   └── JSON: "{invalid}"
│       ├── Show: "Invalid JSON"
│       └── Show: Parser error
│
├── Length validation
    ├── Too short: "ab"
    │   └── Show: "Minimum 3 characters"
    │
    └── Too long: "a" x 1000
        └── Show: "Maximum 500 characters"
│
├── Number validation
    ├── Less than min: 0 (min: 1)
    │   └── Show: "Minimum value is 1"
    │
    └── Greater than max: 10000 (max: 1000)
        └── Show: "Maximum value is 1000"
│
├── Match validation
    ├── Passwords don't match
    │   └── Show: "Passwords do not match"
    │
    └── Email mismatch
        └── Show: "Emails do not match"
│
└── Custom validation
    ├── Username taken
    │   └── Show: "Username already taken"
    │
    ├── Chatbot limit reached
    │   └── Show: "Upgrade plan for more chatbots"
    │
    └── File too large
        └── Show: "File exceeds 50MB limit"
```

### 11.3 File Upload Errors

```
File Upload Error Scenarios
├── Wrong file type
│   ├── Select: .exe file
│   └── Show: "Only PDF, TXT, MD files allowed"
│
├── File too large
│   ├── Select: 100MB file
│   ├── Show: "Maximum file size: 50MB"
│   └── Prevent: Upload start
│
├── Corrupted file
│   ├── Select: Corrupted PDF
│   ├── During: Upload progress
│   ├── Show: "File parsing failed"
│   └── Suggest: "Try another file"
│
├── Network interruption
│   ├── During: Upload (50%)
│   ├── Show: Upload paused
│   ├── Option: Resume
│   └── Option: Cancel and retry
│
├── Virus detected
│   ├── Select: Malicious file
│   ├── Show: "File rejected for security"
│   └── No: Further details (security)
│
└── Storage quota exceeded
    ├── Select: File
    ├── Show: "Storage quota exceeded"
    ├── Link: Upgrade plan
    └── Option: Delete old files
```

### 11.4 Modal/Confirmation Dialogs

```
Modal Flow
├── Open modal
│   ├── Click: Trigger button
│   ├── Assert: Modal visible
│   ├── Assert: Focus trapped in modal
│   ├── Assert: Body scroll locked
│   └── Escape: Close modal
│
├── Close modal
│   ├── Click: btn-close (X)
│   ├── Click: btn-cancel
│   ├── Click: Overlay (outside modal)
│   └── Press: Escape
│
├── Confirmation dialog
│   ├── Show: Warning message
│   ├── Show: Destructive action red
│   ├── May require: Text confirmation
│   ├── btn-confirm: Primary (danger)
│   └── btn-cancel: Secondary
│
├── Unsaved changes
│   ├── Click: btn-cancel with changes
│   ├── Show: `modal-unsaved-changes`
│   ├── Options:
│   │   ├── Discard changes
│   │   ├── Save and continue
│   │   └── Keep editing
│   └── Select: Option
│
└── Loading state
    ├── During: API call
    ├── Show: Spinner
    ├── Disable: Buttons
    └── Prevent: Close
```

---

## 12. Accessibility Tests

### 12.1 Keyboard Navigation

```
Keyboard Navigation Tests
├── Tab order
│   ├── Forward: Tab moves logical order
│   ├── Backward: Shift+Tab reverse order
│   └── Visible: Focus indicator
│
├── Skip links
│   ├── Press: Tab (on page load)
│   ├── Show: "Skip to main content"
│   ├── Click: Skip link
│   └── Focus: Moves to main content
│
├── Focus management
│   ├── Open: Modal → Focus inside modal
│   ├── Close: Modal → Focus returns to trigger
│   ├── Navigate: Page → Focus visible
│   └── Error: Focus not lost
│
├── Shortcuts
│   ├── Ctrl+Enter: Submit form
│   ├── Escape: Close modal/dropdown
│   ├── Arrow keys: Navigate menus
│   └── Space/Enter: Activate button
│
└── Custom elements
    ├── Buttons: Keyboard activatable
    ├── Dropdowns: Arrow key navigation
    ├── Tabs: Arrow key navigation
    ├── Sliders: Arrow key adjustment
    └── Drag-drop: Keyboard alternatives
```

### 12.2 Screen Reader

```
Screen Reader Tests
├── ARIA labels
    ├── Buttons: Label text
    ├── Inputs: Associated label
    ├── Images: Alt text
    ├── Links: Descriptive text
    └── Icons: aria-label
│
├── Live regions
    ├── Toast: aria-live="polite"
    ├── Errors: aria-live="assertive"
    └── Updates: Announced to screen reader
│
├── Semantic HTML
    ├── Headings: h1 > h2 > h3 hierarchy
    ├── Lists: ul/ol > li
    ├── Tables: th with scope
    ├── Forms: fieldset > legend
    └── Buttons: <button> not <div>
│
├── Dynamic content
    ├── Loading: Announced
    ├── Success: Announced
    ├── Error: Announced (assertive)
    └── New content: Announced
│
└── Interactive states
    ├── Expanded/collapsed: aria-expanded
    ├── Selected: aria-selected
    ├── Checked: aria-checked
    ├── Disabled: aria-disabled
    └── Hidden: aria-hidden
```

### 12.3 Color Contrast

```
Color Contrast Tests
├── Text contrast (WCAG AA)
    ├── Normal text: 4.5:1 minimum
    ├── Large text: 3:1 minimum
    └── UI components: 3:1 minimum
│
├── Focus indicators
    ├── Visible: 2px solid outline
    ├── Contrast: 3:1 against background
    └── Offset: 2px from element
│
├── Error states
    ├── Red text: 4.5:1 minimum
    ├── Red background: 3:1 minimum
    └── Icon indicators: Complementary
│
└── Dark mode
    ├── Contrast: Maintained in dark mode
    ├── Text: Readable on dark backgrounds
    └── Focus: Visible on dark backgrounds
```

---

## 13. Performance Tests

### 13.1 Load Time Tests

```
Load Time Benchmarks
├── First Contentful Paint (FCP)
│   ├── Target: < 1.8s
│   └── Test: Measure FCP
│
├── Largest Contentful Paint (LCP)
│   ├── Target: < 2.5s
│   └── Test: Measure LCP
│
├── Time to Interactive (TTI)
│   ├── Target: < 3.8s
│   └── Test: Measure TTI
│
├── Cumulative Layout Shift (CLS)
│   ├── Target: < 0.1
│   └── Test: Measure CLS
│
└── Page load with data
    ├── Dashboard: < 2s
    ├── Chatbot list: < 1.5s
    ├── Chatbot detail: < 2s
    ├── Sources list: < 1.5s
    ├── Playground: < 2s
    └── Settings: < 1.5s
```

### 13.2 API Performance Tests

```
API Performance Benchmarks
├── Chat response
    ├── Target: < 3s for response start
    ├── Target: < 10s for full response
    └── Streaming: < 500ms first chunk
│
├── Source processing
    ├── Small PDF (<1MB): < 30s
    ├── Medium PDF (1-10MB): < 2min
    ├── Large PDF (>10MB): < 5min
    └── URL fetch: < 10s per page
│
├── List endpoints
    ├── Response: < 500ms
    ├── Pagination: < 200ms
    └── Search: < 300ms
│
└── Concurrent requests
    ├── 10 concurrent: No degradation
    ├── 50 concurrent: Acceptable slowdown
    └── 100 concurrent: Graceful degradation
```

### 13.3 Memory Tests

```
Memory Performance Tests
├── Memory leak detection
    ├── Open/close modal: No leak
    ├── Navigate pages: No leak
    ├── Chat messages: No infinite growth
    └── Long session: Stable memory
│
├── Large data handling
    ├── 1000+ chatbots: Render efficiently
    ├── 10000+ messages: Paginate correctly
    └── Large files: Stream properly
│
└── Background processing
    ├── WebSocket: Handle reconnection
    ├── Polling: Clean up intervals
    └── Event listeners: Proper cleanup
```

---

## 14. Test Utilities & Helpers

### 14.1 Common Test Utilities

```typescript
// frontend/e2e/utils/test-helpers.ts

export class TestHelpers {
  // Wait for element to be visible
  static async waitForVisible(page, selector, timeout = 5000) { ... }
  
  // Wait for element to be hidden
  static async waitForHidden(page, selector, timeout = 5000) { ... }
  
  // Wait for API call to complete
  static async waitForResponse(page, urlPattern) { ... }
  
  // Take screenshot on failure
  static async takeScreenshot(page, name) { ... }
  
  // Generate random test data
  static generateTestData() { ... }
  
  // Wait for toast to appear
  static async waitForToast(page, type) { ... }
  
  // Clear all local storage
  static async clearStorage(page) { ... }
  
  // Set authentication state
  static async setAuthState(page, user) { ... }
}
```

### 14.2 Page Objects

```typescript
// frontend/e2e/pages/login.page.ts

export class LoginPage {
  constructor(page) {
    this.page = page;
    this.emailInput = page.locator('[data-testid="input-email"]');
    this.passwordInput = page.locator('[data-testid="input-password"]');
    this.loginButton = page.locator('[data-testid="btn-login"]');
    this.errorToast = page.locator('[data-testid="toast-error"]');
  }
  
  async goto() {
    await this.page.goto('/login');
  }
  
  async fillEmail(email) {
    await this.emailInput.fill(email);
  }
  
  async fillPassword(password) {
    await this.passwordInput.fill(password);
  }
  
  async clickLogin() {
    await this.loginButton.click();
  }
  
  async login(email, password) {
    await this.fillEmail(email);
    await this.fillPassword(password);
    await this.clickLogin();
  }
  
  async getErrorMessage() {
    return this.errorToast.textContent();
  }
}
```

---

## Appendix A: Test Environment Setup

### A.1 Required Environment Variables

```bash
# .env.test
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
TEST_USER_EMAIL=test@example.com
TEST_USER_PASSWORD=Test123@
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=Admin123@
```

### A.2 Browser Support Matrix

| Browser | Version | Support |
|---------|---------|---------|
| Chrome | 120+ | Full |
| Firefox | 120+ | Full |
| Safari | 16+ | Full |
| Edge | 120+ | Full |

### A.3 Test Data Fixtures

| Fixture | Purpose |
|---------|---------|
| `user_free` | Free tier user |
| `user_pro` | Pro tier user |
| `user_admin` | Platform admin |
| `chatbot_with_sources` | Chatbot with training data |
| `chatbot_empty` | Empty chatbot |
| `organization` | Test organization |

---

## Appendix B: CI/CD Integration

### B.1 GitHub Actions Workflow

```yaml
# .github/workflows/e2e-tests.yml
name: E2E Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm run build
      - run: npm run test:e2e
        env:
          VITE_API_URL: ${{ secrets.API_URL }}
```

---

## Appendix C: Reporting

### C.1 Test Results Format

- **HTML Report**: `playwright-report/index.html`
- **JUnit XML**: `test-results/junit.xml`
- **JSON**: `test-results/results.json`
- **Allure**: `allure-results/` directory

### C.2 Coverage Reports

- **Statement Coverage**: `coverage/index.html`
- **Branch Coverage**: `coverage/branch.html`
- **Function Coverage**: `coverage/function.html`

---

*End of Comprehensive Test Paths Documentation*

**Next Steps:**
1. Create page object models for each major section
2. Implement tests following this specification
3. Set up CI/CD pipeline
4. Establish test data fixtures
5. Configure parallel execution
6. Set up reporting and analytics
