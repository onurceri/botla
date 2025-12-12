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

---

### 17.1.2 Docker Compose
**Priority:** High  
**Type:** Deployment Test

```bash
docker-compose up -d
# All services start without error
```

---

### 17.1.3 Database Migrations
**Priority:** Critical  
**Type:** Deployment Test

```bash
make migrate-up
# All migrations applied
```

---

### 17.1.4 Frontend Build
**Priority:** High  
**Type:** Deployment Test

```bash
cd frontend
npm run build
# Build succeeds without errors
```

---

### 17.1.5 Backend Build
**Priority:** High  
**Type:** Deployment Test

```bash
make build
# Binary created successfully
```

---

## How to Run Tests

```bash
docker-compose up -d
make migrate-up
make test
```
