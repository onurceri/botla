> Extracted from: user-profile.md — Section 1.4

# 1.4 Test Scenarios

## Scenario 1: Complete Onboarding Flow

```
  - [ ] Register new user
  - [ ] GET /api/v1/me/onboarding → { "completed": false, "currentStep": 0, "skipped": false }
  - [ ] PUT /api/v1/me/onboarding with { "currentStep": 1, "data": { "botName": "My Bot" } }
  - [ ] Assert: 200 OK
  - [ ] GET /api/v1/me/onboarding → { "currentStep": 1, "data": { "botName": "My Bot" } }
  - [ ] PUT /api/v1/me/onboarding with { "currentStep": 2, "data": { "botName": "My Bot", "personality": "friendly" } }
  - [ ] POST /api/v1/me/onboarding/complete
  - [ ] GET /api/v1/me/onboarding → { "completed": true, "currentStep": 2, "skipped": false }
```

## Scenario 2: Privacy Consent Lifecycle

```
  - [ ] GET /api/v1/me/privacy/consents → default consents (all true or all false depending on policy)
  - [ ] PATCH /api/v1/me/privacy/consents with { "marketing": false, "analytics": true }
  - [ ] Assert: 200 OK
  - [ ] GET /api/v1/me/privacy/consents → { "marketing": false, "analytics": true }
  - [ ] PATCH /api/v1/me/privacy/consents with { "unknown_type": true }
  - [ ] Assert: 400 Bad Request
```

## Scenario 3: Privacy Request Workflow (Export → Download)

```
  - [ ] POST /api/v1/me/privacy/export → 201 Created with request ID
  - [ ] GET /api/v1/me/privacy/requests/{id} → status: "pending"
  - [ ] Admin processes the export request (via admin endpoint or background job)
  - [ ] GET /api/v1/me/privacy/requests/{id} → status: "completed", downloadUrl present
  - [ ] GET download URL → 200 with JSON/ZIP file
  - [ ] Verify export JSON contains user profile, chatbot list, source list
```

## Scenario 4: Account Deletion Request Workflow

```
  - [ ] POST /api/v1/me/privacy/delete → 201 Created
  - [ ] GET /api/v1/me/privacy/requests → list contains deletion request
  - [ ] User still can use the app (request is pending admin processing)
  - [ ] Admin processes deletion
  - [ ] GET /api/v1/me → 401 (soft-deleted)
  - [ ] POST /api/v1/auth/login with same credentials → 403 Forbidden
  - [ ] Attempt to register again with same email → 409 Conflict (email retained for GDPR)
```

## Scenario 5: Tenant Isolation on Privacy Data

```
  - [ ] User A creates a privacy request
  - [ ] User B (different user, same org) calls GET /api/v1/me/privacy/requests → does not see User A's requests
  - [ ] User B calls GET /api/v1/me/privacy/requests/{userA_request_id} → 404
  - [ ] User B calls DELETE /api/v1/me/privacy/requests/{userA_request_id} → 404
```

## Scenario 6: Onboarding Skip and Re-complete

```
  - [ ] POST /api/v1/me/onboarding/skip → 200
  - [ ] GET /api/v1/me/onboarding → { "skipped": true, "completed": false }
  - [ ] POST /api/v1/me/onboarding/complete → 200 (idempotent, overrides skip)
  - [ ] GET /api/v1/me/onboarding → { "completed": true, "skipped": false }
```
