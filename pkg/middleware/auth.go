package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onurceri/botla-co/internal/auth"
)

type contextKey string

const ContextKeyUserID contextKey = "userID"

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
			if h == "" {
				writeAuthError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeAuthError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}
			claims, err := auth.VerifyToken(secret, parts[1])
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
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyUserID)
	s, ok := v.(string)
	return s, ok
}
