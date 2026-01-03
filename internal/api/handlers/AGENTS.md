# AGENTS.md - internal/api/handlers

OVERVIEW: HTTP request handlers for Botla API - 77 files organized by domain (auth, chatbot, chat, source, admin, org).

## WHERE TO LOOK

| Area | Files |
|------|-------|
| Auth handlers | `auth.go` - register, login, refresh, logout |
| Chatbot CRUD | `chatbot.go`, `chatbot_list.go`, `chatbot_item.go` |
| Sources | `source.go`, `source_create.go`, `source_refresh.go`, `source_chunks.go` |
| Chat | `chat.go` - protected and public endpoints |
| Admin | `admin.go`, `admin_health.go`, `admin_queues.go`, `admin_errors.go` |
| Organization | `organization.go` - org and workspace management |
| Public | `public.go` - unauthenticated chatbot chat |
| Router config | `internal/api/router/` - route registration |
| Auth guards | `internal/api/guards/admin.go` - admin-only routes |
| API spec | `api/openapi.yaml` - 534-line OpenAPI 3.0 spec |

## CONVENTIONS

- **Handler structs receive dependencies**: `ChatbotHandlers{ChatbotService, OrgService, Logger}`
- **Domain separation**: each feature has dedicated handler file
- **Route registration in router/**: `routes_auth.go`, `routes_chatbot.go`, etc.
- **Request types defined inline**: e.g., `createChatbotRequest` struct in `chatbot.go`
- **Error handling via `api.WriteErrorCode`**: error codes from `internal/api` package
- **Middleware chain**: `AuthMiddleware` → `ExtractTenantContext` → handler
- **Guards for admin routes**: `RequirePlatformAdmin` in `internal/api/guards/`

## ANTI-PATTERNS

- **Never bypass auth middleware**: `ProtectedHandler` pattern only for testing
- **Don't put business logic in handlers**: delegate to services layer
- **Avoid raw SQL**: use repository pattern (`repository/*_repo.go`)
- **Don't skip guards**: admin routes require `RequirePlatformAdmin`
- **Don't use `api.WriteErrorCode` for auth errors**: use `middleware.WriteAuthError` pattern
