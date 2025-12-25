# KVKK Compliance (Phase 4) — Remaining Parts / Gaps

This document lists remaining work and correctness gaps identified by comparing:
- `docs/plans/admin-plans/04-kvkk-compliance.md` (implementation plan)
- Current backend/frontend code and migrations in this repository (current working tree)

## Status Summary

Core building blocks exist (tables, basic handlers, export/anonymize logic, UI pages, some tests), but the end-to-end KVKK flows described in the plan are not fully correct or complete.

## P0 (Blocking) — Correctness / Crash / Broken Flow

### [x] Export flow does not match plan (no export privacy request lifecycle)
- Plan expects: user requests export → request appears in admin → admin approves → export generated → user gets download link.
- Current behavior: user export request immediately triggers export generation and creates `data_exports`, without creating a `privacy_requests` row of type `export`.
- Impact: admin privacy queue cannot manage export requests as designed; audit trail for export requests is incomplete.
- Acceptance criteria:
  - User export request creates `privacy_requests` (`request_type='export'`, `status='pending'`).
  - Admin approval triggers export generation.
  - Request status transitions are consistent (`pending` → `processing` → `completed/denied`).

### [x] No user-accessible download mechanism for exports (download_url is a storage key)
- Storage `UploadFile` returns an object key, not a signed URL.
- `data_exports.download_url` stores that key, but there is no endpoint to download the export by key, and no signed URL generation.
- Impact: “download link” described in plan and docs is not actually deliverable.
- Acceptance criteria:
  - Provide a secure way for the user (or admin) to download the export (signed URL or authenticated download endpoint).
  - Expiration is enforced (by signed URL TTL and/or server-side checks).

### [x] Potential panics when storage is not configured (nil storage)
- Privacy export path calls `Storage.UploadFile` without guarding `Storage != nil`.
- Retention job calls `Storage.DeleteFile` without guarding `Storage != nil`.
- Impact: runtime panic in environments without R2 configured.
- Acceptance criteria:
  - Exports fail gracefully when storage is unavailable.
  - Retention job handles missing storage gracefully (skip storage deletions or log and continue).

### [x] Consent updates can silently fail due to invalid IP format and ignored DB errors
- DB column type: `INET`. Handler uses `r.RemoteAddr` (often `IP:PORT`), which is not valid `INET`.
- Handler ignores errors from consent upserts.
- Impact: user consent changes may not persist; compliance logging becomes unreliable.
- Acceptance criteria:
  - Correctly parse IP (or store as text if proxies make it unreliable).
  - Surface/handle errors from consent persistence.

## P1 (High) — UI/API Mismatches and Missing Data in Admin Tools

### [ ] Admin Privacy page does not match backend schema
- Frontend expects fields like `details`, `rejection_reason`, and uses mismatched `request_type` values (e.g., `export_data`) and invalid statuses (e.g., `failed`) for `privacy_requests`.
- Backend returns fields like `user_email`, `reason`, `denial_reason`, `request_type` (`export|deletion|correction`), `status` (`pending|processing|completed|denied`).
- Impact: admin UI can show incorrect labels, missing information, and filters that cannot work.
- Acceptance criteria:
  - Align admin UI types and rendering with actual backend response.
  - Show `user_email`, `reason`, `denial_reason`, and correct request types.
  - Remove or implement any unsupported statuses.

### [ ] User privacy page missing “third_party” consent toggle
- Plan includes `third_party` consent type.
- UI currently exposes marketing/analytics/personalization but not third_party.
- Acceptance criteria:
  - UI supports all consent types present in schema and plan, including `third_party`.

### [ ] Users may see undefined consent states (no defaults)
- Backend returns only recorded consents; missing consent types are not defaulted.
- Plan verification expects “new users have default consents”.
- Acceptance criteria:
  - Define and enforce default consents (DB seed, on-user-create initialization, or API defaulting).
  - Ensure UI receives a complete consent map for all known types.

## P2 (Medium) — Retention and Lifecycle Completeness

### [ ] Retention job: audit log retention not implemented
- Config includes `AuditLogRetentionDays` but it is not used.
- Acceptance criteria:
  - Implement retention for audit logs if required by policy.
  - Ensure retention periods match documented policy.

### [ ] Retention job deletes exports by `created_at`, ignores `expires_at`
- Export records have `expires_at`, but cleanup uses `created_at` cutoff.
- Acceptance criteria:
  - Cleanup logic respects expiration semantics (`expires_at < now`) and/or consistent policy.

### [ ] Export → privacy_requests linkage fields are unused
- `privacy_requests` table has `export_url` and `export_expires_at`, but current export completion writes to `data_exports` only.
- Acceptance criteria:
  - Either populate these fields (and use them) or remove them and standardize on `data_exports`.

## P2 (Medium) — Functional Coverage Gaps

### [ ] “Correction” request type exists but has no implementation
- Schema allows `request_type='correction'`.
- No endpoint or processing logic exists for correction.
- Acceptance criteria:
  - Implement correction request workflow or remove it from constraints if not supported.

## P3 (Low) — Tests and Documentation Consistency

### [ ] Integration tests do not validate export completion and download
- Current privacy integration test checks status codes but not that an export becomes downloadable.
- Acceptance criteria:
  - Tests verify: request creation → approval → export completion → download works (or signed URL is returned and valid).

### [ ] Retention integration test quality issues
- The retention export test attempts schema creation/insert patterns that do not reflect actual migrations.
- Acceptance criteria:
  - Tests should operate on the real migrated schema and validate real retention behavior.

### [ ] Docs conflict with implementation (retention schedule/period, notifications)
- Example issues:
  - Docs claim a 2-year retention default and job time of 03:00.
  - Code defaults differ (90/30/365/7) and scheduling is “every 24h from startup”.
  - Docs mention emailing users when export is ready; there is no email notification implementation.
- Acceptance criteria:
  - Align docs with actual behavior, or align behavior to policy and docs.

## Notes / Clarifications Needed (Before Declaring Phase Complete)
- Decide the “source of truth” for export artifacts:
  - `privacy_requests.export_url` vs `data_exports.download_url`
  - Whether exports are always admin-approved or auto-processed for users
- Decide how to implement “download link” securely:
  - signed URL generation or authenticated streaming endpoint
- Decide consent defaults policy (opt-in/opt-out) for each consent type
