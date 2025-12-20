# Tech Stack: Botla-co

## Backend
- **Language:** Go 1.25+
- **Database:** PostgreSQL (using `sqlc` for type-safe SQL queries)
- **Caching & Pub/Sub:** Redis
- **Vector Database:** Qdrant (for RAG and semantic search)
- **Object Storage:** AWS S3 or Cloudflare R2
- **Key Libraries:** `pgx/v5`, `go-redis/v9`, `aws-sdk-go-v2`, `colly` (scraping), `go-fitz` (PDF processing)

## Frontend (Dashboard)
- **Framework:** React 19
- **Build Tool:** Vite
- **Styling:** Tailwind CSS
- **UI Components:** Radix UI Primitives
- **State Management:** TanStack Query (React Query)
- **Routing:** React Router

## Widget
- **Framework:** Preact (for a lightweight footprint)
- **Build Tool:** Vite
- **Styling:** Tailwind CSS

## Infrastructure & Tools
- **Containerization:** Docker & Docker Compose
- **Migrations:** `migrate` CLI
- **Testing:** `testify` (Go), Vitest (Frontend), Playwright (E2E)
- **Linting:** `golangci-lint`, ESLint
