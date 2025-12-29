# Frontend Task 005: Add Error Boundaries

## Background
The application lacks a top-level Error Boundary. Uncaught JS errors in components can crash the entire React component tree, showing a blank screen.

**File:** `src/App.tsx`

## Integration Plan
1.  **Create ErrorBoundary Component**
    - Create `src/components/ErrorBoundary.tsx`.
    - Implement standard React Error Boundary (catch error, show fallback UI).

2.  **Wrap App**
    - In `src/main.tsx` or `src/App.tsx`, wrap the main routes or the entire app with `<ErrorBoundary>`.

3.  **Verify**
    - Intentionally throw an error in a component (temporary).
    - Verify fallback UI is shown instead of white screen.

## Checklist
- [ ] Create `ErrorBoundary` component
- [ ] Integrate into `App.tsx`
- [ ] Verify error catching behavior
