// Package requestid provides a shared context key for request IDs
// to avoid circular dependencies between middleware and logger packages.
package requestid

import "context"

// contextKey is the private type for the request ID context key
type contextKey struct{}

// key is the singleton instance of the context key
var key = contextKey{}

// ToContext adds a request ID to the context
func ToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, key, requestID)
}

// FromContext extracts the request ID from the context
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(key).(string); ok {
		return id
	}
	return ""
}
