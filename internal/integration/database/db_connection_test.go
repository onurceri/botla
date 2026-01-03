package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/integration"
	"github.com/onurceri/botla-co/internal/testdb"
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
		err := env.PGPool.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("connection pool configuration", func(t *testing.T) {
		stat := env.PGPool.Stat()
		assert.Greater(t, stat.AcquireCount(), int64(0))
		assert.GreaterOrEqual(t, stat.AcquiredConns(), int32(0))
		assert.Equal(t, int32(10), stat.MaxConns())
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
		tx, err := env.DB.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		email := "tx-test-" + uuid.New().String()[:8] + "@example.com"
		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, plan_id, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped)
			VALUES (gen_random_uuid(), $1, 'hash', 'Transaction Test', (SELECT id FROM plans LIMIT 1), false, true, 0, false)
		`, email)
		require.NoError(t, err)

		var count int
		err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("rollback undoes changes", func(t *testing.T) {
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		require.NoError(t, err)

		email := "tx-rollback-" + uuid.New().String()[:8] + "@example.com"
		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, plan_id, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped)
			VALUES (gen_random_uuid(), $1, 'hash', 'Rollback Test', (SELECT id FROM plans LIMIT 1), false, true, 0, false)
		`, email)
		require.NoError(t, err)

		// Rollback the transaction
		err = tx.Rollback()
		require.NoError(t, err)

		// Verify the data was not committed
		var count int
		err = env.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Rolled back data should not be visible")
	})
}

func TestRealDatabase_QueryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := context.Background()

	// Create a user first, then use their ID for organizations
	user := testdb.CreateUser(t, env.DB)
	ownerID := user.ID

	// Insert 100 organizations with same owner
	for i := 0; i < 100; i++ {
		_, err := env.DB.ExecContext(ctx, `
			INSERT INTO organizations (id, name, slug, owner_id, plan_id, created_at, updated_at)
			VALUES (gen_random_uuid(), 'Test Org', gen_random_uuid(), $1, (SELECT id FROM plans LIMIT 1), NOW(), NOW())
		`, ownerID)
		require.NoError(t, err)
	}

	t.Run("index usage verification", func(t *testing.T) {
		start := time.Now()

		var count int
		err := env.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations WHERE owner_id = $1", ownerID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 100, count)

		elapsed := time.Since(start)
		assert.Less(t, elapsed.Milliseconds(), int64(500), "Query took too long, index may not be used")
	})

	t.Run("JOIN performance", func(t *testing.T) {
		testOrgID := uuid.New()

		_, err := env.DB.ExecContext(ctx, `
			INSERT INTO organizations (id, name, slug, owner_id, plan_id, created_at, updated_at)
			VALUES ($1, 'Test Org', gen_random_uuid(), $2, (SELECT id FROM plans LIMIT 1), NOW(), NOW())
		`, testOrgID, user.ID)
		require.NoError(t, err)

		start := time.Now()

		rows, err := env.DB.QueryContext(ctx, `
			SELECT o.id, o.name, u.email
			FROM organizations o
			JOIN users u ON o.owner_id = u.id
			WHERE o.id = $1
			LIMIT 1
		`, testOrgID)
		elapsed := time.Since(start)
		require.NoError(t, err)
		rows.Close()

		assert.Less(t, elapsed.Milliseconds(), int64(500), "JOIN query took too long")
	})
}

func TestRealDatabase_MigrationIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
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
			"data_sources",
			"plans",
			"languages",
		}

		for _, table := range requiredTables {
			var exists bool
			err := env.DB.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = $1
				)
			`, table).Scan(&exists)

			assert.NoError(t, err, "Table check failed for %s", table)
			assert.True(t, exists, "Table %s should exist", table)
		}
	})

	t.Run("all required indexes exist", func(t *testing.T) {
		// Check for primary key indexes (always exist) and common indexes
		requiredIndexes := []struct {
			table string
			index string
		}{
			{"users", "users_pkey"},
			{"users", "idx_users_email"},
			{"organizations", "organizations_pkey"},
			{"chatbots", "chatbots_pkey"},
			{"data_sources", "data_sources_pkey"},
			{"workspaces", "workspaces_pkey"},
		}

		for _, idx := range requiredIndexes {
			var exists bool
			err := env.DB.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM pg_indexes
					WHERE schemaname = 'public'
					AND tablename = $1
					AND indexname = $2
				)
			`, idx.table, idx.index).Scan(&exists)

			assert.NoError(t, err)
			assert.True(t, exists, "Index %s on %s should exist", idx.index, idx.table)
		}
	})

	t.Run("all required foreign keys exist", func(t *testing.T) {
		requiredFKs := []struct {
			table string
			fk    string
		}{
			{"organizations", "organizations_owner_id_fkey"},
			{"workspaces", "workspaces_organization_id_fkey"},
			{"chatbots", "chatbots_user_id_fkey"},
			{"data_sources", "data_sources_chatbot_id_fkey"},
		}

		for _, fk := range requiredFKs {
			var exists bool
			err := env.DB.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM pg_constraint
					WHERE conname = $1
				)
			`, fk.fk).Scan(&exists)

			assert.NoError(t, err)
			assert.True(t, exists, "Foreign key %s on %s should exist", fk.fk, fk.table)
		}
	})
}
