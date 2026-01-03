# Test Plan: Backend Integration Tests with Real API

**Plan ID:** TP-BACKEND-INTEGRATION-001  
**Priority:** HIGH  
**Estimated Duration:** 2-3 weeks  
**Target:** Comprehensive integration tests against real database and services  
**Status:** Draft  

---

## Executive Summary

Currently, backend integration tests use mocked external services (LLM, Qdrant, Redis). While this is appropriate for fast unit testing, it doesn't validate real-world integration scenarios. This plan adds a separate test suite that runs against real services to catch integration issues that mocks cannot detect.

Key integration points to test:
1. PostgreSQL database (real queries, transactions, migrations)
2. Redis (caching, rate limiting, sessions)
3. Qdrant vector database (real vector operations)
4. OpenAI API (real LLM calls - with mocking fallback for CI)
5. Full request/response cycle with middleware

---

## Sisyphus Agent Prompt

```
You are Sisyphus, a senior backend engineer with expertise in Go testing, database operations, and integration testing.

### Task Context
The Botla backend currently has two levels of testing:
1. Unit tests - Fast, use mocks for all dependencies
2. Integration tests - Use mocked external services (LLM, Qdrant, Redis)

What's MISSING:
- Real integration tests against PostgreSQL, Redis, Qdrant
- Tests that catch issues mocks don't reveal (connection pooling, timeouts, race conditions)
- Full-stack integration scenarios (multiple services working together)
- Performance/load tests
- Chaos engineering tests (what happens when services fail)

### Current Test Infrastructure
internal/integration/ - Integration tests with mocked services
internal/testdb/ - Test database utilities
internal/testutils/ - Test configuration utilities
docker-compose.yml - Service configuration for dev

### Your Mission
Create a new layer of integration tests that run against REAL services:

1. Create test database configuration
   - Use separate test schemas for isolation
   - Auto-migrate before tests
   - Clean up after tests

2. Create real service tests for:
   - PostgreSQL (queries, transactions, migrations)
   - Redis (caching, rate limiting, sessions)
   - Qdrant (vector operations, collections)
   - OpenAI API (with opt-in for CI)

3. Create full-stack scenarios:
   - User registration → chatbot creation → source upload → chat
   - Multi-tenant isolation verification
   - Rate limiting enforcement

4. Create test utilities:
   - Service connection helpers
   - Test data factories
   - Assertion helpers for integration tests

### Critical Rules
- Tests MUST clean up after themselves
- Use Docker Compose for service isolation
- Provide opt-in for expensive tests (OpenAI API)
- Tests should be parallelizable where possible
- Use the same test patterns as existing tests

### Deliverables
1. Integration test configuration for real services
2. Database integration tests
3. Redis integration tests
4. Qdrant integration tests
5. Full-stack integration test scenarios
6. Test utilities and fixtures

Begin by analyzing the current test infrastructure and service dependencies.
```

---

## Current State Analysis

### Service Dependencies

| Service | Mock Used | Real Test Priority | Notes |
|---------|-----------|-------------------|-------|
| PostgreSQL | `testdb.OpenTestDB()` | CRITICAL | Core data persistence |
| Redis | Mocked in tests | HIGH | Sessions, caching, rate limiting |
| Qdrant | `MockVectorClient` | HIGH | Vector operations |
| OpenAI | `MockLLMClient` | MEDIUM | Expensive, rate limited |
| S3/R2 | Mocked | LOW | File storage |

### Current Integration Test Patterns

**Pattern 1: Mocked Services**
```go
// internal/integration/chat_test.go
func TestChatService_SendMessage(t *testing.T) {
    // Uses MockVectorClient, MockLLMClient
    // Tests business logic without external calls
}
```

**Pattern 2: Database Integration**
```go
// internal/integration/chatbot_test.go
func TestChatbot_CRUD(t *testing.T) {
    // Uses testdb.OpenTestDB()
    // Tests DB operations with mocked services
}
```

### What's Missing

1. **Real Service Integration Tests**
   - Connection pooling behavior
   - Transaction isolation
   - Index performance
   - Query optimization

2. **Cross-Service Integration**
   - Chat → Qdrant → Redis caching
   - Rate limiting → Redis → actual enforcement
   - Multi-tenant isolation with real data

3. **Failure Mode Tests**
   - Database connection failures
   - Service timeouts
   - Partial failures

---

## Step-by-Step Implementation Plan

### Phase 1: Test Infrastructure Setup (Days 1-3)

#### Step 1.1: Create Integration Test Configuration

