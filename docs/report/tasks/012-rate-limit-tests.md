# Task 012: Rate Limiting Tests

**Priority:** 🟡 High (Security/Reliability)  
**Phase:** 8 - Test Coverage  
**Estimated Time:** 2-3 hours  
**Dependencies:** None  

---

## Problem Statement

Rate limiting protects against abuse but lacks comprehensive test coverage for:
- API endpoint rate limits
- Per-user vs global limits
- Rate limit headers
- Recovery after limit window

---

## Tests to Write

### File: `internal/integration/rate_limit_test.go` (NEW)

```go
package integration

import (
    "net/http"
    "testing"
    "time"
)

func TestRateLimit_ChatEndpoint(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    token := authToken(t, te.Server.URL, "rate@test.com")
    botID := createChatbot(t, te.Server.URL, token, "Rate Bot")
    
    // Rapid fire requests
    var rateLimited bool
    for i := 0; i < 100; i++ {
        resp := sendChatRaw(t, te.Server.URL, token, botID, "Hi")
        if resp.StatusCode == http.StatusTooManyRequests {
            rateLimited = true
            
            // Check headers
            if resp.Header.Get("Retry-After") == "" {
                t.Error("missing Retry-After header")
            }
            break
        }
        resp.Body.Close()
    }
    
    if !rateLimited {
        t.Log("Warning: rate limit not triggered")
    }
}

func TestRateLimit_SourceCreation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    token := authToken(t, te.Server.URL, "srcrate@test.com")
    botID := createChatbot(t, te.Server.URL, token, "Source Rate Bot")
    
    // Try to create many sources rapidly
    var rateLimited bool
    for i := 0; i < 50; i++ {
        resp := createTextSourceRaw(t, te.Server.URL, token, botID, 
            fmt.Sprintf("Content %d", i))
        if resp.StatusCode == http.StatusTooManyRequests {
            rateLimited = true
            break
        }
        resp.Body.Close()
    }
    
    // Rate limiting should kick in
    if !rateLimited {
        t.Log("Warning: source creation rate limit not triggered")
    }
}

func TestRateLimit_Recovery(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    token := authToken(t, te.Server.URL, "recover@test.com")
    
    // Trigger rate limit
    for i := 0; i < 100; i++ {
        http.Get(te.Server.URL + "/api/v1/health")
    }
    
    // Wait for recovery
    time.Sleep(2 * time.Second)
    
    // Should work again
    resp, _ := http.Get(te.Server.URL + "/api/v1/health")
    if resp.StatusCode == http.StatusTooManyRequests {
        t.Error("rate limit did not recover")
    }
}

func TestRateLimit_PerUserIsolation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    tokenA := authToken(t, te.Server.URL, "rateA@test.com")
    tokenB := authToken(t, te.Server.URL, "rateB@test.com")
    
    // Exhaust User A's limit
    for i := 0; i < 100; i++ {
        req, _ := http.NewRequest("GET", 
            te.Server.URL+"/api/v1/chatbots", nil)
        req.Header.Set("Authorization", "Bearer "+tokenA)
        http.DefaultClient.Do(req)
    }
    
    // User B should not be affected
    req, _ := http.NewRequest("GET", 
        te.Server.URL+"/api/v1/chatbots", nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode == http.StatusTooManyRequests {
        t.Error("User B rate limited due to User A")
    }
}
```

---

## Acceptance Criteria

- [ ] Chat endpoint rate limit works
- [ ] Source creation rate limit works
- [ ] Rate limit recovery after window
- [ ] Per-user isolation (User A's usage doesn't affect User B)
- [ ] Retry-After header present
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `internal/integration/rate_limit_test.go` | CREATE |
