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

## Development Commands

### Running the Server
```bash
# Run server (with PDF support - requires CGO)
make be-run

# Run server (without PDF support)
make be-run-no-pdf

# Run server with test schema
make be-run-test
```

### Building
```bash
# Build all packages
make build

# Run server directly
make run
```

## Database Operations

### Local Database Access
When checking the database in the local environment, use the PostgreSQL instance running in Docker:

```bash
# Connect to dev database
make psql

# Check Redis connection
make redis-ping
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
# Apply all migrations
make migrate-up

# Rollback all migrations
make migrate-down

# Check current migration version
make migrate-version

# Force migration version (use with caution)
make migrate-force-test v=<version>
```

### Creating New Migrations
```bash
# Create a new migration (run from project root)
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

Then edit the generated `up.sql` and `down.sql` files in `db/migrations/`.

### Docker Database Access
```bash
# Connect to database running in Docker
docker exec -it botla-postgres psql -U botla -d botla_dev

# Run migrations in Docker network
make migrate-up-docker

# Check Docker migration version
make migrate-version-docker
```

### Testing Database
```bash
# Apply migrations to test database
make migrate-up-test

# Rollback test database
make migrate-down-test

# Check test migration version
make migrate-version-test

# Create test schema
make create-test-schema
```

### Common Database Queries (via `make psql`)
```sql
-- List all tables
\dt

-- Describe a table
\d table_name

-- Check migration status
SELECT * FROM schema_migrations;

-- List active connections
SELECT * FROM pg_stat_activity;
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

### Running Specific Tests
```bash
# Run a specific test
go test -v -run TestFunctionName ./path/to/package

# Run integration tests only
go test -v ./internal/integration/...

# Run with race detector
go test -race ./...

# Run handler tests
go test -v ./internal/api/handlers/...
```

### Test Structure
- Unit tests: `*_unit_test.go` files alongside source files
- Integration tests: `internal/integration/*_test.go`
- Table-driven tests are preferred
- Use `internal/testdb` for database test utilities

## Code Quality

### Formatting & Linting
```bash
# Format Go code
make fmt

# Fix imports
make imports

# Run go vet
make vet

# Run golangci-lint
make lint

# Check for vulnerabilities
make vuln

# Full CI check (vet + lint + test)
make ci
```

### Go Module Management
```bash
# Tidy dependencies
make tidy

# Clean build cache
make clean
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

## Admin Operations

### Promote/Demote Admins
```bash
# Make a user an admin
make admin-promote email=user@example.com

# Remove admin from a user
make admin-demote email=user@example.com
```

## Environment Variables

Required in `.env`:
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - JWT signing secret
- `OPENAI_API_KEY` - OpenAI API key
- `QDRANT_URL` - Qdrant vector DB URL
- `AWS_*` - S3 storage credentials (if using)

## Widget Integration

The `widget/` directory contains a Preact-based embeddable chat widget. See `widget/AGENTS.md` for details.

## Troubleshooting

### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check PostgreSQL logs
docker logs botla-postgres

# Verify database exists
docker exec -it botla-postgres psql -U botla -l
```

### Redis Connection Issues
```bash
# Check Redis status
make redis-ping

# Check if Redis is running
docker ps | grep redis

# View Redis logs
docker logs botla-redis
```

### Migration Issues
```bash
# Check migration status in database
make psql -c "SELECT * FROM schema_migrations;"

# Force migration version if stuck
make migrate-force-test v=<version>
```

### Common Errors
- **"connection refused"**: Ensure Docker services are running (`make up`)
- **"relation does not exist"**: Run migrations (`make migrate-up`)
- **"role does not exist"**: Ensure PostgreSQL is initialized with correct user
- **"PDF support disabled"**: Run with `CGO_ENABLED=1` or use `make be-run-no-pdf`

## PR Instructions

- Run `make ci` before committing
- All tests must pass
- Coverage must be ≥ 90%
- No new linter warnings
