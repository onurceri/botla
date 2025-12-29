# Frontend Task 003: Refactor TypeScript Any Types

## Background
Multiple mutation hooks in `src/hooks/mutations/useChatbotMutations.ts` use `any` for payloads. This bypasses type safety and can lead to runtime errors if API contracts change.

**Files:** `src/hooks/mutations/useChatbotMutations.ts`

## Integration Plan
1.  **Define Interfaces**
    - Check `src/types` or `src/api` for existing types for `ChatbotUpdateRequest`, `BasicInfoRequest`, etc.
    - If missing, create them based on the API definition.

2.  **Update Hooks**
    - Replace `(payload: any)` with `(payload: SpecificType)`.
    - Fix any type errors that arise (this catches potential bugs!).

3.  **Verify**
    - Compile project with `tsc --noEmit` (or just check IDE).

## Checklist
- [x] Identify all usages of `any` in `useChatbotMutations.ts`
- [x] Import or define correct request interfaces
- [x] Apply types to mutation functions
- [x] Verify type checking passes

