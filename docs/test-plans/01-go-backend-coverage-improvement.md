# Test Plan: Go Backend Coverage Improvement

**Plan ID:** TP-GO-COVERAGE-001  
**Priority:** CRITICAL  
**Estimated Duration:** 2-3 weeks  
**Target Coverage:** 63.3% → 90%+  
**Status:** Draft  

---

## Executive Summary

This plan addresses the critical coverage gap in the Go backend test suite. Current overall coverage is 63.3%, falling significantly short of the required 90% gate. This plan identifies all low-coverage files, provides a systematic approach to adding tests, and ensures all critical business logic is properly tested with mocks and stubs.

---

## Sisyphus Agent Prompt

```
You are Sisyphus, a senior Go engineer with expertise in testing, test-driven development, and backend architecture.

### Task Context
The Botla backend Go codebase currently has 63.3% test coverage, but requires 90%+ coverage to pass CI gates. Multiple critical files have 0% coverage including:
- cmd/server/main.go (server initialization)
- cmd/cli/main.go (CLI entrypoint)
- pkg/storage/r2.go (R2 storage operations)
- pkg/tokenizer/loader.go (tokenizer loading)

Several other packages also have coverage significantly below 90%:
- internal/scraper/browser.go
- internal/workers/pool.go
- pkg/storage/*.go files

### Your Mission
1. First, run `make cover-func` to get exact coverage numbers per function/file
2. Analyze the coverage report to identify ALL files below 90% coverage
3. For each file below threshold:
   - Understand what the code does (read the source)
   - Identify testable units (functions, methods)
   - Create appropriate unit tests using existing patterns
   - Use mocks from internal/rag/mocks.go, internal/scraper/mock_scraper.go, or create inline mocks as needed
4. Ensure all new tests follow the project's testing conventions:
   - Use `testdb.OpenTestDB()` for DB-dependent tests
   - Use `testutils.TestConfigWith()` for config-dependent tests (NOT t.Setenv)
   - Use existing mocks from internal/integration/fixtures/mocks.go
   - Follow naming convention: `*_unit_test.go` for unit tests
   - Add `t.Parallel()` where appropriate
5. Run `make test` after each batch of changes to verify tests pass
6. Verify coverage improvement with `make cover-func`

### Critical Rules
- NEVER use `t.Setenv()` - use `testutils.TestConfigWith()` instead
- For files with 0% coverage, start with the most critical functionality first
- Use the existing mock infrastructure (RAG mocks, scraper mocks, etc.)
- If a test requires significant setup, create a fixture in internal/testdb/fixtures.go
- Report progress every 10 files completed

### Deliverables
- All test files created/modified
- Coverage report showing 90%+ for all packages
- No failing tests
- All code follows existing patterns

Begin by running `make cover-func` to get the current state.
```

---

## Current Coverage Analysis

### Critical Zero-Coverage Files

| File | Coverage | Priority | Estimated Tests Needed |
|------|----------|----------|------------------------|
| `cmd/server/main.go` | 0% | CRITICAL | 5-8 |
| `cmd/cli/main.go` | 0% | CRITICAL | 5-8 |
| `pkg/storage/r2.go` | 0% | HIGH | 8-12 |
| `pkg/tokenizer/loader.go` | 0% | HIGH | 5-7 |

### High-Priority Low-Coverage Files

| File | Coverage | Priority | Estimated Tests Needed |
|------|----------|----------|------------------------|
| `internal/scraper/browser.go` | <50% | HIGH | 10-15 |
| `internal/workers/pool.go` | <60% | HIGH | 8-12 |
| `pkg/storage/memory.go` | <70% | MEDIUM | 5-7 |
| `pkg/storage/s3.go` | <70% | MEDIUM | 8-10 |

---

## Step-by-Step Implementation Plan

### Phase 1: Coverage Analysis and Baseline (Day 1)

#### Step 1.1: Run Coverage Report
```bash
# Get current coverage state
make cover-func > docs/test-plans/coverage-baseline.txt

# Analyze which packages are below 90%
cat docs/test-plans/coverage-baseline.txt
```

#### Step 1.2: Create Coverage Gap Matrix
Create a spreadsheet/markdown file listing:
- Every file below 90% coverage
- Each function/method in that file
- Whether it's testable in isolation or requires integration setup
- Estimated number of tests needed

#### Step 1.3: Prioritize by Business Impact
Rank files by:
1. Business criticality (user-facing features)
2. Complexity (more complex = more tests needed)
3. Dependencies (files that block other tests)

---

### Phase 2: Zero-Coverage Critical Files (Days 2-5)

#### Step 2.1: cmd/server/main.go Tests

**Target:** 90%+ coverage for server initialization

**Test Scenarios:**
```go
// server_main_unit_test.go

func TestServerMain_Initialization(t *testing.T) {
    // Test config loading
    // Test service initialization order
    // Test graceful shutdown
    // Test panic recovery
    // Test health check endpoint
}

func TestServerMain_ConfigOverride(t *testing.T) {
    // Test environment variable overrides
    // Test config validation
    // Test missing required config error
}

func TestServerMain_SignalHandling(t *testing.T) {
    // Test SIGINT handling
    // Test SIGTERM handling
    // Test graceful shutdown sequence
}
```

