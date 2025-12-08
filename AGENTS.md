# AGENTS.md - Botla Backend

A Go-based backend API for the Botla chatbot platform, providing REST APIs for chatbot management, RAG (Retrieval Augmented Generation), authentication, and analytics.

## Project Structure

```
├── cmd/server/          # Application entrypoint
├── internal/            # Private application code
│   ├── api/             # HTTP handlers and router
│   │   └── handlers/    # Request handlers (auth, chatbot, chat, source, etc.)
│   ├── auth/            # JWT and password utilities
│   ├── db/              # Database layer (sqlc-generated)
│   ├── integration/     # Integration tests
│   ├── models/          # Domain models
│   ├── pdf/             # PDF processing (requires fitz tag)
│   ├── processing/      # Text processing and chunking
│   ├── rag/             # RAG implementation (OpenAI, Qdrant)
│   ├── scraper/         # Web scraping (sitemap, CSS selectors)
│   ├── services/        # Business logic services
│   └── text/            # Text utilities
├── pkg/                 # Public packages
│   ├── config/          # Application configuration
│   ├── langconfig/      # Language/localization config
│   ├── logger/          # Logging utilities
│   ├── middleware/      # HTTP middleware (auth, CORS, rate limiting)
│   └── storage/         # S3/file storage
├── db/                  # Database files
│   └── migrations/      # SQL migration files
└── widget/              # Embeddable chat widget (Preact)
```

## Dev Environment Setup

### Prerequisites
- Go 1.25.4+
- Docker & Docker Compose
- PostgreSQL 15+ (via Docker or local)
- Redis (via Docker or local)
- For PDF support: CGO enabled with `fitz` library

### Start Development Services
```bash
# Start PostgreSQL and Redis
make up

# Run migrations
make migrate-up

# Connect to database
make psql
```

## Build & Run Commands

```bash
# Run server (with PDF support - requires CGO)
make be-run              # CGO_ENABLED=1 go run -tags fitz cmd/server/main.go

# Run server (without PDF support)
make be-run-no-pdf       # go run cmd/server/main.go

# Build all packages
make build

# Run database migrations
make migrate-up          # Apply migrations
make migrate-down        # Rollback migrations
make migrate-version     # Current version
```

## Testing Instructions

### Run Tests
```bash
# Run all tests with coverage (requires CGO for PDF)
make test-all

# Run tests without PDF features
make test-no-pdf

# Quick test without coverage
make test

# View coverage report
make cover-html          # Opens coverage.html
make cover-func          # Console output

# Coverage gate (fails if < 90%)
make cover-gate
```

### Test Structure
- Unit tests: `*_unit_test.go` files alongside source files
- Integration tests: `internal/integration/*_test.go`
- Table-driven tests are preferred
- Use `internal/testdb` for database test utilities

### Running Specific Tests
```bash
# Run a specific test
go test -v -run TestFunctionName ./path/to/package

# Run integration tests only
go test -v ./internal/integration/...

# Run with race detector
go test -race ./...
```

## Code Style Guidelines

### Go Formatting
```bash
make fmt                 # gofmt -s -w .
make imports             # goimports -w .
make vet                 # go vet ./...
make lint                # golangci-lint run ./...
make vuln                # govulncheck ./...

# Full CI check
make ci                  # vet + lint + test-no-pdf
```

### Conventions
- **Error handling**: Always check and propagate errors. Use `fmt.Errorf("context: %w", err)` for wrapping.
- **Logging**: Use `pkg/logger` for structured logging.
- **Configuration**: Access via `pkg/config`. Environment variables loaded from `.env`.
- **Models**: Define in `internal/models/`. Avoid inline struct definitions in handlers.
- **Database**: Define queries in `internal/db/`. Use parameterized queries to prevent SQL injection.
- **Handlers**: Keep thin. Business logic goes in `internal/services/`.
- **Middleware**: Define in `pkg/middleware/`. Use for cross-cutting concerns.

### Enabled Linters (golangci-lint)
- `govet` (with shadow checking)
- `staticcheck`
- `errcheck`
- `gosec`
- `ineffassign`
- `unused`
- `misspell`

## Database & Migrations

### Creating Migrations
```bash
# Create a new migration
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

### Database Queries
Database queries are defined manually in `internal/db/`. Use parameterized queries with `pgx` to prevent SQL injection.

## Security Considerations

- JWT tokens for authentication (access + refresh tokens)
- Passwords hashed with bcrypt
- Rate limiting via `pkg/middleware/`
- CORS configuration in middleware
- SQL injection prevented via sqlc parameterized queries
- Input validation in handlers

## Environment Variables

Required variables (see `.env.example`):
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - JWT signing secret
- `OPENAI_API_KEY` - OpenAI API key
- `QDRANT_URL` - Qdrant vector DB URL
- `AWS_*` - S3 storage credentials (if using AWS)

## Widget Integration

The `widget/` directory contains a Preact-based embeddable chat widget. See `widget/AGENTS.md` for details.

## PR Instructions

- Run `make ci` before committing
- All tests must pass
- Coverage must be ≥ 90%
- No new linter warnings
