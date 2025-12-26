# Task 04: Create Test Fixture Factory

**Priority:** 🟢 Low  
**Effort:** Medium (3-4 hours)  
**Risk Level:** Low (test infrastructure only)

---

## Problem Statement

Test setup logic (user creation, organization creation, plan mocking) is duplicated across multiple test files. When schemas change, developers must update setup functions in many locations.

### Evidence

Functions like `setupTestDB`, `createTestUser`, and `createTestOrg` are redefined in:
- `internal/api/handlers/auth_test.go`
- `internal/api/handlers/organization_mgmt_test.go`
- `internal/integration/source_create_test.go`
- Many other test files

### Current State

The codebase already has `internal/testdb` with:
- `OpenTestDB()` - Opens test database connection
- `OpenParallelTestDB()` - Creates isolated schema for parallel tests
- `WithTx()` - Transaction-based test isolation

**What's Missing:** Model fixture factories for creating test entities.

---

## Acceptance Criteria

- [ ] Centralized fixture factory in `internal/testdb/fixtures.go`
- [ ] Factory functions for all major entities (User, Organization, Workspace, Chatbot, Source)
- [ ] Sensible defaults with optional overrides
- [ ] Duplicate helper functions removed from individual test files
- [ ] All existing tests pass with new fixtures

---

## Implementation Steps

### Step 1: Create Fixture Factory

**File:** `internal/testdb/fixtures.go`

```go
package testdb

import (
    "context"
    "database/sql"
    "testing"
    "time"
    
    "github.com/google/uuid"
    "github.com/onurceri/botla-co/internal/db"
    "github.com/onurceri/botla-co/internal/models"
)

// UserFixture configures a test user
type UserFixture struct {
    ID              string
    Email           string
    Password        string // Plain text, will be hashed
    Name            string
    IsVerified      bool
    IsPlatformAdmin bool
    PlanTier        string
}

// DefaultUserFixture returns sensible defaults
func DefaultUserFixture() UserFixture {
    return UserFixture{
        ID:              uuid.NewString(),
        Email:           "test-" + uuid.NewString()[:8] + "@example.com",
        Password:        "TestPassword123!",
        Name:            "Test User",
        IsVerified:      true,
        IsPlatformAdmin: false,
        PlanTier:        "free",
    }
}

// CreateUser creates a user in the test database
func CreateUser(t *testing.T, dbConn *sql.DB, fixture ...UserFixture) *models.User {
    t.Helper()
    
    f := DefaultUserFixture()
    if len(fixture) > 0 {
        f = mergeUserFixture(f, fixture[0])
    }
    
    ctx := context.Background()
    
    // Hash password
    hashedPassword, err := hashPassword(f.Password)
    if err != nil {
        t.Fatalf("failed to hash password: %v", err)
    }
    
    // Insert user
    user, err := db.CreateUser(ctx, dbConn, db.CreateUserParams{
        ID:              f.ID,
        Email:           f.Email,
        PasswordHash:    hashedPassword,
        Name:            f.Name,
        IsVerified:      f.IsVerified,
        IsPlatformAdmin: f.IsPlatformAdmin,
    })
    if err != nil {
        t.Fatalf("failed to create test user: %v", err)
    }
    
    // Assign plan if needed
    if f.PlanTier != "" {
        if err := assignPlan(ctx, dbConn, user.ID, f.PlanTier); err != nil {
            t.Fatalf("failed to assign plan: %v", err)
        }
    }
    
    return user
}

// mergeUserFixture merges override into defaults
func mergeUserFixture(defaults, override UserFixture) UserFixture {
    if override.ID != "" {
        defaults.ID = override.ID
    }
    if override.Email != "" {
        defaults.Email = override.Email
    }
    if override.Password != "" {
        defaults.Password = override.Password
    }
    if override.Name != "" {
        defaults.Name = override.Name
    }
    if override.PlanTier != "" {
        defaults.PlanTier = override.PlanTier
    }
    defaults.IsVerified = override.IsVerified
    defaults.IsPlatformAdmin = override.IsPlatformAdmin
    return defaults
}
```

### Step 2: Add Organization Fixture

```go
// OrganizationFixture configures a test organization
type OrganizationFixture struct {
    ID      string
    Name    string
    OwnerID string
}

// DefaultOrganizationFixture returns sensible defaults
func DefaultOrganizationFixture() OrganizationFixture {
    return OrganizationFixture{
        ID:   uuid.NewString(),
        Name: "Test Organization " + uuid.NewString()[:8],
    }
}

// CreateOrganization creates an organization in the test database
// If no OwnerID is provided, creates a user first
func CreateOrganization(t *testing.T, dbConn *sql.DB, fixture ...OrganizationFixture) (*models.Organization, *models.User) {
    t.Helper()
    
    f := DefaultOrganizationFixture()
    if len(fixture) > 0 {
        f = mergeOrganizationFixture(f, fixture[0])
    }
    
    // Create owner if not provided
    var owner *models.User
    if f.OwnerID == "" {
        owner = CreateUser(t, dbConn)
        f.OwnerID = owner.ID
    } else {
        // Load existing user
        var err error
        owner, err = db.GetUserByID(context.Background(), dbConn, f.OwnerID)
        if err != nil {
            t.Fatalf("failed to load owner: %v", err)
        }
    }
    
    ctx := context.Background()
    org, err := db.CreateOrganization(ctx, dbConn, db.CreateOrganizationParams{
        ID:      f.ID,
        Name:    f.Name,
        OwnerID: f.OwnerID,
    })
    if err != nil {
        t.Fatalf("failed to create test organization: %v", err)
    }
    
    return org, owner
}
```

