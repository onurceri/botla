# 13. Plan & Quota Tests

> **Priority**: High  
> **Test Count**: 10  
> **Source Files**: `internal/db/plan.go`, plan config in `plans` table

---

## 13.1 Plan Limits

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| PLN-001 | max_chatbots limit | 403 when exceeded | ✅ |
| PLN-002 | max_files_per_chatbot limit | Error when exceeded | ✅ |
| PLN-003 | max_urls_per_chatbot limit | Error when exceeded | ✅ |
| PLN-004 | max_monthly_tokens limit | 429 when exceeded | ✅ |
| PLN-005 | max_monthly_ingestions limit | 429 when exceeded | ✅ |
| PLN-006 | allowed_models enforcement | Error for disallowed model | ✅ |
| PLN-007 | refresh.enabled: false | Refresh blocked | ✅ |
| PLN-008 | branding.can_hide: false | hide_branding blocked | ✅ |
| PLN-009 | Plan upgrade updates limits | New limits immediately | ✅ |
| PLN-010 | Free vs Paid plan features | Correct gating | ✅ |
