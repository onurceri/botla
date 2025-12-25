package testdb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// DefaultTestDBName is the default test database name
	DefaultTestDBName = "botla_test"
	// DefaultTestSchema is the default test schema name (for non-parallel tests)
	DefaultTestSchema = "test"
)

// schemaCreationLock protects concurrent schema creation during parallel tests
var schemaCreationLock sync.Mutex

// cleanupOnce ensures stale schema cleanup only happens once per test run
var cleanupOnce sync.Once

// cleanupStaleSchemas removes any leftover test schemas from previous runs
func cleanupStaleSchemas() {
	// Connect to default DB to perform cleanup
	baseDSN := getTestDSN("")
	db, err := sql.Open("pgx", baseDSN)
	if err != nil {
		// Just log to stderr since we can't use t.Log here easily and don't want to panic
		fmt.Fprintf(os.Stderr, "warning: failed to open db for stale schema cleanup: %v\n", err)
		return
	}
	defer func() { _ = db.Close() }()

	// Find all schemas starting with 'botla_test_'
	rows, err := db.Query(`
		SELECT nspname 
		FROM pg_namespace 
		WHERE nspname LIKE 'botla_test_%' 
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to list stale schemas: %v\n", err)
		return
	}
	defer func() { _ = rows.Close() }()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err == nil {
			schemas = append(schemas, schema)
		}
	}

	for _, schema := range schemas {
		// Drop each stale schema
		_, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to drop stale schema %s: %v\n", schema, err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "cleaned up stale test schema: %s\n", schema)
		}
	}
}

// getTestDSN returns the test database connection string.
func getTestDSN(schema string) string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" || dbName == "botla_dev" {
		dbName = DefaultTestDBName
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "botla"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "botla"
	}

	dsn := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbName + "?sslmode=disable"
	if schema != "" {
		dsn += "&search_path=" + schema
	}
	return dsn
}

// OpenTestDB opens a connection to the test database using the default 'test' schema.
// The returned connection is automatically closed when the test ends via t.Cleanup().
// Use this for tests that run serially or use WithTx for isolation.
func OpenTestDB(t *testing.T) *sql.DB {
	return OpenTestDBWithSchema(t, DefaultTestSchema)
}

// OpenTestDBWithSchema opens a connection to the test database with a specific schema.
// The returned connection is automatically closed when the test ends via t.Cleanup().
func OpenTestDBWithSchema(t *testing.T, schema string) *sql.DB {
	t.Helper()
	if schema == "" {
		schema = DefaultTestSchema
	}

	// Run migrations before connecting
	runMigrations(t, schema)

	dsn := getTestDSN(schema)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	// Register cleanup to ensure DB is closed even if test fails early
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close test db: %v", err)
		}
	})

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("skipping test: database not available: %v", err)
	}

	// Explicitly set search_path
	if _, err := db.Exec("SET search_path TO " + schema); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	return db
}

// OpenParallelTestDB creates an isolated schema for parallel test execution.
// Each call creates a unique schema that is dropped when the test ends.
// Use this for integration tests that need true parallel execution.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    t.Parallel() // Enable parallel execution
//	    db := testdb.OpenParallelTestDB(t)
//	    // ... test code
//	}
func OpenParallelTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Run global cleanup once per test suite execution
	cleanupOnce.Do(cleanupStaleSchemas)

	// Generate unique schema name
	schema := generateUniqueSchema(t)

	// Create the schema (needs lock to prevent race in schema creation)
	schemaCreationLock.Lock()
	db := createSchemaAndConnect(t, schema)
	schemaCreationLock.Unlock()

	// Register cleanup to drop schema when test ends
	t.Cleanup(func() {
		dropSchema(t, db, schema)
	})

	return db
}

// generateUniqueSchema creates a unique schema name for parallel tests
func generateUniqueSchema(t *testing.T) string {
	t.Helper()
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatalf("failed to generate random bytes: %v", err)
	}
	return fmt.Sprintf("botla_test_%s", hex.EncodeToString(bytes))
}

// createSchemaAndConnect creates a new schema, runs migrations, and returns a connection
func createSchemaAndConnect(t *testing.T, schema string) *sql.DB {
	t.Helper()

	// First, connect without schema to create it
	baseDSN := getTestDSN("")
	baseDB, err := sql.Open("pgx", baseDSN)
	if err != nil {
		t.Fatalf("open base db: %v", err)
	}

	// Create the schema
	if _, err = baseDB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q", schema)); err != nil {
		_ = baseDB.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}
	_ = baseDB.Close()

	// Run migrations for the new schema (this seeds plans/languages via migration)
	runMigrations(t, schema)

	// Connect with the new schema
	dsn := getTestDSN(schema)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test db with schema %s: %v", schema, err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("skipping test: database not available: %v", err)
	}

	// Explicitly set search_path
	if _, err := db.Exec("SET search_path TO " + schema); err != nil {
		_ = db.Close()
		t.Fatalf("set search_path to %s: %v", schema, err)
	}

	return db
}

// dropSchema drops the test schema and cleans up
func dropSchema(t *testing.T, db *sql.DB, schema string) {
	t.Helper()

	// Close the connection first
	if err := db.Close(); err != nil {
		t.Logf("warning: failed to close test db: %v", err)
	}

	// Connect without schema to drop it
	baseDSN := getTestDSN("")
	baseDB, err := sql.Open("pgx", baseDSN)
	if err != nil {
		t.Logf("warning: failed to open base db for schema cleanup: %v", err)
		return
	}
	defer func() { _ = baseDB.Close() }()

	// Force disconnect any remaining connections
	_, _ = baseDB.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = current_database()
		AND pid <> pg_backend_pid()
		AND query LIKE '%%search_path%%' || '%s' || '%%'
	`, schema))

	// Drop the schema
	if _, err := baseDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err != nil {
		t.Logf("warning: failed to drop schema %s: %v", schema, err)
	}
}

func runMigrations(t *testing.T, schema string) {
	t.Helper()

	wd, _ := os.Getwd()
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			projectRoot = "../../"
			break
		}
		projectRoot = parent
	}
	migrationsPath := filepath.Join(projectRoot, "db/migrations")

	dbURL := getTestDSN(schema)

	//nolint:gosec // this is a test helper using dynamic path
	cmd := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("migration failed: %s (error: %v)", string(output), err)
	}
}

// WithTx opens a transaction, runs fn, and rolls back afterwards.
// This ensures test data never persists to the database.
func WithTx(t *testing.T, db *sql.DB, fn func(ctx context.Context, tx *sql.Tx)) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	// Always rollback - this is intentional for test isolation
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Logf("warning: tx rollback failed: %v", err)
		}
	})
	fn(ctx, tx)
}
