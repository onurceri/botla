> Extracted from: chatbot-management.md — Section 4

# 4. Test Scenarios

## 4.1 Happy Path — Create and Configure a Chatbot

- [ ] **Precondition**: User has a valid account on Starter plan or above
- [ ] **Step 1**: `POST /api/v1/chatbots` with `{"name": "Support Bot", "language": "en"}`
- [ ] **Expect**: `201 Created` — chatbot created with default values; `language` is `"en-US"`; `model` is the platform default
- [ ] **Step 2**: `GET /api/v1/chatbots` — verify chatbot appears in list
- [ ] **Step 3**: `PUT /api/v1/chatbots/{id}/appearance` with `{"theme_color": "#FF6600", "position": "bottom-left"}`
- [ ] **Expect**: `200 OK` — theme_color updated to `"#FF6600"`, welcome_message unchanged
- [ ] **Step 4**: `PUT /api/v1/chatbots/{id}/basic-info` with `{"name": "Support Bot v2", "description": "Updated description"}`
- [ ] **Expect**: `200 OK` — name updated, description set
- [ ] **Step 5**: `PUT /api/v1/chatbots/{id}/model` with `{"temperature": 0.5, "max_tokens": 2048}`
- [ ] **Expect**: `200 OK` — temperature and max_tokens updated
- [ ] **Step 6**: `DELETE /api/v1/chatbots/{id}`
- [ ] **Expect**: `204 No Content`
- [ ] **Step 7**: `GET /api/v1/chatbots/{id}`
- [ ] **Expect**: `404 Not Found`

## 4.2 Plan Limit Enforcement

- [ ] **Precondition**: User on Free plan (limit: 1 chatbot)
- [ ] **Step 1**: Create first chatbot → `201 Created`
- [ ] **Step 2**: Create second chatbot → `403 Forbidden` with `ERR_MAX_CHATBOTS_EXCEEDED`
- [ ] **Step 3**: `PUT /api/v1/chatbots/{id}/model` with `{"model": "gpt-4"}` → `403` — model not in `ChatAllowedModels`
- [ ] **Step 4**: `PUT /api/v1/chatbots/{id}/refresh` with `{"refresh_policy": "auto"}` → `403` — `RefreshEnabled: false`
- [ ] **Step 5**: `PUT /api/v1/chatbots/{id}/security` with `{"secure_embed_enabled": true}` → `403` — feature gated
- [ ] **Step 6**: `PUT /api/v1/chatbots/{id}/handoff` with `{"handoff_enabled": true}` → `403` — feature gated
- [ ] **Step 7**: `PUT /api/v1/chatbots/{id}/appearance` with `{"hide_branding": true}` → `403` — `BrandingCanHideBranding: false`
- [ ] **Step 8**: Upgrade to Pro plan, retry all above → success expected

## 4.3 Workspace Isolation

- [ ] **Precondition**: Two workspaces (WS-A, WS-B); User belongs to both
- [ ] **Step 1**: Create chatbot in WS-A via header `X-Workspace-ID: WS-A`
- [ ] **Expect**: Chatbot created with `workspace_id: WS-A`
- [ ] **Step 2**: `GET /api/v1/chatbots` with `X-Workspace-ID: WS-B`
- [ ] **Expect**: WS-A chatbot NOT in the response list
- [ ] **Step 3**: `GET /api/v1/chatbots/{id-from-WS-A}` with `X-Workspace-ID: WS-B`
- [ ] **Expect**: `403 Forbidden`
- [ ] **Step 4**: `PUT /api/v1/chatbots/{id-from-WS-A}/basic-info` with `X-Workspace-ID: WS-B`
- [ ] **Expect**: `403 Forbidden`
- [ ] **Step 5**: `DELETE /api/v1/chatbots/{id-from-WS-A}` with `X-Workspace-ID: WS-B`
- [ ] **Expect**: `403 Forbidden`

## 4.4 Organization Isolation

- [ ] **Precondition**: Two orgs (ORG-1, ORG-2); User is member of both
- [ ] **Step 1**: Create chatbot with `X-Organization-ID: ORG-1`
- [ ] **Expect**: `organization_id` set to ORG-1; `workspace_id` is nil
- [ ] **Step 2**: Access chatbot with `X-Organization-ID: ORG-2`
- [ ] **Expect**: `403 Forbidden`

## 4.5 Input Validation and Edge Cases

### Name Validation
- [ ] `POST /api/v1/chatbots` with empty body → `400`
- [ ] `POST /api/v1/chatbots` with `{"name": ""}` → `400`
- [ ] `POST /api/v1/chatbots` with `{"name": "  "}` → `400`
- [ ] `POST /api/v1/chatbots` with name > 100 chars → `400`
- [ ] `PUT /api/v1/chatbots/{id}/basic-info` with `{"name": ""}` → `400`

