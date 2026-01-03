# Test Plan: Go Backend Parallelization Fix

**Plan ID:** TP-GO-PARALLEL-001  
**Priority:** CRITICAL  
**Estimated Duration:** 1-2 weeks  
**Target:** Remove all t.Setenv() blocking patterns, enable true parallel test execution  
**Status:** Draft  

---

## Executive Summary

This plan addresses the critical anti-pattern of using `t.Setenv()` in Go tests, which blocks parallel test execution. Currently, 42+ test files use this pattern, preventing the test suite from running efficiently in parallel. By migrating to `testutils.TestConfigWith()`, we can maintain test isolation while enabling full parallelization.

---

## Sisyphus Agent Prompt

```
You are Sisyphus, a senior Go engineer with expertise in testing patterns, parallel execution, and Go testing best practices.

### Task Context
The Botla backend test suite has 42+ files using `t.Setenv()` which blocks parallel test execution. This is a critical anti-pattern that prevents:
1. Efficient CI/CD pipeline execution
2. Fast feedback loops for developers
3. True test isolation

The current pattern blocks all parallel tests when ANY test sets an environment variable:
```go
// WRONG - This blocks all parallel tests in the package
t.Setenv("DATABASE_URL", "postgres://...")

// CORRECT - This allows parallel execution
cfg := testutils.TestConfigWith(func(c *config.Config) {
    c.DatabaseURL = "postgres://..."
})
```

### Your Mission
1. Search ALL Go files for `t.Setenv` usage patterns
2. For each occurrence:
   - Understand what environment variable is being set
   - Find the corresponding config field in pkg/config/config.go
   - Create a test config override using testutils.TestConfigWith()
   - Remove the t.Setenv() call
3. Verify tests still pass after changes
4. Ensure new pattern uses t.Parallel() for parallelization
5. Run tests in parallel to verify improvement

### Critical Rules
- NEVER remove tests - only refactor the pattern
- Use testutils.TestConfigWith() for config overrides
- For tests that MUST use environment variables (rare), document why
- Maintain the same test behavior and assertions
- Run make test after each batch of changes

### Files to Focus On (Priority Order)
1. internal/integration/*_test.go (highest priority)
2. internal/rag/*_test.go
3. internal/services/*_test.go
4. internal/repository/*_test.go
5. pkg/*/*_test.go

### Deliverables
- All t.Setenv() calls refactored to testutils.TestConfigWith()
- All refactored tests include t.Parallel()
- All tests pass after refactoring
- Test execution time reduced by 50%+
- Documentation of any legitimate t.Setenv() usage

Begin by searching for all t.Setenv occurrences in the codebase.
```

---

## Current State Analysis

### Files with t.Setenv() Usage

#### Critical Priority (Integration Tests)

| File | Count | Primary Variables | Business Impact |
|------|-------|-------------------|-----------------|
| `internal/integration/auth_test.go` | 5+ | DATABASE_URL, JWT_SECRET | Authentication flows |
| `internal/integration/chat_test.go` | 5+ | DATABASE_URL, REDIS_URL | Chat operations |
| `internal/integration/sources_test.go` | 3+ | DATABASE_URL | Source management |
| `internal/integration/chatbot_test.go` | 3+ | DATABASE_URL | Chatbot CRUD |

#### High Priority (RAG Tests)

| File | Count | Primary Variables | Business Impact |
|------|-------|-------------------|-----------------|
| `internal/rag/chunker_test.go` | 2+ | QDRANT_URL, OPENAI_API_KEY | Text chunking |
| `internal/rag/embed_openai_test.go` | 2+ | OPENAI_API_KEY | Embeddings |

#### Medium Priority (Service Tests)

| File | Count | Primary Variables | Business Impact |
|------|-------|-------------------|-----------------|
| `internal/services/action_service_impl_test.go` | 2+ | DATABASE_URL | Action CRUD |
| `internal/services/plan_service_test.go` | 2+ | Various config | Plan caching |

---

## Environment Variable to Config Mapping

### Common Mappings

```go
// BEFORE (blocking)
t.Setenv("DATABASE_URL", "postgres://...")

// AFTER (parallel-safe)
cfg := testutils.TestConfigWith(func(c *config.Config) {
    c.DatabaseURL = "postgres://..."
})
```

### Complete Mapping Table

| Environment Variable | Config Field | Type | Notes |
|---------------------|--------------|------|-------|
| `DATABASE_URL` | `c.DatabaseURL` | string | PostgreSQL connection |
| `DATABASE_HOST` | `c.DBHost` | string | Host only |
| `DATABASE_PORT` | `c.DBPort` | int | Port only |
| `REDIS_URL` | `c.RedisURL` | string | Redis connection |
| `JWT_SECRET` | `c.JWTSecret` | string | JWT signing key |
| `OPENAI_API_KEY` | `c.OpenAIAPIKey` | string | OpenAI access |
| `QDRANT_URL` | `c.QdrantURL` | string | Vector DB |
| `AWS_ACCESS_KEY_ID` | `c.AWSAccessKeyID` | string | AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | `c.AWSSecretAccessKey` | string | AWS credentials |
| `AWS_REGION` | `c.AWSRegion` | string | AWS region |
| `R2_ACCOUNT_ID` | `c.R2AccountID` | string | Cloudflare R2 |
| `R2_ACCESS_KEY_ID` | `c.R2AccessKeyID` | string | R2 credentials |
| `R2_SECRET_ACCESS_KEY` | `c.R2SecretAccessKey` | string | R2 credentials |
| `ENVIRONMENT` | `c.Environment` | string | dev/staging/prod |

