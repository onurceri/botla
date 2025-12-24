package auth

import (
	"testing"
	"time"
)

func TestGenerateAndVerifyToken(t *testing.T) {
	secret := "s"
	id := "user-1"
	tok, err := GenerateToken(secret, id, false, "access", time.Minute)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	claims, err := VerifyToken(secret, tok)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if claims.UserID != id {
		t.Fatalf("want %s got %s", id, claims.UserID)
	}
	if claims.TokenType != "access" {
		t.Fatalf("want access got %s", claims.TokenType)
	}
	if claims.ExpiresAt == nil {
		t.Fatalf("expected exp claim to be set")
	}
	if claims.IssuedAt == nil {
		t.Fatalf("expected iat claim to be set")
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatalf("expected exp in the future, got %v", claims.ExpiresAt.Time)
	}
	if claims.IssuedAt.Time.After(time.Now()) {
		t.Fatalf("expected iat in the past, got %v", claims.IssuedAt.Time)
	}
}

func TestExpiredToken(t *testing.T) {
	secret := "s"
	id := "user-1"
	tok, err := GenerateToken(secret, id, false, "access", -time.Minute)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	claims, err := VerifyToken(secret, tok)
	if err == nil || claims != nil {
		t.Fatalf("expected expired token error")
	}
}
