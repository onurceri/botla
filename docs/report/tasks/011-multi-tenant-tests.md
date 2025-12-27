# Task 011: Multi-Tenant Isolation Tests

**Priority:** 🔴 Critical (Security)  
**Phase:** 8 - Test Coverage  
**Estimated Time:** 3-4 hours  
**Dependencies:** None  

---

## Problem Statement

Multi-tenant isolation is critical for SaaS security. Need tests verifying:
- User A cannot access User B's chatbots
- User A cannot access User B's sources
- User A cannot access User B's conversations
- Organization isolation across workspaces

---

## Objective

Create comprehensive tests for multi-tenant data isolation.

---

## Tests to Write

### File: `internal/integration/multi_tenant_test.go` (NEW)

```go
package integration

import (
    "net/http"
    "testing"
)

func TestMultiTenant_ChatbotIsolation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    // User A creates chatbot
    tokenA := authToken(t, te.Server.URL, "userA@test.com")
    botA := createChatbot(t, te.Server.URL, tokenA, "User A Bot")
    
    // User B tries to access
    tokenB := authToken(t, te.Server.URL, "userB@test.com")
    
    req, _ := http.NewRequest("GET", 
        te.Server.URL+"/api/v1/chatbots/"+botA, nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ := http.DefaultClient.Do(req)
    
    // Should be 403 or 404
    if resp.StatusCode != http.StatusForbidden && 
       resp.StatusCode != http.StatusNotFound {
        t.Errorf("expected 403/404, got %d", resp.StatusCode)
    }
}

func TestMultiTenant_SourceIsolation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    tokenA := authToken(t, te.Server.URL, "srcA@test.com")
    botA := createChatbot(t, te.Server.URL, tokenA, "Bot A")
    sourceA := createTextSource(t, te.Server.URL, tokenA, botA, "Secret content")
    
    tokenB := authToken(t, te.Server.URL, "srcB@test.com")
    
    // Try to access source directly
    req, _ := http.NewRequest("GET", 
        te.Server.URL+"/api/v1/sources/"+sourceA, nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode != http.StatusForbidden && 
       resp.StatusCode != http.StatusNotFound {
        t.Errorf("source leak: got %d", resp.StatusCode)
    }
    
    // Try to delete source
    req, _ = http.NewRequest("DELETE", 
        te.Server.URL+"/api/v1/sources/"+sourceA, nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ = http.DefaultClient.Do(req)
    if resp.StatusCode != http.StatusForbidden && 
       resp.StatusCode != http.StatusNotFound {
        t.Errorf("could delete other user's source: %d", resp.StatusCode)
    }
}

func TestMultiTenant_ConversationIsolation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    tokenA := authToken(t, te.Server.URL, "convA@test.com")
    botA := createChatbot(t, te.Server.URL, tokenA, "Convo Bot")
    
    // Create conversation via chat
    chatResp := sendChat(t, te.Server.URL, tokenA, botA, "Hello")
    convID := chatResp["conversation_id"].(string)
    
    // User B tries to access conversation
    tokenB := authToken(t, te.Server.URL, "convB@test.com")
    
    req, _ := http.NewRequest("GET", 
        te.Server.URL+"/api/v1/conversations/"+convID, nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode != http.StatusForbidden && 
       resp.StatusCode != http.StatusNotFound {
        t.Errorf("conversation leak: %d", resp.StatusCode)
    }
}

func TestMultiTenant_JobIsolation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    tokenA := authToken(t, te.Server.URL, "jobA@test.com")
    botA := createChatbot(t, te.Server.URL, tokenA, "Job Bot")
    sourceA := createTextSource(t, te.Server.URL, tokenA, botA, "Content")
    
    tokenB := authToken(t, te.Server.URL, "jobB@test.com")
    
    // Try to access job status
    req, _ := http.NewRequest("GET", 
        te.Server.URL+"/api/v1/sources/"+sourceA+"/job", nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode != http.StatusForbidden {
        t.Errorf("job status leak: %d", resp.StatusCode)
    }
    
    // Try to retry job
    req, _ = http.NewRequest("POST", 
        te.Server.URL+"/api/v1/sources/"+sourceA+"/job/retry", nil)
    req.Header.Set("Authorization", "Bearer "+tokenB)
    
    resp, _ = http.DefaultClient.Do(req)
    if resp.StatusCode != http.StatusForbidden {
        t.Errorf("could retry other user's job: %d", resp.StatusCode)
    }
}

func TestMultiTenant_ListingIsolation(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    // User A creates resources
    tokenA := authToken(t, te.Server.URL, "listA@test.com")
    createChatbot(t, te.Server.URL, tokenA, "A's Bot 1")
    createChatbot(t, te.Server.URL, tokenA, "A's Bot 2")
    
    // User B creates resources
    tokenB := authToken(t, te.Server.URL, "listB@test.com")
    createChatbot(t, te.Server.URL, tokenB, "B's Bot")
    
    // User B lists chatbots
    bots := listChatbots(t, te.Server.URL, tokenB)
    
    // Should only see their own
    for _, bot := range bots {
        if name, ok := bot["name"].(string); ok {
            if strings.Contains(name, "A's Bot") {
                t.Error("User B can see User A's chatbots")
            }
        }
    }
}
```

---

## Acceptance Criteria

- [ ] Cannot access other user's chatbots
- [ ] Cannot access other user's sources
- [ ] Cannot access other user's conversations
- [ ] Cannot access other user's job status
- [ ] Listings only show own resources
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `internal/integration/multi_tenant_test.go` | CREATE |
