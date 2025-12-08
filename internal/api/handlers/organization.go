package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type OrganizationHandlers struct {
	OrgService *services.OrganizationService
	DB         *sql.DB // Added DB to look up users
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type updateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type createWorkspaceRequest struct {
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	ClientName *string `json:"client_name"`
}

type updateWorkspaceRequest struct {
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	ClientName *string `json:"client_name"`
}

type addMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type updateMemberRoleRequest struct {
	Role string `json:"role"`
}

func (h *OrganizationHandlers) ListOrCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		orgs, err := h.OrgService.GetUserOrganizations(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orgs)

	case http.MethodPost:
		var req createOrgRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Slug == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		org, err := h.OrgService.CreateOrganization(r.Context(), req.Name, req.Slug, userID)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(org)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *OrganizationHandlers) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req updateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.OrgService.UpdateOrganization(r.Context(), orgID, req.Name, req.Slug); err != nil {
		if strings.Contains(err.Error(), "exists") {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *OrganizationHandlers) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.OrgService.DeleteOrganization(r.Context(), orgID); err != nil {
		if strings.Contains(err.Error(), "cannot delete the last") {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *OrganizationHandlers) Workspaces(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		workspaces, err := h.OrgService.GetWorkspaces(r.Context(), orgID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(workspaces)

	case http.MethodPost:
		var req createWorkspaceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Slug == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ws, err := h.OrgService.CreateWorkspace(r.Context(), orgID, req.Name, req.Slug, req.ClientName)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ws)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *OrganizationHandlers) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID := r.PathValue("wsID")
	if wsID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.OrgService.UpdateWorkspace(r.Context(), wsID, req.Name, req.Slug, req.ClientName); err != nil {
		if strings.Contains(err.Error(), "exists") {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *OrganizationHandlers) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID := r.PathValue("wsID")
	if wsID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.OrgService.DeleteWorkspace(r.Context(), wsID); err != nil {
		if strings.Contains(err.Error(), "cannot delete the last") {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// membersResponse wraps member list with caller's role for frontend RBAC
type membersResponse struct {
	Members    any    `json:"members"`
	CallerRole string `json:"caller_role"`
}

func (h *OrganizationHandlers) GetMembers(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || callerID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	members, err := h.OrgService.GetMembers(r.Context(), orgID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Find caller's role from members list
	var callerRole string
	for _, m := range members {
		if m.UserID == callerID {
			callerRole = m.Role
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(membersResponse{
		Members:    members,
		CallerRole: callerRole,
	})
}

func (h *OrganizationHandlers) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Look up user by email
	user, err := db.GetUserByEmail(r.Context(), h.DB, req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		// Or create a pending invitation (not in scope yet, so just 404)
		return
	}

	if err := h.OrgService.AddMember(r.Context(), orgID, user.ID, req.Role); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *OrganizationHandlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || callerID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	targetUserID := r.PathValue("userID")

	if err := h.OrgService.RemoveMember(r.Context(), orgID, callerID, targetUserID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "cannot remove the last owner") {
			w.WriteHeader(http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not a member") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *OrganizationHandlers) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || callerID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	targetUserID := r.PathValue("userID")

	var req updateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.OrgService.UpdateMemberRole(r.Context(), orgID, callerID, targetUserID, req.Role); err != nil {
		w.Header().Set("Content-Type", "application/json")
		// Return appropriate status codes based on error type
		if strings.Contains(err.Error(), "invalid role") ||
			strings.Contains(err.Error(), "cannot promote") ||
			strings.Contains(err.Error(), "cannot demote") ||
			strings.Contains(err.Error(), "only owners can") {
			w.WriteHeader(http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not a member") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}
