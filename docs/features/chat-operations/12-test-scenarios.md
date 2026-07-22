> Extracted from: chat-operations.md — Section: 3. Test Scenarios

# 3. Test Scenarios

## Scenario 1: Basic Authenticated Chat Flow

**Preconditions:** Valid JWT token, existing public chatbot with trained data, sessionID is new.

- [ ] Send POST to `/api/v1/chatbots/{id}/chat` with `{"message": "What is your return policy?", "sessionID": "test-session-1"}`
- [ ] Verify response status is 200
- [ ] Verify response body contains `response` (string, markdown-formatted)
- [ ] Verify response body contains `messageID` (valid UUID v4)
- [ ] Verify response body contains `conversationID` (valid UUID v4)
- [ ] Verify response body contains `sources` (array, possibly empty)
- [ ] Verify response body contains `tokensUsed` (positive integer)
- [ ] Verify response body contains `handoffRequestID` (null or string UUID)
- [ ] Verify a Conversation was created in the database with the correct sessionID and chatbotID
- [ ] Verify a Message was created with role=user
- [ ] Verify a Message was created with role=assistant

## Scenario 2: Conversation Continuity (Same Session)

**Preconditions:** Existing conversation from Scenario 1.

- [ ] Send POST to `/api/v1/chatbots/{id}/chat` with `{"message": "Can you tell me more?", "sessionID": "test-session-1"}`
- [ ] Verify response status is 200
- [ ] Verify `conversationID` matches the one from Scenario 1
- [ ] Verify `messageID` is new (different from Scenario 1)
- [ ] Verify the conversation now has 4 messages (2 user, 2 assistant)
- [ ] Verify `LastMessageAt` on the Conversation was updated

## Scenario 3: New Session Creates New Conversation

**Preconditions:** Same chatbot as Scenario 1.

- [ ] Send POST to `/api/v1/chatbots/{id}/chat` with `{"message": "Hello!", "sessionID": "test-session-2"}`
- [ ] Verify response status is 200
- [ ] Verify `conversationID` is different from Scenario 1
- [ ] Verify a new Conversation record was created

## Scenario 4: Public Chat (Unauthenticated)

**Preconditions:** Chatbot has `isPublic: true`.

- [ ] Send POST to `/api/v1/public/chatbots/{id}/chat` with `{"message": "Hi there", "sessionID": "pub-session-1"}` (no auth headers)
- [ ] Verify response status is 200
- [ ] Verify response format matches authenticated chat (messageID, conversationID, response, sources, tokensUsed)
- [ ] Verify Conversation has `UserID = null`
- [ ] Send a follow-up with same sessionID and verify conversation continuity

## Scenario 5: Public Chat on Non-Public Chatbot

- [ ] Send POST to `/api/v1/public/chatbots/{id}/chat` where chatbot `isPublic: false`
- [ ] Verify response status is 404
- [ ] Verify no information leak (response body does not indicate the chatbot exists)

## Scenario 6: Public Config Endpoint

- [ ] Send GET to `/api/v1/public/chatbots/{id}/config` for a public chatbot
- [ ] Verify response status is 200
- [ ] Verify appearance fields: `theme`, `primaryColor`, `position`, `title`, `description`
- [ ] Verify `suggestions` is an array of strings
- [ ] Verify `branding` contains `logoUrl`, `companyName`, `hideBranding`
- [ ] Verify no sensitive fields are exposed
- [ ] Verify Cache-Control header is present
- [ ] Repeat for non-public chatbot, expect 404

## Scenario 7: Chat Feedback

- [ ] Send POST to `/api/v1/public/chatbots/{id}/feedback` with `{"messageID": "<valid-uuid>", "score": true}`
- [ ] Verify 200 response
- [ ] Verify `feedback` field on the Message is now `true`
- [ ] Send same request with `"score": false`
- [ ] Verify `feedback` is now `false` (overwrites)
- [ ] Send with invalid UUID → 404
- [ ] Send with missing `score` → 400
- [ ] Send with non-boolean `score` → 400

## Scenario 8: Message Validation

- [ ] Send request with empty message → 400
- [ ] Send request with message > 4000 chars → 422 or 400
- [ ] Send request with exactly 4000 chars → 200 (boundary)
- [ ] Send request with missing `message` field → 400
- [ ] Send request with extra unknown fields → 200 (ignored, not rejected)
- [ ] Send request with non-string `message` → 400

