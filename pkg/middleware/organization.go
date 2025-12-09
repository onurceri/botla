package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/services"
)

const (
	ContextKeyOrgID      contextKey = "organization_id"
	ContextKeyMembership contextKey = "membership"
)

// RequireOrganizationAccess checks if user has access to organization
func RequireOrganizationAccess(orgService *services.OrganizationService, minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok || userID == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Extract OrgID from path
			// Assuming path is like /api/organizations/:orgId/...
			orgID := extractOrgIDFromPath(r.URL.Path)
			if orgID == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			membership, err := orgService.CheckMembership(r.Context(), userID, orgID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if membership == nil {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error": "user is not a member of this organization"}`))
				return
			}

			if !hasMinRole(membership.Role, minRole) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = fmt.Fprintf(w, `{"error": "insufficient role: have %s, need %s"}`, membership.Role, minRole)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyOrgID, orgID)
			ctx = context.WithValue(ctx, ContextKeyMembership, membership)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractOrgIDFromPath(path string) string {
	// Pattern: /api/v1/organizations/:orgId/...
	parts := strings.Split(path, "/")
	// ["", "api", "v1", "organizations", "orgId", ...]
	for i, part := range parts {
		if part == "organizations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func hasMinRole(userRole, minRole string) bool {
	roles := map[string]int{
		"member": 1,
		"admin":  2,
		"owner":  3,
	}
	return roles[userRole] >= roles[minRole]
}

func OrgIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyOrgID)
	s, ok := v.(string)
	return s, ok
}

// Header names for tenant context
const (
	HeaderOrgID       = "X-Organization-ID"
	HeaderWorkspaceID = "X-Workspace-ID"
)

// Context key for workspace ID
const ContextKeyWorkspaceID contextKey = "workspace_id"

// ExtractTenantContext reads X-Organization-ID and X-Workspace-ID headers
// and adds them to request context. Unlike RequireOrganizationAccess,
// this doesn't enforce membership - it just makes IDs available.
// Use this for endpoints that should work with or without tenant context.
func ExtractTenantContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract organization ID from header
			if orgID := r.Header.Get(HeaderOrgID); orgID != "" {
				ctx = context.WithValue(ctx, ContextKeyOrgID, orgID)
			}

			// Extract workspace ID from header
			if wsID := r.Header.Get(HeaderWorkspaceID); wsID != "" {
				ctx = context.WithValue(ctx, ContextKeyWorkspaceID, wsID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WorkspaceIDFromContext extracts workspace ID from context
func WorkspaceIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ContextKeyWorkspaceID)
	s, ok := v.(string)
	return s, ok
}
