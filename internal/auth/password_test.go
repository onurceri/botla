package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	h, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if h == "" {
		t.Fatalf("empty hash")
	}
	if !VerifyPassword(h, "secret") {
		t.Fatalf("verify failed")
	}
	if VerifyPassword(h, "wrong") {
		t.Fatalf("verify should fail with wrong password")
	}
}
