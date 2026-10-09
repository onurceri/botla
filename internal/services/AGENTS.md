# AGENTS.md - internal/services

OVERVIEW: Business logic layer implementing chat, chatbot, RAG, analytics, handoff, and plan management.

## WHERE TO LOOK

- **Chat flow**: `chat_service.go`, `chat_pipeline.go`, `chat_context_builder.go`
- **LLM behavior rules**: `chat_prompts.go` (DO NOT, NEVER, ALWAYS patterns)
- **AI models**: `model_registry.go` (resolves model names to API IDs)
- **Token quotas**: `quota_enforcer.go` (reserves/adjusts usage)
- **Human handoff**: `handoff_service.go`; operator emails go through the `EmailSender` interface (`email_sender.go`). No mail provider is integrated yet, so the default `NoopEmailSender` only logs and `email_sent` is false
- **Guardrails**: `guardrail_service.go` (topic restrictions)
- **Analytics**: `analytics_service.go` (usage tracking)

## CONVENTIONS

- Services receive `*sql.DB` and `*logger.Logger` in constructors
- Business logic stays in services, handlers remain thin
- Use `pkgerrors.Wrapf` for error context
- Language prompts always in English (`chat_prompts.go` line 10)
- Channel buffer sizes: 512 for work channels, 256 for result channels
- Context always passed as first parameter

## ANTI-PATTERNS

- DON'T put database queries in services; delegate to `internal/db`
- DON'T return raw SQL errors to users; wrap with user-friendly messages
- DON'T skip quota reservation before chat operations
- DON'T hardcode model names; use `model_registry.go` lookup
- DON'T send email directly in handoff; implement `EmailSender` and inject it via `HandoffService.EmailSender`