**Setup Requirements:**
- Use `testutils.TestConfigWith()` for config overrides
- Mock external dependencies (database, redis, qdrant)
- Use `internal/integration/fixtures/mocks.go` for RAG mocks

#### Step 2.2: cmd/cli/main.go Tests

**Target:** 90%+ coverage for CLI entrypoint

**Test Scenarios:**
```go
// cli_main_unit_test.go

func TestCLIMain_CommandRouting(t *testing.T) {
    // Test root command
    // Test serve command
    // Test migrate command
    // Test version command
}

func TestCLIMain_ArgumentParsing(t *testing.T) {
    // Test valid arguments
    // Test missing required arguments
    // Test invalid flags
}

func TestCLIMain_EnvironmentValidation(t *testing.T) {
    // Test required env vars
    // Test env var format validation
}
```

#### Step 2.3: pkg/storage/r2.go Tests

**Target:** 90%+ coverage for R2 storage operations

**Test Scenarios:**
```go
// storage_r2_unit_test.go

func TestR2Storage_Upload(t *testing.T) {
    // Test successful upload
    // Test upload with progress callback
    // Test upload cancellation
    // Test upload error handling
}

func TestR2Storage_Download(t *testing.T) {
    // Test successful download
    // Test download with range requests
    // Test download not found
    // Test download error handling
}

func TestR2Storage_Delete(t *testing.T) {
    // Test successful delete
    // Test delete non-existent file
    // Test delete error handling
}

func TestR2Storage_List(t *testing.T) {
    // Test list all files
    // Test list with prefix
    // Test list pagination
    // Test list empty result
}
```

**Mock Requirements:**
- Create `MockR2Client` implementing `pkg/storage.R2Client` interface
- Use interface-based testing to allow mock injection
- Reference `internal/rag/mocks.go` for mock patterns

#### Step 2.4: pkg/tokenizer/loader.go Tests

**Target:** 90%+ coverage for tokenizer loading

**Test Scenarios:**
```go
// tokenizer_loader_unit_test.go

func TestTokenizerLoader_Load(t *testing.T) {
    // Test successful load from file
    // Test load from embedded data
    // Test load error handling
    // Test cached tokenizer reuse
}

func TestTokenizerLoader_Reload(t *testing.T) {
    // Test hot reload functionality
    // Test reload error handling
    // Test reload with new tokenizer
}
```

---

### Phase 3: High-Priority Low-Coverage Files (Days 6-10)

#### Step 3.1: internal/scraper/browser.go Tests

**Target:** 85%+ coverage for browser scraper

**Test Scenarios:**
```go
// scraper_browser_unit_test.go

func TestBrowserScraper_Initialize(t *testing.T) {
    // Test browser launch
    // Test browser context creation
    // Test browser options configuration
}

func TestBrowserScraper_Navigate(t *testing.T) {
    // Test successful navigation
    // Test navigation timeout
    // Test navigation error handling
    // Test page load waiting
}

func BrowserScraper_Extract(t *testing.T) {
    // Test content extraction
    // Test JavaScript execution
    // Test dynamic content loading
    // Test element selection
}

func TestBrowserScraper_Close(t *testing.T) {
    // Test graceful close
    // Test resource cleanup
    // Test close error handling
}
```

**Mock Requirements:**
- Use `internal/scraper/mock_scraper.go` as reference
- Mock browser automation library (Playwright/CDP)
- Test both success and error paths

#### Step 3.2: internal/workers/pool.go Tests

**Target:** 85%+ coverage for worker pool

**Test Scenarios:**
```go
// workers_pool_unit_test.go

func TestWorkerPool_Submit(t *testing.T) {
    // Test successful task submission
    // Test task queue overflow
    // Test concurrent submissions
}

func TestWorkerPool_WorkerLifecycle(t *testing.T) {
    // Test worker startup
    // Test worker idle timeout
    // Test worker panic recovery
    // Test worker shutdown
}

func TestWorkerPool_TaskExecution(t *testing.T) {
    // Test task success
    // Test task failure
    // Test task timeout
    // Test task retry
}

func TestWorkerPool_Scaling(t *testing.T) {
    // Test auto-scale up
    // Test auto-scale down
    // Test scale limits
}
```

---

### Phase 4: Medium-Priority Files (Days 11-15)

#### Step 4.1: Remaining pkg/storage Tests
- `pkg/storage/s3.go`
- `pkg/storage/gcs.go`
- `pkg/storage/azure.go`

#### Step 4.2: Remaining internal/* Tests
- Complete coverage for all handlers
- Complete coverage for middleware
- Complete coverage for validators

#### Step 4.3: Package-Level Tests
- `pkg/errors/*.go`
- `pkg/httputil/*.go`
- `pkg/middleware/*.go`
- `pkg/policy/*.go`

---

### Phase 5: Verification and Gate (Days 16-18)

