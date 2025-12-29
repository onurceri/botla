# Frontend Task 004: Fix E2E Auth Bypass Typing

## Background
In `src/App.tsx`, the E2E check uses `@ts-ignore` to access `import.meta.env.VITE_E2E`. This suppresses type checking and is bad practice.

**File:** `src/App.tsx`
**Location:** Lines 52-53

## Integration Plan
1.  **Extend ImportMeta Usage**
    - Create or update `src/vite-env.d.ts` (or similar declaration file).
    - Add `readonly VITE_E2E: string | boolean` to `ImportMetaEnv` interface.

2.  **Remove ts-ignore**
    - Remove `// @ts-ignore` and the cast to `any`.
    - Access `import.meta.env.VITE_E2E` directly.

## Checklist
- [ ] Update TypeScript definitions for `ImportMetaEnv`
- [ ] Remove `@ts-ignore` in `src/App.tsx`
- [ ] Compile check
