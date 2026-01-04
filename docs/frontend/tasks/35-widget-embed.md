# Task: Implement Widget Embed Tests

> **Task ID**: 35-widget-embed  
> **Source**: TEST_PATHS.md Section 10.1  
> **Priority**: Medium (Widget Integration)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Widget Embed Code Generation including script tag, iframe, React component options and configuration.

### Reference Specifications (Section 10.1)

- Deploy tab shows embed code section with generated code
- Embed code options: Script tag (default) with copy button, Iframe tag, React component
- Configuration options: Position, Theme color, Welcome message, Language, Custom branding
- Preview widget modal showing widget in isolation with chat interaction
- Test on site opens test page with embed
- Copy embed code shows success toast

### Implementation Requirements

1. `frontend/e2e/deploy-embed.spec.ts`
2. `frontend/e2e/pages/deploy-embed.page.ts`
3. `frontend/e2e/mocks/deploy.mocks.ts`

---

## Implementation Plan

- Embed code display tests
- Code type selection tests
- Configuration options tests
- Preview widget tests
- Copy code tests
- Test on site tests

---

## Dependencies

- **Prerequisites**: 11-chatbots-detail.md (Deploy tab)

---

## Related Tasks

- 11-chatbots-detail.md - Chatbot detail with Deploy tab
- 36-widget-preview.md - Widget configuration preview

---

*Task created from: docs/frontend/TEST_PATHS.md Section 10.1*
