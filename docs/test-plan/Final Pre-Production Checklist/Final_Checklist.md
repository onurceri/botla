# Final Pre-Production Checklist

## Overview
This is the final checklist before deploying to production.

---

## Critical Items

- [ ] All critical bugs fixed
- [ ] All security vulnerabilities addressed
- [ ] All tests passing (unit, integration, E2E)
- [ ] Code coverage >= 90%

**Verification Steps:**
1. Check Issue Tracker for open `priority:critical` bugs.
2. Run `scripts/audit.sh` (or `govulncheck` + `npm audit`).
3. Run `make test` and `cd frontend && npm run test:e2e`.
4. Run `make cover` and check report.

---

## Performance

- [ ] Performance benchmarks met
- [ ] Load testing completed
- [ ] Database queries optimized
- [ ] Indexes verified

**Verification Steps:**
1. Run `go test -bench=. ./...`.
2. Run `k6 run scripts/load_test.js`.
3. Check `pg_stat_statements` for slow queries.
4. Run `scripts/check_indexes.sh` (if available) or manual `\d`.

---

## Infrastructure

- [ ] Production environment configured
- [ ] Database backups configured
- [ ] Monitoring configured (e.g., Sentry)
- [ ] Logging configured
- [ ] SSL/TLS configured

**Verification Steps:**
1. Verify `PROD.env` values.
2. Verify backup schedule in cron/cloud provider.
3. Trigger test error and verify alert in Sentry.
4. Check logs in Kibana/CloudWatch.
5. Verify `https://api.botla.co` works.

---

## Documentation

- [ ] API documentation complete
- [ ] Deployment guide complete
- [ ] Rollback plan documented
- [ ] Team trained on procedures

**Verification Steps:**
1. Review `docs/api/`.
2. Review `docs/deploy.md`.
3. Review `docs/rollback.md`.
4. Confirm with team lead.

---

## Final Verification

- [ ] Staging environment tested end-to-end
- [ ] All plan limits enforced
- [ ] All security tests pass
- [ ] Widget works on allowed domains
- [ ] Analytics tracking verified

**Verification Steps:**
1. Run full E2E suite against Staging URL.
2. Manually attempt to bypass Free plan limits on Staging.
3. Run `internal/integration/plan_enforcement_security_test.go` against Staging.
4. Embed widget on test page.
5. Generate traffic and check `/analytics` endpoint.

---

## Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Developer | | | |
| QA Lead | | | |
| Security | | | |
| Product | | | |

---

## Go/No-Go Decision

- [ ] **GO** - Approved for production
- [ ] **NO-GO** - Issues must be resolved

**Notes:**
