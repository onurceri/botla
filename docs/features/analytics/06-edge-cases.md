> Extracted from: analytics.md — Section: 4. Edge Cases

# 4. Edge Cases

### 4.1 Analytics

- [ ] **Zero analytics data:** New chatbot with no messages returns all-zero overview, not null
- [ ] **Large date range (365 days):** Trends query for max window
- [ ] **Date range beyond existing data:** Gaps in daily entries — missing dates excluded from response (no zero-fill)
- [ ] **Malformed chatbot ID:** Returns 404 via route dispatch
- [ ] **Concurrent analytics writes:** Two messages processed at same time — no data corruption expected (aggregate per day)
- [ ] **Feedback rate division by zero:** `TotalConversations = 0` — `feedback_rate` should be 0, not NaN
- [ ] **Extremely high token counts:** Values exceeding int range for `TotalTokensUsed` — capped at INT max

---

