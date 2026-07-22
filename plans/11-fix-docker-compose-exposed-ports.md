# Fix 4.2 — Lock down exposed DB/Qdrant/Redis ports in compose files

- **Source:** `PROJECT_ISSUES_REPORT.md` §4.2
- **Severity:** Low (dev) / High (if the prod compose is used as-is)
- **Validation:** ✅ Confirmed TRUE
- **Type:** Infrastructure security

## Context

The compose files publish Postgres (5432), Qdrant (6333), and Redis (6379) to the host. Convenient
for local dev, but a real exposure in staging/production — these should stay on the private Docker
network (or be bound to a restricted interface) so only the app container can reach them.

## Validated current state

- `docker-compose.yml` (production — includes Caddy for HTTPS, confirming prod intent):
  - Postgres `ports: - "5432:5432"` (**lines 11–12**)
  - Qdrant `ports: - "6333:6333"` (**lines 25–26**)
- `docker-compose.dev.yml` (development):
  - Postgres `"5432:5432"` (**lines 9–10**)
  - Redis `"6379:6379"` (**lines 21–22**)
  - Qdrant `"6333:6333"` (**lines 45–46**)
- `docker-compose.integration.yml` already uses offset host ports (5433/6380/6384) to avoid
  conflicts — a separate concern, leave as-is unless desired.

## Plan

1. **`docker-compose.yml` (prod):** remove the `ports:` mappings for Postgres and Qdrant entirely —
   the app container reaches them over the compose network by service name. If host access is ever
   required, bind to loopback only (e.g. `"127.0.0.1:5432:5432"`).
2. **`docker-compose.dev.yml` (dev):** keep host exposure for local tooling, but optionally bind to
   loopback (`"127.0.0.1:5432:5432"`, etc.) to avoid exposing dev DBs on the LAN. At minimum,
   document that these are dev-only.
3. Ensure all services that need DB/Qdrant/Redis are on the same Docker network so removing host
   ports doesn't break app connectivity (the app already connects by service name).
4. Leave `docker-compose.integration.yml` unchanged (offset ports are intentional).

## Files to modify
- `docker-compose.yml`
- `docker-compose.dev.yml` (optional loopback bind / docs)

## Verification
- `docker compose -f docker-compose.yml config` validates the file.
- `docker compose up -d` and confirm the app still connects to Postgres/Qdrant/Redis by service name.
- From the host, confirm the ports are no longer externally reachable
  (`nc -zv localhost 5432` should fail for prod; succeed only on loopback if bound there).
