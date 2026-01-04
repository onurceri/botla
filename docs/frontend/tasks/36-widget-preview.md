# Task: Implement Widget Configuration Preview Tests

> **Task ID**: 36-widget-preview  
> **Source**: TEST_PATHS.md Section 10.2  
> **Priority**: Medium (Widget Integration)  
> **Estimated Effort**: 4-6 hours  

---

## Detailed Prompt

Implement E2E tests for Widget Configuration Preview including live preview, configuration form, and export/import.

### Reference Specifications (Section 10.2)

- Configuration form with: Position select, Color pickers (theme, header, bot message, user message), Font family select, Toggles (auto open, hide branding), Inputs (welcome message, bot display name, custom CSS)
- Live preview updates in real-time as config changes
- Preview widget interactions: Toggle open/close, Send test message
- Reset to defaults button restores all values
- Export configuration as JSON file
- Import configuration from JSON file
- Save configuration updates embed code

### Implementation Requirements

1. `frontend/e2e/deploy-preview.spec.ts`
2. `frontend/e2e/pages/deploy-preview.page.ts`
3. `frontend/e2e/mocks/deploy.mocks.ts`

---

## Implementation Plan

- Configuration form tests
- Live preview tests
- Configuration changes tests
- Reset defaults tests
- Export config tests
- Import config tests
- Save configuration tests

---

## Dependencies

- **Prerequisites**: 35-widget-embed.md (embed code)

---

## Related Tasks

- 35-widget-embed.md - Embed code generation

---

*Task created from: docs/frontend/TEST_PATHS.md Section 10.2*
