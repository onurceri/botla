# Remaining Admin Dashboard Tasks

This document outlines the unfinished tasks for the Botla Admin Dashboard. Although initial routes and files were created, some additional features may still be needed.

## ✅ Completed Tasks (2025-12-25)

### Phase 1: Backend API Enhancements
- [x] **Chatbot Management API**
    - [x] Created `AdminListChatbots` handler to return a paginated list of every chatbot on the platform.
    - [x] Created `ForceRefreshChatbot` endpoint to allow admins to manually trigger a re-index for a specific bot.
- [x] **Data Source Monitoring API**
    - [x] Created `AdminListSources` handler to monitor the status (Ready, Processing, Failed) of all data sources.
    - [x] Created `ReprocessSource` endpoint to help recover stuck or failed indexing jobs.
    - [x] Created `GetSourceStats` endpoint to get aggregated statistics by status.

### Phase 2: Frontend API Client (admin.ts)
- [x] **API Functions**
    - [x] Added `listChatbots(params)` to `frontend/src/api/admin.ts`.
    - [x] Added `listSources(params)` to `frontend/src/api/admin.ts`.
    - [x] Added `forceRefreshChatbot(id)` and `reprocessSource(id)` functions.
    - [x] Added `getSourceStats()` function.
- [x] **Type Safety**
    - [x] Defined strict TypeScript interfaces for `AdminChatbot`, `AdminSource`.

### Phase 3: UI Implementation (Replacing Placeholders)
- [x] **Admin Users Page (`AdminUsersPage.tsx`)**
    - [x] Implemented a dynamic table showing Email, Plan, Status, and Created Date.
    - [x] Added a search bar to filter users by e-mail.
    - [x] Added dropdown filters for "Plan".
    - [x] Added an "Actions" menu to toggle admin status.
- [x] **Admin Organizations Page (`AdminOrganizationsPage.tsx`)**
    - [x] Implemented a card-based grid showing Name, Plan, and creation date.
    - [x] Added a search bar to filter by organization name.
    - [x] Added plan filter dropdown.
- [x] **Admin Chatbots Page (`AdminChatbotsPage.tsx`)**
    - [x] Built the full management UI.
    - [x] Created a table listing all platform bots with their parent organization and status.
    - [x] Added a "Force Refresh" button for troubleshooting bot indexing issues.
- [x] **Admin Sources Page (`AdminSourcesPage.tsx`)**
    - [x] Built the full monitoring UI.
    - [x] Created a status board for data sources showing stats by status.
    - [x] Included an action to "Reprocess" failed or stuck jobs.

### Phase 4: UI Components
- [x] **DropdownMenu Component**
    - [x] Created reusable `components/ui/DropdownMenu.tsx` using Radix UI.

---

## 🔄 Remaining Tasks (Lower Priority)

### UX, Navigation & Polish
- [ ] **Detail Pages**
    - [ ] Implement `AdminUserDetailPage` to show deep-dive statistics and bot ownership.
    - [ ] Implement `AdminOrgDetailPage` to manage members and see organization-wide usage.
- [ ] **Feedback Systems**
    - [ ] Standardize "Toast" notifications across all admin pages for every action.
- [ ] **Privacy/KVKK Enhancements**
    - [ ] Improve the "Privacy Request" dialog to show user context and history before an admin approves data deletion or export.

### Backend Enhancements (Future)
- [ ] **Detailed Aggregation API**
    - [ ] Implement `GetAdminUserDetail` logic to fetch aggregated stats like total tokens used, number of bots, and last activity timestamp.
    - [ ] Implement `GetAdminOrgDetail` logic to fetch member lists and workspace summaries for a specific organization.

---

## Summary

The core admin dashboard functionality is now complete:
- ✅ Backend APIs for chatbots and sources management
- ✅ Frontend pages with full data integration
- ✅ Search, filtering, and pagination on all pages
- ✅ Admin actions (Force Refresh, Reprocess, Toggle Admin)
- ✅ Reusable DropdownMenu component

The remaining tasks are lower priority enhancements for the detail pages and improved UX polish.
