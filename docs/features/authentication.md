# Authentication

## Purpose
User registration and login, token issuance, refresh rotation, logout, and protected endpoints.

## User Flow
1. Register via `POST /api/v1/auth/register`.
2. Login via `POST /api/v1/auth/login` → receive `token` and `refresh_token`.
3. Store tokens in `localStorage` (`botla_token`, `botla_refresh_token`).
4. Access protected routes with `Authorization: Bearer <token>`.
5. Auto-refresh on 401 via `POST /api/v1/auth/refresh`.
6. Logout revokes refresh token.

## Backend Interfaces
- Register: `cmd/server/main.go:45`, `internal/api/handlers/auth.go:39-82`.
- Login: `cmd/server/main.go:46`, `internal/api/handlers/auth.go:84-116`.
- Refresh: `cmd/server/main.go:47`, `internal/api/handlers/auth.go:118-155`.
- Logout: `cmd/server/main.go:48`, `internal/api/handlers/auth.go:157-176`.
- Protected ping: `cmd/server/main.go:49`, `internal/api/handlers/auth.go:202-211`.
- Middleware: `pkg/middleware/auth.go:15-37`.
- Tokens: `internal/auth/jwt.go:15-45`.

## Frontend Interfaces
- Route guard: `frontend/src/App.tsx:11-18,30-39`.
- Auth hook: `frontend/src/hooks/useAuth.ts:4-45` (`signIn`, `signOut`, `protectedPing`).
- Axios client: `frontend/src/api/client.ts:8-16` (request), `:17-46` (response refresh). 
- Pages: `frontend/src/pages/LoginPage.tsx`, `frontend/src/pages/RegisterPage.tsx`.

## Error Handling
- 400 for invalid inputs; 401 for auth failures; 409 for existing email.
- Frontend refresh failures redirect to `/login`.

## Testing
- Backend: auth integration tests in `internal/integration/auth_*.go`.
- Frontend: `frontend/src/pages/__tests__/LoginPage.test.tsx` and added route guard tests.

