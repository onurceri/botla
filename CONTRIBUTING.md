# Contributing to Botla

Thanks for your interest. Bug reports, fixes and focused improvements are all welcome. For larger changes, please open an issue first so we can agree on the approach before you write the code.

## Setup

You need Go 1.25+, Node.js 20.19+ with npm, and Docker with Compose.

```bash
cp .env.example .env              # set OPENAI_API_KEY and JWT_SECRET
cp frontend/.env.example frontend/.env
make up                           # PostgreSQL, Redis, Qdrant; applies migrations
make be-run-no-pdf                # API on :8080 (`make be-run` adds PDF support via CGO)
npm ci && npm run dev             # dashboard on :5173, widget on :5174
```

The [README](README.md#quick-start) has more detail. PDF extraction needs CGO and the `fitz` build tag; everything else builds without CGO.

## Conventions

[AGENTS.md](AGENTS.md) is the contributor guide for the backend: the layer rules (thin handlers, business logic in services, SQL only in repositories), error wrapping, logging and the gotchas. Read the per-area guide before you change that area: `frontend/AGENTS.md`, `internal/api/handlers/AGENTS.md`, `internal/services/AGENTS.md`, `internal/rag/AGENTS.md` and `internal/integration/AGENTS.md`.

## Checks to run

Run the checks for the parts you touched before you open a pull request. CI runs the same lint, typecheck and test steps.

**Backend**

```bash
make ci                           # go vet, golangci-lint, tests with coverage
make test-fast                    # quicker: -short tests (needs the PostgreSQL from `make up`)
```

Keep coverage of the packages you touch at or above 90% (`make test-all && make cover-gate`).

**Dashboard and widget**

```bash
npm run ci --workspace=frontend   # ESLint, tsc, Vitest with coverage
npm run ci --workspace=widget     # tsc, ESLint, Vitest with coverage
```

If you change user flows, also run the Playwright suites (`npm run e2e --workspace=frontend`, `npm run test:e2e --workspace=widget`).

**RAG and chat behaviour**

Changes to retrieval, prompts or the chat pipeline can change answer quality without breaking a single unit test. Run `make eval-test` (offline) and, if you have an API key, `make eval-quick`. To have CI run the full eval suite on your pull request, ask a maintainer to add the `run-evals` label. See [evals/README.md](evals/README.md).

## Commits

Commits follow [Conventional Commits](https://www.conventionalcommits.org/): `type(scope): summary`, with the summary in the imperative mood and without a trailing period.

- Types used in this repo: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `build`, `ci`.
- The scope is optional and names the area: `feat(evals): ...`, `fix(widget): ...`, `build(widget): ...`.
- Keep each commit to one logical change, and explain *why* in the body when it isn't obvious.

Example: `fix(rag): skip empty chunks before embedding`.

## Pull requests

- Keep pull requests focused, and describe what changed and how you tested it.
- Add or update tests with behaviour changes.
- Don't commit secrets or `.env` files; `.env.example` is the template.
- New code must not add linter warnings or TypeScript errors.

By contributing, you agree that your contributions are licensed under the [AGPL-3.0-or-later](LICENSE), the same license as the project.
