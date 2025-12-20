package testdb

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestWithTx(t *testing.T) {
	db := OpenTestDB(t)
	called := false
	WithTx(t, db, func(ctx context.Context, tx *sql.Tx) {
		called = true
		// simple query that should work in tx context: no-op select
		var one int
		_ = tx.QueryRowContext(ctx, `SELECT 1`).Scan(&one)
	})
	if !called {
		t.Fatalf("callback not invoked")
	}
}

func TestWithTx_RollbackVerification(t *testing.T) {
	db := OpenTestDB(t)

	// Create a temporary table for this test
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS _test_rollback_check (id SERIAL PRIMARY KEY, value TEXT)`)
	if err != nil {
		t.Fatalf("failed to create temp table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS _test_rollback_check`)
	})

	// Insert inside WithTx - should be rolled back
	WithTx(t, db, func(ctx context.Context, tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO _test_rollback_check (value) VALUES ('should_not_persist')`)
		if err != nil {
			t.Fatalf("insert in tx failed: %v", err)
		}

		// Verify row exists within transaction
		var count int
		err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM _test_rollback_check`).Scan(&count)
		if err != nil {
			t.Fatalf("count in tx failed: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected 1 row in tx, got %d", count)
		}
	})

	// After WithTx, the row should NOT exist (rollback happened)
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM _test_rollback_check WHERE value = 'should_not_persist'`).Scan(&count)
	if err != nil {
		t.Fatalf("count after tx failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after rollback, got %d - rollback did not work!", count)
	}
}

func TestOpenTestDB_Cleanup(t *testing.T) {
	// This test verifies that t.Cleanup() properly closes the connection
	// We can't easily test this directly, but we can verify the pattern works
	db := OpenTestDB(t)

	// Connection should be usable
	var one int
	err := db.QueryRow(`SELECT 1`).Scan(&one)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if one != 1 {
		t.Fatalf("expected 1, got %d", one)
	}
	// Note: db.Close() is called automatically by t.Cleanup()
	// We don't call defer db.Close() here - that's the point of the improvement
}

func TestOpenParallelTestDB(t *testing.T) {
	db := OpenParallelTestDB(t)

	// Verify we can query
	var one int
	err := db.QueryRow(`SELECT 1`).Scan(&one)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if one != 1 {
		t.Fatalf("expected 1, got %d", one)
	}

	// Verify plans were seeded (by migration)
	var planCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM plans`).Scan(&planCount)
	if err != nil {
		t.Fatalf("plan count query failed: %v", err)
	}
	if planCount < 2 {
		t.Fatalf("expected at least 2 plans (seeded by migration), got %d", planCount)
	}
}

func TestOpenParallelTestDB_Isolation(t *testing.T) {
	// Test that two parallel subtests each get their own isolated schema
	// and cannot see each other's data

	t.Run("test_a", func(t *testing.T) {
		t.Parallel()
		db := OpenParallelTestDB(t)

		// Create a unique table in this schema
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS _isolation_test (value TEXT)`)
		if err != nil {
			t.Fatalf("create table failed: %v", err)
		}

		// Insert a value unique to this test
		_, err = db.Exec(`INSERT INTO _isolation_test (value) VALUES ($1)`, "test_a")
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}

		// Query should only return our value
		var value string
		err = db.QueryRow(`SELECT value FROM _isolation_test`).Scan(&value)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}

		if value != "test_a" {
			t.Fatalf("expected 'test_a', got '%s'", value)
		}
	})

	t.Run("test_b", func(t *testing.T) {
		t.Parallel()
		db := OpenParallelTestDB(t)

		// Create a unique table in this schema
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS _isolation_test (value TEXT)`)
		if err != nil {
			t.Fatalf("create table failed: %v", err)
		}

		// Insert a value unique to this test
		_, err = db.Exec(`INSERT INTO _isolation_test (value) VALUES ($1)`, "test_b")
		if err != nil {
			t.Fatalf("insert failed: %v", err)
		}

		// Query should only return our value (not test_a's value)
		var value string
		err = db.QueryRow(`SELECT value FROM _isolation_test`).Scan(&value)
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}

		if value != "test_b" {
			t.Fatalf("expected 'test_b', got '%s' (isolation failure!)", value)
		}
	})
}

func TestGenerateUniqueSchema(t *testing.T) {
	schemas := make(map[string]bool)
	for i := 0; i < 10; i++ {
		schema := generateUniqueSchema(t)
		if schemas[schema] {
			t.Fatalf("duplicate schema generated: %s", schema)
		}
		schemas[schema] = true

		// Verify format
		if len(schema) != 13 { // "test_" (5) + 8 hex chars
			t.Fatalf("unexpected schema length: %s (len=%d)", schema, len(schema))
		}
	}
}
