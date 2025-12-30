package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type AdminHandlers struct {
	DB           *sql.DB
	AdminService *services.AdminService
}

func NewAdminHandlers(db *sql.DB, adminSvc *services.AdminService) *AdminHandlers {
	return &AdminHandlers{
		DB:           db,
		AdminService: adminSvc,
	}
}

// GetOverviewStats returns high-level platform metrics.
func (h *AdminHandlers) GetOverviewStats(w http.ResponseWriter, r *http.Request) {
	stats, err := db.GetPlatformOverviewStats(r.Context(), h.DB)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, stats)
}

// ListUsers returns a paginated list of all users.
func (h *AdminHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filter := db.UserFilter{}
	if email := r.URL.Query().Get("email"); email != "" {
		filter.Email = &email
	}
	if isAdmin := r.URL.Query().Get("is_platform_admin"); isAdmin != "" {
		b, _ := strconv.ParseBool(isAdmin)
		filter.IsPlatformAdmin = &b
	}
	if planID := r.URL.Query().Get("plan_id"); planID != "" {
		filter.PlanID = &planID
	}

	users, total, err := db.AdminListUsers(r.Context(), h.DB, filter, limit, offset)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"users": users,
		"total": total,
	})
}

// GetUser returns details for a single user.
func (h *AdminHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	user, err := db.GetUserByID(r.Context(), h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, user)
}

// UpdateUser updates a user's details or admin status.
func (h *AdminHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	err := db.AdminUpdateUser(r.Context(), h.DB, id, updates)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	// Log the action
	adminID, _ := middleware.UserIDFromContext(r.Context())
	_ = h.AdminService.LogAction(r.Context(), adminID, "update_user", "user", &id, updates, r)

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ListOrganizations returns a paginated list of all organizations.
func (h *AdminHandlers) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filter := db.OrganizationFilter{}
	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}
	if planID := r.URL.Query().Get("plan_id"); planID != "" {
		filter.PlanID = &planID
	}

	orgs, total, err := db.AdminListOrganizations(r.Context(), h.DB, filter, limit, offset)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"organizations": orgs,
		"total":         total,
	})
}

// GetOrganization returns details for a single organization.
func (h *AdminHandlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrCodeBadRequest)
		return
	}

	org, err := db.GetOrganizationByID(r.Context(), h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteErrorCode(w, http.StatusNotFound, api.ErrCodeNotFound)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrCodeInternalError)
		return
	}

	api.WriteJSON(w, http.StatusOK, org)
}
