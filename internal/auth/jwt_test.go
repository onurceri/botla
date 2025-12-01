package auth

import (
    "testing"
    "time"
)

func TestGenerateAndVerifyToken(t *testing.T) {
    secret := "s"
    id := "user-1"
    tok, err := GenerateToken(secret, id, "access", time.Minute)
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
}

func TestExpiredToken(t *testing.T) {
    secret := "s"
    id := "user-1"
    tok, err := GenerateToken(secret, id, "access", -time.Minute)
    if err != nil {
        t.Fatalf("generate error: %v", err)
    }
    claims, err := VerifyToken(secret, tok)
    if err == nil || claims != nil {
        t.Fatalf("expected expired token error")
    }
}

