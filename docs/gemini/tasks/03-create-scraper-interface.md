# Task 03: Create Scraper Interface for Testability

**Priority:** 🟢 Low  
**Effort:** Medium (4-6 hours)  
**Risk Level:** Low (internal refactor)

---

## Problem Statement

The `URLProcessor` and related processing logic depend on package-level functions (like `scraper.ScrapeURLWithFallback`) instead of interfaces. This makes unit testing difficult and forces integration tests to use real infrastructure.

### Evidence

From `internal/processing/processor_unit_test.go` (line 68):
```go
// Note: URLProcessor uses scraper.ScrapeURLWithFallback which is package-level and hard to mock without more refactoring.
```

### Why This Matters

1. **Slow CI**: Integration tests with real scraping are slow
2. **Flaky Tests**: Real HTTP calls can fail intermittently
3. **Edge Case Testing**: Hard to simulate scraper failures, timeouts, specific HTML
4. **Coupling**: Processing logic is tightly coupled to scraper implementation

---

## Acceptance Criteria

- [ ] New `Scraper` interface defined in `internal/scraper/`
- [ ] `URLProcessor` accepts `Scraper` interface as a dependency
- [ ] Mock implementation exists for testing
- [ ] Existing functionality unchanged
- [ ] Unit tests use mock scraper
- [ ] All existing tests pass

---

## Implementation Steps

### Step 1: Define Scraper Interface

Create interface in `internal/scraper/interface.go`:

```go
package scraper

import "context"

// ScrapedContent represents the result of scraping a URL
type ScrapedContent struct {
    URL         string
    Title       string
    Content     string
    Links       []string
    StatusCode  int
    ContentType string
}

// Scraper defines the interface for web scraping operations
type Scraper interface {
    // ScrapeURL scrapes a single URL and returns its content
    ScrapeURL(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error)
    
    // ScrapeURLWithFallback tries multiple scraping strategies
    ScrapeURLWithFallback(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error)
}

// ScrapeOptions configures scraping behavior
type ScrapeOptions struct {
    UserAgent       string
    Timeout         time.Duration
    FollowRedirects bool
    MaxDepth        int
    SelectorWhitelist []string
}
```

### Step 2: Create Default Implementation

Wrap existing functions in a struct that implements the interface:

**File:** `internal/scraper/default_scraper.go`

```go
package scraper

import "context"

// DefaultScraper implements Scraper using existing scraping logic
type DefaultScraper struct {
    cache *Cache
}

// NewDefaultScraper creates a new DefaultScraper instance
func NewDefaultScraper() *DefaultScraper {
    return &DefaultScraper{
        cache: NewCache(),
    }
}

// ScrapeURL implements Scraper.ScrapeURL
func (s *DefaultScraper) ScrapeURL(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error) {
    // Delegate to existing implementation
    return scrapeURLInternal(ctx, url, opts, s.cache)
}

// ScrapeURLWithFallback implements Scraper.ScrapeURLWithFallback
func (s *DefaultScraper) ScrapeURLWithFallback(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error) {
    // Delegate to existing implementation
    return scrapeURLWithFallbackInternal(ctx, url, opts, s.cache)
}
```

### Step 3: Create Mock Implementation

**File:** `internal/scraper/mock_scraper.go`

```go
package scraper

import "context"

// MockScraper is a configurable mock for testing
type MockScraper struct {
    // ScrapeURLFunc allows custom behavior per test
    ScrapeURLFunc func(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error)
    
    // Fixed responses for simple cases
    Responses map[string]*ScrapedContent
    Errors    map[string]error
    
    // Call tracking
    Calls []string
}

// NewMockScraper creates a mock with default behavior
func NewMockScraper() *MockScraper {
    return &MockScraper{
        Responses: make(map[string]*ScrapedContent),
        Errors:    make(map[string]error),
        Calls:     make([]string, 0),
    }
}

// ScrapeURL implements Scraper.ScrapeURL
func (m *MockScraper) ScrapeURL(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error) {
    m.Calls = append(m.Calls, url)
    
    if m.ScrapeURLFunc != nil {
        return m.ScrapeURLFunc(ctx, url, opts)
    }
    
    if err, ok := m.Errors[url]; ok {
        return nil, err
    }
    
    if resp, ok := m.Responses[url]; ok {
        return resp, nil
    }
    
    return &ScrapedContent{
        URL:     url,
        Title:   "Mock Title",
        Content: "Mock content for " + url,
    }, nil
}

// ScrapeURLWithFallback implements Scraper.ScrapeURLWithFallback
func (m *MockScraper) ScrapeURLWithFallback(ctx context.Context, url string, opts ScrapeOptions) (*ScrapedContent, error) {
    return m.ScrapeURL(ctx, url, opts)
}

// SetResponse configures a response for a specific URL
func (m *MockScraper) SetResponse(url string, content *ScrapedContent) {
    m.Responses[url] = content
}

// SetError configures an error for a specific URL
func (m *MockScraper) SetError(url string, err error) {
    m.Errors[url] = err
}
```

