# Widget Task 006: Sanitize Regex Improvement

## Background
The user-agent based regex for removing `<script>` tags in `widget/src/utils/sanitize.ts` is complex and potentially bypassable. `DOMParser` provides a standard browser-based way to parse and sanitize HTML safe from most XSS vectors.

**File:** `widget/src/utils/sanitize.ts`
**Location:** Line 101

## Integration Plan
1.  **Switch to DOMParser**
    - Use `new DOMParser().parseFromString(html, 'text/html')`.
    - Traverse the DOM and remove `<script>`, `<iframe>`, etc. nodes.
    - Return `body.innerHTML` (or `textContent` if stripping tags).

2.  **Fallback**
    - Keep regex for environments where DOMParser might be missing (unlikely in browsers, but maybe test environment).

3.  **Verify**
    - Test with tricky vectors: `<script src=x>` inside other tags, etc.

## Checklist
- [ ] Implement `DOMParser` based sanitization
- [ ] Replace or augment existing regex
- [ ] Update XSS tests
