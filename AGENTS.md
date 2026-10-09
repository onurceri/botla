# AGENTS.md - Botla

Botla is an AI chatbot platform: it ingests content (URLs, sitemaps, PDFs, raw text), embeds it into Qdrant, and answers end-user questions via a RAG pipeline (OpenAI/OpenRouter LLMs). It ships a management dashboard and an embeddable chat widget. This file is the primary guide for the Go backend; frontend and widget have their own `AGENTS.md`.

## Monorepo Layout

This is an **npm-workspaces monorepo** with a Go backend at the root plus three JS workspaces:

| Path | Stack | Purpose |
|---|---|---|
| `cmd/`, `internal/`, `pkg/` | Go 1.25 | Backend HTTP API + ingestion workers |
| `frontend/` | React 19, Vite, TS, TanStack Query | Management dashboard |
| `widget/` | Preact, Vite | Embeddable chat widget (deployed to Cloudflare Pages) |
| `packages/ui-shared/` | TS | UI components shared by frontend + widget (`@botla/ui-shared`) |

Detailed conventions live in **per-area `AGENTS.md` files** — read the relevant one before working in that area: `frontend/AGENTS.md`, `internal/api/handlers/AGENTS.md`, `internal/services/AGENTS.md`, `internal/rag/AGENTS.md`, `internal/integration/AGENTS.md`.

## Backend Structure

```
├── cmd/
│   ├── server/          # HTTP API entrypoint (dependency wiring)
│   ├── cli/             # Admin CLI (promote/demote platform admins)
│   └── eval/            # RAG eval runner (see evals/README.md, `make eval`)
├── internal/            # Private application code
│   ├── api/             # HTTP layer
│   │   ├── handlers/    # Thin request handlers (auth, chatbot, chat, source, admin, org)
│   │   ├── router/      # Route registration + handler construction
│   │   └── guards/      # Admin-only route guards
│   ├── auth/            # JWT and password utilities
│   ├── db/              # DB connection pool setup ONLY (DSN, pooling)
│   ├── repository/      # Data access: interfaces + Postgres impls + mocks (squirrel)
│   ├── services/        # Business logic
│   ├── models/          # Domain models
│   ├── rag/             # RAG implementation (OpenAI/OpenRouter embeddings+LLM, Qdrant)
│   ├── processing/      # Source ingestion queue + per-type processors
│   ├── workers/         # Bounded worker pool for async/analytics jobs
│   ├── scraper/         # Web scraping (sitemap, CSS selectors, go-rod browser)
│   ├── pdf/             # PDF processing (requires `fitz` build tag)
│   ├── validation/      # Input validation
│   ├── evals/           # RAG eval framework: dataset, metrics, judge, harness, gate, runner
│   ├── integration/     # Integration tests
│   └── testdb/          # Database test utilities
├── pkg/                 # Public packages
│   ├── config/          # Env-driven configuration (loads .env)
│   ├── middleware/      # HTTP middleware (auth, CORS, rate limiting, plan loader)
│   ├── ratelimit/       # Redis-backed limiter
│   ├── storage/         # R2/S3 file storage
│   ├── tokenizer/       # Sentence-aware token counting
│   ├── urlutil/         # URL helpers + SSRF protection
│   └── logger/          # Structured logging
├── db/migrations/       # SQL migrations (golang-migrate)
├── api/openapi.yaml     # Public API spec
└── widget/              # Embeddable chat widget (Preact)
```

## Backend Architecture

Request flow is a strict layered pipeline. The wiring in `cmd/server/main.go` + `internal/api/router/router.go` is the fastest way to see how the pieces connect.

```
HTTP → middleware chain → router → handlers (thin) → services (business logic) → repository (SQL) → Postgres
                                                            ↓
                                              rag / processing / workers (async)
```

**Layer rules (enforced by convention — see per-area AGENTS.md):**
- `internal/api/handlers/` — thin HTTP only: parse/validate, delegate to services, respond via `api.WriteErrorCode`. No business logic, no raw SQL.
- `internal/services/` — business logic. Constructors receive repositories/interfaces + `*logger.Logger`; `context.Context` is always the first param. No SQL here — delegate to repositories.
- `internal/repository/` — data access. **There is NO ORM and NO sqlc.** Repositories hand-write SQL over `database/sql` using the **squirrel** query builder + `pgx` stdlib driver. Each repo has an interface in `repository/interfaces.go`, a `Postgres*Repo` implementation, and a hand-written `mock_*_repo.go` for tests.
- `internal/db/` — connection pool setup only: DSN building (simple_protocol to avoid prepared-statement collisions), pooling limits, schema override via `DB_SCHEMA`.
- `internal/models/` — domain types shared across layers.

**Dependency wiring:** `main.go` builds everything in phases (`initInfrastructure` → `initRateLimiting` → `initProcessing` → `initSchedulers`) into an `application` struct, then `router.New(...)` constructs all repos → services → handlers and registers routes. Adding a handler means threading it through `router.New`.

