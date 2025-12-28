package guards

import (
	"net/http"

	"github.com/onurceri/botla-co/pkg/middleware"
)

// RequirePlatformAdmin ensures the request is from a platform admin.
// This middleware must be used after AuthMiddleware.
func RequirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !middleware.IsPlatformAdminFromContext(r.Context()) {
			// Using the existing writeAuthError pattern from pkg/middleware/auth.go
			// but since it's not exported, we'll just use http.Error for now or
			// we can make it accessible if needed.
			// Actually, the plan suggested http.Error.
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
