# 3.3 Chatbot Update Test Plan

## Overview
This test plan covers all chatbot update scenarios including field validation and plan restrictions.

---

## Test Cases

### 3.3.1 Update Basic Fields
**Priority:** High  
**Type:** Integration Test

| Field | Test |
|-------|------|
| name | Update succeeds |
| description | Update succeeds |
| custom_instruction | Update succeeds |
| welcome_message | Update succeeds |
| language | Update succeeds |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Send `PUT` with updated `name`, `description`, `custom_instruction`, `welcome_message`, `language_code`.
  2. Verify 200 OK.
  3. Fetch bot and verify fields match.

---

### 3.3.2 Update Theme Settings
**Priority:** Medium  
**Type:** Integration Test

| Field | Test |
|-------|------|
| theme_color | Valid hex color succeeds |
| bot_message_color | Update succeeds |
| user_message_color | Update succeeds |
| bot_message_text_color | Update succeeds |
| user_message_text_color | Update succeeds |
| chat_font_family | Update succeeds |
| chat_header_color | Update succeeds |
| chat_header_text_color | Update succeeds |
| chat_background_color | Update succeeds |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Send `PUT` with all theme fields set to new hex values.
  2. Verify 200 OK.
  3. Verify persistence.

---

### 3.3.3 Update Model with Plan Validation
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free user updates to gpt-4o-mini | 200 OK |
| 2 | Free user updates to gpt-4o | 403 Forbidden |
| 3 | Pro user updates to gpt-4o | 200 OK |
| 4 | Pro user updates to claude | 403 Forbidden |
| 5 | Ultra user updates to claude | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create users for each plan (Free, Pro, Ultra).
  - Create a bot for each.
- **Steps:**
  1. Execute the matrix of updates (User x Model) as described in the table.
  2. Verify expected HTTP status codes.

---

### 3.3.4 Update Temperature
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update to 0.0 | 200 OK |
| 2 | Update to 2.0 | 200 OK |
| 3 | Update to -0.1 | 400 Bad Request |
| 4 | Update to 2.1 | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Update `temperature` to `0.0` -> 200.
  2. Update `temperature` to `2.0` -> 200.
  3. Update `temperature` to `-0.1` -> 400.
  4. Update `temperature` to `2.1` -> 400.

---

### 3.3.5 Update Branding Settings
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: hide_branding = true | 403 Forbidden |
| 2 | Pro: hide_branding = true | 200 OK |
| 3 | Pro: custom_branding | 403 Forbidden |
| 4 | Ultra: custom_branding | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Users on Free, Pro, Ultra.
- **Steps:**
  1. Execute matrix of updates for `custom_branding` object (specifically `hide_branding` flag and other fields).
  2. Verify expected status codes.

---

### 3.3.6 Update Secure Embed Settings
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: secure_embed_enabled = true | 403 Forbidden |
| 2 | Pro: secure_embed_enabled = true | 200 OK |
| 3 | Pro: allowed_domains | 200 OK |
| 4 | Pro: embed_secret | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Users on Free, Pro.
- **Steps:**
  1. Free user sets `secure_embed_enabled=true`. Expect 403.
  2. Pro user sets `secure_embed_enabled=true`. Expect 200.
  3. Pro user sets `allowed_domains=["example.com"]`. Expect 200.

---

### 3.3.7 Update Refresh Settings
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: refresh_policy = "auto" | 403 Forbidden |
| 2 | Pro: refresh_policy = "auto" | 200 OK |
| 3 | Pro: refresh_frequency = "daily" | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Users on Free, Pro.
- **Steps:**
  1. Free user sets `refresh_policy="auto"`. Expect 403.
  2. Pro user sets `refresh_policy="auto"`. Expect 200.

---

### 3.3.8 Update Discovery Mode
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: discovery_mode = "auto" | 403 Forbidden |
| 2 | Pro: discovery_mode = "auto" | 200 OK |
| 3 | Pro: discovery_mode = "pending" | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Users on Free, Pro.
- **Steps:**
  1. Free user sets `discovery_mode="auto"`. Expect 403.
  2. Pro user sets `discovery_mode="auto"`. Expect 200.

---

### 3.3.9 Update Guardrails Settings
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update threshold_config | Success per plan |
| 2 | Update fallback_messages | Success per plan |
| 3 | Update topic_restrictions | Success per plan |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Pro user (assuming Free is restricted per 2.1).
- **Steps:**
  1. Update `threshold_config`, `fallback_messages`, `topic_restrictions`.
  2. Verify 200 OK and persistence.

---

### 3.3.10 Update Handoff Settings
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | handoff_enabled = true | 200 OK |
| 2 | handoff_type = "email" | 200 OK |
| 3 | handoff_config with email | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Update `handoff_enabled=true`, `handoff_type="email"`, `handoff_config={"email": "test@example.com"}`.
  2. Verify 200 OK.

---

### 3.3.11 Cannot Update Another User's Chatbot
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as User A | Token A |
| 2 | PUT `/api/v1/chatbots/{user_b_bot}` | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - User A, User B with Bot B.
- **Steps:**
  1. Login as A.
  2. Send `PUT /api/v1/chatbots/{bot_b_id}`.
  3. Verify 403.

---

### 3.3.12 Updated_at Timestamp Updates
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note current updated_at | Timestamp T1 |
| 2 | Update chatbot | 200 OK |
| 3 | Verify updated_at | Timestamp T2 > T1 |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot. Store `updated_at` as T1.
- **Steps:**
  1. Sleep 1 second.
  2. Update name.
  3. Fetch bot. Store `updated_at` as T2.
  4. Verify T2 > T1.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "UpdateChatbot"
```
