> Extracted from: auth.md — Section 2

# 2. Cross-Cutting Edge Cases

## 2.1 Authentication + Authorization across Features

- [ ] Verify authenticated user cannot access another org's resources (403)
- [ ] Verify platform admin can access all orgs via admin endpoints
- [ ] Verify non-admin cannot access admin endpoints (403)
- [ ] Verify deleted user's access tokens are invalidated mid-session (401)
- [ ] Verify plan limits affect endpoint behavior (e.g., max orgs, max members)
- [ ] Verify rate limiting across all endpoints (per-plan limits)

## 2.2 Cookie and Session Security

- [ ] Verify `botla_token` cookie is `HttpOnly`, `Secure`, `SameSite=Lax`
- [ ] Verify `botla_refresh_token` cookie is `HttpOnly`, `Secure`, `SameSite=Lax`
- [ ] Verify cookies have `Path=/api/v1` (or appropriate scope)
- [ ] Verify cookies are NOT accessible via `document.cookie`
- [ ] Verify refresh token is never exposed in response body
- [ ] Verify access token is never exposed in response body (only in cookie)

## 2.3 Request Validation

- [ ] Test all endpoints with missing `Content-Type: application/json`
- [ ] Test all endpoints with malformed JSON body
- [ ] Test all POST/PATCH/PUT endpoints with empty body
- [ ] Test all endpoints with oversized payload (>52MB, MaxBytes middleware)
- [ ] Test all endpoints with invalid UUID in path parameters
- [ ] Test all endpoints with extra unknown fields (should be ignored, not fail)
- [ ] Test all endpoints with SQL injection attempts in string fields
- [ ] Test all endpoints with XSS payloads in string fields (stored, not executed)

## 2.4 Concurrency

- [ ] Create org and add member simultaneously — no duplicate member
- [ ] Register same email simultaneously — only one succeeds (unique constraint)
- [ ] Delete org while member is being added — consistent state
- [ ] Multiple session refresh calls — only one succeeds per token
- [ ] Concurrent privacy request creation — only one pending per type

## 2.5 Data Integrity

- [ ] Verify cascading deletes: org → workspaces → chatbots (or error prevention)
- [ ] Verify soft delete: user marked `deleted_at` but data preserved
- [ ] Verify unique constraints: email, org slug, workspace slug (per org)
- [ ] Verify audit logging: admin actions tracked in audit logs
- [ ] Verify timestamps: `createdAt`, `updatedAt`, `deletedAt` are UTC ISO 8601