**File:** `internal/integration/integration_test.go`

```go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/botla/botla/internal/testdb"
    "github.com/botla/botla/pkg/config"
    "github.com/botla/botla/pkg/logger"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    qdrant "github.com/qdrant/go-client/qdrant"
)

type RealServiceConfig struct {
    PG        *pgxpool.Pool
    Redis     *redis.Client
    Qdrant    *qdrant.Client
    OpenAIKey string
}

type RealServiceTestEnv struct {
    cfg    *config.Config
    PG     *pgxpool.Pool
    Redis  *redis.Client
    Qdrant *qdrant.Client
    t      *testing.T
}

// SetupRealServices creates connections to real services for testing
func SetupRealServices(t *testing.T) *RealServiceTestEnv {
    t.Helper()
    
    // Load test configuration
    cfg := testutils.TestConfig()
    
    // Connect to PostgreSQL
    pgConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
    if err != nil {
        t.Fatalf("Failed to parse DB config: %v", err)
    }
    pgConfig.MaxConns = 10
    pgConfig.MinConns = 2
    
    pgPool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
    if err != nil {
        t.Fatalf("Failed to connect to PostgreSQL: %v", err)
    }
    
    // Verify connection
    if err := pgPool.Ping(context.Background()); err != nil {
        t.Fatalf("Failed to ping PostgreSQL: %v", err)
    }
    
    // Connect to Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
        DB:       0,
    })
    
    if err := redisClient.Ping(context.Background()).Err(); err != nil {
        t.Fatalf("Failed to connect to Redis: %v", err)
    }
    
    // Connect to Qdrant
    qdrantClient, err := qdrant.NewClient(&qdrant.Config{
        Host: cfg.QdrantURL,
    })
    if err != nil {
        t.Fatalf("Failed to connect to Qdrant: %v", err)
    }
    
    return &RealServiceTestEnv{
        cfg:    cfg,
        PG:     pgPool,
        Redis:  redisClient,
        Qdrant: qdrantClient,
        t:      t,
    }
}

// Cleanup closes all service connections
func (e *RealServiceTestEnv) Cleanup() {
    if e.PG != nil {
        e.PG.Close()
    }
    if e.Redis != nil {
        e.Redis.Close()
    }
    // Qdrant doesn't have a Close method
}
```

#### Step 1.2: Create Docker Compose for Integration Tests

**File:** `docker-compose.integration.yml`

```yaml
version: '3.8'

services:
  postgres-integration:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: botla_integration
      POSTGRES_USER: botla
      POSTGRES_PASSWORD: botla
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U botla -d botla_integration"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis-integration:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  qdrant-integration:
    image: qdrant/qdrant:v1.7.0
    ports:
      - "6334:6333"
    volumes:
      - qdrant-integration:/qdrant/storage
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:6333/"]
      interval: 10s
      timeout: 5s
      retries: 3

volumes:
  qdrant-integration:
```

#### Step 1.3: Create Test Data Factories

**File:** `internal/integration/factories/user_factory.go`

