> Extracted from: chatbot-management.md — Section 5

# 5. Edge Cases

## Authentication & Authorization
- [ ] No auth cookie → all endpoints return `401`
- [ ] Expired/invalid JWT → all endpoints return `401`
- [ ] Deleted user account → `403` via `DeletedAccountMiddleware`
- [ ] Chatbot owned by another user (no workspace/org overlap) → `403`
- [ ] Access with different HTTP methods on dispatch handlers (`/api/v1/chatbots/{id}` only handles GET/PUT/DELETE; others return `405`)

## Chatbot State
- [ ] Soft-deleted chatbot returns `404` on all endpoints
- [ ] Chatbot with no sources — suggestions job completes with empty questions
- [ ] Chatbot with no workspace or org — accessible only by direct owner

## Request Payload Edge Cases
- [ ] `PUT` with malformed JSON → `400`
- [ ] `PUT` with unexpected fields → fields silently ignored (no strict schema validation)
- [ ] `PUT` with `null` value for a field → field not updated (pointer is nil)
- [ ] `PUT /api/v1/chatbots/{id}/security` with extremely long domain list → `500` or truncation (no explicit limit)
- [ ] `PUT /api/v1/chatbots/{id}/guardrails` with `topic_restrictions: null` → clears existing restrictions
- [ ] `POST /api/v1/chatbots` with all optional fields provided → all applied correctly

## Plan Edge Cases
- [ ] User with no plan record → plan fetch fails → `500` or feature returns upgrade-required
- [ ] Plan with zero limits → all features effectively disabled
- [ ] Downgrade plan while chatbot has restricted features enabled → features remain set but cannot be changed
- [ ] Custom branding on non-Enterprise plan → `403` with `upgrade_required: true`

## Suggestion Job Edge Cases
- [ ] No chatbot exist for job status → `404`
- [ ] Multiple rapid regenerate calls → each creates a new job (no dedup)
- [ ] Suggestion job from a previous run persists → status returns latest job

## Database & Concurrency
- [ ] Concurrent create requests near plan limit → race condition possible (no atomic count+create)
- [ ] Concurrent delete + update → update operates on non-deleted chatbot, succeeds
- [ ] Database connection failure → `500` on all endpoints
