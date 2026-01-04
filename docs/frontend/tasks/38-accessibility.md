# Task: Implement Accessibility Tests

> **Task ID**: 38-accessibility  
> **Source**: TEST_PATHS.md Section 12  
> **Priority**: Medium-Low (Quality Assurance)  
> **Estimated Effort**: 8-10 hours  

---

## Detailed Prompt

Implement E2E tests for Accessibility including keyboard navigation, screen reader support, and color contrast.

### Reference Specifications (Section 12)

**Keyboard Navigation:**
- Tab order follows logical sequence with visible focus indicator
- Skip links appear on page load, skip to main content
- Focus management: Modal opens → focus inside, Modal closes → focus returns
- Keyboard shortcuts: Ctrl+Enter submit, Escape close modal/dropdown, Arrow keys navigate menus, Space/Enter activate buttons
- Custom elements: Buttons keyboard activatable, Dropdowns arrow key navigation, Tabs arrow key navigation, Sliders arrow key adjustment

**Screen Reader:**
- ARIA labels: Buttons have label text, Inputs have associated label, Images alt text, Links descriptive text, Icons aria-label
- Live regions: Toast aria-live="polite", Errors aria-live="assertive", Updates announced
- Semantic HTML: Headings h1>h2>h3 hierarchy, Lists ul/ol>li, Tables th with scope, Forms fieldset>legend, Buttons <button> not <div>
- Dynamic content: Loading, Success, Error, New content all announced
- Interactive states: Expanded/collapsed aria-expanded, Selected aria-selected, Checked aria-checked, Disabled aria-disabled, Hidden aria-hidden

**Color Contrast:**
- Text contrast meets WCAG AA: Normal text 4.5:1, Large text 3:1, UI components 3:1
- Focus indicators: 2px solid outline, 3:1 contrast, 2px offset
- Error states: Red text 4.5:1, Red background 3:1, Icon indicators complementary
- Dark mode: Contrast maintained, Text readable, Focus visible

### Implementation Requirements

1. `frontend/e2e/accessibility.spec.ts`
2. `frontend/e2e/utils/accessibility-helper.ts`
3. Use Playwright's built-in accessibility testing

---

## Implementation Plan

- Keyboard navigation tests
- Focus management tests
- Screen reader ARIA tests
- Semantic HTML tests
- Color contrast tests
- Focus indicator tests

---

## Dependencies

- **Prerequisites**: All UI component tests

---

## Related Tasks

- 37-edge-cases.md - Error handling
- 39-performance.md - Performance tests

---

*Task created from: docs/frontend/TEST_PATHS.md Section 12*
