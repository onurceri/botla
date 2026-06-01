# Botla

**AI-powered chatbot platform** that turns your website, PDFs, and text into a knowledgeable assistant you can embed anywhere.

Botla crawls your content, indexes it in a vector database, and answers user questions in real time using a Retrieval-Augmented Generation (RAG) pipeline. You get a full dashboard for managing bots and an embeddable widget that drops into any site with a single script tag.

---

## Highlights

- **Train on your data** — ingest URLs, sitemaps, PDFs, or raw text
- **RAG pipeline** — vector search via Qdrant, LLM responses via OpenAI or OpenRouter
- **Embeddable widget** — small Preact bundle, fully themable, drop-in `<script>` install
- **Multi-tenant** — organizations, workspaces, role-based access
- **Streaming chat** — Server-Sent Events for low-latency responses
- **Tiered plans** — Free / Pro / Ultra with quotas, model access, and feature gates
- **Analytics** — message volume, token usage, feedback, and conversation history
- **Privacy-aware** — KVKK / GDPR data export and account deletion flows
- **Multi-language** — English and Turkish UI out of the box

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25+, net/http, sqlc (no ORM) |
| Database | PostgreSQL 15+ |
| Cache / Queue | Redis 7 |
| Vector DB | Qdrant |
| Object Storage | Cloudflare R2 / S3-compatible |
| Frontend | React 19, Vite, Tailwind CSS, Radix UI, TanStack Query |
| Widget | Preact, Vite, Tailwind CSS |
| LLM Providers | OpenAI, OpenRouter |
| Auth | JWT (HttpOnly cookies) with refresh-token rotation |
| Web Scraping | go-rod (Chromium), colly, sitemap parser |
| PDF / OCR | go-fitz (Mumford), tesseract |
| Infra | Docker Compose, Cloudflare Pages, Caddy |

---

## Repository Layout

```
.
├── cmd/
│   ├── server/          # HTTP API entrypoint
│   └── cli/             # Admin CLI (promote/demote platform admins)
├── internal/
│   ├── api/             # HTTP handlers, router, middleware
│   ├── auth/            # JWT, password hashing
│   ├── db/              # sqlc-generated queries
│   ├── models/          # Domain types
│   ├── services/        # Business logic
│   ├── processing/      # Background jobs, source queue
│   ├── rag/             # Embeddings, vector search, LLM clients
│   ├── scraper/         # Web crawling, sitemap, CSS selectors
│   ├── pdf/             # PDF text extraction
│   ├── repository/      # Data access
│   ├── workers/         # Worker pool for async jobs
│   ├── models/          # Domain entities
│   ├── testdb/          # Test database utilities
│   └── integration/     # Integration test suite
├── pkg/                 # Public packages
│   ├── config/          # Env-driven configuration
│   ├── middleware/      # Auth, CORS, rate limiting
│   ├── ratelimit/       # Redis-backed limiter
│   ├── storage/         # S3/R2 abstraction
│   ├── tokenizer/       # Sentence-aware token counting
│   ├── logger/          # Structured logging
│   └── urlutil/         # URL helpers + SSRF protection
├── db/migrations/       # SQL migrations (golang-migrate)
├── api/openapi.yaml     # Public API spec
├── frontend/            # React dashboard
├── widget/              # Embeddable Preact widget
├── packages/ui-shared/  # Shared UI components
└── docker-compose*.yml  # Local dev / integration stacks
```

---

## How It Works

```
  [Your site / PDFs / text]
           |
           v
  +------------------+
  |   Scraper / OCR  |   (URL crawl, sitemap, PDF text extraction)
  +------------------+
           |
           v
  +------------------+
  |     Chunker      |   (sentence-aware, 512 tokens, ~15% overlap)
  +------------------+
           |
           v
  +------------------+
  |   Embeddings     |   (OpenAI text-embedding-3)
  +------------------+
           |
           v
  +------------------+
  |     Qdrant       |   (vector storage, per-chatbot collection)
  +------------------+

  User question ---> Embed ---> Vector search ---> Top-K chunks
                                                       |
                                                       v
                                              LLM (OpenAI / OpenRouter)
                                                       |
                                                       v
                                              Streamed answer
```

