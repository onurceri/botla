> Extracted from: user-profile.md — Section 1.5

# 1.5 Edge Cases

| Edge Case | Expected Behavior |
|-----------|------------------|
| Usage exceeds plan limit | Return zero or negative `tokensRemaining`; quota enforcement blocks new operations |
| Onboarding data with special characters | Stored and returned as-is; properly escaped in JSON |
| Privacy request download link expiry | Link expires after configurable period; expired link returns 410 Gone |
| Concurrent privacy request creation | DB unique constraint on pending requests; second request returns 409 |
| User has no chatbots/usage yet | Usage endpoint returns zero values, not errors |
| Extremely large consent payload | Rejected with 413 Payload Too Large (MaxBytes middleware) |
| Download export for non-export request type | 400 Bad Request |
| Delete already-deleted privacy request | 404 Not Found |
| Update onboarding with null data | Treated as no-op; existing data preserved |