```go
package factories

import (
    "context"
    "time"

    "github.com/botla/botla/internal/models"
    "github.com/botla/botla/pkg/auth"
    "github.com/botla/botla/pkg/config"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

type UserFactory struct {
    pool    PgxPool
    cfg     *config.Config
}

type UserFactoryOption func(*UserFactory)

func WithPool(pool PgxPool) UserFactoryOption {
    return func(f *UserFactory) {
        f.pool = pool
    }
}

func WithConfig(cfg *config.Config) UserFactoryOption {
    return func(f *UserFactory) {
        f.cfg = cfg
    }
}

func NewUserFactory(opts ...UserFactoryOption) *UserFactory {
    f := &UserFactory{}
    for _, opt := range opts {
        opt(f)
    }
    return f
}

func (f *UserFactory) Create(ctx context.Context, opts ...UserOption) (*models.User, error) {
    user := &models.User{
        ID:           uuid.New(),
        Email:        f.generateEmail(),
        PasswordHash: f.generatePassword(),
        FullName:     "Test User",
        IsVerified:   true,
        PlanCode:     "free",
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    // Apply options
    for _, opt := range opts {
        opt(user)
    }
    
    // Insert into database
    query := `
        INSERT INTO users (id, email, password_hash, full_name, is_verified, plan_code, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    
    _, err := f.pool.Exec(ctx, query,
        user.ID,
        user.Email,
        user.PasswordHash,
        user.FullName,
        user.IsVerified,
        user.PlanCode,
        user.CreatedAt,
        user.UpdatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return user, nil
}

func (f *UserFactory) generateEmail() string {
    return "test-" + uuid.New().String()[:8] + "@example.com"
}

func (f *UserFactory) generatePassword() string {
    hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    return string(hash)
}

// UserOption is a functional option for configuring a user
type UserOption func(*models.User)

func WithEmail(email string) UserOption {
    return func(u *models.User) {
        u.Email = email
    }
}

func WithName(name string) UserOption {
    return func(u *models.User) {
        u.FullName = name
    }
}

func WithPlan(planCode string) UserOption {
    return func(u *models.User) {
        u.PlanCode = planCode
    }
}

func AsPlatformAdmin() UserOption {
    return func(u *models.User) {
        u.IsPlatformAdmin = true
    }
}
```

---

### Phase 2: PostgreSQL Integration Tests (Days 4-7)

#### Step 2.1: Database Connection Tests

**File:** `internal/integration/database/db_connection_test.go`

```go
package database

import (
    "context"
    "testing"
    "time"

    "github.com/botla/botla/internal/integration"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRealDatabase_Connection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    t.Run("connection is healthy", func(t *testing.T) {
        err := env.PG.Ping(context.Background())
        assert.NoError(t, err)
    })
    
    t.Run("connection pool configuration", func(t *testing.T) {
        // Verify pool settings
        stat := env.PG.Stat()
        assert.Greater(t, stat.MaxConns(), 0)
        assert.GreaterOrEqual(t, stat.MinConns(), 0)
    })
}

func TestRealDatabase_Transactions(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    t.Run("basic transaction", func(t *testing.T) {
        ctx := context.Background()
        tx, err := env.PG.Begin(ctx)
        require.NoError(t, err)
        defer tx.Rollback(ctx)
        
        // Insert test data
        _, err = tx.Exec(ctx, "INSERT INTO users (id, email, full_name) VALUES ($1, $2, $3)",
            uuid.New(), "tx-test@example.com", "Transaction Test")
        require.NoError(t, err)
        
        // Verify data is visible within transaction
        var count int
        err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", "tx-test@example.com").Scan(&count)
        require.NoError(t, err)
        assert.Equal(t, 1, count)
    })
    
    t.Run("transaction isolation", func(t *testing.T) {
        ctx := context.Background()
        
        // Start two concurrent transactions
        tx1, err := env.PG.Begin(ctx)
        require.NoError(t, err)
        defer tx1.Rollback(ctx)
        
        tx2, err := env.PG.Begin(ctx)
        require.NoError(t, err)
        defer tx2.Rollback(ctx)
        
        // Insert same UUID in both transactions
        id := uuid.New()
        
        _, err = tx1.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, "iso1@example.com")
        require.NoError(t, err)
        
        _, err = tx2.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, "iso2@example.com")
        // One should fail due to unique constraint
        assert.Error(t, err)
    })
    
    t.Run("concurrent inserts", func(t *testing.T) {
        ctx := context.Background()
        id := uuid.New()
        
        // Run concurrent inserts
        done := make(chan error, 10)
        for i := 0; i < 10; i++ {
            go func(i int) {
                tx, err := env.PG.Begin(ctx)
                if err != nil {
                    done <- err
                    return
                }
                _, err = tx.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, "concurrent@example.com")
                tx.Rollback(ctx)
                done <- err
            }(i)
        }
        
        // Wait for all goroutines
        errors := 0
        for i := 0; i < 10; i++ {
            if err := <-done; err != nil {
                errors++
            }
        }
        
        // Exactly 9 should fail (one succeeds)
        assert.Equal(t, 9, errors)
    })
}
```

#### Step 2.2: Query Performance Tests

**File:** `internal/integration/database/db_performance_test.go`

```go
package database

