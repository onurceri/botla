package router

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/api/middleware"
	pkgMiddleware "github.com/onurceri/botla-co/pkg/middleware"
)

func registerAdminRoutes(mux *http.ServeMux, ah *handlers.AdminHandlers, secret string) {
	// Base admin handler with Auth and Admin middleware
	adminChain := func(h http.HandlerFunc) http.Handler {
		return pkgMiddleware.AuthMiddleware(secret)(
			middleware.RequirePlatformAdmin(h),
		)
	}

	// Stats & Overview
	mux.Handle("GET /api/v1/admin/stats/overview", adminChain(ah.GetOverviewStats))

	// Users
	mux.Handle("GET /api/v1/admin/users", adminChain(ah.ListUsers))
	mux.Handle("GET /api/v1/admin/users/{id}", adminChain(ah.GetUser))
	mux.Handle("PATCH /api/v1/admin/users/{id}", adminChain(ah.UpdateUser))

	// Organizations
	mux.Handle("GET /api/v1/admin/organizations", adminChain(ah.ListOrganizations))
	mux.Handle("GET /api/v1/admin/organizations/{id}", adminChain(ah.GetOrganization))

	// System Health
	mux.Handle("GET /api/v1/admin/health/detailed", adminChain(ah.GetDetailedHealth))
}
