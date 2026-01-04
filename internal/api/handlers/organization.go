package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/middleware"
)

type OrganizationHandlers struct {
	OrgService *services.OrganizationService
	UserRepo   repository.UserRepository
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type updateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
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
		api.WriteJSON(w, http.StatusOK, orgs)

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
			if errors.Is(err, services.ErrOrgSlugExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		api.WriteJSON(w, http.StatusCreated, org)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *OrganizationHandlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	org, err := h.OrgService.GetOrganization(r.Context(), orgID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if org == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	api.WriteJSON(w, http.StatusOK, org)
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
		if errors.Is(err, services.ErrOrgSlugExists) {
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
		if errors.Is(err, services.ErrLastOrganization) {
			api.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
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

	api.WriteJSON(w, http.StatusOK, membersResponse{
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
	user, err := h.UserRepo.GetByEmail(r.Context(), req.Email)
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
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, services.ErrLastOwner):
			status = http.StatusForbidden
		case errors.Is(err, services.ErrNotMember):
			status = http.StatusNotFound
		}
		api.WriteJSON(w, status, map[string]string{"error": err.Error()})
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
		status := http.StatusInternalServerError
		// Return appropriate status codes based on error type
		switch {
		case errors.Is(err, services.ErrInvalidRole),
			errors.Is(err, services.ErrCannotPromoteSelf),
			errors.Is(err, services.ErrCannotDemoteOwner),
			errors.Is(err, services.ErrOnlyOwnersCanGrant):
			status = http.StatusForbidden
		case errors.Is(err, services.ErrNotMember):
			status = http.StatusNotFound
		}
		api.WriteJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}
