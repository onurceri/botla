package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/onurceri/botla-app/pkg/requestid"
)

// RequestIDHeader is the header name for request ID
const RequestIDHeader = "X-Request-ID"

// RequestID middleware generates or extracts a request ID for each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get existing request ID from header
		reqID := r.Header.Get(RequestIDHeader)

		// Generate new one if not provided
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set(RequestIDHeader, reqID)

		// Add to context
		ctx := requestid.ToContext(r.Context(), reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext extracts request ID from context
func RequestIDFromContext(ctx context.Context) string {
	return requestid.FromContext(ctx)
}
