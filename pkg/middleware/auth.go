package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onurceri/botla-app/internal/auth"
)

type contextKey string

const (
	ContextKeyUserID          contextKey = "userID"
	ContextKeyIsPlatformAdmin contextKey = "isPlatformAdmin"
)

type authErrorResponse struct {
	Error string `json:"error"`
}

func writeAuthError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(authErrorResponse{Error: message})
}

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			var tokenString string
			if h != "" {
				parts := strings.SplitN(h, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					tokenString = parts[1]
				}
			}

			if tokenString == "" {
				// Try cookie
				c, err := r.Cookie("botla_token")
				if err == nil {
					tokenString = c.Value
				}
			}

			if tokenString == "" {
				writeAuthError(w, http.StatusUnauthorized, "Missing authorization header or cookie")
				return
			}

			claims, err := auth.VerifyToken(secret, tokenString)
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					writeAuthError(w, http.StatusUnauthorized, "Token expired")
					return
				}
				writeAuthError(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			if claims.UserID == "" || claims.TokenType != "access" {
				writeAuthError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// If the writer is our statusRecorder (from RequestLogger), set the userID directly
			// so it's available for logging even after this handler returns
			if sr, ok := w.(*statusRecorder); ok {
				sr.SetUserID(claims.UserID)
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyIsPlatformAdmin, claims.IsPlatformAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware attempts to extract the user ID from the token if present,
// but does not enforce authentication.
func OptionalAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			var tokenString string
			if h != "" {
				parts := strings.SplitN(h, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					tokenString = parts[1]
				}
			}

			if tokenString == "" {
				// Try cookie
				c, err := r.Cookie("botla_token")
				if err == nil {
					tokenString = c.Value
				}
			}

			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := auth.VerifyToken(secret, tokenString)
			if err != nil {
				// Ignore invalid tokens in optional auth
				next.ServeHTTP(w, r)
				return
			}
			if claims.UserID == "" || claims.TokenType != "access" {
				next.ServeHTTP(w, r)
				return
			}

			// If the writer is our statusRecorder (from RequestLogger), set the userID directly
			if sr, ok := w.(*statusRecorder); ok {
				sr.SetUserID(claims.UserID)
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyIsPlatformAdmin, claims.IsPlatformAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyUserID)
	s, ok := v.(string)
	return s, ok
}

func IsPlatformAdminFromContext(ctx context.Context) bool {
	v := ctx.Value(ContextKeyIsPlatformAdmin)
	b, ok := v.(bool)
	return ok && b
}