---

## Step-by-Step Implementation Plan

### Phase 1: Inventory and Analysis (Day 1)

#### Step 1.1: Find All t.Setenv() Occurrences
```bash
# Search for all t.Setenv usage
grep -rn "t\.Setenv" --include="*_test.go" /Users/onur/Documents/workspace/botla-co

# Output to file for analysis
grep -rn "t\.Setenv" --include="*_test.go" /Users/onur/Documents/workspace/botla-co > docs/test-plans/tsetenv-inventory.txt

# Count occurrences
wc -l docs/test-plans/tsetenv-inventory.txt
```

#### Step 1.2: Categorize by Type
Create a categorization:

1. **Config-based** (90%) - Can be replaced with `testutils.TestConfigWith()`
2. **External dependency** (5%) - May need custom mock
3. **Legacy pattern** (5%) - No longer needed

#### Step 1.3: Identify Blockers
- Tests that absolutely require environment variables
- Tests with complex setup dependencies
- Tests with unclear variable purpose

---

### Phase 2: Integration Tests Refactoring (Days 2-4)

#### Step 2.1: Pattern Transformation Template

**Before:**
```go
func TestChatService_SendMessage(t *testing.T) {
    t.Setenv("DATABASE_URL", testDBURL)
    t.Setenv("JWT_SECRET", testJWTSecret)
    
    db := testdb.OpenTestDB(t)
    // ... test implementation
}
```

**After:**
```go
func TestChatService_SendMessage(t *testing.T) {
    t.Parallel() // Enable parallelization
    
    cfg := testutils.TestConfigWith(func(c *config.Config) {
        c.DatabaseURL = testDBURL
        c.JWTSecret = testJWTSecret
    })
    
    db := testdb.OpenTestDB(t)
    // ... test implementation
}
```

#### Step 2.2: Refactor internal/integration/auth_test.go

**File Analysis:**
- Lines with t.Setenv: 5+
- Config variables: DATABASE_URL, JWT_SECRET, REDIS_URL
- Tests affected: 3 major test functions

**Refactoring Steps:**
1. Add `t.Parallel()` to each test function
2. Replace each `t.Setenv()` with `testutils.TestConfigWith()`
3. Move config setup before DB initialization
4. Verify tests pass

**Expected Result:**
```go
func TestAuthService_Login(t *testing.T) {
    t.Parallel()
    
    cfg := testutils.TestConfigWith(func(c *config.Config) {
        c.DatabaseURL = postgresTestURL
        c.JWTSecret = "test-secret-key"
        c.RedisURL = redisTestURL
    })
    
    db := testdb.OpenTestDBWithConfig(t, cfg)
    svc := auth.NewService(db, cfg)
    
    // Test implementation
}
```

#### Step 2.3: Refactor internal/integration/chat_test.go

**File Analysis:**
- Lines with t.Setenv: 5+
- Config variables: DATABASE_URL, REDIS_URL, QDRANT_URL
- Tests affected: 5+ test functions

**Key Considerations:**
- Chat tests may require specific Redis/Qdrant setup
- Use `testutils.TestRAGConfig()` for RAG-related tests

---

### Phase 3: RAG Tests Refactoring (Day 5)

#### Step 3.1: Pattern for RAG Tests

```go
func TestChunker_Split(t *testing.T) {
    t.Parallel()
    
    cfg := testutils.TestRAGConfig()
    
    chunker := rag.NewChunker(cfg)
    
    tests := []struct {
        name  string
        input string
        want  int
    }{
        {"short text", "Hello world", 1},
        {"long text", strings.Repeat("a", 10000), 10},
    }
    
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            chunks := chunker.Split(tt.input)
            assert.Len(t, chunks, tt.want)
        })
    }
}
```

#### Step 3.2: Files to Refactor
- `internal/rag/chunker_test.go`
- `internal/rag/embed_openai_test.go`
- Any other RAG-related test files

---

### Phase 4: Service Tests Refactoring (Day 6-7)

#### Step 4.1: Service Layer Pattern

**Before (blocking):**
```go
func TestActionService_Create(t *testing.T) {
    t.Setenv("DATABASE_URL", testDBURL)
    
    db := testdb.OpenTestDB(t)
    repo := action_repo.New(db)
    svc := action_service.New(repo)
    
    // Test implementation
}
```

**After (parallel-safe):**
```go
func TestActionService_Create(t *testing.T) {
    t.Parallel()
    
    cfg := testutils.TestConfigWith(func(c *config.Config) {
        c.DatabaseURL = testDBURL
    })
    
    db := testdb.OpenTestDBWithConfig(t, cfg)
    repo := action_repo.New(db)
    svc := action_service.New(repo, cfg)
    
    // Test implementation
}
```

