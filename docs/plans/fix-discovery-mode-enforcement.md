# Plan: Fix Discovery Mode Enforcement

## Problem
"Discovery Mode" (finding sub-pages during crawling) is a Pro+ feature. Free plans are limited to `max_pages_per_crawl: 0`. The Frontend disables the selection, but the Backend API accepts `discovery_mode` values other than 'disabled' (e.g., 'auto', 'pending') without validating the plan's crawling limits.

## Analysis
- **File:** `internal/api/handlers/chatbot_item.go`
- **Function:** `updateChatbot`
- **Current State:** Accepts `DiscoveryMode` string directly.
- **Requirement:** Ensure that if `DiscoveryMode` is NOT 'disabled', the plan allows sub-page crawling.

## Proposed Changes

### 1. Update `updateChatbot` Handler
- Retrieve plan using `db.GetPlanByUserID`.
- Check if `req.DiscoveryMode` is provided and is NOT "disabled".
- Check `plan.Config.Scraping.MaxPagesPerCrawl`.
- If `MaxPagesPerCrawl <= 0` (meaning no crawling allowed), reject the request if they try to enable discovery.

### 2. Code Snippet
```go
if req.DiscoveryMode != nil && *req.DiscoveryMode != "disabled" {
    plan, err := db.GetPlanByUserID(r.Context(), h.DB, c.UserID)
    // Assuming MaxPagesPerCrawl > 0 means discovery is allowed
    if err != nil || plan == nil || plan.Config.Scraping.MaxPagesPerCrawl <= 0 {
         w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Discovery mode is not available on your plan",
            "upgrade_required": true,
        })
        return
    }
}
```

## Verification Plan
1.  **Unit Test:** Test setting `discovery_mode: "auto"` on a Free plan user. Expect 403.
2.  **Manual Verification:** API call verification.
