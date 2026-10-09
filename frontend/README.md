# Botla dashboard

The React 19 + Vite + TypeScript app where users create chatbots, add sources, tune retrieval and guardrails, test bots in a playground, review handoff requests and analytics, and where platform admins manage users, plans and queues.

It talks to the Go API with HttpOnly cookies (`withCredentials: true`) and uses TanStack Query for server state. Conventions live in [AGENTS.md](AGENTS.md); end-to-end testing standards live in [e2e/TESTING_STANDARDS.md](e2e/TESTING_STANDARDS.md).

## Develop

Install dependencies once from the repository root (`npm ci`), then:

```bash
cp .env.example .env     # VITE_API_BASE_URL, VITE_WIDGET_URL, ...
npm run dev              # http://localhost:5173 (expects the API on :8080)
```

## Scripts

| Script | What it does |
|---|---|
| `npm run ci` | ESLint, `tsc --noEmit`, Vitest with coverage thresholds |
| `npm run test` | Vitest |
| `npm run e2e` | Playwright (Chromium); needs the dev server running |
| `npm run build` | Production build to `dist/` |
| `npm run generate-types` | Regenerate API types from `../api/openapi.yaml` |
