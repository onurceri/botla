package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JWTIssuer   = "botla-app"
	JWTAudience = "botla-api"
)

type Claims struct {
	UserID          string
	TokenType       string
	IsPlatformAdmin bool
	jwt.RegisteredClaims
}

func GenerateToken(secret string, userID string, isPlatformAdmin bool, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	jti := hex.EncodeToString(randomBytes)
	claims := Claims{
		UserID:          userID,
		TokenType:       tokenType,
		IsPlatformAdmin: isPlatformAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			Issuer:    JWTIssuer,
			Audience:  jwt.ClaimStrings{JWTAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return s, nil
}

func VerifyToken(secret string, tokenString string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithIssuer(JWTIssuer), jwt.WithAudience(JWTAudience))
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !parsed.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return c, nil
}