**Middleware chain order matters** (outermost first, `main.go` `start()`):
`RequestID → SecurityHeaders → Recovery → RequestLogger → MaxBytes(52MB) → PlanLoader → DeletedAccount → RateLimit → mux`. Routes then wrap individual handlers with `AuthMiddleware` → `DeletedAccountMiddleware` → `ExtractTenantContext`. Admin routes additionally require the `RequirePlatformAdmin` guard (`internal/api/guards/`). The `/api/v1/chatbots/` subtree uses a dispatch handler (`ChatbotsDispatchHandler`) to route sub-paths.

**Background processing:**
- `internal/processing/` — `SourceQueue` runs ingestion jobs (scrape/PDF/text → chunk → embed → upsert to Qdrant → summarize for suggestions) in a bounded worker pool. `url_processor.go`, `pdf_processor.go`, `text_processor.go` are the per-source-type processors.
- `internal/workers/` — separate `WorkerPool` for analytics/async jobs.
- Schedulers started in `main.go`: `RefreshScheduler` (re-ingest) and `RetentionJob` (daily at 03:00).

**RAG pipeline (`internal/rag/`):** interface-driven (`EmbeddingClient`, `VectorClient`, `LLMClient`). OpenAI embeddings (`text-embedding-3-small`, 25-chunk batches, rate-limited), Qdrant HTTP client, tiered similarity search (`search.go`, thresholds from env `RAG_TOPK`/`RAG_MAX_CONTEXT_TOKENS`), and all LLM calls routed through a gobreaker **circuit breaker** (`circuit_breaker.go`). LLM provider resolved via `client_factory.go` (OpenAI or OpenRouter); model names resolved through `services/model_registry.go`. Chunk IDs are deterministic (`MakePointID(sourceID, index)`).

**Cross-cutting concerns:**
- **Auth is HttpOnly-cookie based** (`botla_token` / `botla_refresh_token`, refresh-token rotation) — not localStorage/bearer. Frontend uses `withCredentials: true` and treats `/api/v1/me` as the auth source of truth. Don't reintroduce token-in-localStorage.
- **Multi-tenancy**: orgs → workspaces; tenant scope comes from `ExtractTenantContext` middleware.
- **Plan enforcement**: quotas/model access/feature gates checked centrally (`services/plan_service.go`, `quota_enforcer.go`, `models/plan_limits.go`). Plans are validated at startup — invalid plan config in the DB fails boot.
- **Config**: `pkg/config` loads env from `.env` (godotenv). All config is env-driven; see `.env.example`.
- **Rate limiting**: Redis-backed (`pkg/ratelimit` + `pkg/middleware`), plan-based with per-endpoint overrides. Redis is required to boot.
- **SSRF protection**: `pkg/urlutil` blocks private IP ranges during URL ingestion.
- **Logging**: `pkg/logger` structured logging everywhere.

## Dev Environment Setup

### Prerequisites
- Go 1.25.4+
- Docker & Docker Compose
- PostgreSQL 15+ (via Docker or local)
- Redis (via Docker or local)
- For PDF support: CGO enabled with `fitz` library

### Start Development Services
```bash
# Start PostgreSQL + Redis + Qdrant (docker-compose.dev.yml)
make up

# Run migrations
make migrate-up

# Connect to database
make psql
```

## Development Commands

The backend has two build modes. **PDF support requires CGO + the `fitz` build tag**; everything else builds without CGO.

### Running the Server
```bash
make be-run           # WITH PDF/OCR (CGO_ENABLED=1, -tags fitz)
make be-run-no-pdf    # without CGO
make be-run-test      # against the `test` schema (DB_SCHEMA=test)
```

### Building
```bash
make build            # go build ./...
make run              # go run ./cmd/server/main.go (no PDF)
```

## Frontend & Widget Commands

```bash
# Frontend dashboard
cd frontend
npm run dev           # :5173
npm run ci            # lint + typecheck + test:coverage  <-- run before committing
npm run e2e           # Playwright (needs dev server running)
npx vitest run src/pages/__tests__/LoginPage.test.tsx   # single test file
npx vitest run -t "should display"                      # by name

# Embeddable widget
cd widget
npm run dev           # :5174
npm run ci            # typecheck + lint + test:coverage
npm run build         # embeddable bundle
```

OpenAPI types are generated into the frontend from `api/openapi.yaml` via `cd frontend && npm run generate-types`.

## Database Operations

### Local Database Access
```bash
make psql             # connect to dev DB (botla/botla @ botla_dev)
make redis-ping       # check Redis
```

**Database Credentials (Docker):**
```
DB_HOST=botla-postgres
DB_PORT=5432
DB_NAME=botla_dev
DB_USER=botla
DB_PASSWORD=botla
```

### Database Migrations
```bash
make migrate-up                # apply all
make migrate-down              # rollback all
make migrate-version           # current version
make migrate-force v=<version> # force version (use with caution)

# Create a new migration (from project root)
migrate create -ext sql -dir db/migrations -seq <migration_name>
# then edit the generated up.sql / down.sql in db/migrations/
```

