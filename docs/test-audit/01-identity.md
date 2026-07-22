# Test Coverage Audit — Identity Features (auth, user-profile, organization-workspace)

**Scope:** `docs/features/auth/` (2 files), `docs/features/user-profile/` (21 files), `docs/features/organization-workspace/` (2 files)
**Layers mapped:** Go unit (`internal/**`, `pkg/**`) · Go integration (`internal/integration/**`) · Frontend Vitest (`frontend/src/**/__tests__/`) · Playwright e2e (`frontend/e2e/*.spec.ts`)
**Method:** read all 25 spec docs, extracted every checklist group / Scenario block / edge-case row cluster, verified each against actual test assertions (not just name similarity). No test suites executed.

**Legend:** ✅ COVERED · 🟡 PARTIAL · ❌ MISSING

---

## 1. AUTH (`docs/features/auth/`)

### 1.2.1 Registration `POST /api/v1/auth/register` — 🟡 PARTIAL

| Sub-scenario | Status | Evidence |
|---|---|---|
| Valid register (201, tokens) | ✅ | int: `TestAuth_Register_Login_Protected` (auth_test.go:34); unit: `TestAuth_Cookies` (auth_cookie_test.go:15) |
| Sub-addressing email (`user+tag@`) | ✅ | All integration tests register with `+`-suffixed emails |
| Minimum-valid password (8 chars, all classes) | ✅ | int: `TestAuth_Register_Login_Protected` uses `Test@123` (8 chars, all classes) |
| Duplicate email → 409 `ERR_EMAIL_EXISTS` | 🟡 | int: `TestAuth_Register_Login_Protected` (exact-case only). **Case-variant duplicate untested — impl does case-sensitive `SELECT … WHERE email=$1` pre-check while login is case-insensitive; likely bug** |
| Password rule rejections (upper/lower/digit/special matrix) | 🟡 | Only `TestAuth_Register_WeakPassword` (`"short"`→400 `ERR_PASSWORD_TOO_SHORT`). `isStrongPassword` (auth.go:35) has **no unit test**; Scenario-3 matrix (no-upper, no-lower, no-digit, no-special, whitespace-only, leading-space) ❌ |
| Malformed email → 400 | ✅ | int: `TestAuth_Register_InvalidEmailFormat` |
| Empty body / missing fields → 400 | ✅ | unit: `TestRegisterHandler_BadJSON`, `TestRegisterHandler_MethodNotAllowed`; int: `TestAuth_Register_MissingEmail`, `TestAuth_Register_MissingPassword` |
| Email >254 chars / null bytes / unsupported encoding | ❌ | No tests found |
| HttpOnly `botla_token` + `botla_refresh_token` cookies | ✅ | unit: `TestAuth_Cookies`, `TestAuth_CookieSecureFlag_Development`, `TestAuth_CookieSecureFlag_Production` — **note spec says SameSite=Lax; impl/tests pin Lax(dev)/None(prod)** |
| Response body `user{id,email,fullName,createdAt}` | 🟡 | Impl returns `{token, refresh_token}` JSON; tests assert tokens only, never a `user` object. Spec/impl mismatch (see §6) |
| Default org + workspace created | ✅ | int: `TestAuth_RegistrationCreatesDefaultOrgAndWorkspace` |
| bcrypt hash in DB (`$2a$/$2b$`, not plaintext) | ✅ | int: `TestAuth_Register_PasswordHashing`; unit: `TestHashAndVerifyPassword` |
| Rate limit 5 req/min → 429 | 🟡 | unit: `TestRateLimitMiddleware_AuthEndpointsStrictLimits` (asserts config header=**3**/min register, 5 login, 10 refresh — spec says 5 for register). No Nth-request-429 or window-recovery test for auth endpoints |

### 1.2.2 Login `POST /api/v1/auth/login` — 🟡 PARTIAL

