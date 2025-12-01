# Analytics

## Purpose
Daily aggregates (messages, conversations) over last 7 days for a user’s chatbots.

## User Flow
1. Navigate to Analytics page.
2. Fetch series via API.
3. Visualize charts; view totals and KPIs.

## Backend Interfaces
- Endpoint: `cmd/server/main.go:86-88`, `internal/api/handlers/analytics.go:21-77`.
- SQL aggregates over `analytics` table joined with user’s chatbots.

## Frontend Interfaces
- Page: `frontend/src/pages/AnalyticsPage.tsx` with charts via `recharts`.
- API: `frontend/src/api/analytics.ts:3-5`.
- States: loading, empty, error, and filtered chart view.

## Testing
- Backend SQL aggregation tests.
- Frontend rendering and state transitions.

