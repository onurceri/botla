package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type ActionHandlers struct {
	DB                *sql.DB
	ToolNameGenerator *rag.ToolNameGenerator
}

type createActionRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ActionType  string          `json:"action_type"`
	Config      json.RawMessage `json:"config"`
	Parameters  json.RawMessage `json:"parameters"`
	Enabled     bool            `json:"enabled"`
}

func (h *ActionHandlers) authorize(w http.ResponseWriter, r *http.Request) (string, *models.Chatbot, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return "", nil, false
	}
	botID := r.PathValue("id")
	if botID == "" {
		w.WriteHeader(http.StatusNotFound)
		return "", nil, false
	}

	bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", nil, false
	}
	if bot == nil {
		w.WriteHeader(http.StatusNotFound)
		return "", nil, false
	}
	if bot.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return "", nil, false
	}
	return botID, bot, true
}

func (h *ActionHandlers) List(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

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

func (h *ActionHandlers) Create(w http.ResponseWriter, r *http.Request) {
	botID, bot, ok := h.authorize(w, r)
	if !ok {
		return
	}

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

	// Generate tool_name using LLM
	toolName, err := h.ToolNameGenerator.Generate(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate tool name: %v", err), http.StatusInternalServerError)
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
		ToolName:    &toolName,
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

func (h *ActionHandlers) Get(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	actionID := r.PathValue("actionId")
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

func (h *ActionHandlers) Update(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	actionID := r.PathValue("actionId")
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

	// Check if name or description changed - OR if tool_name is currently empty (migration case)
	nameChanged := req.Name != "" && req.Name != action.Name
	descChanged := req.Description != "" && (action.Description == nil || req.Description != *action.Description)
	toolNameMissing := action.ToolName == nil || *action.ToolName == ""

	if nameChanged || descChanged || toolNameMissing {
		newName := action.Name
		if req.Name != "" {
			newName = req.Name
		}
		newDesc := ""
		if req.Description != "" {
			newDesc = req.Description
		} else if action.Description != nil {
			newDesc = *action.Description
		}

		toolName, err := h.ToolNameGenerator.Generate(r.Context(), newName, newDesc)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate tool name: %v", err), http.StatusInternalServerError)
			return
		}
		action.ToolName = &toolName
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

func (h *ActionHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	actionID := r.PathValue("actionId")
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

func (h *ActionHandlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit
	logs, err := db.GetActionLogs(r.Context(), h.DB, botID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	if logs == nil {
		logs = []*models.ActionExecutionLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"logs":  logs,
		"page":  page,
		"limit": limit,
	})
}
