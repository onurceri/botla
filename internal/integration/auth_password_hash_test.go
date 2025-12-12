package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAuth_Register_PasswordHashing(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)

	email := "hashcheck+" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com"
	password := "mypassword123"
	regBody := map[string]string{"email": email, "password": password, "full_name": "Hash Check User"}
	body, err := json.Marshal(regBody)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	res, err := http.Post(te.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", res.StatusCode)
	}

	var storedHash string
	err = te.DB.QueryRow("SELECT password_hash FROM users WHERE email=$1", email).Scan(&storedHash)
	if err == sql.ErrNoRows {
		t.Fatalf("expected user row for %s, got none", email)
	}
	if err != nil {
		t.Fatalf("query for stored password_hash failed: %v", err)
	}

	if storedHash == "" {
		t.Fatalf("expected non-empty password_hash")
	}
	if storedHash == password {
		t.Fatalf("password_hash should not equal plaintext password")
	}
	if !strings.HasPrefix(storedHash, "$2a$") && !strings.HasPrefix(storedHash, "$2b$") {
		t.Fatalf("expected bcrypt hash starting with $2a$ or $2b$, got %q", storedHash)
	}
}