#### Step 5.1: Run Full Test Suite
```bash
# Run all tests with coverage
make test-all

# Verify coverage gate
make cover-gate
```

#### Step 5.2: Fix Any Failures
- Address any test failures immediately
- Fix coverage regressions
- Ensure all tests pass

#### Step 5.3: Final Verification
```bash
# Final coverage report
make cover-func

# Save final report
make cover-func > docs/test-plans/coverage-final.txt
```

---

## Test Creation Guidelines

### Naming Conventions

```go
// Unit tests: *_unit_test.go
// Integration tests: *_test.go in internal/integration/
// Table-driven tests: Use test tables

func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Use table-driven tests for multiple scenarios
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"scenario1", input1, expected1, false},
        {"scenario2", input2, expected2, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Mock Usage Pattern

```go
// Use existing mocks when available
mockLLM := rag.NewMockLLMClient(t)
mockLLM.On("Complete", mock.Anything).Return(response, nil)

// Create inline mocks for repository interfaces
type mockActionRepository struct {
    mock.Mock
}

func (m *mockActionRepository) Create(ctx context, action *models.Action) error {
    args := m.Called(ctx, action)
    return args.Error(0)
}

// Use testutils.TestConfigWith for config
cfg := testutils.TestConfigWith(func(c *config.Config) {
    c.MaxWorkers = 10
})
```

### Database Test Pattern

```go
func TestActionService_CRUD(t *testing.T) {
    t.Parallel()
    
    db := testdb.OpenTestDB(t)
    t.Cleanup(func() { db.Close() })
    
    repo := action_repo.New(db)
    mockLog := scraper.NewMockActionLogStore(t)
    
    svc := action_service.New(repo, mockLog)
    
    // Test cases
    tests := []struct {
        name string
        test func(t *testing.T)
    }{
        {"Create", func(t *testing.T) {
            // Create test implementation
        }},
        {"GetByID", func(t *testing.T) {
            // GetByID test implementation
        }},
    }
    
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            tt.test(t)
        })
    }
}
```

### Error Testing Pattern

```go
func TestService_Errors(t *testing.T) {
    tests := []struct {
        name       string
        setup      func(t *testing.T) *Service
        operation  func(t *testing.T, s *Service) error
        wantErr    error
    }{
        {
            name: "not found error",
            setup: func(t *testing.T) *Service {
                repo := NewMockRepository(t)
                repo.On("GetByID", "nonexistent").Return(nil, errors.New("not found"))
                return NewService(repo)
            },
            operation: func(t *testing.T, s *Service) error {
                return s.Delete("nonexistent")
            },
            wantErr: ErrNotFound,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := tt.setup(t)
            err := tt.operation(t, svc)
            assert.ErrorIs(t, err, tt.wantErr)
        })
    }
}
```

---

## Progress Tracking

### Daily Checklist

- [ ] Run coverage report
- [ ] Create N new test files
- [ ] Update N existing test files
- [ ] Verify tests pass
- [ ] Update coverage matrix
- [ ] Document any blockers

### Milestone Reviews

| Milestone | Target Date | Coverage Target | Status |
|-----------|-------------|-----------------|--------|
| Phase 1 Complete | Day 1 | Baseline established | ⏳ |
| Phase 2 Complete | Day 5 | Critical files 90%+ | ⏳ |
| Phase 3 Complete | Day 10 | High-priority 85%+ | ⏳ |
| Phase 4 Complete | Day 15 | All files 80%+ | ⏳ |
| Final Gate | Day 18 | All files 90%+ | ⏳ |

---

## Success Criteria

### Functional Requirements
- [ ] All files have 90%+ coverage
- [ ] All tests pass (no skipped tests)
- [ ] No test flakiness (tests pass consistently)
- [ ] Tests run in reasonable time (< 5 minutes for full suite)

### Non-Functional Requirements
- [ ] Follow existing code patterns
- [ ] Use appropriate mocks and stubs
- [ ] Well-documented test cases
- [ ] Maintainable test code

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Tests take too long | Schedule delay | Run in parallel, prioritize critical paths |
| Mocking complexity | Development slow | Use existing mocks, create shared mock factories |
| Coverage gaming | Poor test quality | Review test quality, require assertions not just calls |
| Test maintenance burden | Long-term debt | Document patterns, refactor tests regularly |

---

## Dependencies

- `internal/testdb` - Database test utilities
- `internal/testutils` - Test configuration utilities
- `internal/rag/mocks.go` - RAG layer mocks
- `internal/scraper/mock_scraper.go` - Scraper mocks
- `internal/integration/fixtures/mocks.go` - Integration test mocks

---

## References

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Test-Driven Development by Example - Kent Beck](https://books.google.com/books?id=KfRYBwAAQBAJ)
- [Google Testing Blog](https://testing.googleblog.com/)
- Project AGENTS.md - Testing conventions
- Project CLAUDE.md - Development guidelines

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Sisyphus | Initial plan |

---

*This plan is part of the comprehensive test improvement initiative. For questions or clarifications, refer to the project documentation or consult with the team lead.*
