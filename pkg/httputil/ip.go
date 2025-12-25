package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ExtractIP extracts the client IP from the request.
// It handles X-Forwarded-For and X-Real-IP headers for proxy support,
// and falls back to RemoteAddr.
func ExtractIP(r *http.Request) string {
	// Check X-Forwarded-For header (can contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If RemoteAddr doesn't have a port (e.g. in some tests or local dev)
		return r.RemoteAddr
	}
	return host
}