#### Step 4.2: Files to Refactor
- `internal/services/action_service_impl_test.go`
- `internal/services/plan_service_test.go`
- `internal/services/guardrail_service_test.go`
- Any other service test files

---

### Phase 5: Package Tests Refactoring (Day 8)

#### Step 5.1: Package-Level Pattern
```go
func TestConfig_Load(t *testing.T) {
    t.Parallel()
    
    cfg := testutils.TestConfigWith(func(c *config.Config) {
        c.Environment = "test"
        c.LogLevel = "debug"
    })
    
    // Test implementation
}
```

#### Step 5.2: Files to Refactor
- `pkg/config/config_test.go`
- `pkg/ratelimit/config_test.go`
- Any other package test files

---

### Phase 6: Verification (Day 9-10)

#### Step 6.1: Run Full Test Suite
```bash
# Run tests in parallel mode
go test -race -count=1 ./... 2>&1 | tee test-output.txt

# Measure execution time
time go test -race -count=1 ./internal/... 2>&1
```

#### Step 6.2: Compare Before/After

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Test Duration | ~5 min | ~2 min | 60% faster |
| Parallel Workers | 1 | 8+ | Full parallel |
| Resource Usage | High | Normal | 50% reduction |

#### Step 6.3: Verify No Regression
- All tests pass
- No flaky tests
- Coverage maintained or improved

---

## Testutils.TestConfigWith() Usage Guide

### Basic Usage

```go
// Override single value
cfg := testutils.TestConfigWith(func(c *config.Config) {
    c.DatabaseURL = "postgres://test:test@localhost:5432/botla_test"
})

// Override multiple values
cfg := testutils.TestConfigWith(func(c *config.Config) {
    c.DatabaseURL = testDBURL
    c.JWTSecret = testJWTSecret
    c.RedisURL = testRedisURL
    c.Environment = "test"
})

// Use with testdb
db := testdb.OpenTestDBWithConfig(t, cfg)
```

### Specialized Configs

```go
// RAG-optimized config
cfg := testutils.TestRAGConfig()

// Fast rate limit testing
cfg := testutils.FastRateLimitTestConfig()

// Custom config
cfg := testutils.TestConfigWith(func(c *config.Config) {
    c.MaxWorkers = 10
    c.RateLimitRequests = 1000
    c.CacheTTL = 60 * time.Second
})
```

---

## Progress Tracking

### Daily Checklist

- [ ] Refactor N files with t.Setenv()
- [ ] Add t.Parallel() to N test functions
- [ ] Run tests to verify no regression
- [ ] Update inventory document
- [ ] Document any blockers

### Milestone Reviews

| Milestone | Target Date | Files Refactored | Status |
|-----------|-------------|------------------|--------|
| Phase 1 Complete | Day 1 | Inventory done | ⏳ |
| Phase 2 Complete | Day 4 | Integration tests | ⏳ |
| Phase 3 Complete | Day 5 | RAG tests | ⏳ |
| Phase 4 Complete | Day 7 | Service tests | ⏳ |
| Phase 5 Complete | Day 8 | Package tests | ⏳ |
| Phase 6 Complete | Day 10 | Verification | ⏳ |

---

## Success Criteria

### Functional Requirements
- [ ] Zero `t.Setenv()` calls in test files
- [ ] All tests include `t.Parallel()`
- [ ] All tests pass without regression
- [ ] Test execution time reduced by 50%+

### Non-Functional Requirements
- [ ] Follow existing code patterns
- [ ] Tests remain readable and maintainable
- [ ] Documentation updated with parallel patterns

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Tests fail after refactoring | Schedule delay | Run tests after each file |
| Config override doesn't work | Test broken | Verify config structure first |
| Some tests need real env vars | Cannot parallelize | Document exceptions, create issue |
| Test timing issues | Flaky tests | Add proper synchronization |

---

## Legacy Pattern Exceptions

Some tests may legitimately need `t.Setenv()` for:
1. **External binary paths** - When testing CLI tools
2. **System-level configuration** - When testing OS interactions
3. **Third-party service discovery** - When testing service discovery

**Process for exceptions:**
1. Document the exception in code comment
2. Create issue to address later
3. Add `// TODO: Convert to config override` comment

Example:
```go
// LEGACY: t.Setenv required for CLI binary path testing
// TODO: Refactor to use config-based approach
// See issue: https://github.com/botla/botla/issues/XXX
t.Setenv("PATH", "/usr/local/bin:"+os.Getenv("PATH"))
```

---

## Dependencies

- `pkg/config/config.go` - Configuration structure
- `internal/testutils/config.go` - Test config utilities
- `internal/testdb/testdb.go` - Database test utilities
- `Makefile` - Test commands

---

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Test Utilities Pattern](internal/testutils/AGENTS.md)
- [Database Testing Pattern](internal/testdb/AGENTS.md)
- [Integration Testing Pattern](internal/integration/AGENTS.md)

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Sisyphus | Initial plan |

---

*This plan is part of the comprehensive test improvement initiative. For questions or clarifications, refer to the project documentation or consult with the team lead.*