Every source is also summarized for **capability metadata** and **suggested questions**, surfaced to the end user as a carousel inside the chat.

---

## Quick Start

### Prerequisites
- Go 1.25.4+
- Docker and Docker Compose
- Node.js 18+ (for the frontend and widget)

### 1. Clone and configure
```bash
git clone https://github.com/onurceri/botla.git
cd botla
cp .env.example .env       # then edit secrets
```

### 2. Start dependencies
```bash
make up                    # PostgreSQL + Redis + Qdrant via docker-compose.dev.yml
make migrate-up            # apply DB migrations
```

### 3. Run the backend
```bash
make be-run                # with PDF/OCR support (requires CGO)
# or
make be-run-no-pdf         # without CGO
```

### 4. Run the dashboard
```bash
cd frontend
npm install
npm run dev                # http://localhost:5173
```

### 5. Run the widget (for local embed testing)
```bash
cd widget
npm install
npm run dev                # http://localhost:5174
```

### 6. Open the app
Visit `http://localhost:5173`, register an account, create a chatbot, add a URL or PDF source, and chat with it.

---

## Environment Variables

All configuration is environment-driven. See `.env.example` for the full list. Key variables:

| Variable | Purpose |
|---|---|
| `DATABASE_URL` / `DB_*` | PostgreSQL connection |
| `REDIS_URL` | Redis for rate limiting and queues |
| `QDRANT_URL` / `QDRANT_API_KEY` | Vector database |
| `OPENAI_API_KEY` | Embeddings + LLM |
| `OPENROUTER_API_KEY` | Optional alternative LLM provider |
| `JWT_SECRET` | Token signing secret (32+ chars) |
| `R2_*` | Cloudflare R2 / S3 credentials for source files |

The frontend and widget have their own `.env.example` files for `VITE_API_BASE_URL` and related variables.

---

## Development

### Backend
```bash
make test          # unit tests
make lint          # golangci-lint
make vet           # go vet
make ci            # vet + lint + tests
make cover-gate    # enforce 90% coverage threshold
```

See [`AGENTS.md`](./AGENTS.md) for full backend conventions, the layered architecture, and the testing strategy.

### Frontend
```bash
cd frontend
npm run dev        # dev server
npm run ci         # lint + typecheck + coverage
npm run e2e        # Playwright e2e
```

See [`frontend/AGENTS.md`](./frontend/AGENTS.md) for frontend conventions and the HttpOnly-cookie auth model.

### Widget
```bash
cd widget
npm run dev
npm run build      # build embeddable bundle
```

See [`widget/AGENTS.md`](./widget/AGENTS.md).

---

## API

A public OpenAPI 3.0 spec lives at [`api/openapi.yaml`](./api/openapi.yaml). Generate TypeScript types with:

```bash
cd frontend
npm run generate-types
```

---

## Architecture Notes

- **No ORM** — type-safe queries via [sqlc](https://github.com/sqlc-dev/sqlc) over `database/sql`
- **Layered structure** — handlers stay thin, business logic lives in `internal/services`
- **Worker pool** — async ingestion jobs (scraping, embedding, refresh) run in a bounded goroutine pool
- **SSRF protection** — `pkg/urlutil` blocks private IP ranges during URL ingestion
- **Rate limiting** — Redis-backed, per-user and per-endpoint overrides
- **Cookie-based auth** — `HttpOnly` + `Secure` cookies, automatic refresh-token rotation
- **Plan enforcement** — every action checks quota / plan limits in one place

---

## Roadmap Highlights

Currently shipped: multi-tenant orgs/workspaces, RAG with tiered confidence, embeddings + vector search, embedded widget, KVKK export/delete, plan gating, analytics, human handoff (email capture).

See the in-repo `docs/` for deeper feature, system, and admin references.

---

## Contributing

PRs welcome. Before opening one:

1. Run `make ci` (backend) or `npm run ci` (frontend)
2. Keep coverage at or above 90% for the package you touch
3. Follow conventions in the per-area `AGENTS.md`
4. Don't commit secrets — `.env` is gitignored, keep it that way

---

## License

See `LICENSE` if present, or contact the maintainers.