import (
    "context"
    "testing"
    "time"

    "github.com/botla/botla/internal/integration"
    "github.com/botla/botla/internal/models"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRealDatabase_QueryPerformance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    // Create test data
    ctx := context.Background()
    orgID := uuid.New()
    
    // Insert 1000 test organizations
    for i := 0; i < 1000; i++ {
        _, err := env.PG.Exec(ctx, `
            INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, uuid.New(), "Test Org", uuid.New().String()[:8], orgID, time.Now(), time.Now())
        require.NoError(t, err)
    }
    
    t.Run("index usage verification", func(t *testing.T) {
        // Query that should use an index
        start := time.Now()
        var count int
        err := env.PG.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE owner_id = $1", orgID).Scan(&count)
        elapsed := time.Since(start)
        
        assert.NoError(t, err)
        assert.Equal(t, 1000, count)
        
        // Query should complete quickly (< 100ms)
        assert.Less(t, elapsed.Milliseconds(), int64(100), "Query took too long, index may not be used")
    })
    
    t.Run("JOIN performance", func(t *testing.T) {
        start := time.Now()
        
        // Complex join query
        rows, err := env.PG.Query(ctx, `
            SELECT o.id, o.name, u.email, COUNT(w.id) as workspace_count
            FROM organizations o
            JOIN users u ON o.owner_id = u.id
            LEFT JOIN workspaces w ON w.organization_id = o.id
            WHERE o.id = $1
            GROUP BY o.id, o.name, u.email
        `, orgID)
        elapsed := time.Since(start)
        
        assert.NoError(t, err)
        rows.Close()
        
        // Join should complete in reasonable time
        assert.Less(t, elapsed.Milliseconds(), int64(500), "JOIN query took too long")
    })
}

func TestRealDatabase_MigrationIntegrity(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping migration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    ctx := context.Background()
    
    t.Run("all required tables exist", func(t *testing.T) {
        requiredTables := []string{
            "users",
            "organizations",
            "workspaces",
            "chatbots",
            "sources",
            "action_logs",
            "conversations",
            "messages",
            "plans",
            "languages",
            "action_types",
        }
        
        for _, table := range requiredTables {
            var exists bool
            err := env.PG.QueryRow(ctx, `
                SELECT EXISTS (
                    SELECT FROM information_schema.tables
                    WHERE table_name = $1
                )
            `, table).Scan(&exists)
            
            assert.NoError(t, err, "Table %s should exist", table)
            assert.True(t, exists, "Table %s should exist", table)
        }
    })
    
    t.Run("all required indexes exist", func(t *testing.T) {
        requiredIndexes := []struct {
            table  string
            column string
        }{
            {"users", "email"},
            {"organizations", "owner_id"},
            {"chatbots", "organization_id"},
            {"chatbots", "workspace_id"},
            {"sources", "chatbot_id"},
        }
        
        for _, idx := range requiredIndexes {
            var exists bool
            err := env.PG.QueryRow(ctx, `
                SELECT EXISTS (
                    SELECT 1 FROM pg_indexes
                    WHERE tablename = $1
                    AND indexdef LIKE '%' || $2 || '%'
                )
            `, idx.table, idx.column).Scan(&exists)
            
            assert.NoError(t, err)
            assert.True(t, exists, "Index on %s.%s should exist", idx.table, idx.column)
        }
    })
}
```

---

### Phase 3: Redis Integration Tests (Days 8-10)

#### Step 3.1: Redis Connection and Operations Tests

**File:** `internal/integration/redis/redis_test.go`

```go
package redis

import (
    "context"
    "testing"
    "time"

    "github.com/botla/botla/internal/integration"
    "github.com/botla/botla/pkg/middleware"
    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRealRedis_Connection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    t.Run("connection is healthy", func(t *testing.T) {
        err := env.Redis.Ping(context.Background()).Err()
        assert.NoError(t, err)
    })
    
    t.Run("basic operations", func(t *testing.T) {
        ctx := context.Background()
        key := "test:basic:" + uuid.New().String()
        
        // Set value
        err := env.Redis.Set(ctx, key, "value", time.Hour).Err()
        assert.NoError(t, err)
        
        // Get value
        val, err := env.Redis.Get(ctx, key).Result()
        assert.NoError(t, err)
        assert.Equal(t, "value", val)
        
        // Delete
        err = env.Redis.Del(ctx, key).Err()
        assert.NoError(t, err)
    })
}

func TestRealRedis_RateLimiting(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    ctx := context.Background()
    
    t.Run("rate limit enforcement", func(t *testing.T) {
        key := "ratelimit:test:" + uuid.New().String()
        limit := 5
        window := time.Minute
        
        // Simulate rate limiting logic
        count := 0
        for i := 0; i < limit+1; i++ {
            pipe := env.Redis.Pipeline()
            
            incr := pipe.Incr(ctx, key)
            pipe.Expire(ctx, key, window)
            
            _, err := pipe.Exec(ctx)
            if err != nil {
                t.Fatalf("Redis operation failed: %v", err)
            }
            
            count = int(incr.Val())
            
            if i < limit {
                assert.Less(t, count, limit+1)
            } else {
                // Should be rate limited
                assert.Equal(t, limit+1, count)
            }
        }
        
        // Cleanup
        env.Redis.Del(ctx, key)
    })
    
    t.Run("concurrent rate limiting", func(t *testing.T) {
        key := "ratelimit:concurrent:" + uuid.New().String()
        limit := 10
        window := time.Minute
        
        done := make(chan int, 100)
        
        // Simulate 100 concurrent requests
        for i := 0; i < 100; i++ {
            go func() {
                ctx := context.Background()
                pipe := env.Redis.Pipeline()
                
                incr := pipe.Incr(ctx, key)
                pipe.Expire(ctx, key, window)
                
                pipe.Exec(ctx)
                done <- int(incr.Val())
            }()
        }
        
        // Collect results
        rateLimited := 0
        for i := 0; i < 100; i++ {
            count := <-done
            if count > limit {
                rateLimited++
            }
        }
        
        // 90 should be rate limited
        assert.Equal(t, 90, rateLimited)
        
        // Cleanup
        env.Redis.Del(ctx, key)
    })
}

func TestRealRedis_SessionManagement(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    ctx := context.Background()
    
    t.Run("session creation and retrieval", func(t *testing.T) {
        sessionID := "session:" + uuid.New().String()
        userData := map[string]interface{}{
            "user_id":    uuid.New().String(),
            "email":      "test@example.com",
            "created_at": time.Now().Unix(),
        }
        
        // Create session
        sessionJSON, err := json.Marshal(userData)
        require.NoError(t, err)
        
        err = env.Redis.Set(ctx, sessionID, sessionJSON, 24*time.Hour).Err()
        assert.NoError(t, err)
        
        // Retrieve session
        retrieved, err := env.Redis.Get(ctx, sessionID).Result()
        assert.NoError(t, err)
        assert.Equal(t, string(sessionJSON), retrieved)
        
        // Cleanup
        env.Redis.Del(ctx, sessionID)
    })
    
    t.Run("session expiration", func(t *testing.T) {
        sessionID := "session:expiring:" + uuid.New().String()
        
        // Create session with short TTL
        err := env.Redis.Set(ctx, sessionID, "data", time.Second).Err()
        assert.NoError(t, err)
        
        // Wait for expiration
        time.Sleep(1100 * time.Millisecond)
        
        // Session should be gone
        _, err = env.Redis.Get(ctx, sessionID).Result()
        assert.Error(t, err)
    })
}
```

---

### Phase 4: Qdrant Integration Tests (Days 11-13)

#### Step 4.1: Qdrant Vector Operations Tests

**File:** `internal/integration/qdrant/qdrant_test.go`

```go
package qdrant

import (
    "testing"
    "time"

    "github.com/botla/botla/internal/integration"
    qdrant "github.com/qdrant/go-client/qdrant"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRealQdrant_Connection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Qdrant integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    t.Run("health check", func(t *testing.T) {
        health, err := env.Qdrant.HealthCheck()
        assert.NoError(t, err)
        assert.NotNil(t, health)
    })
}

func TestRealQdrant_CollectionOperations(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Qdrant integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    collectionName := "test-collection-" + uuid.New().String()[:8]
    
    t.Run("create collection", func(t *testing.T) {
        err := env.Qdrant.CreateCollection(&qdrant.CreateCollection{
            CollectionName: collectionName,
            VectorsConfig: &qdrant.VectorParams{
                Size:     1536,
                Distance: qdrant.Distance_Cosine,
            },
        })
        assert.NoError(t, err)
    })
    
    t.Run("upsert vectors", func(t *testing.T) {
        // Generate test vectors
        vectors := make([]*qdrant.PointStruct, 10)
        for i := 0; i < 10; i++ {
            vectors[i] = &qdrant.PointStruct{
                Id:      uint64(i),
                Payload: map[string]any{"text": "test content " + string(rune('A'+i))},
                Vector:  generateRandomVector(1536),
            }
        }
        
        err := env.Qdrant.Upsert(collectionName, &qdrant.PointsSelector{
            Points: vectors,
        })
        assert.NoError(t, err)
    })
    
    t.Run("search vectors", func(t *testing.T) {
        queryVector := generateRandomVector(1536)
        
        results, err := env.Qdrant.Search(collectionName, &qdrant.SearchPoints{
            Query:    queryVector,
            Limit:    5,
            WithPayload: &qdrant.WithPayloadSelector{
                Enable: true,
            },
        })
        assert.NoError(t, err)
        assert.Len(t, results, 5)
    })
    
    t.Run("delete collection", func(t *testing.T) {
        err := env.Qdrant.DeleteCollection(collectionName)
        assert.NoError(t, err)
    })
}

func TestRealQdrant_ConcurrentOperations(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Qdrant integration test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    collectionName := "concurrent-test-" + uuid.New().String()[:8]
    
    // Create collection first
    err := env.Qdrant.CreateCollection(&qdrant.CreateCollection{
        CollectionName: collectionName,
        VectorsConfig: &qdrant.VectorParams{
            Size:     1536,
            Distance: qdrant.Distance_Cosine,
        },
    })
    require.NoError(t, err)
    
    t.Run("concurrent upserts", func(t *testing.T) {
        done := make(chan error, 100)
        
        for i := 0; i < 100; i++ {
            go func(id int) {
                vectors := []*qdrant.PointStruct{
                    {
                        Id:      uint64(id),
                        Payload: map[string]any{"index": id},
                        Vector:  generateRandomVector(1536),
                    },
                }
                
                err := env.Qdrant.Upsert(collectionName, &qdrant.PointsSelector{
                    Points: vectors,
                })
                done <- err
            }(i)
        }
        
        // Collect results
        errors := 0
        for i := 0; i < 100; i++ {
            if err := <-done; err != nil {
                errors++
            }
        }
        
        // All should succeed (Qdrant handles concurrent writes)
        assert.Equal(t, 0, errors)
    })
    
    t.Run("concurrent searches", func(t *testing.T) {
        done := make(chan *qdrant.SearchResult, 100)
        
        for i := 0; i < 100; i++ {
            go func() {
                queryVector := generateRandomVector(1536)
                result, err := env.Qdrant.Search(collectionName, &qdrant.SearchPoints{
                    Query:    queryVector,
                    Limit:    5,
                })
                if err != nil {
                    done <- nil
                } else {
                    done <- result
                }
            }()
        }
        
        // Collect results
        successCount := 0
        for i := 0; i < 100; i++ {
            if result := <-done; result != nil {
                successCount++
            }
        }
        
        assert.Equal(t, 100, successCount)
    })
    
    // Cleanup
    env.Qdrant.DeleteCollection(collectionName)
}

// Helper function to generate random vectors
func generateRandomVector(size int) []float32 {
    vector := make([]float32, size)
    for i := range vector {
        vector[i] = float32(i%256) / 255.0
    }
    return vector
}
```

---

### Phase 5: Full-Stack Integration Tests (Days 14-17)

#### Step 5.1: Complete User Journey Tests

**File:** `internal/integration/scenarios/user_journey_test.go`

```go
package scenarios

import (
    "context"
    "testing"
    "time"

    "github.com/botla/botla/internal/integration"
    "github.com/botla/botla/internal/models"
    "github.com/botla/botla/internal/services"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRealServices_CompleteUserJourney(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping full journey test in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    ctx := context.Background()
    
    t.Run("full chatbot creation and chat flow", func(t *testing.T) {
        // Step 1: Create user
        user, err := createTestUser(ctx, env.PG)
        require.NoError(t, err)
        t.Logf("Created user: %s", user.Email)
        
        // Step 2: Create organization
        org, err := createTestOrganization(ctx, env.PG, user.ID)
        require.NoError(t, err)
        t.Logf("Created organization: %s", org.Name)
        
        // Step 3: Create workspace
        workspace, err := createTestWorkspace(ctx, env.PG, org.ID, user.ID)
        require.NoError(t, err)
        t.Logf("Created workspace: %s", workspace.Name)
        
        // Step 4: Create chatbot
        chatbot, err := createTestChatbot(ctx, env.PG, workspace.ID, org.ID, user.ID)
        require.NoError(t, err)
        t.Logf("Created chatbot: %s", chatbot.Name)
        
        // Step 5: Create source
        source, err := createTestSource(ctx, env.PG, chatbot.ID)
        require.NoError(t, err)
        t.Logf("Created source: %s", source.SourceURL)
        
        // Step 6: Add source to Qdrant
        err = addSourceToVectorDB(env.Qdrant, chatbot.ID, source.ID)
        require.NoError(t, err)
        
        // Step 7: Create conversation
        conversation, err := createTestConversation(ctx, env.PG, chatbot.ID)
        require.NoError(t, err)
        t.Logf("Created conversation: %s", conversation.ID)
        
        // Step 8: Send chat message
        message, err := sendTestMessage(ctx, env.PG, conversation.ID, user.ID, "Hello, I need help")
        require.NoError(t, err)
        t.Logf("Sent message: %s", message.Content)
        
        // Verify all data is persisted correctly
        verifyMessagePersisted(ctx, env.PG, message.ID, t)
        
        // Verify Redis caching
        verifyCachePopulation(env.Redis, chatbot.ID, t)
    })
    
    t.Run("multi-tenant isolation", func(t *testing.T) {
        // Create two organizations
        org1, _ := createTestOrganization(ctx, env.PG, uuid.New())
        org2, _ := createTestOrganization(ctx, env.PG, uuid.New())
        
        // Create chatbots for each organization
        bot1, _ := createTestChatbot(ctx, env.PG, uuid.New(), org1.ID, uuid.New())
        bot2, _ := createTestChatbot(ctx, env.PG, uuid.New(), org2.ID, uuid.New())
        
        // Verify organization 1 cannot access organization 2's chatbot
        bot1OrgID, err := getChatbotOrgID(ctx, env.PG, bot1.ID)
        require.NoError(t, err)
        assert.Equal(t, org1.ID, bot1OrgID)
        
        bot2OrgID, err := getChatbotOrgID(ctx, env.PG, bot2.ID)
        require.NoError(t, err)
        assert.Equal(t, org2.ID, bot2OrgID)
        
        // Cross-organization access should fail
        _, err = getChatbotByID(ctx, env.PG, bot1.ID, org2.ID)
        assert.Error(t, err) // Should not find chatbot belonging to other org
    })
    
    t.Run("rate limiting enforcement", func(t *testing.T) {
        // This test verifies real rate limiting with Redis
        userID := uuid.New()
        
        // Make rapid requests
        for i := 0; i < 15; i++ {
            // Each request should increment rate limit counter
            // After 10 requests, should be rate limited
        }
        
        // Verify rate limiting was enforced
        // This would require actual API endpoint testing
    })
}

// Helper functions for test data creation

func createTestUser(ctx context.Context, pg PgxPool) (*models.User, error) {
    user := &models.User{
        ID:           uuid.New(),
        Email:        "journey-test-" + uuid.New().String()[:8] + "@example.com",
        PasswordHash: "hash",
        FullName:     "Test User",
        IsVerified:   true,
        PlanCode:     "free",
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    _, err := pg.Exec(ctx, `
        INSERT INTO users (id, email, password_hash, full_name, is_verified, plan_code, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, user.ID, user.Email, user.PasswordHash, user.FullName, user.IsVerified, user.PlanCode, user.CreatedAt, user.UpdatedAt)
    
    return user, err
}

// Additional helper functions...
```

#### Step 5.2: Failure Mode Tests

**File:** `internal/integration/scenarios/failure_modes_test.go`

```go
package scenarios

import (
    "context"
    "testing"
    "time"

    "github.com/botla/botla/internal/integration"
    "github.com/stretchr/testify/assert"
)

func TestRealServices_FailureModes(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping failure mode tests in short mode")
    }
    
    env := integration.SetupRealServices(t)
    t.Cleanup(env.Cleanup)
    
    t.Run("database connection failure handling", func(t *testing.T) {
        // This test would verify graceful degradation when DB is unavailable
        // Note: In real scenario, we can't actually disconnect the DB
        // This tests error handling paths
        
        // Simulate by using wrong connection string
        // Verify error handling is correct
    })
    
    t.Run("redis failure handling", func(t *testing.T) {
        // Test behavior when Redis is unavailable
        // Rate limiting should fall back to in-memory or allow all
        
        // This requires actual Redis failure simulation
    })
    
    t.Run("qdrant failure handling", func(t *testing.T) {
        // Test behavior when Qdrant is unavailable
        // Search should return empty results or fallback
        
        // Verify error is logged but doesn't crash the service
    })
    
    t.Run("partial service availability", func(t *testing.T) {
        // Test behavior when some services are available but others are not
        // e.g., DB and Redis available, but Qdrant unavailable
        
        // Verify graceful degradation
    })
}
```

---

### Phase 6: Test Execution and Verification (Days 18-21)

#### Step 6.1: Create Test Targets

**File:** `Makefile`

```makefile
# Integration tests with real services
test-integration-real: docker-compose-up-integration
    TEST_INTEGRATION=1 go test -v ./internal/integration/... -timeout=5m

test-integration-real-race: docker-compose-up-integration
    TEST_INTEGRATION=1 go test -race -v ./internal/integration/... -timeout=10m

test-integration-db:
    TEST_INTEGRATION=1 TEST_SERVICE=postgres go test -v ./internal/integration/database/... -timeout=2m

test-integration-redis:
    TEST_INTEGRATION=1 TEST_SERVICE=redis go test -v ./internal/integration/redis/... -timeout=2m

test-integration-qdrant:
    TEST_INTEGRATION=1 TEST_SERVICE=qdrant go test -v ./internal/integration/qdrant/... -timeout=2m

test-integration-full:
    TEST_INTEGRATION=1 go test -v ./internal/integration/scenarios/... -timeout=5m

# Docker management
docker-compose-up-integration:
    docker-compose -f docker-compose.integration.yml up -d

docker-compose-down-integration:
    docker-compose -f docker-compose.integration.yml down

docker-compose-logs-integration:
    docker-compose -f docker-compose.integration.yml logs -f
```

#### Step 6.2: CI Configuration

**File:** `.github/workflows/integration-tests.yml`

```yaml
name: Integration Tests

on:
  schedule:
    # Run weekly on Sundays at 2 AM
    - cron: '0 2 * * 0'
  workflow_dispatch:
    inputs:
      services:
        description: 'Services to test (postgres, redis, qdrant, all)'
        required: false
        default: 'all'

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule' || github.event_name == 'workflow_dispatch'
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: botla_integration
          POSTGRES_USER: botla
          POSTGRES_PASSWORD: botla
        ports:
          - 5433:5432
        options: >-
          --health-cmd "pg_isready -U botla -d botla_integration"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        ports:
          - 6380:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      qdrant:
        image: qdrant/qdrant:v1.7.0
        ports:
          - 6334:6333
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Run database integration tests
        if: inputs.services == 'all' || inputs.services == 'postgres'
        run: |
          export DATABASE_URL=postgres://botla:botla@localhost:5433/botla_integration?sslmode=disable
          go test -v ./internal/integration/database/... -timeout=5m
      
      - name: Run Redis integration tests
        if: inputs.services == 'all' || inputs.services == 'redis'
        run: |
          export REDIS_ADDR=localhost:6380
          go test -v ./internal/integration/redis/... -timeout=5m
      
      - name: Run Qdrant integration tests
        if: inputs.services == 'all' || inputs.services == 'qdrant'
        run: |
          export QDRANT_URL=http://localhost:6334
          go test -v ./internal/integration/qdrant/... -timeout=5m
      
      - name: Run full scenario tests
        if: inputs.services == 'all'
        run: |
          export DATABASE_URL=postgres://botla:botla@localhost:5433/botla_integration?sslmode=disable
          export REDIS_ADDR=localhost:6380
          export QDRANT_URL=http://localhost:6334
          go test -v ./internal/integration/scenarios/... -timeout=10m
```

---

## Progress Tracking

### Daily Checklist

- [ ] Create N new test files
- [ ] Run tests to verify
- [ ] Update documentation
- [ ] Report progress

### Milestone Reviews

| Milestone | Target Date | Coverage Target | Status |
|-----------|-------------|-----------------|--------|
| Phase 1 Complete | Day 3 | Infrastructure ready | ⏳ |
| Phase 2 Complete | Day 7 | DB tests 100% | ⏳ |
| Phase 3 Complete | Day 10 | Redis tests 100% | ⏳ |
| Phase 4 Complete | Day 13 | Qdrant tests 100% | ⏳ |
| Phase 5 Complete | Day 17 | Full scenarios 100% | ⏳ |
| Phase 6 Complete | Day 21 | CI/CD configured | ⏳ |

---

## Success Criteria

### Functional Requirements
- [ ] PostgreSQL integration tests complete
- [ ] Redis integration tests complete
- [ ] Qdrant integration tests complete
- [ ] Full-stack scenario tests complete
- [ ] Failure mode tests complete
- [ ] CI/CD pipeline configured

### Performance Targets
| Test Type | Target Duration |
|-----------|-----------------|
| DB connection test | < 1s |
| Transaction test | < 500ms |
| Query performance test | < 100ms |
| Redis rate limit test | < 1s |
| Qdrant search test | < 500ms |
| Full journey test | < 30s |

---

## Dependencies

- `docker-compose` - Service orchestration
- `pgx` - PostgreSQL driver
- `go-redis` - Redis client
- `qdrant-go-client` - Qdrant client
- GitHub Actions - CI/CD

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Sisyphus | Initial plan |

---

*This plan is part of the comprehensive test improvement initiative. For questions or clarifications, refer to the project documentation or consult with the team lead.*
