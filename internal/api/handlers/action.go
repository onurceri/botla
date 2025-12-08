package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type ActionHandlers struct {
	DB *sql.DB
}

type createActionRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ActionType  string          `json:"action_type"`
	Config      json.RawMessage `json:"config"`
	Parameters  json.RawMessage `json:"parameters"`
	Enabled     bool            `json:"enabled"`
}

// Dispatch handles requests to /api/v1/chatbots/:id/actions...
func (h *ActionHandlers) Dispatch(w http.ResponseWriter, r *http.Request) {
	// Path is like /api/v1/chatbots/:id/actions...
	// We need to extract botID
	path := r.URL.Path
	parts := strings.Split(path, "/")
	// ["", "api", "v1", "chatbots", "{id}", "actions", ...]
	if len(parts) < 6 || parts[5] != "actions" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	botID := parts[4]

	// Check permissions
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if bot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Route based on suffix
	// /api/v1/chatbots/{id}/actions
	if len(parts) == 6 {
		if r.Method == http.MethodGet {
			h.List(w, r, botID)
			return
		}
		if r.Method == http.MethodPost {
			h.Create(w, r, botID)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// /api/v1/chatbots/{id}/actions/{actionId}
	actionID := parts[6]
	if len(parts) == 7 {
		switch r.Method {
		case http.MethodGet:
			h.Get(w, r, botID, actionID)
		case http.MethodPut:
			h.Update(w, r, botID, actionID)
		case http.MethodDelete:
			h.Delete(w, r, botID, actionID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// /api/v1/chatbots/{id}/actions/{actionId}/test
	if len(parts) == 8 && parts[7] == "test" && r.Method == http.MethodPost {
		h.Test(w, r, botID, actionID)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *ActionHandlers) List(w http.ResponseWriter, r *http.Request, botID string) {
	actions, err := db.GetActions(r.Context(), h.DB, botID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if actions == nil {
		actions = []*models.ChatbotAction{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"actions": actions})
}

func (h *ActionHandlers) Create(w http.ResponseWriter, r *http.Request, botID string) {
	bot, _ := db.GetChatbotByID(r.Context(), h.DB, botID)
	base := "tr"
	if bot != nil {
		base = api.BaseLang(bot.LanguageCode)
	}
	cfg := api.ConfigFromBase(base)
	var req createActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.ActionType == "" {
		api.WriteLocalizedError(w, http.StatusBadRequest, api.ErrNameAndActionTypeRequired, cfg)
		return
	}

	var config *json.RawMessage
	if len(req.Config) > 0 {
		config = &req.Config
	}
	var params *json.RawMessage
	if len(req.Parameters) > 0 {
		params = &req.Parameters
	}
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	action := &models.ChatbotAction{
		ChatbotID:   botID,
		Name:        req.Name,
		Description: desc,
		ActionType:  models.ActionType(req.ActionType),
		Config:      config,
		Parameters:  params,
		Enabled:     req.Enabled,
	}

	if err := db.CreateAction(r.Context(), h.DB, action); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(action)
}

func (h *ActionHandlers) Get(w http.ResponseWriter, r *http.Request, botID, actionID string) {
	action, err := db.GetActionByID(r.Context(), h.DB, actionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if action == nil || action.ChatbotID != botID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(action)
}

func (h *ActionHandlers) Update(w http.ResponseWriter, r *http.Request, botID, actionID string) {
	action, err := db.GetActionByID(r.Context(), h.DB, actionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if action == nil || action.ChatbotID != botID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var req createActionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		action.Name = req.Name
	}
	if req.Description != "" {
		action.Description = &req.Description
	}
	if req.ActionType != "" {
		action.ActionType = models.ActionType(req.ActionType)
	}
	if len(req.Config) > 0 {
		action.Config = &req.Config
	}
	if len(req.Parameters) > 0 {
		action.Parameters = &req.Parameters
	}
	action.Enabled = req.Enabled

	if err = db.UpdateAction(r.Context(), h.DB, action); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(action)
}

func (h *ActionHandlers) Delete(w http.ResponseWriter, r *http.Request, botID, actionID string) {
	action, err := db.GetActionByID(r.Context(), h.DB, actionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if action == nil || action.ChatbotID != botID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err = db.DeleteAction(r.Context(), h.DB, actionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ActionHandlers) Test(w http.ResponseWriter, r *http.Request, botID, actionID string) {
	// TODO: Implement test logic (execute action with test params)
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = fmt.Fprintf(w, "Test not implemented yet")
}
