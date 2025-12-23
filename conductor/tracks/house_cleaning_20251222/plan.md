# Plan: House Cleaning Audit & Roadmap

This plan outlines the steps to perform a comprehensive technical audit of the Botla-co project and establish a central debt registry.

## Phase 1: Preparation
- [ ] Task: Create Central Debt Registry `docs/technical_debt.md` with required structure.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Preparation' (Protocol in workflow.md)

## Phase 2: Backend Audit (Go)
- [ ] Task: Audit `internal/api` and `internal/services` for unused handlers and logic.
- [ ] Task: Audit `internal/db` and migrations for redundant queries or schema elements.
- [ ] Task: Audit `pkg/` for unused utilities or legacy patterns.
- [ ] Task: Document all Backend findings in `docs/technical_debt.md`.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Backend Audit (Go)' (Protocol in workflow.md)

## Phase 3: Frontend & Widget Audit (React/Preact)
- [ ] Task: Audit `frontend/src` and `widget/src` for unused components and hooks.
- [ ] Task: Audit `package.json` files for redundant dependencies.
- [ ] Task: Document all Frontend/Widget findings in `docs/technical_debt.md`.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Frontend & Widget Audit' (Protocol in workflow.md)

## Phase 4: Infrastructure & Configuration Audit
- [ ] Task: Audit root directory configuration files (Docker, Caddy, Env) for redundancy.
- [ ] Task: Document all Infrastructure findings in `docs/technical_debt.md`.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Infrastructure & Configuration Audit' (Protocol in workflow.md)

## Phase 5: Finalization
- [ ] Task: Review and categorize all entries in `docs/technical_debt.md` by severity and type.
- [ ] Task: Create a summary "Roadmap" section in `docs/technical_debt.md` for cleanup prioritization.
- [ ] Task: Conductor - User Manual Verification 'Phase 5: Finalization' (Protocol in workflow.md)
