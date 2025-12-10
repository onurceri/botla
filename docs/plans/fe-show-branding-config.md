# Plan: Frontend - Show Branding Config

## Problem
The `PlanPage` details view currently lists storage, scraping, and chat limits, but does not explicitly show "Branding" capabilities (e.g., "Remove Branding", "Custom Branding") which are key selling points for Pro and Ultra plans.

## Analysis
- **File:** `frontend/src/pages/PlanPage.tsx`
- **Current UI:** "System Usage" card is sparse. "Branding" is only mentioned as a badge in the `ChatbotDetailPage`.
- **Requirement:** Add a visual indicator of Branding capabilities in the Plan Overview.

## Proposed Changes

### 1. Update UI Layout
- In the "Detailed Limits" section, add a "Branding & Customization" card (or add to an existing card).
- Display rows for:
    - **Remove 'Powered by Botla':** Checkmark if `planConfig.branding.can_hide_branding` is true, else "Pasif".
    - **Custom Branding (Logo/Link):** Checkmark if `planConfig.branding.can_custom_branding` is true, else "Pasif".

### 2. Implementation Details
- Reuse the `InactiveBadge` component.
- Add `branding` to the `PlanConfig` interface in `PlanPage.tsx`.

## Verification Plan
1.  **Visual Inspection:** Verify the new rows appear and correctly reflect the current plan's capabilities (Free vs Pro vs Ultra).