## Scenario 9: Authentication & Authorization

- [ ] Send POST to `/api/v1/chatbots/{id}/chat` without JWT → 401
- [ ] Send with expired JWT → 401
- [ ] Send with malformed JWT → 401
- [ ] Send with valid JWT but chatbot belongs to different org → 403
- [ ] Send with valid JWT but chatbot is disabled → 403 or 423
- [ ] Send with valid JWT and valid chatbot → 200

## Scenario 10: Source Citations

**Preconditions:** Chatbot has ingested content that is relevant to the query.

- [ ] Send a question that matches trained content (e.g., "What is the price of plan X?")
- [ ] Verify `sources` array is non-empty
- [ ] Each source object contains required fields (title, url, content snippet)
- [ ] Sources are rendered in the assistant response (cited via markdown links)
- [ ] Verify source relevance scores are above the high threshold (≥ 0.50)
- [ ] Send a question with no relevant content → empty or low-relevance sources

## Scenario 11: Markdown Rendering in Responses

- [ ] Send a query expected to produce a formatted response
- [ ] Verify response string contains valid markdown (headings, lists, bold, links, code blocks)
- [ ] Verify code blocks use proper markdown fencing (```language)
- [ ] Verify tables are rendered as markdown tables
- [ ] Verify markdown is not double-escaped

## Scenario 12: RAG Tiered Search

- [ ] Send query that matches content with high relevance (score ≥ 0.50) → sources populated
- [ ] Send query with partial match (score 0.30–0.49) → medium-relevance sources included
- [ ] Send query with weak match (score < 0.30) → sources may be empty or low relevance
- [ ] Verify `RAG_TOPK` limit is respected (no more than K chunks returned)
- [ ] Verify context window does not exceed `RAG_MAX_CONTEXT_TOKENS`

## Scenario 13: Fallback Strategies

- [ ] Configure chatbot with **static fallback** and send an off-topic query
- [ ] Verify response is the configured static message
- [ ] Configure chatbot with **smart fallback** and send an off-topic query
- [ ] Verify LLM generates a response from general knowledge
- [ ] Configure chatbot with **escalate fallback** and send an off-topic query
- [ ] Verify response suggests handoff or includes handoff prompt
- [ ] Verify fallback type is logged appropriately

## Scenario 14: Guardrails — Low Confidence

- [ ] Send a query that produces a low-confidence response (< confidence threshold)
- [ ] Verify response is blocked or flagged
- [ ] Verify configured refusal message is returned (if applicable)
- [ ] Verify the incident is logged

## Scenario 15: Guardrails — Topic Restrictions

- [ ] Configure a topic restriction (e.g., block "pricing" questions)
- [ ] Send a query about the restricted topic
- [ ] Verify refusal message is returned
- [ ] Verify no RAG results or LLM response about the restricted topic is leaked
- [ ] Send a query about an allowed topic → normal response

## Scenario 16: Circuit Breaker Protection

**Preconditions:** LLM provider is configured to fail (e.g., invalid API key or network block).

- [ ] Send multiple chat requests until the 50% failure threshold is exceeded
- [ ] Verify subsequent requests return a 503 or fallback error response
- [ ] Verify circuit breaker open state is logged
- [ ] Wait for the 30s timeout and verify requests resume (circuit half-open/closed)
- [ ] Verify normal behavior restores after the timeout if provider recovers

## Scenario 32: Rate Limiting

- [ ] Send rapid successive requests to the chat endpoint
- [ ] Verify after N requests (configurable limit), subsequent requests return 429
- [ ] Verify `Retry-After` header is set
- [ ] Verify rate limit resets after the window expires
- [ ] Verify rate limiting is per-session and/or per-IP
- [ ] Repeat for handoff endpoint
- [ ] Repeat for action endpoints

## Scenario 33: Performance — Response Time

- [ ] Send chat request and measure response time
- [ ] Verify p50 response time < 2s
- [ ] Verify p95 response time < 5s
- [ ] Verify p99 response time < 10s
- [ ] Test with large context (many prior messages)
- [ ] Test with many retrieved chunks (RAG_TOPK at maximum)

## Scenario 34: Concurrent Sessions

- [ ] Open multiple sessions against the same chatbot simultaneously
- [ ] Send messages to each session concurrently
- [ ] Verify each conversation is isolated (no cross-session data leakage)
- [ ] Verify all responses are correct and timely
- [ ] Verify database state is consistent after concurrent writes
