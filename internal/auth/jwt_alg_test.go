package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyToken_RejectsUnexpectedAlgorithms(t *testing.T) {
	t.Parallel()

	const secret = "test-secret-0123456789abcdef0123456789"
	claims := func() Claims {
		now := time.Now()
		return Claims{
			UserID:    "user-1",
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-1",
				Issuer:    JWTIssuer,
				Audience:  jwt.ClaimStrings{JWTAudience},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
			},
		}
	}

	sign := func(t *testing.T, method jwt.SigningMethod, key any) string {
		t.Helper()
		s, err := jwt.NewWithClaims(method, claims()).SignedString(key)
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		return s
	}

	tests := []struct {
		name    string
		token   func(t *testing.T) string
		wantErr bool
	}{
		{name: "HS256 accepted", token: func(t *testing.T) string { return sign(t, jwt.SigningMethodHS256, []byte(secret)) }},
		{name: "HS384 rejected", token: func(t *testing.T) string { return sign(t, jwt.SigningMethodHS384, []byte(secret)) }, wantErr: true},
		{name: "HS512 rejected", token: func(t *testing.T) string { return sign(t, jwt.SigningMethodHS512, []byte(secret)) }, wantErr: true},
		{name: "none rejected", token: func(t *testing.T) string {
			return sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType)
		}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := VerifyToken(secret, tt.token(t))
			if tt.wantErr {
				if err == nil || c != nil {
					t.Fatalf("expected rejection, got claims=%v err=%v", c, err)
				}
				return
			}
			if err != nil || c == nil || c.UserID != "user-1" {
				t.Fatalf("expected valid token, got claims=%v err=%v", c, err)
			}
		})
	}
}
