# 10. Widget & Embed Tests

> **Priority**: High  
> **Test Count**: 14  
> **Source Files**: `widget/`, `internal/api/handlers/public.go`

---

## 10.1 Widget Loading

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| WGT-001 | Load widget from build | Script loads | ✅ |
| WGT-002 | Bubble renders in corner | Visible and clickable | ✅ |
| WGT-003 | Drawer opens on click | Chat UI visible | ✅ |
| WGT-004 | Config fetched from API | Theme applied | ✅ |
| WGT-005 | Shadow DOM isolation | Styles isolated | ✅ |

---

## 10.2 Secure Embed

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| SEC-001 | Token issuance | Valid token returned | ✅ |
| SEC-002 | Widget with embed_secret | Validates signature | ✅ |
| SEC-003 | Invalid signature | Widget fails to load | ✅ |
| SEC-004 | allowed_domains validation | Cross-origin blocked | ✅ |
| SEC-005 | secure_embed_enabled: false | No validation | ✅ |

---

## 10.3 Branding

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| BRD-001 | Default branding | "Powered by Botla" visible | ✅ |
| BRD-002 | hide_branding: true | Logo hidden | ✅ |
| BRD-003 | custom_branding.logo_url | Custom logo shown | ✅ |
| BRD-004 | custom_branding.text | Custom text shown | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/chatbot_secure_embed_test.go` | Secure embed config |
| `internal/integration/public_secure_embed_test.go` | Secure embed enforcement |
| `internal/integration/public_endpoints_test.go` | Public API |
| `frontend/e2e/widget-embed.spec.ts` | Basic widget flow |
| `frontend/e2e/widget-embed-secure.spec.ts` | Secure embed widget |
| `frontend/e2e/widget-branding.spec.ts` | Branding options |
