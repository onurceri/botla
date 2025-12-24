package router

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/api/middleware"
	pkgMiddleware "github.com/onurceri/botla-co/pkg/middleware"
)

func registerAdminRoutes(mux *http.ServeMux, ah *handlers.AdminHandlers, adhh *handlers.AdminHealthHandlers, aqh *handlers.AdminQueueHandlers, aeh *handlers.AdminErrorHandlers, aah *handlers.AdminAuditHandlers, secret string) {
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
	mux.Handle("GET /api/v1/admin/health/detailed", adminChain(adhh.GetDetailedHealth))

	// Queues
	mux.Handle("GET /api/v1/admin/queues", adminChain(aqh.GetQueues))
	mux.Handle("GET /api/v1/admin/queues/stuck", adminChain(aqh.GetStuckJobs))
	mux.Handle("POST /api/v1/admin/queues/{id}/retry", adminChain(aqh.RetryJob))
	mux.Handle("DELETE /api/v1/admin/queues/{id}", adminChain(aqh.DeleteJob))

	// Errors
	mux.Handle("GET /api/v1/admin/errors", adminChain(aeh.ListErrors))
	mux.Handle("GET /api/v1/admin/errors/stats", adminChain(aeh.GetErrorStats))
	mux.Handle("GET /api/v1/admin/errors/{id}", adminChain(aeh.GetError))

	// Audit Logs
	mux.Handle("GET /api/v1/admin/audit-logs", adminChain(aah.ListAuditLogs))
}

