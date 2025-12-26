package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/services"
)

type ActionHandlers struct {
	DB                *sql.DB
	ToolNameGenerator *rag.ToolNameGenerator
	WorkspaceService  *services.WorkspaceService
	OrgService        *services.OrganizationService
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
	bot, botID, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
	return botID, bot, ok
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

	api.WriteJSON(w, http.StatusOK, map[string]any{"actions": actions})
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

	api.WriteJSON(w, http.StatusCreated, action)
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

	api.WriteJSON(w, http.StatusOK, action)
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

		var toolName string
		toolName, err = h.ToolNameGenerator.Generate(r.Context(), newName, newDesc)
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
		if errors.Is(err, db.ErrVersionConflict) {
			http.Error(w, "Action was modified by another request, please refresh and try again", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, action)
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

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"logs":  logs,
		"page":  page,
		"limit": limit,
	})
}
