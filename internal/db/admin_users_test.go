package db_test

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
)

func TestAdminUpdateUser(t *testing.T) {
	conn := testdb.OpenTestDB(t)
	uid := createUser(t, conn)

	ctx := context.Background()

	// Test valid update
	updates := map[string]any{
		"full_name": "Updated Name",
	}
	err := db.AdminUpdateUser(ctx, conn, uid, updates)
	assert.NoError(t, err)

	// Verify update
	var name string
	err = conn.QueryRowContext(ctx, "SELECT full_name FROM users WHERE id = $1", uid).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", name)

	// Test ignored key (simulation of injection attempt)
	// Even though the current implementation is "safe" via whitelist, we want to ensure any future implementation is also safe
	// and that invalid keys are simply ignored.
	updates = map[string]any{
		"full_name":                         "Updated Name 2",
		"invalid_col; DROP TABLE users; --": "some value",
	}
	err = db.AdminUpdateUser(ctx, conn, uid, updates)
	assert.NoError(t, err)

	// Verify update
	err = conn.QueryRowContext(ctx, "SELECT full_name FROM users WHERE id = $1", uid).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name 2", name) // Should be updated

	// Verify table still exists (injection failed)
	var count int
	err = conn.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&count)
	assert.NoError(t, err)
	assert.Greater(t, count, 0)
}
