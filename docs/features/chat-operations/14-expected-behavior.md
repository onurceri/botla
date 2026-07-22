> Extracted from: chat-operations.md — Section: 5. Expected Behavior

# 5. Expected Behavior

## 5.1 Success Conditions

| Scenario | Expected Result |
|----------|----------------|
| Valid chat request with matching content | 200 with markdown response, sources, tokensUsed |
| Valid chat request continued session | 200 with same conversationID, history preserved |
| New session | 201/200 with new conversationID |
| Public chat on public chatbot | 200, full response, UserID is null |
| Feedback submission | 200, Message.feedback updated |
| Action CRUD | 201/200/204 as appropriate, DB persisted |
| Action execution | Action invoked server-side, result in response |
| Handoff request | 201 with handoffRequestID, status=pending |
| Contact email submission | 200, status=contacted, ContactedAt set |
| Handoff management (list/get/update) | 200, correct data returned/updated |
| Valid status transition | 200, status updated, timestamps set |

## 5.2 Failure Conditions

| Scenario | Expected Result |
|----------|----------------|
| No auth token | 401, no data returned |
| Wrong org / no access | 403, no data returned |
| Non-existent chatbot/action/handoff | 404, descriptive error message |
| Validation failure (missing/invalid fields) | 400 or 422, field-level error description |
| Rate limit exceeded | 429, Retry-After header, rate limit message |
| Disabled/paused chatbot | 403 or 423, informative message |
| Invalid status transition (handoff) | 422, allowed transitions listed |
| Concurrent modification conflict | 409 or last-write-wins with appropriate handling |
| Circuit breaker open | 503, service temporarily unavailable |
| Internal server error | 500, error logged, generic message to client |
| Duplicate action name | 422 or 409, duplicate name error |
| Action on private/non-public chatbot via public endpoint | 404 (information hiding) |

## 5.3 Idempotency

- [ ] Repeated feedback with same messageID and score → same result, no duplicate
- [ ] Repeated handoff request from same session → returns existing requestId
- [ ] Repeated contact submission with same email → idempotent, no error
- [ ] Repeated action DELETE → second call returns 404 (already deleted)
- [ ] Repeated action PUT with same values → 200, no side effects