| Sub-scenario | Status | Evidence |
|---|---|---|
| Correct credentials → 200 | ✅ | int: `TestAuth_Register_Login_Protected` |
| Case-insensitive email | ✅ | int: `TestAuth_Login_CaseInsensitiveEmail` |
| Whitespace-trimmed email | ❌ | Impl trims (auth.go:190); no test |
| Wrong password → 401 | ✅ | int: `TestAuth_Login_InvalidPassword` |
| Non-existent email → 401 | ✅ | int: `TestAuth_Login_InvalidEmail` |
| Soft-deleted account → 403 `ACCOUNT_DELETED` | ❌ | **No test. Impl gap: LoginHandler query (auth.go:198) has no `deleted_at` filter** — privacy-deleted users only fail because email is anonymized (→401). Unit `TestDeletedAccountMiddleware` (403) covers the middleware, not login |
| Empty credentials → 400 | ✅ | int: `TestAuth_Login_EmptyEmail`, `TestAuth_Login_EmptyPassword` |
| Cookie attributes (1h/7d, HttpOnly, Secure, SameSite, Path, Domain) | 🟡 | Access-token 60m TTL asserted (`TestAuth_Register_Login_Protected`); cookie flags in `auth_cookie_test.go` (Path=`/`, spec says `/api/v1` — mismatch); 7d refresh TTL & Domain never asserted |
| Refresh token SHA-256 in DB | ✅ | int: `TestAuth_RevokedRefreshToken_401`, `TestAuth_Login_MultipleSessions`, `TestAuth_Login_RefreshTokenTracking` (hash lookups) |
| Response body excludes password hash | 🟡 | Response is token-only JSON; no explicit negative assertion |
| Login rate limit (5/min; 3/min after failures) | 🟡 | Config-header unit test only; "3 req/min after 3 failures" adaptive behavior not implemented/tested |
| Remember-me refresh expiry | ❌ | Backend has no remember-me parameter; e2e `sessions.spec.ts` remember-me tests are mock/localStorage-based (stale, see §6) |

### 1.2.3 Token Refresh `POST /api/v1/auth/refresh` — 🟡 PARTIAL

| Sub-scenario | Status | Evidence |
|---|---|---|
| Valid refresh → new pair | ✅ | int: `TestAuth_Refresh_GeneratesNewAccessToken`, `TestAuth_RefreshRotationAndLogout`; unit cookie-path: `TestAuth_Cookies` |
| Rotation: old revoked, new differs | ✅ | int: `TestAuth_RefreshRotationAndLogout` (old→401, new≠old) |
| Expired refresh → 401 | ✅ | int: `TestAuth_RefreshWithExpiredToken` |
| Reused/rotated token → 401 | ✅ | int: `TestAuth_RevokedRefreshToken_401`, `TestAuth_RefreshRotationAndLogout` |
| Missing cookie/token → 400 | 🟡 | Impl 400; only `TestRefreshHandler_BadJSON` (malformed body) |
| Access-token-as-refresh → 401 | 🟡 | Token-type check exists in impl; middleware-side covered by `TestAuthMiddleware_InvalidTokenType`; refresh-endpoint side untested |
| **Concurrent refresh — exactly 1 succeeds (Scenario 2)** | ❌ | No concurrency test; rotation is check-then-act (`SELECT revoked` → `UPDATE`) — race plausible and unpinned |
| Refresh rate limit | 🟡 | Config header unit test only |

### 1.2.4 Logout `POST /api/v1/auth/logout` — 🟡 PARTIAL

| Sub-scenario | Status | Evidence |
|---|---|---|
| Logout → refresh revoked in DB | ✅ | int: `TestAuth_RefreshRotationAndLogout` (post-logout refresh→401), `TestAuth_Login_MultipleSessions` (`revoked=true`, other session survives) |
| Cookies cleared (MaxAge=-1) | ✅ | unit: `TestAuth_Cookies`, `TestAuth_LogoutHandler_CookieSecureFlag` |
| Subsequent API calls → 401 | 🟡 | Refresh-after-logout ✅; **old access token after logout → 401 ❌** (stateless JWT, no denylist — Scenario 1 Step 4 cannot pass as specced; untested) |
| Logout without cookies (idempotent 200) | 🟡 | Impl returns **400**; pinned by `TestAuth_LogoutHandler_CookieSecureFlag` — spec mismatch |
| Logout rate limit | ❌ | Not in strict-limits test set |


### 1.2.5 Protected Endpoint — ✅ mostly COVERED

✅ valid 200 / no-auth 401 / expired 401 / malformed 401 / tampered 401 / refresh-as-access 401 / wrong secret — `TestAuthMiddleware_*` (auth_test.go, auth_format_test.go), `TestAuth_InvalidAccessToken_Me401`, `TestAuth_TamperedAccessToken_Me401`, `TestAuth_MissingAuthorizationHeader_Me401`, `TestAuth_ValidAccessToken_Me200`, `TestAuth_ExpiredAccessToken_Me401`, `TestVerifyToken_WrongSecret`, `TestVerifyToken_Malformed`, int `TestAuth_InvalidBearerFormat_Protected401`, `TestAuth_BearerWithoutToken_Protected401`, `TestAuth_BasicScheme_Protected401`.
🟡 soft-deleted user via middleware — unit `TestDeletedAccountMiddleware` only; no full-stack test. ❌ `WWW-Authenticate` header on 401 (not implemented anywhere). 🟡 consistent 401 error-body format — not asserted.

### 1.2.6 Current User `GET /api/v1/me` — 🟡 PARTIAL

