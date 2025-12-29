# Frontend Task 007: Refactor Global Redirect

## Background
`src/api/client.ts` uses a mutable global `_redirectToLogin` variable. This is bad for testing isolation as tests running in parallel (or sequentially without cleanup) can interfere with each other.

**File:** `src/api/client.ts`
**Location:** Lines 17-24

## Integration Plan
1.  **Use Context/Singleton Pattern**
    - Instead of a global export, consider using a singleton class or dependency injection if possible.
    - Or, ensure `_redirectToLogin` is reset strictly in test teardowns.

2.  **Alternatives**
    - Pass the redirect function into the API client creator (factory pattern).

3.  **Verify**
    - Run tests in parallel if possible.

## Checklist
- [ ] Analyze usages of `_redirectToLogin`
- [ ] Refactor to safer pattern (e.g. factory or strict reset)
- [ ] Update tests
