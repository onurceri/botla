package middleware

import "net/http"

// SecurityHeadersMiddleware adds essential security headers to all responses.
// Headers added:
// - X-Frame-Options: DENY - Prevents clickjacking
// - X-Content-Type-Options: nosniff - Prevents MIME type sniffing
// - X-XSS-Protection: 1; mode=block - Enables XSS filter in browsers
// - Referrer-Policy: strict-origin-when-cross-origin - Controls referrer information
// - Strict-Transport-Security (HSTS): Only set for HTTPS connections
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Only set HSTS for HTTPS connections (including behind proxy)
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}
