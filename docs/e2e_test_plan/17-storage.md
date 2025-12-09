# 17. Storage Tests

> **Priority**: Medium  
> **Test Count**: 6  
> **Source Files**: `pkg/storage/`

---

## 17.1 R2 Integration

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| R2-001 | Upload small file | Key generated | ⬜ |
| R2-002 | Download by key | Content returned | ⬜ |
| R2-003 | Delete by key | File removed | ⬜ |
| R2-004 | Key isolation (org/bot scoped) | No cross-access | ⬜ |
| R2-005 | Missing key download | 404 returned | ⬜ |
| R2-006 | Storage used MB tracking | Usage calculated | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/r2_env_negative_test.go` | Env missing |
| `internal/integration/storage_usage_test.go` | R2-006: Storage usage tracking |