### ID Format
- [ ] `GET /api/v1/chatbots/not-a-uuid` → `400` with `ErrInvalidIDFormat`
- [ ] `GET /api/v1/chatbots/new` → `400` (reserved path value)
- [ ] `GET /api/v1/chatbots/00000000-0000-0000-0000-000000000000` → `404`

### Color Validation
- [ ] `PUT /api/v1/chatbots/{id}/appearance` with `{"chat_background_color": "not-a-color"}` → `403` with `"feature": "color"`
- [ ] Valid HEX: `#FF6600`, `#F60`, `#FF6600CC` → `200`
- [ ] Valid RGBA: `rgba(255, 174, 0, 1)` → `200`
- [ ] Invalid: `"red"`, `"rgb(255,0,0)"`, `"transparent"` → `403`

### Model Settings
- [ ] `temperature` < 0 → currently not validated server-side (accepts negative; should add validation)
- [ ] `temperature` > 2 → not validated server-side
- [ ] `max_tokens` beyond plan max → `403` with `upgrade_required: true`
- [ ] `max_tokens` below plan min → `403` with `upgrade_required: false`

### Guardrails — Fallback Mode
- [ ] `fallback_mode: "static"` → always allowed (no plan check)
- [ ] `fallback_mode: "smart"` → requires `GuardrailsCanUseSmartFallback`
- [ ] `fallback_mode: "escalate"` → requires `GuardrailsCanUseEscalateFallback`
- [ ] `fallback_mode: "unknown"` → accepted (no server-side enum validation)

### Security — FlexibleStringSlice
- [ ] `PUT /api/v1/chatbots/{id}/security` with `{"allowed_domains": ["example.com", "test.org"]}` → `200`
- [ ] `PUT /api/v1/chatbots/{id}/security` with `{"allowed_domains": "example.com, test.org"}` → `200` (comma-separated string accepted)
- [ ] `PUT /api/v1/chatbots/{id}/security` with `{"allowed_domains": ""}` → `200` (empty string → empty slice)
- [ ] `PUT /api/v1/chatbots/{id}/security` with `{"allowed_domains": null}` → `200` (null → no update)

### Discovery Response for Deleted Chatbots
- [ ] Accessing any endpoint for a soft-deleted chatbot → `404 Not Found`

## 4.6 Suggestions Flow

- [ ] **Step 1**: `POST /api/v1/chatbots/{id}/suggestions/regenerate`
- [ ] **Expect**: `202 Accepted` with `{"job_id": "<uuid>"}`
- [ ] **Step 2**: `GET /api/v1/chatbots/{id}/suggestions/status`
- [ ] **Expect**: `200 OK` with `{"job_id": "...", "status": "pending"}`
- [ ] **Step 3**: Wait for worker to process (poll status every 2s up to 30s timeout)
- [ ] **Expect**: `status` transitions to `"running"` then `"completed"`; `suggested_questions` populated
- [ ] **Edge case**: No sources exist → job completes with empty `suggested_questions`
- [ ] **Edge case**: Worker crash → job stays `"running"`; no timeout mechanism evident — may need monitoring

## 4.7 Sitemap Discovery

- [ ] **Step 1**: `POST /api/v1/chatbots/{id}/sitemap/discover`
- [ ] **Expect**: `202 Accepted`
- [ ] **Step 2**: Verify pending URLs appear via `GET /api/v1/chatbots/{id}/pending-urls`
- [ ] **Edge case**: Invalid sitemap URL (in source) → discovery still returns 202; failure captured in source status
- [ ] **Edge case**: No sitemap found at domain → discovery completes with 0 URLs

## 4.8 Concurrent Operations

- [ ] **Scenario**: Two simultaneous `PUT` requests to different sub-resources on the same chatbot
- [ ] **Expect**: Both succeed (no locking — last-write-wins on shared fields)
- [ ] **Scenario**: Delete chatbot while a suggestion job is running
- [ ] **Expect**: Delete succeeds (soft-delete); job may continue or fail silently

## 4.9 Refresh Policy Transitions

- [ ] **Step 1**: `PUT /api/v1/chatbots/{id}/refresh` with `{"refresh_policy": "auto", "refresh_frequency": "weekly"}`
- [ ] **Expect**: `NextRefreshAt` is set to a future date (now + 7 days)
- [ ] **Step 2**: `PUT /api/v1/chatbots/{id}/refresh` with `{"refresh_policy": "manual"}`
- [ ] **Expect**: `NextRefreshAt` is `null`
- [ ] **Step 3**: `PUT /api/v1/chatbots/{id}/refresh` with `{"refresh_frequency": "daily"}` (while policy is manual)
- [ ] **Expect**: `NextRefreshAt` stays `null` (frequency only computed when policy is auto)