### Step 3: Add Workspace Fixture

```go
// WorkspaceFixture configures a test workspace
type WorkspaceFixture struct {
    ID             string
    Name           string
    OrganizationID string
}

// CreateWorkspace creates a workspace in the test database
func CreateWorkspace(t *testing.T, dbConn *sql.DB, fixture ...WorkspaceFixture) (*models.Workspace, *models.Organization, *models.User) {
    t.Helper()
    
    f := DefaultWorkspaceFixture()
    if len(fixture) > 0 {
        f = mergeWorkspaceFixture(f, fixture[0])
    }
    
    // Create org if not provided
    var org *models.Organization
    var owner *models.User
    if f.OrganizationID == "" {
        org, owner = CreateOrganization(t, dbConn)
        f.OrganizationID = org.ID
    }
    
    ctx := context.Background()
    workspace, err := db.CreateWorkspace(ctx, dbConn, db.CreateWorkspaceParams{
        ID:             f.ID,
        Name:           f.Name,
        OrganizationID: f.OrganizationID,
    })
    if err != nil {
        t.Fatalf("failed to create test workspace: %v", err)
    }
    
    return workspace, org, owner
}
```

### Step 4: Add Chatbot Fixture

```go
// ChatbotFixture configures a test chatbot
type ChatbotFixture struct {
    ID             string
    Name           string
    WorkspaceID    string
    UserID         string
    Model          string
    WelcomeMessage string
}

// DefaultChatbotFixture returns sensible defaults
func DefaultChatbotFixture() ChatbotFixture {
    return ChatbotFixture{
        ID:             uuid.NewString(),
        Name:           "Test Bot " + uuid.NewString()[:8],
        Model:          "gpt-4o-mini",
        WelcomeMessage: "Hello! How can I help you?",
    }
}

// CreateChatbot creates a chatbot with full hierarchy (workspace, org, user)
func CreateChatbot(t *testing.T, dbConn *sql.DB, fixture ...ChatbotFixture) *ChatbotTestContext {
    t.Helper()
    
    f := DefaultChatbotFixture()
    if len(fixture) > 0 {
        f = mergeChatbotFixture(f, fixture[0])
    }
    
    // Create full hierarchy if not provided
    var workspace *models.Workspace
    var org *models.Organization
    var user *models.User
    
    if f.WorkspaceID == "" {
        workspace, org, user = CreateWorkspace(t, dbConn)
        f.WorkspaceID = workspace.ID
        f.UserID = user.ID
    }
    
    ctx := context.Background()
    chatbot, err := db.CreateChatbot(ctx, dbConn, db.CreateChatbotParams{
        ID:             f.ID,
        Name:           f.Name,
        WorkspaceID:    f.WorkspaceID,
        UserID:         f.UserID,
        Model:          f.Model,
        WelcomeMessage: f.WelcomeMessage,
    })
    if err != nil {
        t.Fatalf("failed to create test chatbot: %v", err)
    }
    
    return &ChatbotTestContext{
        Chatbot:   chatbot,
        Workspace: workspace,
        Org:       org,
        User:      user,
    }
}

// ChatbotTestContext provides all related entities for testing
type ChatbotTestContext struct {
    Chatbot   *models.Chatbot
    Workspace *models.Workspace
    Org       *models.Organization
    User      *models.User
}
```

### Step 5: Migrate Existing Tests

Update test files to use the new fixtures:

**Before:**
```go
func TestSomething(t *testing.T) {
    db := setupTestDB(t)
    user := createTestUser(t, db, "test@example.com") // Local helper
    org := createTestOrg(t, db, user.ID)              // Local helper
    // ...
}
```

**After:**
```go
func TestSomething(t *testing.T) {
    db := testdb.OpenTestDB(t)
    ctx := testdb.CreateChatbot(t, db) // Creates user, org, workspace, chatbot
    
    // Access all entities
    user := ctx.User
    chatbot := ctx.Chatbot
    // ...
}
```

### Step 6: Delete Duplicated Helpers

Remove local `createTestUser`, `createTestOrg` functions from individual test files after migration.

---

## Testing Checklist

- [ ] `go build ./...` succeeds
- [ ] `make test-no-pdf` passes
- [ ] `make lint` passes
- [ ] Fixtures work with parallel tests
- [ ] All existing tests migrate without behavior change

---

## Files to Create/Modify

| File | Change |
|------|--------|
| `internal/testdb/fixtures.go` | Create new file with fixture factory |
| `internal/testdb/fixtures_test.go` | Tests for the fixtures themselves |
| `internal/api/handlers/*_test.go` | Migrate to use fixtures |
| `internal/integration/*_test.go` | Migrate to use fixtures |

---

## Fixture Usage Examples

```go
// Simple: Just need a user
user := testdb.CreateUser(t, db)

// With overrides
admin := testdb.CreateUser(t, db, testdb.UserFixture{
    IsPlatformAdmin: true,
    PlanTier: "business",
})

// Full hierarchy for chatbot tests
ctx := testdb.CreateChatbot(t, db)
// ctx.Chatbot, ctx.Workspace, ctx.Org, ctx.User all available

// Custom chatbot
ctx := testdb.CreateChatbot(t, db, testdb.ChatbotFixture{
    Model: "gpt-4",
    Name:  "Custom Bot",
})
```

---

## Related Issues

- Code Audit Finding #5: "High Duplication in Test Infrastructure"
- Existing `internal/testdb` package needs enhancement
