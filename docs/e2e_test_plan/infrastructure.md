# Test Infrastructure

## Tooling

| Component | Tool | Notes |
|-----------|------|-------|
| Dashboard E2E | Playwright | `http://localhost:5173`, `VITE_E2E=1` |
| Widget E2E | Playwright | Shadow DOM assertions |
| Backend Integration | Go `httptest` | `internal/integration/testserver.go` |
| Database | PostgreSQL | Test schema via `testutils.go` |
| Stubs | Custom HTTP servers | OpenAI, Qdrant, HTML |

## Execution Modes

### 1. Stubbed Mode (Default)
```bash
export E2E_MODE=stubbed
export OPENAI_API_BASE=stub_url
export QDRANT_URL=stub_url
```

### 2. Real Backend Mode
```bash
export E2E_API_BASE=http://localhost:8080
```

### 3. Real Provider Mode (Manual)
```bash
export E2E_USE_REAL=true
export OPENAI_API_KEY=sk-...
```

## Stubs

| Stub | Location | Purpose |
|------|----------|---------|
| `startOpenAIStub()` | `internal/integration/testutils.go` | Mock embeddings/completions |
| `startOpenAIErrorStub()` | Same | 500 errors |
| `startOpenAITimeoutStub()` | Same | Timeout simulation |
| `startQdrantStub()` | Same | Mock vector search |
| `startHTMLStub()` | Same | Test pages for URL ingestion |

## Run Commands

```bash
# All backend tests
make test

# With coverage
make coverage

# Specific integration test
go test -v ./internal/integration/... -run TestChat

# Frontend tests
cd frontend && npm test

# E2E tests
cd frontend && npm run test:e2e
```
