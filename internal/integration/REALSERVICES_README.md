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
# Start integration services
docker compose -f docker-compose.integration.yml up -d

# Wait for services to be healthy
docker compose -f docker-compose.integration.yml ps

# Run all integration tests
make test-integration-real

# Run specific service tests
make test-integration-db
make test-integration-redis
make test-integration-qdrant
make test-integration-full

# Stop services when done
docker compose -f docker-compose.integration.yml down
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

Tests use environment variables or defaults:

| Variable | Default | Description |
|-----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5433` | PostgreSQL port (integration) |
| `DB_NAME` | `botla_integration` | Database name |
| `REDIS_URL` | `redis://localhost:6380` | Redis URL (integration) |
| `QDRANT_URL` | `http://localhost:6334` | Qdrant URL (integration) |
| `OPENAI_API_KEY` | (optional) | OpenAI API key for optional tests |

## CI/CD

Tests run weekly via GitHub Actions: `.github/workflows/integration-tests.yml`

## Notes

- Tests create isolated schemas per test for PostgreSQL
- Redis uses separate keys with random prefixes
- Qdrant creates temporary collections for test isolation
- All tests clean up after themselves
- Use `-short` flag to skip integration tests during unit test runs