### Step 4: Update URLProcessor

Modify `internal/processing/processor.go` to accept interface:

**Before:**
```go
type URLProcessor struct {
    db      *sql.DB
    storage storage.StorageService
    // Uses scraper package functions directly
}

func (p *URLProcessor) Process(ctx context.Context, url string) error {
    content, err := scraper.ScrapeURLWithFallback(ctx, url) // Package-level call
    // ...
}
```

**After:**
```go
type URLProcessor struct {
    db      *sql.DB
    storage storage.StorageService
    scraper scraper.Scraper // Interface dependency
}

func NewURLProcessor(db *sql.DB, storage storage.StorageService, s scraper.Scraper) *URLProcessor {
    if s == nil {
        s = scraper.NewDefaultScraper()
    }
    return &URLProcessor{
        db:      db,
        storage: storage,
        scraper: s,
    }
}

func (p *URLProcessor) Process(ctx context.Context, url string) error {
    content, err := p.scraper.ScrapeURLWithFallback(ctx, url, scraper.ScrapeOptions{})
    // ...
}
```

### Step 5: Update Production Wiring

Update `cmd/server/main.go` or wherever `URLProcessor` is created:

```go
scraper := scraper.NewDefaultScraper()
processor := processing.NewURLProcessor(db, storage, scraper)
```

### Step 6: Write Unit Tests with Mock

**File:** `internal/processing/processor_unit_test.go`

```go
func TestURLProcessor_HandlesScraperError(t *testing.T) {
    mock := scraper.NewMockScraper()
    mock.SetError("https://example.com", errors.New("connection refused"))
    
    processor := NewURLProcessor(testDB, nil, mock)
    
    err := processor.Process(context.Background(), "https://example.com")
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "connection refused")
    assert.Equal(t, 1, len(mock.Calls))
}

func TestURLProcessor_ExtractsContent(t *testing.T) {
    mock := scraper.NewMockScraper()
    mock.SetResponse("https://example.com", &scraper.ScrapedContent{
        URL:     "https://example.com",
        Title:   "Test Page",
        Content: "This is the page content for testing.",
    })
    
    processor := NewURLProcessor(testDB, nil, mock)
    
    err := processor.Process(context.Background(), "https://example.com")
    
    assert.NoError(t, err)
    // Assert content was processed correctly
}
```

---

## Testing Checklist

- [ ] `go build ./...` succeeds
- [ ] `make test-no-pdf` passes
- [ ] `make lint` passes
- [ ] Unit tests use mock scraper
- [ ] Integration tests still work with real scraper
- [ ] Edge cases tested (timeout, 404, empty content)

---

## Files to Modify

| File | Change |
|------|--------|
| `internal/scraper/interface.go` | Create new file with interface definition |
| `internal/scraper/default_scraper.go` | Create new file wrapping existing logic |
| `internal/scraper/mock_scraper.go` | Create new file with mock implementation |
| `internal/processing/processor.go` | Accept Scraper interface |
| `internal/processing/processor_unit_test.go` | Use mock scraper |
| `cmd/server/main.go` | Wire up default scraper |

---

## Benefits After Completion

| Aspect | Before | After |
|--------|--------|-------|
| Unit test speed | Slow (real HTTP) | Fast (mock) |
| Test reliability | Flaky | Deterministic |
| Edge case coverage | Limited | Comprehensive |
| Dependency coupling | High | Low |

---

## Related Issues

- Code Audit Finding #3: "Tight Coupling and Poor Testability in Processing Pipelines"
- Test TODO in `processor_unit_test.go`
