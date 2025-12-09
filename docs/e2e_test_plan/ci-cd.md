# CI/CD Guidelines

## Test Suites

| Suite | When | Content | Duration |
|-------|------|---------|----------|
| Smoke | Every PR | Login, create bot, chat, analytics | ~2 min |
| Full | Daily | All stubs, all features | ~15 min |
| External | Nightly | Real LLM, R2 tiny files | ~5 min |

## Cost Controls

- Always use stubs unless `E2E_USE_REAL=true`
- Token budget: abort if >10,000 tokens
- Minimal content sizes
- Low parallelism for heavy tests

## GitHub Actions Example

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  smoke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run Smoke Tests
        run: |
          make test-smoke
        env:
          E2E_MODE: stubbed

  full:
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'
    steps:
      - uses: actions/checkout@v4
      - name: Run Full Suite
        run: make test
```

## Environment Matrix

| Env Var | Smoke | Full | External |
|---------|-------|------|----------|
| `E2E_MODE` | stubbed | stubbed | real |
| `OPENAI_API_KEY` | - | - | ✓ |
| `QDRANT_URL` | stub | stub | prod |
