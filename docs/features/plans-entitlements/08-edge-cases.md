> Extracted from: plans-entitlements.md — Section: 4. Edge Cases

# 4. Edge Cases

### 4.1 Plans & Entitlements

- [ ] **Missing cache (cold start):** First-ever plan request hits DB; response cached after fetch
- [ ] **Redis connection drop during cache read:** Falls back to in-memory cache; if also expired, falls through to DB
- [ ] **Redis connection drop during cache write:** Logs error, continues (non-fatal)
- [ ] **Plan with `Price = 0`:** Free plan should display correctly
- [ ] **Plan with `TrialDays = 0`:** No trial period (standard plans)
- [ ] **Billing cycle variations:** Plans with `monthly` vs `yearly` billing
- [ ] **Currency variations:** Plans priced in different currencies (TRY/USD)
- [ ] **Concurrent plan limit updates:** Two admins updating same plan simultaneously
- [ ] **Token quota at exactly the limit:** User with 100,000/100,000 tokens — reservation returns quota exceeded
- [ ] **Token quota overflow safety:** Reserving > `MaxMonthlyTokens` when usage is 0
- [ ] **Chatbot creation race condition:** Two concurrent requests creating chatbots when `MaxChatbots` reached — only one should succeed
- [ ] **File upload exactly at size limit:** 5MB file on Free (FilesMaxSizeMB=5) — should succeed
- [ ] **File upload 1 byte over limit:** Should fail
- [ ] **Rate limit exactly at boundary:** `RateLimitsRequestsPerMinute` requests in first 59 seconds, then 1 more at second 60 — should count as next window
- [ ] **Iyzico payment timeout:** Payment gateway unresponsive — graceful failure
- [ ] **Plan assigned to user deleted from plans table:** Orphaned reference — `getPlanInfo` returns plan not found error

---

