package auth

import "testing"

func TestVerifyToken_WrongSecret(t *testing.T) {
	tok, err := GenerateToken("secret1", "u", "access", 0)
	if err != nil {
		t.Fatalf("gen err: %v", err)
	}
	c, err := VerifyToken("secret2", tok)
	if err == nil || c != nil {
		t.Fatalf("expected error with wrong secret")
	}
}

func TestVerifyToken_Malformed(t *testing.T) {
	c, err := VerifyToken("secret", "not-a-jwt")
	if err == nil || c != nil {
		t.Fatalf("expected parse error for malformed token")
	}
}
