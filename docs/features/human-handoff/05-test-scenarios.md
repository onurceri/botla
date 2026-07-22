> Extracted from: human-handoff.md — Section: 3. Test Scenarios

# 3. Test Scenarios

## Scenario 26: Human Handoff — Request Flow

- [ ] POST `/api/v1/public/chatbots/{id}/handoff` with `{"sessionID": "handoff-test-1"}`
- [ ] Verify 201 response with `handoffRequestID` (UUID)
- [ ] Verify a HandoffRequest record is created with status `pending`
- [ ] Verify a Message with role=system is created recording the handoff event
- [ ] Verify subsequent chat requests for the same session include `handoffRequestID` in responses

## Scenario 27: Human Handoff — Submit Contact

- [ ] From Scenario 26, POST `/api/v1/public/chatbots/{id}/handoff/{requestId}/contact` with `{"email": "user@example.com"}`
- [ ] Verify 200 response
- [ ] Verify status is now `contacted`
- [ ] Verify `ContactedAt` is set
- [ ] Verify `ContactEmail` matches the submitted email
- [ ] Submit with invalid email → 400
- [ ] Submit for non-existent requestId → 404
- [ ] Submit for already-resolved request → 409

## Scenario 28: Human Handoff — Management List

- [ ] GET `/api/v1/chatbots/{id}/handoff-requests`
- [ ] Verify array of handoff requests with all fields
- [ ] Verify pagination works
- [ ] Verify filtering by status (e.g., `?status=pending`)
- [ ] Verify sorting (most recent first)
- [ ] Without auth → 401
- [ ] Wrong org → 403

## Scenario 29: Human Handoff — Get Detail

- [ ] GET `/api/v1/chatbots/{id}/handoff-requests/{requestId}`
- [ ] Verify response includes all model fields: ID, ChatbotID, ConversationID, MessageID, Status, ContactEmail, ContactedAt, ResolvedAt, Notes
- [ ] Non-existent requestId → 404
- [ ] Without auth → 401

## Scenario 30: Human Handoff — Status Transitions

- [ ] Create handoff request → status is `pending`
- [ ] PATCH to update status to `contacted` (with `{"status": "contacted"}`) → 200
- [ ] PATCH to update status to `resolved` (with `{"status": "resolved", "notes": "Issue resolved via email"}`) → 200
- [ ] Verify `ResolvedAt` is set
- [ ] Verify `Notes` is stored
- [ ] PATCH from `pending` directly to `resolved` → 422 (invalid transition)
- [ ] PATCH from `resolved` back to `contacted` → 422
- [ ] PATCH with invalid status value → 400
- [ ] PATCH non-existent request → 404
- [ ] PATCH without auth → 401
- [ ] PATCH from wrong org → 403

## Scenario 31: Full Handoff Lifecycle

- [ ] User sends chat messages in a session
- [ ] User requests handoff (via handoff endpoint or escalate fallback)
- [ ] Handoff request is created (status: pending)
- [ ] Chat response now includes `handoffRequestID`
- [ ] User submits contact email (status: contacted)
- [ ] Agent views handoff in dashboard (GET list)
- [ ] Agent views handoff detail (GET by ID)
- [ ] Agent updates status to `resolved` with notes
- [ ] Verify full timeline: ContactedAt, ResolvedAt
- [ ] Verify conversation messages remain intact and linked to the handoff
