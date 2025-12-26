package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/pkg/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
	userID string
}

func (sr *statusRecorder) SetUserID(id string) {
	sr.userID = id
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	n, err := sr.ResponseWriter.Write(b)
	sr.bytes += n
	if err != nil {
		return n, fmt.Errorf("write response: %w", err)
	}
	return n, nil
}

func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(sr, r)
			dur := time.Since(start)

			// Use userID from recorder if set (by AuthMiddleware), otherwise try context
			uid := sr.userID
			if uid == "" {
				uid, _ = UserIDFromContext(r.Context())
			}

			log.Info("http_request", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      sr.status,
				"bytes":       sr.bytes,
				"duration_ms": dur.Milliseconds(),
				"userID":      uid,
			})
		})
	}
}
