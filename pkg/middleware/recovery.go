package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/onurceri/botla-co/pkg/logger"
)

func RecoveryMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					// Log the error
					stack := string(debug.Stack())
					log.Error("panic_recovered", map[string]any{
						"error": fmt.Sprintf("%v", err),
						"stack": stack,
					})

					// In dev, show stack trace in response
					if os.Getenv("GO_ENV") == "development" {
						w.Header().Set("Content-Type", "text/plain")
						_, _ = fmt.Fprintf(w, "Panic recovered: %v\n\n%s", err, stack)
					} else {
						// In prod, generic JSON error
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{"error": "Internal Server Error", "code": "INTERNAL_ERROR"}`))
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
