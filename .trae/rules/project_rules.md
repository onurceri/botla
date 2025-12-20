# Project Workflow Rules & Guidelines

## Backend (Go)

### Code Quality & Standards
Before merging or finishing a backend task, ensure code quality by running the following. All commands must pass without error.

```bash
make fmt      # Format code
make imports  # Organize imports
make vet      # Run go vet
make lint     # Run golangci-lint
make shadow   # Check for variable shadowing (CRITICAL)
make vuln     # Check for vulnerabilities
```

**Critical Rule:** Do not shadow variables. The `make shadow` command will catch this.
*Avoid:* `shadow: declaration of "err" shadows declaration at ...`

### Testing
- **Run all tests (includes PDF/CGO):** `make test-all`
- **Run fast tests (no PDF/CGO):** `make test-no-pdf`
- **Check coverage:**
  - `make cover-func` (Summary)
  - `make cover-html` (HTML report)
  - `make cover-gate` (Fails if coverage < 90%)

### Running the Server
- **With PDF Support (requires CGO):** `make be-run`
- **Without PDF Support:** `make be-run-no-pdf`

## Frontend (React) - `frontend/`

### Development
- **Start Dev Server:** `make fe-run` (or `cd frontend && npm run dev`)

### Quality & Testing
Run these commands within the `frontend/` directory:
- **Lint:** `npm run lint`
- **Typecheck:** `npm run typecheck`
- **Format:** `npm run format`
- **Unit Tests:** `npm run test` (Vitest)
- **E2E Tests:** `npm run e2e` (Playwright)
- **Full CI Check:** `npm run ci` (Lint + Typecheck + Test Coverage)

## Widget - `widget/`

- **Development:** `cd widget && npm install && npm run dev`
- **Deploy:** `make widget-deploy` (Builds & Deploys via Wrangler)

## Infrastructure & Database

### Docker Services (Postgres & Redis)
- **Start Services:** `make up`
- **Stop Services:** `make down`

### Database Management
- **Connect to DB (CLI):** `make psql`
  - *Alternative:* `docker exec -it botla-postgres psql -U botla -d botla_dev`
- **Migrations:**
  - `make migrate-up`
  - `make migrate-down`
  - `make migrate-version`

## General Workflow
1.  Start infrastructure: `make up`
2.  Run migrations if needed: `make migrate-up`
3.  Develop features (Backend: `make be-run`, Frontend: `make fe-run`)
4.  **Verify before commit:**
    - Backend: `make fmt imports vet lint shadow test-no-pdf`
    - Frontend: `cd frontend && npm run ci`
