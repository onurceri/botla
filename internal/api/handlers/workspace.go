package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type WorkspaceHandlers struct {
	WorkspaceService *services.WorkspaceService
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

func (h *WorkspaceHandlers) Workspaces(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.OrgIDFromContext(r.Context())
	if !ok || orgID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		workspaces, err := h.WorkspaceService.GetWorkspaces(r.Context(), orgID)
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

		ws, err := h.WorkspaceService.CreateWorkspace(r.Context(), orgID, req.Name, req.Slug, req.ClientName)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(ws)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *WorkspaceHandlers) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
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

	if err := h.WorkspaceService.UpdateWorkspace(r.Context(), wsID, req.Name, req.Slug, req.ClientName); err != nil {
		if strings.Contains(err.Error(), "exists") {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WorkspaceHandlers) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID := r.PathValue("wsID")
	if wsID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.WorkspaceService.DeleteWorkspace(r.Context(), wsID); err != nil {
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
