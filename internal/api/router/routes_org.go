package router

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func registerOrgRoutes(mux *http.ServeMux, secret string, userRepo *repository.PostgresUserRepo, log *logger.Logger, orgSvc *services.OrganizationService, oh *handlers.OrganizationHandlers, wh *handlers.WorkspaceHandlers) {
	auth := middleware.AuthMiddleware(secret)
	deletedCheck := middleware.DeletedAccountMiddleware(userRepo, log)
	requireMember := middleware.RequireOrganizationAccess(orgSvc, "member")
	requireAdmin := middleware.RequireOrganizationAccess(orgSvc, "admin")
	requireOwner := middleware.RequireOrganizationAccess(orgSvc, "owner")

	// Org List/Create
	mux.Handle("GET /api/v1/organizations", auth(deletedCheck(http.HandlerFunc(oh.ListOrCreate))))
	mux.Handle("POST /api/v1/organizations", auth(deletedCheck(http.HandlerFunc(oh.ListOrCreate))))

	// Org Management
	mux.Handle("PATCH /api/v1/organizations/{id}", auth(deletedCheck(requireOwner(http.HandlerFunc(oh.UpdateOrganization)))))
	mux.Handle("DELETE /api/v1/organizations/{id}", auth(deletedCheck(requireOwner(http.HandlerFunc(oh.DeleteOrganization)))))

	// Workspaces
	mux.Handle("GET /api/v1/organizations/{id}/workspaces", auth(deletedCheck(requireMember(http.HandlerFunc(wh.Workspaces)))))
	mux.Handle("POST /api/v1/organizations/{id}/workspaces", auth(deletedCheck(requireAdmin(http.HandlerFunc(wh.Workspaces)))))
	mux.Handle("PATCH /api/v1/organizations/{id}/workspaces/{wsID}", auth(deletedCheck(requireAdmin(http.HandlerFunc(wh.UpdateWorkspace)))))
	mux.Handle("DELETE /api/v1/organizations/{id}/workspaces/{wsID}", auth(deletedCheck(requireAdmin(http.HandlerFunc(wh.DeleteWorkspace)))))

	// Members
	mux.Handle("GET /api/v1/organizations/{id}/members", auth(deletedCheck(requireMember(http.HandlerFunc(oh.GetMembers)))))
	mux.Handle("POST /api/v1/organizations/{id}/members", auth(deletedCheck(requireAdmin(http.HandlerFunc(oh.AddMember)))))
	mux.Handle("DELETE /api/v1/organizations/{id}/members/{userID}", auth(deletedCheck(requireAdmin(http.HandlerFunc(oh.RemoveMember)))))
	mux.Handle("PATCH /api/v1/organizations/{id}/members/{userID}", auth(deletedCheck(requireAdmin(http.HandlerFunc(oh.UpdateMemberRole)))))
}