### Docker / Test Database
```bash
make migrate-up-docker         # migrate via the Docker network
make migrate-up-test           # apply to test DB (botla_test, search_path=test)
make create-test-schema        # CREATE SCHEMA IF NOT EXISTS test
```

## Testing Instructions

```bash
make test             # go test ./... (fast, no coverage)
make test-all         # coverage + fitz tag (CGO)
make test-no-pdf      # coverage, no CGO
make test-fast        # -short, skips integration tests
make cover-html       # render coverage.html
make cover-gate       # FAILS if coverage < 90%
```

### Running Specific Tests
```bash
go test -v -run TestFunctionName ./internal/services/...   # single test
go test -v ./internal/integration/...                       # integration suite
go test -race ./...                                         # race detector
```

### Test Structure & Gating
- Unit tests: `*_unit_test.go` / `*_test.go` alongside source files; table-driven preferred.
- Integration tests live in `internal/integration/` and are **skipped under `-short`**.
- Real-service suites (`internal/integration/{database,redis,qdrant,scenarios}`) additionally require `TEST_INTEGRATION=1` and the `docker-compose.integration.yml` stack — use `make test-integration-real`.
- Use `internal/testdb` for DB test utilities.

## Code Quality

```bash
make fmt              # gofmt -s
make imports          # goimports
make vet              # go vet
make lint             # golangci-lint
make vuln             # govulncheck
make ci               # vet + lint + test-no-pdf  <-- run before committing
make tidy             # go mod tidy
```

### Conventions
- **Error handling**: check and wrap errors with `pkg/errors` (`pkgerrors.Wrapf`) for context.
- **Logging**: `pkg/logger` structured logging.
- **Configuration**: access via `pkg/config`; env loaded from `.env`.
- **Models**: define in `internal/models/`; avoid inline structs in handlers.
- **Database**: repository pattern in `internal/repository/`; parameterized queries only.
- **Handlers**: keep thin; business logic in `internal/services/`.
- **Middleware**: cross-cutting concerns in `pkg/middleware/`.

### Enabled Linters (golangci-lint)
`govet` (with shadow), `staticcheck`, `errcheck`, `gosec`, `ineffassign`, `unused`, `misspell`, `gocritic`, `wrapcheck`. `wrapcheck` is configured (`.golangci.yml`) to ignore the repo's own packages + testify.

## Gotchas

- **`-tags fitz` + CGO** for anything touching PDF (`internal/pdf`, PDF ingestion tests). Without it, use the `*-no-pdf` make targets.
- **Coverage gate is 90%** (`make cover-gate`); keep touched packages at/above it.
- **Integration tests leak goroutines** if you use `res.Body.Close()` or `http.Get` directly — use `drainBody(res)` and the `testHTTP*` helpers (`internal/integration/http.go`). See `internal/integration/AGENTS.md`.
- **`.bak` files** (e.g. `admin_plan.go.bak`) exist in the tree but are not compiled — ignore them.
- LLM/topic-extraction prompts are **always written in English** (`services/chat_prompts.go`, `rag/topic_extractor.go`); output language is controlled by a directive, not prompt language.
- **No sqlc/ORM** despite older docs — data access is the squirrel-based repository pattern (see Architecture above).

## Admin Operations

```bash
make admin-promote email=user@example.com   # grant platform admin
make admin-demote email=user@example.com    # revoke platform admin
```

## Environment Variables

All config is env-driven (see `.env.example`). Key variables:
- `DATABASE_URL` / `DB_*` — PostgreSQL connection
- `REDIS_URL` — Redis (required to boot; rate limiting + queues)
- `JWT_SECRET` — token signing secret (32+ chars)
- `OPENAI_API_KEY` — embeddings + LLM
- `OPENROUTER_API_KEY` — optional alternative LLM provider
- `QDRANT_URL` / `QDRANT_API_KEY` — vector database
- `R2_*` — Cloudflare R2 / S3 credentials for source files

## Widget Integration

The `widget/` directory contains a Preact-based embeddable chat widget. Deploy with `make widget-deploy` (builds and publishes to Cloudflare Pages).

## Troubleshooting

- **"connection refused"**: ensure Docker services are running (`make up`).
- **"relation does not exist"**: run migrations (`make migrate-up`).
- **"role does not exist"**: ensure PostgreSQL is initialized with the correct user.
- **"PDF support disabled"**: run with `CGO_ENABLED=1 -tags fitz` or use `make be-run-no-pdf`.
- DB/Redis logs: `docker logs botla-postgres` / `docker logs botla-redis`.

## PR Instructions

- Run `make ci` (backend) or `npm run ci` (frontend/widget) before committing.
- All tests must pass; coverage must be ≥ 90% for packages you touch.
- No new linter warnings / TypeScript errors.
- Don't commit secrets — `.env` is gitignored.
