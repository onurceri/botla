# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Botla Overview

Botla is a Go-based SaaS chatbot platform with three main components:
- **Backend API** (Go) - REST APIs for chatbot management, RAG, and authentication
- **Frontend Admin** (React/TypeScript/Vite) - Administrative dashboard at `/frontend`
- **Chat Widget** (Preact) - Embeddable customer-facing widget at `/widget`

## Common Commands

### Backend Development
```bash
# Start PostgreSQL and Redis services
make up

# Run backend server (requires CGO for PDF support)
make be-run

# Run tests (requires CGO enabled)
make test-all

# Run specific test
go test -v ./internal/services -run TestChatService

# Database migrations
make migrate-up
make migrate-down

# Full CI checks (vet, lint, test)
make ci
```

### Frontend Development
```bash
# Run frontend dev server
make fe-run

# Build frontend
make fe-build
```

### Widget Development
```bash
# Build widget
make widget-build

# Deploy widget (through Cloudflare Pages)
make widget-deploy
```

## Architecture & Key Patterns

### Service Layer Architecture
The codebase follows a service layer pattern where business logic is isolated in `/internal/services/`. Each service handles a specific domain:
- `chat_service.go` - Chatbot interactions and message handling
- `rag_service.go` - Document ingestion and retrieval
- `auth_service.go` - Authentication and authorization
- `scraper_service.go` - Web scraping and crawling

### Handler → Service → Repository Flow
1. HTTP handlers in `/internal/api/` validate requests and delegate to services
2. Services in `/internal/services/` contain business logic
3. Database queries in `/internal/db/` use Go 1.16+ embed for SQL queries

### Multi-tenant Architecture
The system uses organizations and workspaces for tenant isolation. Always filter by organization_id when querying data:
```go
// Example pattern
WHERE organization_id = $1 AND workspace_id = $2
```

### RAG System Implementation
The RAG system uses:
- **Qdrant** for vector storage and similarity search
- **OpenAI** for embeddings (text-embedding-3-small)
- **OpenRouter** for LLM completions
- Document processing pipeline in `/internal/processing/`

Key files:
- `/internal/rag/` - RAG client implementations
- `/internal/models/rag_source.go` - Document metadata
- `/internal/processing/chunker.go` - Text chunking strategies

### Authentication Flow
1. JWT tokens with bcrypt password hashing
2. Middleware chain: Recovery → Logger → PlanLoader → RateLimit → Auth
3. Organization/workspace context loaded in middleware
4. Plan-based feature gating enforced throughout

### Recent Development Context
Based on git status, recent work involves:
- AI model management in `/internal/models/ai_model.go`
- Chat service optimizations with debouncing and ref-based payload handling
- Widget script loading improvements
- Plan features and pricing alignment

### Testing Approach
- Unit tests alongside source files (`*_unit_test.go`)
- Integration tests in `/internal/integration/`
- Test utilities/helper functions in `/internal/testdb/`
- Coverage gate at 90% - ensure new code is well-tested

### Key Dependencies
- **Go 1.25+** with CGO enabled (required for PDF processing)
- **PostgreSQL 15+** with pgx driver
- **Redis** for caching and rate limiting
- **Qdrant** for vector search
- Optional: **Cloudflare R2** for file storage

When modifying the codebase, ensure compatibility with these dependencies and maintain the existing architectural patterns.