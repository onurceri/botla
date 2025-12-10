# Plan: Frontend - Show Hidden Limits

## Problem
The Backend enforces limits for "Monthly Ingestions" (max 50) and "Monthly Embedding Tokens" (max 250k). However, the Frontend's `PlanPage.tsx` does not display these limits or the user's current usage. This leads to a poor user experience where ingestions might fail silently or with generic errors without the user knowing they hit a limit.

## Analysis
- **File:** `frontend/src/pages/PlanPage.tsx`
- **Data Source:** `/api/v1/me` endpoint returns `usage` object.
- **Missing Data:** The `Usage` interface in FE and the UI cards need to be updated to show these specific counters.

## Proposed Changes

### 1. Update `Usage` Interface
In `frontend/src/pages/PlanPage.tsx`:
```typescript
interface Usage {
  // ... existing fields
  ingestions_count?: number // Need to confirm exact field name from BE response
  embedding_tokens_used?: number // Need to confirm exact field name
}
```
*Verification needed:* Check `internal/api/handlers/me.go` to see what JSON fields are actually returned for usage.

### 2. Update `PlanConfig` Interface
Add the missing limit fields:
```typescript
interface PlanConfig {
  // ...
  max_monthly_ingestions: number
  max_monthly_embedding_tokens: number
}
```

### 3. Update UI
- Add a new section or row in the "Web Scraping" or "Files" card (or a general "System Limits" card) to display:
    - **Monthly Ingestions:** `usage.ingestions_count` / `plan.max_monthly_ingestions`
    - **Embedding Tokens:** `usage.embedding_tokens` / `plan.max_monthly_embedding_tokens`
- Use the `Progress` component to visualize usage.

## Verification Plan
1.  **Mock Data:** Temporarily hardcode high usage values in the FE to verify the UI renders correctly.
2.  **End-to-End:** Check the actual `/api/v1/me` response in the browser network tab to ensure the fields match.
