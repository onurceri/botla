package testdb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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

// activeSchemas tracks schemas currently in use by this process to prevent cleanup
var activeSchemas = struct {
	sync.RWMutex
	schemas map[string]time.Time
}{schemas: make(map[string]time.Time)}

// migrationOnce tracks which schemas have been migrated in this process
var migrationOnce = struct {
	sync.Mutex
	migrated map[string]bool
}{migrated: make(map[string]bool)}

// registerActiveSchema marks a schema as in-use by this process
func registerActiveSchema(schema string) {
	activeSchemas.Lock()
	defer activeSchemas.Unlock()
	activeSchemas.schemas[schema] = time.Now()
}

// unregisterActiveSchema removes a schema from the active list
func unregisterActiveSchema(schema string) {
	activeSchemas.Lock()
	defer activeSchemas.Unlock()
	delete(activeSchemas.schemas, schema)
}

// CleanupAllTestSchemas removes ALL test schemas (for use in CI or explicit cleanup)
// This is more aggressive than cleanupStaleSchemas and ignores the time threshold.
func CleanupAllTestSchemas() error {
	baseDSN := getTestDSN("")
	db, err := sql.Open("pgx", baseDSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(`
		SELECT nspname 
		FROM pg_namespace 
		WHERE nspname LIKE 'botla_test_%'
		  AND nspname NOT IN ('botla_test')
	`)
	if err != nil {
		return fmt.Errorf("list schemas: %w", err)
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
		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: failed to drop schema %s: %v\n", schema, err)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "cleaned up test schema: %s\n", schema)
		}
	}

	return nil
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
		dsn += "&options=" + url.QueryEscape("-c search_path="+schema)
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

	// Acquire lock to prevent concurrent schema creation/migration
	schemaCreationLock.Lock()
	defer schemaCreationLock.Unlock()

	// Check if we've already migrated this schema in this process
	migrationOnce.Lock()
	alreadyMigrated := migrationOnce.migrated[schema]
	migrationOnce.Unlock()

	// Create schema if it doesn't exist
	baseDSN := getTestDSN("")
	baseDB, err := sql.Open("pgx", baseDSN)
	if err != nil {
		t.Fatalf("open base db for schema creation: %v", err)
	}
	if _, err = baseDB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q", schema)); err != nil {
		_ = baseDB.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}
	_ = baseDB.Close()

	// Run migrations only if not already done
	if !alreadyMigrated {
		runMigrations(t, schema)
		migrationOnce.Lock()
		migrationOnce.migrated[schema] = true
		migrationOnce.Unlock()
	}

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
// Use this for tests that need true parallel execution.
//
// Note: Stale schemas from previous runs are NOT automatically cleaned up.
// Use CleanupAllTestSchemas() at test suite boundaries for explicit cleanup.
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

	// Generate unique schema name
	schema := generateUniqueSchema(t)

	// Register this schema as active BEFORE creation to prevent race conditions
	// (only works within this process, but better than nothing)
	registerActiveSchema(schema)

	// Create the schema (needs lock to prevent race in schema creation)
	schemaCreationLock.Lock()
	defer schemaCreationLock.Unlock()
	db := createSchemaAndConnect(t, schema)

	// Register cleanup to drop schema when test ends
	t.Cleanup(func() {
		dropSchema(t, db, schema)
		unregisterActiveSchema(schema)
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

	// Drop the schema with retries
	for i := 0; i < 3; i++ {
		if _, err := baseDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema)); err != nil {
			if i == 2 {
				t.Logf("warning: failed to drop schema %s after 3 attempts: %v", schema, err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
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

	for i := 0; i < 3; i++ {
		//nolint:gosec // this is a test helper using dynamic path
		cmd := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "up")
		output, err := cmd.CombinedOutput()
		if err == nil {
			return
		}

		outputStr := string(output)

		// Check if the error is due to a dirty database version
		// This can happen if a previous migration was interrupted mid-way
		if dirtyVersion := extractDirtyVersion(outputStr); dirtyVersion > 0 {
			t.Logf("detected dirty database version %d, forcing version to %d and retrying", dirtyVersion, dirtyVersion-1)
			if forceErr := forceMigrationVersion(migrationsPath, dbURL, dirtyVersion-1); forceErr != nil {
				t.Logf("warning: failed to force migration version: %v", forceErr)
			} else {
				// Retry immediately after forcing version
				continue
			}
		}

		if i == 2 {
			t.Fatalf("migration failed: %s (error: %v)", outputStr, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// extractDirtyVersion parses the dirty version number from migrate error output.
// Returns 0 if no dirty version is found.
// Example: "error: Dirty database version 16. Fix and force version." -> 16
func extractDirtyVersion(output string) int {
	// Look for pattern "Dirty database version N"
	prefix := "Dirty database version "
	idx := strings.Index(output, prefix)
	if idx == -1 {
		return 0
	}

	// Extract the version number
	versionStart := idx + len(prefix)
	versionEnd := versionStart
	for versionEnd < len(output) && output[versionEnd] >= '0' && output[versionEnd] <= '9' {
		versionEnd++
	}

	if versionEnd == versionStart {
		return 0
	}

	version, err := strconv.Atoi(output[versionStart:versionEnd])
	if err != nil {
		return 0
	}
	return version
}

// forceMigrationVersion forces the database to a specific migration version.
// This is used to recover from a dirty state.
func forceMigrationVersion(migrationsPath, dbURL string, version int) error {
	//nolint:gosec // this is a test helper using dynamic path
	cmd := exec.Command("migrate", "-path", migrationsPath, "-database", dbURL, "force", strconv.Itoa(version))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("force version %d failed: %s (error: %w)", version, string(output), err)
	}
	return nil
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
