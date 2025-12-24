# Plan: Comprehensive Feature Audit & Documentation Update

This plan outlines the steps to perform a deep-dive audit of the Botla-co project to produce an exhaustive and accurate `existing_features.md` document.

## Phase 1: Database and Backend Audit
- [x] Task: Audit database migrations (`db/migrations/`) and `internal/db/` to map out the schema, plan configurations (Free, Pro, Ultra), and defined limits.
- [x] Task: Audit `internal/models/` and `pkg/config/` for plan-specific constants, model availability matrices, and feature flags.
- [x] Task: Investigate `internal/rag/`, `internal/scraper/`, and `internal/processing/` to document chunking strategies, scraping depth, and RAG behaviors.
- [x] Task: Audit `internal/api/` and `internal/services/` for hidden "small features" like smart naming, execution logs, and handoff logic details.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Database and Backend Audit' (Protocol in workflow.md)

## Phase 2: Frontend and Widget Audit
- [x] Task: Audit `frontend/src/` components and state management to identify all dashboard-controlled settings and user workflows.
- [x] Task: Audit `widget/src/` to document embed security, UI customization options, and end-user interactions (feedback, sources).
- [x] Task: Audit `data/sentences/` and frontend localization logic to confirm multi-language support details.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Frontend and Widget Audit' (Protocol in workflow.md)

## Phase 3: Documentation Synthesis & Finalization
- [x] Task: Consolidate all technical findings into a draft "Full Documentation" style `existing_features.md`.
- [x] Task: Cross-reference the draft with the codebase to ensure no specific values (limits, model IDs) were missed.
- [x] Task: Finalize and replace the content of `conductor/existing_features.md`.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Documentation Synthesis & Finalization' (Protocol in workflow.md)
