# Specification: House Cleaning Audit & Roadmap

## Overview
This track focuses on a comprehensive "house cleaning" audit of the Botla-co project. Over time, the codebase has accumulated unused code, technical debt, and redundant configurations. The objective is not to perform the cleanup yet, but to meticulously identify, document, and categorize these issues to create a detailed remediation roadmap.

## Type
- [ ] Feature
- [ ] Bug Fix
- [x] Chore / Refactor Audit

## Scope

### Functional Requirements
- **Backend Audit (Go):** 
    - Identify unused handlers, services, and internal helpers.
    - Locate redundant `sqlc` queries or unused database fields in the schema.
    - Document technical debt such as "god objects," high coupling, or legacy error handling patterns.
- **Frontend/Widget Audit (React/Preact):**
    - Identify unused components, hooks, or utility functions.
    - Audit `package.json` for unused or redundant dependencies.
    - Document state management bloat or inconsistent styling patterns.
- **Infrastructure Audit:**
    - Identify redundant environment variables in `.env.example` and production templates.
    - Check for deprecated or unused Docker configurations/images.
    - Review database migrations for legacy structures that could be streamlined.

### Documentation Requirements
- **Central Debt Registry:** Create and populate `docs/technical_debt.md`.
- **Categorization:** Issues must be categorized by Module (Backend, Frontend, Widget, Infra), Severity (High, Medium, Low), and Issue Type (Unused Code, Tech Debt, Configuration).
- **Precision:** Each entry must include the file path and a specific description of the issue to ensure it is actionable.

## Acceptance Criteria
1.  **Creation of `docs/technical_debt.md`:** A structured document that serves as the project's technical debt source of truth.
2.  **Comprehensive Coverage:** The audit must cover at least 3 high-level areas: Backend, Frontend (Dashboard & Widget), and Infrastructure.
3.  **Actionable Roadmap:** The documented issues must be detailed enough that a developer can pick them up and perform the cleanup without further investigation.
4.  **No Functional Changes:** The track is complete when the audit is documented; no application code should be modified or deleted during this phase.

## Out of Scope
- Actual deletion of code or refactoring (this will be handled in subsequent tracks based on this roadmap).
- Performance benchmarking.