✅ basic fetch + `isPlatformAdmin` bool: unit `TestMe_Success`, `TestMe_IsPlatformAdmin_DefaultFalse/WhenTrue`; int `TestMe_ReturnsUserProfile`, `TestMe_ProfileBasicInfo`, `TestMe_ProfileIncludesOrganizations`, `TestMe_DBErrorReturns500`, `TestMe_CrossUserProfileIsolation`.
🟡 full field set (`avatarUrl`, `languageCode`, `onboardingData`, `planId`) — never all asserted; ❌ `languageCode` default `"en"`; ❌ explicit exclusion of `passwordHash`/`deletedAt`; ❌ fetch-after-soft-delete (unit middleware only, and middleware returns 403 vs spec's 401); 🟡 CORS `withCredentials` — CORS unit tests (`TestCORSMiddleware_*`) + int cors specs exist, but `Access-Control-Allow-Credentials` is never asserted.

### 1.2.7 Frontend Auth Source of Truth — 🟡 PARTIAL

✅ refresh queue / single refresh on concurrent 401s: `client.test.ts` — "coalesces concurrent 401 refresh into a single request", "retries all 5 concurrent requests after single token refresh". ✅ 401→/login: `App.auth-redirect.test.tsx`, `client.test.ts` ("redirects to login on refresh failure…", "/me 401 new visitor"). ✅ session-expiry modal: `SessionExpiryHandler.test.tsx` (5 tests), `client.test.ts` wasAuthenticated gating, ERR_ACCOUNT_DELETED handling. 🟡 "/me called on mount and cached as auth state" — `AuthContext.tsx` (React Query, staleTime 5m) has **no dedicated test**. ❌ cookie-based reload persistence & multi-tab — see Scenario 7.

### Auth Scenarios 1–7

| Scenario | Status | Notes |
|---|---|---|
| S1 Register→Login→Refresh→Logout cycle | 🟡 | All steps covered piecewise (tests above) **except** Step 4 "old access token → 401 after logout" (no denylist; untestable as specced) |
| S2 Refresh race (3 concurrent → 1×200, 2×401) | ❌ | No concurrency test anywhere |
| S3 Password strength matrix | 🟡 | 1 of 8 cases (`TestAuth_Register_WeakPassword`) |
| S4 Email case-insensitivity | 🟡 | Login lower/upper ✅ (`TestAuth_Login_CaseInsensitiveEmail`); re-register 409 + lowercase DB storage ❌ (impl stores original case; dup check case-sensitive) |
| S5 Auth rate limiting (6th→429, window recovery) | 🟡 | Config-level unit test only; integration 429 tests exist for chat/source (`TestRateLimit_ChatEndpoint`, `TestRateLimit_Recovery`) but not auth |
| S6 Soft-deleted user cannot authenticate | ❌ | No test of login→403 / refresh→401 / me→401 for deleted users |
| S7 Frontend persistence & multi-tab | 🟡 | Persistence e2e exist (`sessions.spec.ts`, `logout.spec.ts`, `register.spec.ts`) but **all mock-based, asserting `localStorage.botla_token` — the production app is HttpOnly-cookie-based and never sets that key**; BroadcastChannel multi-tab logout has **no production implementation** (e2e tests instantiate their own channel — false confidence) |

### Auth §1.5 Edge Cases

✅ SQL injection email (`TestAuth_Register_SQLInjectionEmail`, incl. DB integrity check) · ✅ XSS fullName (`TestAuth_Register_XSSFullName` — asserts sanitized `/me` output) · 🟡 stolen-refresh reuse (reuse→401 ✅; token-family revocation ❌ not implemented) · 🟡 duplicate-email race (sequential ✅; concurrent ❌) · 🟡 CSRF/SameSite (cookie attrs unit-tested; prod SameSite=None weakens the spec claim) · ❌ `alg:none` JWT · ❌ RS256/HS256 confusion (VerifyToken lacks `WithValidMethods`; library default likely safe but unpinned) · ❌ `__Host-` prefix · ❌ null bytes in email · ❌ bcrypt >72-char truncation · ❌ blocked-IP (feature absent) · ❌ browser without cookies.

### `auth/02-cross-cutting-edge-cases.md`

✅ cross-org 403 (RBAC `GetOrg_Outsider`; `TestMultiTenantIsolation`; `TestMultiTenant_*` 11 isolation suites) · ✅ admin accesses orgs / non-admin 403 (`TestAdminFlow` 403 sweep; `TestRequirePlatformAdmin` unit) · ✅ oversized payload (`TestMaxBytesMiddleware_*`) · ✅ invalid UUIDs (`TestRBAC_Extended` InvalidOrgID/WorkspaceID/ChatbotID → 400 `ERR_INVALID_ID_FORMAT`) · 🟡 deleted-user mid-session (unit only) · 🟡 malformed JSON (auth only, not "all endpoints") · ❌ missing Content-Type sweep · ❌ unknown-fields-ignored · ❌ all concurrency rows (org+member race, email race, org-delete race, refresh race, privacy race) · ❌ org/member plan-limit effects (not implemented — no org/member/workspace quotas exist in `plan_limits.go`) · 🟡 cascade org→ws→chatbots (workspace-level only, WSC-004) · ✅ user soft-delete preserved (privacy anonymization test) · 🟡 audit logging (frontend `AdminAuditPage.test.tsx` only; no backend write-verification) · ❌ UTC ISO-8601 timestamp format assertions.


---

## 2. USER-PROFILE (`docs/features/user-profile/`)

| # | Feature group | Status | Evidence / gaps |
|---|---|---|---|
| 1.2.1 | `GET /me` | 🟡 | Covered as in Auth §1.2.6; ❌ avatarUrl null/set, languageCode, soft-delete 401 |
| 1.2.2 | `GET /me/plan` | 🟡 | ✅ unit `TestGetPlan_Success`; int `TestFreePlan_DefaultAssignment_RegisterThenMe`, `TestGetPlan_ReturnsSplitConfig`; ✅ cache logic unit `TestPlanService_MemoryCache_*`, `TestPlanService_ConcurrentAccess`; ❌ **Redis-layer plan caching (`pkg/middleware/plan_loader.go`) has zero tests**; 401 covered via middleware tests |
| 1.2.3 | `GET /me/usage` | ❌ | **No HTTP-level test exists** (no handler unit, no integration). Only repo aggregation: int `TestMonthlyTokenUsageTracking` (incl. month-boundary reset USG-005), `TestIngestionUsageTracking`. Response shape (`messageCount/tokenUsage/tokensRemaining/periodStart/periodEnd`), zero-values, 401 — all untested |
| 1.2.4 | `GET /me/onboarding` | ✅ | unit `TestGetOnboardingState_Success`, `TestGetOnboardingState_Unauthorized`; int `TestOnboardingFlow` (initial `{completed:false, step:0, skipped:false}`), `TestOnboardingStatePersistedAcrossSessions` |
| 1.2.5 | `PUT /me/onboarding` | 🟡 | ✅ success: unit `TestUpdateOnboardingState_Success` (+ DB verify), `TestUpdateOnboardingState_InvalidJSON`; int `TestOnboardingFlow`. ❌ invalid step rejection (**impl accepts any int — no validation**); ❌ oversized payload; ⚠️ **merge-vs-replace: spec requires merge; handler replaces `OnboardingData` wholesale, and null `data` → empty struct (likely wipes saved data) — untested, probable bug** |
| 1.2.6 | `POST /onboarding/skip` | 🟡 | ✅ unit `TestSkipOnboarding_Success`; int `TestOnboardingSkip` (skipped=true persists). ❌ idempotency (skip when completed / already skipped) |
| 1.2.7 | `POST /onboarding/complete` | 🟡 | ✅ unit `TestCompleteOnboarding_Success`; int `TestOnboardingFlow` (completed=true). ❌ complete-after-skip override (Scenario 6); ❌ idempotency. Note impl requires `bot_id` (400 without) — spec scenario omits this |
| 1.2.8 | `GET /privacy/consents` | 🟡 | ✅ int `TestPrivacyFlow/User Consents` (PATCH then GET verified). ❌ defaults for brand-new user (never fetched before first PATCH) |
| 1.2.9 | `PATCH /privacy/consents` | 🟡 | ✅ multi-consent update (same test); ✅ consent IP recording int `TestConsentIPExtraction`. ❌ **unknown consent type → 400 (impl uses fixed struct; unknown keys silently ignored — spec violation)**; ❌ non-boolean → 400; ❌ `updatedAt` refresh; ❌ no-op same values |
| 1.2.10 | `POST /privacy/export` | ✅ | int `TestPrivacyFlow`: "User Request Export" (201/pending), duplicate-pending → 409 subtest; unit `TestPrivacyService_RequestExport_RateLimit` (4 subtests incl. 24h cooldown); frontend `PrivacySettingsPage.test.tsx` (409 + 429 handling) |
| 1.2.11 | `POST /privacy/correction` | 🟡 | ✅ create w/ reason (`TestPrivacyFlow/User Request Correction`). ❌ missing reason → 400 (impl exists, user_privacy.go:224); ❌ max-length rejection (frontend has char-count/maxlength tests: `PrivacySettingsPage.test.tsx` — unit-level only) |
| 1.2.12 | `POST /privacy/delete` | 🟡 | ✅ create pending + admin processing + anonymization + `deleted_at` (`TestPrivacyFlow` "User Request Deletion", "Admin Process Request"). ❌ "cannot authenticate after processing" (no login/refresh attempt post-deletion); 🟡 duplicate-pending deletion 409 (export-only tested) |
| 1.2.13 | `GET /privacy/requests` | 🟡 | ✅ list + pagination (`TestPrivacyFlow` "User List Privacy Requests with Pagination", pages 1–2). ❌ createdAt-desc ordering assertion; ❌ downloadUrl-in-list when completed |
| 1.2.14 | `GET /privacy/requests/{id}` | 🟡 | Cross-user 404 covered for DELETE ("User Cannot Delete Another User's Request"); single-GET happy-path and 404s not directly tested |
| 1.2.15 | `DELETE /privacy/requests/{id}` | 🟡 | ✅ delete own (soft-delete verified in DB), other user's → 404. ❌ delete completed request → 409; ❌ delete already-deleted → 404 |
| 1.2.16 | `GET …/download` | 🟡 | ✅ 200 download (`TestPrivacyFlow` "User Download Export"); ✅ 410-expiry implemented (user_privacy.go:354,413) but ❌ untested; ❌ not-completed (impl 400 vs spec 404 — mismatch); ❌ non-export type → 400; ❌ another user's export → 404; 🟡 Content-Type/Disposition set in impl, weakly asserted |
| S1 | Complete onboarding flow | ✅ | `TestOnboardingFlow` + unit tests |
| S2 | Consent lifecycle | 🟡 | Except unknown-type 400 (above) |
| S3 | Export→download workflow | ✅ | `TestPrivacyFlow` incl. admin approve → completed → `export_url` → download |
| S4 | Deletion workflow | 🟡 | Up to anonymization ✅; **login→403 after processing ❌; re-register same email → 409 ❌ (impl anonymizes email, so re-register would succeed — spec conflict)** |
| S5 | Privacy tenant isolation | 🟡 | Cross-user DELETE 404 ✅; list-invisibility and GET 404 not asserted |
| S6 | Skip→re-complete | ❌ | No test |
| §1.5 | Edge cases | 🟡 | ✅ usage-exceeds-plan → quota blocks (quota_enforcement_test.go); ✅ month reset (USG-005); ❌ negative `tokensRemaining`; ❌ special chars in onboarding data; ❌ 410 download expiry; ❌ concurrent privacy creation (sequential 409 only); ❌ zero-usage endpoint response; ❌ 413 large consent payload (global MaxBytes unit only); ❌ download non-export 400; ❌ delete-already-deleted 404; ❌ null-data no-op (**impl likely wipes data**) |


---

## 3. ORGANIZATION-WORKSPACE (`docs/features/organization-workspace/`)

| # | Feature group | Status | Evidence / gaps |
|---|---|---|---|
| 1.2.1 | List orgs | 🟡 | ✅ default org + owner role: `TestAuth_RegistrationCreatesDefaultOrgAndWorkspace`, `TestServiceLevelConstraints/SVC-007`; ✅ 401 (`TestRBAC_Extended/MissingAuthHeader`). ❌ `memberCount`/full field shape; ❌ owned-vs-member mix; ❌ pagination (not implemented) |
| 1.2.2 | Create org | 🟡 | ✅ success 201 (RBAC setup, `organization_role_test.go`); ✅ creator=owner (membership insert, asserted via role="owner"); ✅ duplicate slug → 409 (`ErrOrgSlugExists`, bare body — no `SLUG_TAKEN` code asserted). ❌ **server-side slug auto-generation (impl 400s on missing slug; spec Scenario 5 untestable — FE `CreateOrganizationDialog` generates slug but has no tests)**; ❌ **unique-suffix on collision (spec Scenario 1 Step 3 expects 201 `acme-corp-1`; impl returns 409)**; ❌ **default workspace in new org (`CreateOrganization` does NOT create one — registration path only)**; ❌ empty-name/long-name/invalid-slug-chars 400; ❌ org limit from plan (not implemented) |
| 1.2.3 | Update org | 🟡 | ✅ 200 owner + DB verify: `TestOrganizationManagement/UpdateOrganization`; ✅ 403 admin/member: RBAC `PatchOrg_Admin/Member`. ❌ duplicate-slug → 409 test; ❌ empty name 400; 🟡 404 non-existent (cross-org → 403 covered via outsider) |
| 1.2.4 | Delete org | 🟡 | ✅ owner 200/204 + DB gone: `TestOrganizationManagement/DeleteOrganization`, `TestRBAC_DeleteOrg`; ✅ non-owner 403 (admin case); ✅ last-org 400: `SVC-007`, `DeleteLastOrganization_Fail`. ❌ **409 `ORG_HAS_CHATBOTS` (not implemented — no chatbot check in `DeleteOrganization`)**; 🟡 cascade ws→chatbots (workspace-level WSC-004 only); 🟡 other-orgs-unchanged (implicit) |
| 1.2.5 | List members | 🟡 | ✅ 200 member: RBAC `GetMembers_Member`; shape `{members, caller_role}` decoded in tests. ❌ owner-first/joinedAt-sort/pagination; ❌ outsider → 403 on members endpoint specifically |
| 1.2.6 | Add member | 🟡 | ✅ 201 by email (org_role_test, RBAC `AddMember_Admin`, helper `addMember`). ❌ **non-existent email → 400**; ❌ **duplicate member → 409 `ALREADY_MEMBER` (impl is `ON CONFLICT DO UPDATE` upsert — spec violation)**; 🟡 add by non-owner/non-admin → 403 (member-PATCH 403 exists; member-POST-members not in RBAC matrix); ❌ member limit from plan (not implemented); 🟡 default role `member` (not asserted in response); ❌ batch add |
| 1.2.7 | Remove member | 🟡 | ✅ 200 admin: RBAC `RemoveMember_Admin`; ✅ last-owner 400/403: `SVC-006`. ❌ remove-by-non-owner 403 (dedicated); ❌ 404 non-existent member/org; ❌ removed-member-loses-access; ❌ user-not-deleted assertion; ❌ self-removal 409 (only last-owner protected) |
| 1.2.8 | Update member role | ✅ | `TestOrganization_UpdateMemberRole_Permissions` (member 403, owner 200); `TestServiceLevelConstraints`: SVC-002 self-promotion, SVC-003 admin-assigns-owner 403, SVC-005 demote-last-owner, SVC-009 two-owners-demote-one, SVC-010 invalid role 400. ❌ 404 non-existent member |

| 1.2.9 | List workspaces | 🟡 | ✅ default workspace (`TestAuth_RegistrationCreatesDefaultOrgAndWorkspace`, SVC-008); ✅ member 200 (RBAC `GetWS_Member`). ❌ non-member → 403 (Scenario 3 Step 5); 🟡 field shape |
| 1.2.10 | Create workspace | 🟡 | ✅ 201 admin (RBAC `CreateWS_Admin`, helper); ✅ 403 member (`CreateWS_Member`); ✅ slug-dup error type exists (`ErrWorkspaceSlugExists`). ❌ duplicate-name/slug → 409 test; ❌ workspace plan limit (not implemented) |
| 1.2.11 | Update workspace | 🟡 | ✅ 200 admin / 403 member (RBAC `PatchWS_*`); ✅ invalid UUID 400 (`InvalidWorkspaceID`). ❌ 404 valid-but-missing ws/org |
| 1.2.12 | Delete workspace | 🟡 | ✅ admin 200/204 (RBAC-017 `DeleteWorkspace_Admin`); ✅ last-workspace 400 (`SVC-008` + `TestOrganizationManagement/DeleteWorkspace` — **spec expects 409 `LAST_WORKSPACE`; impl/test pin 400**); ✅ chatbot cleanup (WSC-004). ❌ delete-by-non-admin → 403 |
| S1 | Full org lifecycle | 🟡 | All steps piecewise ✅ except Step 3 slug-collision-expects-201 (impl 409 — spec internal contradiction with 1.2.2) |
| S2 | Member management workflow | 🟡 | Promote/remove/cannot-remove-owner ✅ piecewise; add-non-existent-email 400 ❌; member-adds-member 403 ❌; post-removal list verification ❌ |
| S3 | Workspace isolation | 🟡 | WSC-001/002/004 ✅ (scoped create/list/delete-cleanup); non-member workspace listing 403 ❌ |
| S4 | Tenant data isolation | 🟡 | Outsider GET org → 403 ✅ (RBAC matrix); cross-org PATCH/DELETE/members/workspaces for outsider — only GET covered |
| S5 | Slug auto-generation | ❌ | Not implemented server-side; no tests |
| §1.5 | Edge-case table | 🟡 | ✅ last-workspace delete (400); last-owner constraints (SVC-005/006). ❌ slug race → 409 unique-constraint; ❌ already-member 409 (upsert); ❌ remove-self 409; ❌ special-char/100%-special slug sanitize + fallback; ❌ >255 name 400; ❌ member limit 409; ❌ add-by-UUID; ❌ ownership transfer (unsupported); ❌ concurrent add; ❌ deleted-org-in-list |
| 02 | Cross-cutting doc | — | Identical to auth/02 — see §1 assessment above |


---

## 4. Prioritized Gap List for PROD Readiness

### P0 — Security / auth / tenant-isolation / data-loss risk

1. **Soft-deleted user authentication (Auth S6; profile S4)** — No test that a deleted user gets 403 on login / 401 on refresh & `/me`. Worse, `LoginHandler`'s SQL lacks a `deleted_at` filter — only email anonymization prevents privacy-deleted users from logging in. Add integration test + fix query. *(docs: auth 1.2.2, Scenario 6; user-profile 13, Scenario 4)*
2. **Refresh-token rotation race (Auth S2)** — `RefreshHandler` does check-then-act revoke without locking/atomicity; no concurrent-refresh test. Token-theft family revocation also unimplemented/untested (§1.5 stolen-token row).
3. **E2E auth suites are architecturally stale** — `sessions.spec.ts`, `logout.spec.ts`, remember-me portions of `auth.spec.ts`/`register.spec.ts` run entirely against route mocks that inject `localStorage.botla_token`, which the production cookie-based app never writes. They cannot catch real cookie-flow regressions (HttpOnly/SameSite/expiry, `withCredentials`). Need at least one real-backend Playwright path (register→reload→refresh→logout) plus cookie-attribute assertions via `browserContext.cookies()`.
4. **Multi-tab session termination (Auth S7)** — BroadcastChannel logout propagation is "tested" by e2e that build their own channel; **no production code uses BroadcastChannel**. Either implement + test, or delete the spec/tests.
5. **Tenant-isolation breadth** — Cross-org access tested only for GET-org (outsider). Add outsider/cross-org 403|404 tests for PATCH/DELETE org, GET members, GET/POST workspaces, and workspace-delete-by-non-admin (Org S4, 1.2.9, 1.2.12).
6. **`PUT /me/onboarding` data-loss risk** — Handler replaces the whole `OnboardingData` and maps `data:null` → empty struct; spec requires merge/no-op. Untested. Verify + test before PROD.
7. **`GET /api/v1/me/usage` has zero HTTP-level tests** — a user-facing quota endpoint (tokensRemaining, period bounds) with no handler or integration test; only repo aggregation is covered.
8. **JWT algorithm pinning tests** — `VerifyToken` doesn't restrict valid methods (relies on library defaults). Add alg:none and RS256/HS256-confusion negative tests (§1.5).

### P1 — Core user flows break if regressed

9. **Password strength matrix** — `isStrongPassword` has no unit test; only too-short is integration-tested. Table-drive all 8 spec cases (Auth S3).
10. **Email case semantics** — Case-variant duplicate registration (409) and lowercase normalization untested; impl dup-check is case-sensitive while login is case-insensitive → possible duplicate-account bug (Auth S4).
11. **Onboarding idempotency & completion rules** — skip/complete idempotency, complete-after-skip override (profile S6), invalid-step rejection — all untested.
12. **Consent update validation** — unknown consent type should 400 (currently silently ignored); non-boolean → 400; defaults-for-new-user; `updatedAt` refresh (profile 1.2.8/1.2.9, S2).
13. **Privacy request lifecycle edges** — delete-completed → 409, delete-already-deleted → 404, download-not-ready / non-export-type → 400, expired-link → 410 (implemented, untested), cross-user GET/download 404 (profile 1.2.14–1.2.16, §1.5).
14. **Org spec mismatches needing a decision + tests** — slug auto-generation & unique-suffix (S5, S1-step-3: spec expects 201, impl 409s with empty body and no `SLUG_TAKEN` code); default workspace on POST /organizations (not created); duplicate-member 409 vs current upsert; last-workspace 409 vs current 400; `ORG_HAS_CHATBOTS` 409 (unimplemented — accidental-deletion guard). Docs and code disagree; pin whichever is chosen.
15. **Member management negatives** — add-non-existent-email 400, member-adds-member 403, remove-by-non-owner 403, 404s, removed-member-loses-access (Org S2).
16. **Auth rate-limit behavior tests** — only config-header unit tests exist; add 429-on-Nth-request + window-recovery integration for register/login/refresh (Auth S5). Also reconcile spec "5/min register" vs impl "3/min".
17. **Post-deletion re-registration (profile S4)** — spec says 409 (email retained); impl anonymizes email so re-register succeeds. Decide policy; test it.
18. **Frontend org feature has no Vitest coverage** — `features/organization/` (OrganizationContext, switcher, create dialogs incl. client-side slug generation, settings pages) is completely untested; only mock-e2e switcher tests in `dashboard.spec.ts`.

### P2 — Nice-to-have / hardening

19. `WWW-Authenticate` header on 401s (unimplemented) and consistent 401 error-body assertions.
20. Cookie `Domain`/`Path` scope assertions (spec `Path=/api/v1` vs impl `/`); prod `SameSite=None` CSRF review; `__Host-` prefix decision.
21. bcrypt >72-char password, null bytes, >254-char email, whitespace-only password registration edges.
22. Full field-shape contract tests for `/me` (avatarUrl/languageCode/onboardingData), org list (`memberCount`), members sort/pagination.
23. Request-validation sweeps beyond auth (missing Content-Type, unknown-fields-ignored, SQLi/XSS in org/workspace name fields).
24. Concurrency suite: duplicate-email race, concurrent member add, org-delete-vs-member-add, concurrent privacy-request creation (DB unique constraints).
25. Data-integrity assertions: UTC ISO-8601 timestamp formats, audit-log write verification for admin actions, Redis plan-cache behavior (`plan_loader.go` untested).


---

## 5. Coverage Summary by Area

| Area | Scenario groups | ✅ Covered | 🟡 Partial | ❌ Missing |
|---|---|---|---|---|
| Auth (7 sub-features + 7 scenarios + edge cases + cross-cutting) | ~30 | 8 | 15 | 7 |
| User-profile (16 endpoints + 6 scenarios + edge cases) | ~24 | 4 | 16 | 4 |
| Org-workspace (12 sub-features + 5 scenarios + edge cases + cross-cutting) | ~20 | 1 | 15 | 4 |

**Overall:** Backend auth/session/token lifecycle and RBAC/role-constraint logic are genuinely well covered (integration depth: DB-state verification, hash lookups, revocation). The weakest spots are **HTTP-level coverage of `/me/usage` (none), soft-delete auth flows (none), concurrency (none), frontend organization feature unit tests (none), and the entire e2e auth layer being mock-based and inconsistent with the cookie architecture**. Several spec docs also describe behavior the code doesn't implement (slug auto-gen/suffixing, plan org/member limits, 409 `ORG_HAS_CHATBOTS`, merge-onboarding, 409 `ALREADY_MEMBER`, remember-me) — these need spec↔code reconciliation, not just tests.

---

## 6. Appendix — Spec ↔ Implementation Mismatches Discovered

These are cases where the docs and the code disagree; tests currently pin the implementation (or nothing). Each needs a product decision.

| Topic | Spec says | Implementation does | Reference |
|---|---|---|---|
| Tokens in response body | Cross-cutting 2.2: "access/refresh token never exposed in response body (only in cookie)" | Register/login/refresh return `{token, refresh_token}` JSON **and** set cookies; integration tests assert body tokens present | auth 02 §2.2 vs `tokenResp` in auth_test.go |
| Register/login response shape | `{ "user": {id, email, fullName, createdAt} }` | `{token, refresh_token}` (no user object) | auth §1.6 |
| SameSite | `SameSite=Lax` | Lax in dev, **None** in prod (pinned by `TestAuth_CookieSecureFlag_Production`) | auth 1.2.2 |
| Cookie Path | `Path=/api/v1` | `Path=/` (pinned by `TestAuth_Cookies`) | auth 02 §2.2 |
| Register rate limit | 5 req/min | 3 req/min (`TestRateLimitMiddleware_AuthEndpointsStrictLimits`) | auth 1.2.1 / S5 |
| Slug auto-generation | Server auto-generates from name; unique suffix `-1` on collision (201) | 400 when slug missing; 409 bare body on collision; FE dialog generates slug (untested) | org 1.2.2, S1-step-3, S5 |
| Default workspace on org create | Created inside new org | Only created during registration, not in `CreateOrganization` | org 1.2.2 |
| Duplicate member add | 409 `ALREADY_MEMBER` | `ON CONFLICT DO UPDATE` upsert (role silently updated) | org 1.2.6 / §1.5 |

| Last workspace delete | 409 `LAST_WORKSPACE` | 400 (pinned by SVC-008 + handler unit) | org 1.2.12 / §1.6 |
| Org delete with chatbots | 409 `ORG_HAS_CHATBOTS` | No chatbot check; hard delete (only last-org guard) | org 1.2.4 / §1.6 |
| Plan org/member/workspace limits | Enforced (409 with plan-limit message) | No such quotas exist in `plan_limits.go`; nothing enforced | org 1.2.2/1.2.6/1.2.10; auth 02 §2.1 |
| Onboarding update | Merge partial data; `null` data = no-op | Wholesale replace; `null` → empty struct (likely wipes data) | user-profile 1.2.5 / §1.5 |
| Onboarding complete | POST without body | Requires `bot_id` (400 without) | user-profile 1.2.7 / S1 |
| Consent unknown type | 400 | Silently ignored (fixed struct decode) | user-profile 1.2.9 / S2 |
| Download not-ready export | 404 | 400 | user-profile 1.2.16 |
| Deleted-user login | 403 `ACCOUNT_DELETED` | Login query lacks `deleted_at` filter; anonymized email → 401; middleware 403 only on subsequent calls | auth 1.2.2 / S6 |
| Deleted-user `/me` | 401 | `DeletedAccountMiddleware` returns 403 `ERR_ACCOUNT_DELETED` | auth 1.2.6 |
| Re-register after deletion | 409 (email retained for GDPR) | Email anonymized → re-register succeeds | user-profile S4 |
| Logout idempotency | 200 without cookies | 400 (pinned by `TestAuth_LogoutHandler_CookieSecureFlag`) | auth 1.2.4 |
| Access-token invalidation on logout | Old access token → 401 | Stateless JWT remains valid until expiry (no denylist) | auth S1 Step 4 |
| Multi-tab logout detection | BroadcastChannel redirect to /login | No BroadcastChannel in production code; e2e self-mocks only | auth S7 |
| Remember-me | Refresh expiry set appropriately | No backend parameter; e2e mock/localStorage-only | auth 1.2.2 |

