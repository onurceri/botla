> Extracted from: chatbot-management.md — Section 2.1

# 2.1 Chatbot CRUD

## List Chatbots
- [ ] `GET /api/v1/chatbots` returns all chatbots accessible to the authenticated user
- [ ] When a workspace context header (`X-Workspace-ID`) is present, returns only chatbots in that workspace
- [ ] Non-members of the workspace receive `403 Forbidden`
- [ ] Response is a JSON array of chatbot objects
- [ ] Empty array returned when no chatbots exist

## Create Chatbot
- [ ] `POST /api/v1/chatbots` creates a new chatbot with a `name` (required, min 1, max 100 chars)
- [ ] Optional fields accepted: `description`, `language`, `custom_instruction`, `model`, `temperature`, `max_tokens`, all appearance fields, security fields, suggestions, discovery, refresh, guardrails, handoff, branding
- [ ] Missing `name` returns `400 Bad Request`
- [ ] Empty/whitespace-only `name` returns `400 Bad Request`
- [ ] `language` defaults to `tr-TR` when omitted; locale normalization maps `tr` → `tr-TR`, `en` → `en-US`
- [ ] Plan limits enforced: `MaxChatbots` check — `403 Forbidden` with `ERR_MAX_CHATBOTS_EXCEEDED` when at limit
- [ ] Workspace context verified via `X-Workspace-ID` header; non-members get `403`
- [ ] Organization context verified via `X-Organization-ID` header; non-members get `403`
- [ ] Default values applied for all optional fields not provided (see `buildNewChatbot`)
- [ ] Response is `201 Created` with full chatbot object
- [ ] WorkspaceID and OrganizationID are set correctly based on context headers

## Get Chatbot by ID
- [ ] `GET /api/v1/chatbots/{id}` returns a single chatbot object
- [ ] `404 Not Found` for non-existent ID
- [ ] `400 Bad Request` for non-UUID ID format
- [ ] `403 Forbidden` when chatbot belongs to a different workspace/org the user does not have access to
- [ ] `401 Unauthorized` without valid auth cookie
- [ ] Full chatbot object returned (all 61 fields)

## Update Chatbot (full PUT)
- [ ] `PUT /api/v1/chatbots/{id}` accepts partial updates via `ChatbotUpdateRequest`
- [ ] Empty body returns `400 Bad Request`
- [ ] Invalid JSON body returns `400 Bad Request`
- [ ] All 61 fields are updatable through this endpoint
- [ ] Plan validation runs for: model, branding, refresh, discovery, secure embed, guardrails, handoff, topic restrictions, colors, max tokens, manual questions
- [ ] Plan-restricted features return `403 Forbidden` with `FeatureError` body containing `upgrade_required`, `feature`, and `error`
- [ ] `200 OK` with updated chatbot object on success
- [ ] `404 Not Found` for non-existent chatbot
- [ ] Access control enforced (same as GET)

## Delete Chatbot
- [ ] `DELETE /api/v1/chatbots/{id}` soft-deletes the chatbot (sets `deleted_at`)
- [ ] Returns `204 No Content`
- [ ] Associated source vectors cleaned up from vector store (best-effort)
- [ ] `404 Not Found` for non-existent ID
- [ ] `403 Forbidden` for unauthorized access
- [ ] Deleting an already-deleted chatbot returns `404 Not Found`
