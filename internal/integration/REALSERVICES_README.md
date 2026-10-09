# Backend Integration Tests with Real Services

This directory contains integration tests that run against **real services** (PostgreSQL, Redis, Qdrant) instead of mocked dependencies.

## Purpose

These tests catch issues that mocks cannot detect:
- Connection pooling behavior
- Transaction isolation levels
- Query performance with real indexes
- Distributed rate limiting
- Vector similarity search accuracy
- Cross-service integration scenarios

## Running Tests

### Quick Test (without services)
```bash
# Run tests without starting services (will skip)
go test -v ./internal/integration/database/... -short
go test -v ./internal/integration/redis/... -short
go test -v ./internal/integration/qdrant/... -short
```

### Full Test (with services)
```bash
# One command: starts docker-compose.integration.yml, waits until every service is
# healthy, then runs the database/redis/qdrant/scenarios suites against it
make test-integration-real

# Run specific service tests (each also starts the stack)
make test-integration-db
make test-integration-redis
make test-integration-qdrant
make test-integration-full

# Every package under internal/integration (in-process API suite included),
# also against the integration stack; needs the golang-migrate CLI on PATH
make test-integration

# Stop services and drop their volumes when done
make test-integration-down
```

## Test Structure

### Database Tests (`database/`)
- `db_connection_test.go` - Connection pooling, health checks, configuration
- Tests PostgreSQL connection pool with 10 max connections, 2 min connections

### Redis Tests (`redis/`)
- `redis_test.go` - Connection, rate limiting, session management, TTL
- Tests Redis operations with real client

### Qdrant Tests (`qdrant/`)
- `qdrant_test.go` - Collection operations, vector upsert/search, concurrency
- Tests Qdrant vector database with real client

### Scenario Tests (`scenarios/`)
- `user_journey_test.go` - Complete user flow, multi-tenant isolation, database constraints
- Tests end-to-end scenarios across multiple services

## Configuration

Tests use environment variables or defaults. The `make test-integration*` targets
pass the values below (the integration stack's host ports); override them with
`INTEGRATION_DB_PORT=...`, `INTEGRATION_REDIS_URL=...` and so on. The in-process
suite in `internal/integration` (`fixtures/env.go`) reads the same variables but
defaults to the dev stack (`DB_PORT=5432`, `DB_NAME=botla_test`, Redis `6379`,
Qdrant `6333`) when they are unset.

| Variable | Default | Description |
|-----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5433` | PostgreSQL port (integration) |
| `DB_NAME` | `botla_integration` | Database name |
| `REDIS_URL` | `redis://localhost:6380` | Redis URL (integration) |
| `QDRANT_URL` | `http://localhost:6334` | Qdrant URL (integration) |
| `OPENAI_API_KEY` | (optional) | OpenAI API key for optional tests |
| `TEST_INTEGRATION` | (unset) | **Required to run** these tests; set to `1`. When unset the tests skip, so a plain `go test ./...` won't fail if the stack is down. The `make test-integration-*` targets set it. |

## CI/CD

Tests run weekly (Mondays) and on-demand via GitHub Actions: `.github/workflows/integration-tests.yml`.
The workflow runs `make test-integration-up` (start the stack and wait for it to be healthy), then `make test-integration-real`, then `make test-integration-down`.

## Notes

- Tests create isolated schemas per test for PostgreSQL
- Redis uses separate keys with random prefixes
- Qdrant creates temporary collections for test isolation
- All tests clean up after themselves
- These tests are opt-in: they skip unless `TEST_INTEGRATION=1` is set (or `-short` is passed), so a plain `go test ./...` won't fail when the integration stack is down
