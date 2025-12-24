# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Botla is a chatbot platform with:
- **Go backend**: REST API for chatbot management, RAG, authentication, analytics
- **React frontend**: Dashboard for chatbot configuration and analytics
- **Preact widget**: Embeddable chat widget (deployed to Cloudflare Pages)

Key services: PostgreSQL, Redis, Qdrant (vector DB), OpenAI for LLM/RAG

## Commands

### Backend (Go)

```bash
# Start dev services (PostgreSQL, Redis)
make up

# Run migrations
make migrate-up

# Run server (with PDF support - requires CGO)
make be-run

# Run tests with coverage (90% gate)
make test-all

# CI check (vet + lint + test)
make ci
```

### Frontend (React + Vite)

```bash
cd frontend

# Dev server
npm run dev

# Build
npm run build

# Lint & typecheck
npm run lint
npm run typecheck

# Tests
npm run test              # unit tests
npm run test:coverage     # with coverage
npm run e2e               # Playwright e2e tests
npm run e2e:headed        # headed mode

# Format
npm run format            # write
npm run format:check      # check
```

### Widget (Preact)

```bash
cd widget

# Dev server
npm run dev

# Build for production
npm run build

# Deploy to Cloudflare Pages
npm run build && npx wrangler pages deploy dist --project-name botla-widget
```

## Architecture

### Backend Structure (`internal/`)

```
├── api/handlers/    # HTTP request handlers (auth, chatbot, chat, source, etc.)
├── auth/            # JWT and password utilities
├── db/              # Database layer (pgx, parameterized queries)
├── models/          # Domain models
├── processing/      # Text processing and chunking
├── rag/             # RAG implementation (OpenAI, Qdrant)
├── scraper/         # Web scraping (sitemap, CSS selectors)
├── services/        # Business logic
└── api/             # Router setup
```

`pkg/` contains public packages: config, langconfig, logger, middleware, storage.

### Frontend Structure (`frontend/src/`)

```
├── api/           # API client (axios)
├── components/    # Shared UI components
├── features/      # Feature modules (analytics, chatbot, organization)
├── hooks/         # Custom hooks
├── lib/           # Utilities (api client, utils)
├── pages/         # Route pages (DashboardPage, ChatbotsPage, etc.)
└── providers/     # Context providers (QueryClient, Auth)
```

Uses TanStack Query for data fetching, React Router for routing, Radix UI primitives, Tailwind CSS v4.

### Data Flow

1. Frontend calls backend REST API (axios)
2. Backend handlers delegate to services in `internal/services/`
3. Services use models, db queries, and external services (OpenAI, Qdrant, S3)
4. Responses returned to frontend

## Conventions

### Backend (Go)
- Error handling: `fmt.Errorf("context: %w", err)` for wrapping
- Logging: Use `pkg/logger` for structured logs
- Handlers: Keep thin, business logic in services
- Database: Parameterized queries with pgx

### Frontend (React)
- Feature-based organization in `features/`
- Components: CVA (class-variance-authority) for variants
- Data fetching: TanStack Query hooks
- Styling: Tailwind CSS + tailwind-merge

## Database

PostgreSQL with migrations in `db/migrations`. Connect locally:
```
DB_HOST=localhost
DB_PORT=5432
DB_NAME=botla_dev
DB_USER=botla
DB_PASSWORD=botla
```

## Environment Variables

Required in `.env`:
- `DATABASE_URL` - PostgreSQL
- `REDIS_URL` - Redis
- `JWT_SECRET` - JWT signing
- `OPENAI_API_KEY` - OpenAI
- `QDRANT_URL` - Vector DB
- `AWS_*` - S3 storage (if using)
