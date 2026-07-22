> Extracted from: user-profile.md — Section 1.6

# 1.6 Expected Behavior

| Condition | HTTP Status | Response Body |
|-----------|-------------|---------------|
| Profile fetch success | 200 | `{ "id", "email", "fullName", "avatarUrl", "planId", "isPlatformAdmin", "languageCode", "onboardingData", "createdAt" }` |
| Plan fetch success | 200 | `{ "id", "name", "code", "limits": { ... } }` |
| Usage fetch success | 200 | `{ "messageCount", "tokenUsage", "chatbotCount", "sourceCount", "tokensRemaining", "periodStart", "periodEnd" }` |
| Onboarding get | 200 | `{ "completed": bool, "currentStep": int, "skipped": bool, "data": object }` |
| Onboarding update success | 200 | `{ "completed": bool, "currentStep": int, "skipped": bool, "data": object }` |
| Onboarding skip | 200 | `{ "message": "Onboarding skipped" }` or updated state |
| Onboarding complete | 200 | `{ "message": "Onboarding completed" }` or updated state |
| Consent list | 200 | `{ "consents": [{ "type": "...", "granted": bool, "updatedAt": "..." }] }` |
| Consent update success | 200 | `{ "consents": [{ "type": "...", "granted": bool, "updatedAt": "..." }] }` |
| Privacy export created | 201 | `{ "id": "...", "type": "export", "status": "pending", "createdAt": "..." }` |
| Privacy correction created | 201 | `{ "id": "...", "type": "correction", "status": "pending", "createdAt": "..." }` |
| Privacy deletion created | 201 | `{ "id": "...", "type": "deletion", "status": "pending", "createdAt": "..." }` |
| Privacy request list | 200 | `{ "requests": [{ "id", "type", "status", "createdAt", "updatedAt" }], "total": int }` |
| Privacy request get | 200 | `{ "id", "type", "status", "description", "createdAt", "updatedAt", "downloadUrl": "..." }` |
| Privacy request delete | 200 | `{ "message": "Request deleted" }` |
| Export download | 200 | File with `Content-Type: application/json` or `application/zip` |
| Unauthenticated (any) | 401 | `{ "error": { "code": "UNAUTHORIZED" } }` |
| Not found (privacy request) | 404 | `{ "error": { "code": "NOT_FOUND" } }` |
| Duplicate export/deletion | 409 | `{ "error": { "code": "DUPLICATE_REQUEST" } }` |
