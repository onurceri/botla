> Extracted from: analytics.md — Section: 3. Test Scenarios

# 3. Test Scenarios

### 3.1 Global & Chatbot Analytics

**Scenario 3.1.1 — Global analytics (workspace scope)**
- [ ] Authenticate as user with workspace membership
- [ ] Set workspace context
- [ ] Send `GET /api/v1/analytics`
- [ ] Verify 200 OK with analytics data scoped to workspace
- [ ] Verify response structure matches `[]AnalyticsPoint`

**Scenario 3.1.2 — Global analytics (org scope)**
- [ ] Authenticate as user with org membership (no workspace context)
- [ ] Send `GET /api/v1/analytics`
- [ ] Verify 200 OK with analytics scoped to org
- [ ] Verify membership validated

**Scenario 3.1.3 — Global analytics — unauthorized for non-member**
- [ ] Send `GET /api/v1/analytics` with workspace/org context user does not belong to
- [ ] Verify 403 Forbidden

**Scenario 3.1.4 — Chatbot analytics overview**
- [ ] Authenticate as owner of a chatbot with activity
- [ ] Send `GET /api/v1/chatbots/{id}/analytics/overview`
- [ ] Verify 200 OK
- [ ] Verify response includes `total_messages`, `total_conversations`, `total_tokens_used`, `positive_feedback`, `negative_feedback`, `feedback_rate`, `handoff_count`
- [ ] Verify `feedback_rate` = `(positive / total) * 100`

**Scenario 3.1.5 — Chatbot analytics overview (no data)**
- [ ] Send for chatbot with zero activity
- [ ] Verify 200 OK with all-zero overview (not null/error)

**Scenario 3.1.6 — Chatbot analytics trends**
- [ ] Send `GET /api/v1/chatbots/{id}/analytics/trends`
- [ ] Verify 200 OK with `daily` array
- [ ] Default to 30 days
- [ ] Verify each daily entry has `date`, `total_messages`, `total_conversations`, `total_tokens_used`, `thumbs_up_count`, `thumbs_down_count`, `handoff_count`, `avg_response_time_ms`

**Scenario 3.1.7 — Trends with custom days parameter**
- [ ] Send `GET /api/v1/chatbots/{id}/analytics/trends?days=7`
- [ ] Verify 7 days of data
- [ ] Send `?days=365`
- [ ] Verify 365 days returned
- [ ] Send `?days=0` — verify defaults to 30
- [ ] Send `?days=999` — verify capped at 365
- [ ] Send `?days=invalid` — verify defaults to 30

**Scenario 3.1.8 — Source usage breakdown**
- [ ] Send `GET /api/v1/chatbots/{id}/analytics/sources`
- [ ] Verify 200 OK with source usage stats
- [ ] Verify days parameter works (same as trends)

**Scenario 3.1.9 — Analytics access denied (wrong owner)**
- [ ] Authenticate as user who does not own the chatbot
- [ ] Send analytics request for that chatbot
- [ ] Verify 403 or 404

**Scenario 3.1.10 — Analytics for non-existent chatbot**
- [ ] Send `GET /api/v1/chatbots/00000000-0000-0000-0000-000000000000/analytics/overview`
- [ ] Verify 404 Not Found

---

