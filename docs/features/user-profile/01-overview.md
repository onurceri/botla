> Extracted from: user-profile.md — Section 1

# User & Profile Management — Feature Testing

> **Document Version**: 1.0
> **Last Updated**: July 2026
> **Scope**: User & Profile Management
> **Backend**: Go 1.25+ API
> **Test Frameworks**: Go `testing` (backend unit/integration), Playwright (frontend E2E)

---

# 1.1 Feature Overview

User profiles are managed through the `/api/v1/me` subtree. This includes profile retrieval, plan/usage information, onboarding state, and GDPR/KVKK privacy workflows (consents, data export, correction, deletion).
