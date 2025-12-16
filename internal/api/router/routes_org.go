package router

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func registerOrgRoutes(mux *http.ServeMux, secret string, orgSvc *services.OrganizationService, oh *handlers.OrganizationHandlers, wh *handlers.WorkspaceHandlers) {
	auth := middleware.AuthMiddleware(secret)
	requireMember := middleware.RequireOrganizationAccess(orgSvc, "member")
	requireAdmin := middleware.RequireOrganizationAccess(orgSvc, "admin")
	requireOwner := middleware.RequireOrganizationAccess(orgSvc, "owner")

	// Org List/Create
	mux.Handle("GET /api/v1/organizations", auth(http.HandlerFunc(oh.ListOrCreate)))
	mux.Handle("POST /api/v1/organizations", auth(http.HandlerFunc(oh.ListOrCreate)))

	// Org Management
	mux.Handle("PATCH /api/v1/organizations/{id}", auth(requireOwner(http.HandlerFunc(oh.UpdateOrganization))))
	mux.Handle("DELETE /api/v1/organizations/{id}", auth(requireOwner(http.HandlerFunc(oh.DeleteOrganization))))

	// Workspaces
	mux.Handle("GET /api/v1/organizations/{id}/workspaces", auth(requireMember(http.HandlerFunc(wh.Workspaces))))
	mux.Handle("POST /api/v1/organizations/{id}/workspaces", auth(requireAdmin(http.HandlerFunc(wh.Workspaces))))
	mux.Handle("PATCH /api/v1/organizations/{id}/workspaces/{wsID}", auth(requireAdmin(http.HandlerFunc(wh.UpdateWorkspace))))
	mux.Handle("DELETE /api/v1/organizations/{id}/workspaces/{wsID}", auth(requireAdmin(http.HandlerFunc(wh.DeleteWorkspace))))

	// Members
	mux.Handle("GET /api/v1/organizations/{id}/members", auth(requireMember(http.HandlerFunc(oh.GetMembers))))
	mux.Handle("POST /api/v1/organizations/{id}/members", auth(requireAdmin(http.HandlerFunc(oh.AddMember))))
	mux.Handle("DELETE /api/v1/organizations/{id}/members/{userID}", auth(requireAdmin(http.HandlerFunc(oh.RemoveMember))))
	mux.Handle("PATCH /api/v1/organizations/{id}/members/{userID}", auth(requireAdmin(http.HandlerFunc(oh.UpdateMemberRole))))
}
