package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/auth"
)

type contextKey string

const ContextKeyUserID contextKey = "userID"

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			claims, err := auth.VerifyToken(secret, parts[1])
			if err != nil || claims.UserID == "" || claims.TokenType != "access" {
				w.WriteHeader(http.StatusUnauthorized)
				return
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
