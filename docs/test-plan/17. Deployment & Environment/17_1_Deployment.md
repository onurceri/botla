# 17.1 Deployment Test Plan

## Overview
This test plan covers deployment and environment configuration.

---

## Test Cases

### 17.1.1 Environment Variables
**Priority:** Critical  
**Type:** Deployment Test

| Variable | Test |
|----------|------|
| DATABASE_URL | App connects to DB |
| OPENROUTER_API_KEY | AI calls work |
| QDRANT_URL | Vector search works |
| JWT_SECRET | Auth works |

**Implementation Plan:**
- **Test Script:** `scripts/verify_env.sh`
- **Steps:**
  1. Check if required vars are set in current env.
  2. Attempt dry-run connection to DB using `psql` or helper.
  3. Verify `JWT_SECRET` is not empty.

---

### 17.1.2 Docker Compose
**Priority:** High  
**Type:** Deployment Test

```bash
docker-compose up -d
# All services start without error
```

**Implementation Plan:**
- **Test Script:** `scripts/test_deploy.sh`
- **Steps:**
  1. `docker-compose up -d`
  2. `docker-compose ps` -> Verify state is "Up".
  3. `curl localhost:8080/health` -> Verify 200.

---

### 17.1.3 Database Migrations
**Priority:** Critical  
**Type:** Deployment Test

```bash
make migrate-up
# All migrations applied
```

**Implementation Plan:**
- **Test Script:** `scripts/test_deploy.sh`
- **Steps:**
  1. `make migrate-up`.
  2. Verify exit code 0.

---

### 17.1.4 Frontend Build
**Priority:** High  
**Type:** Deployment Test

```bash
cd frontend
npm run build
# Build succeeds without errors
```

**Implementation Plan:**
- **Test Script:** `scripts/test_build.sh`
- **Steps:**
  1. `cd frontend && npm install && npm run build`.
  2. Verify `dist/index.html` exists.

---

### 17.1.5 Backend Build
**Priority:** High  
**Type:** Deployment Test

```bash
make build
# Binary created successfully
```

**Implementation Plan:**
- **Test Script:** `scripts/test_build.sh`
- **Steps:**
  1. `make build`.
  2. Verify `bin/server` exists and is executable.
  3. Run `bin/server --version` (if supported) or check `file bin/server`.

---

## How to Run Tests

```bash
docker-compose up -d
make migrate-up
make test
```
