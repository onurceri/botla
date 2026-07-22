package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
)

// TestAuth_SoftDeletedEmailReuse verifies the partial unique index on
// users(email) WHERE deleted_at IS NULL (migration 000003) together with the
// deleted_at IS NULL filters on the register pre-check and login query:
//   - a soft-deleted user can no longer log in;
//   - a soft-deleted user's email can be reused for a fresh registration;
//   - the partial index ignores soft-deleted rows (multiple may share an email)
//     while still enforcing uniqueness among active rows.
func TestAuth_SoftDeletedEmailReuse(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	email := fmt.Sprintf("reuse_%d@example.com", time.Now().UnixNano())
	const password = "Test@123"

	// 1. Register user A.
	userA := registerUser(t, te.DB, te.Server.URL, email, password)
	if userA == "" {
		t.Fatal("register user A failed")
	}

	// 2. Soft-delete A WITHOUT anonymizing the email. (The GDPR path anonymizes the
	//    email too, but the constraint must not depend on that side effect.)
	if _, err := te.DB.Exec("UPDATE users SET deleted_at = NOW() WHERE id = $1", userA); err != nil {
		t.Fatalf("soft-delete user A: %v", err)
	}

	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})

	// 3. Login as A must now fail — soft-deleted accounts cannot authenticate.
	loginA, err := testHTTPPost(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login A: %v", err)
	}
	drainBody(loginA)
	if loginA.StatusCode != http.StatusUnauthorized {
		t.Fatalf("login for soft-deleted user: expected 401, got %d", loginA.StatusCode)
	}

	// 4. Re-register the same email as user B — must succeed (the partial index
	//    ignores A's soft-deleted row and the pre-check filters deleted_at IS NULL).
	regB, _ := json.Marshal(map[string]string{"email": email, "password": password, "full_name": "User B"})
	regRes, err := testHTTPPost(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regB))
	if err != nil {
		t.Fatalf("register B: %v", err)
	}
	drainBody(regRes)
	if regRes.StatusCode != http.StatusCreated {
		t.Fatalf("re-register a soft-deleted email: expected 201, got %d", regRes.StatusCode)
	}

	// 5. Login as B must succeed.
	loginB, err := testHTTPPost(te.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login B: %v", err)
	}
	drainBody(loginB)
	if loginB.StatusCode != http.StatusOK {
		t.Fatalf("login for new user B: expected 200, got %d", loginB.StatusCode)
	}

	// 6. The partial index must allow multiple soft-deleted rows sharing the email
	//    (only active rows are constrained). Reuse A's plan_id to satisfy NOT NULL.
	var planID string
	if err := te.DB.QueryRow("SELECT plan_id FROM users WHERE id = $1", userA).Scan(&planID); err != nil {
		t.Fatalf("get plan_id: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := te.DB.Exec(
			`INSERT INTO users (email, password_hash, full_name, plan_id, deleted_at)
			 VALUES ($1, 'hash', 'soft-deleted', $2, NOW())`,
			email, planID,
		); err != nil {
			t.Fatalf("insert soft-deleted duplicate %d should be allowed by the partial index: %v", i, err)
		}
	}

	// 7. A second ACTIVE row with the same email must still be rejected.
	if _, err := te.DB.Exec(
		`INSERT INTO users (email, password_hash, full_name, plan_id) VALUES ($1, 'hash', 'active-dup', $2)`,
		email, planID,
	); err == nil {
		t.Fatal("inserting a second active row with the same email should violate the partial unique index")
	}
}
