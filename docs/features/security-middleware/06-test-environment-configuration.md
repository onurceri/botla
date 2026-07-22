> Extracted from: security-middleware.md — Section: 6. Test Environment Configuration

# 6. Test Environment Configuration

## 6.1 Required Services
- PostgreSQL 15+ (with migrations applied)
- Redis 7+ (for rate limiting + queues)
- Qdrant (for vector storage)
- OpenAI API key (for embeddings + LLM)
- Docker & Docker Compose (for local dev)

## 6.2 Test Data Setup
- Seed database with test users, orgs, workspaces, chatbots
- Seed test plans with various rate limits and feature flags
- Ingest test content (HTML, PDF, text sources)
- Create test JWT tokens with known expiry times

## 6.3 Test Tools
- Go test suite (`internal/integration/`)
- Playwright (widget E2E tests in `widget/`)
- `curl` / `httpie` for manual API testing
- Browser DevTools for widget inspection
- `redis-cli` for rate limit state inspection
- Qdrant REST API for vector store inspection

## 6.4 Test Credentials
```
Test User:     test@botla.app / Test123!@#
Admin User:    admin@botla.app / Admin123!@#
Test Chatbot:  test-chatbot-1 (ID from seed data)
```
