package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
)

type ActionHandlers struct {
	ActionService    services.ActionService
	ChatbotRepo      repository.ChatbotRepository
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
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
	bot, botID, ok := getChatbotContextWithRepo(w, r, h.ChatbotRepo, h.WorkspaceService, h.OrgService)
	return botID, bot, ok
}

func (h *ActionHandlers) List(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	actions, err := h.ActionService.ListActions(r.Context(), botID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{"actions": actions})
}

func (h *ActionHandlers) Create(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	var req createActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := services.CreateActionInput{
		Name:        req.Name,
		Description: req.Description,
		ActionType:  req.ActionType,
		Config:      req.Config,
		Parameters:  req.Parameters,
		Enabled:     req.Enabled,
	}

	action, err := h.ActionService.CreateAction(r.Context(), botID, input)
	if err != nil {
		if errors.Is(err, services.ErrActionNameRequired) || errors.Is(err, services.ErrActionTypeRequired) {
			api.WriteErrorCode(w, http.StatusBadRequest, api.ErrNameAndActionTypeRequired)
			return
		}
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
	action, err := h.ActionService.GetAction(r.Context(), actionID)
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

	existingAction, err := h.ActionService.GetAction(r.Context(), actionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingAction == nil || existingAction.ChatbotID != botID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var req createActionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := services.UpdateActionInput{}
	if req.Name != "" {
		input.Name = &req.Name
	}
	if req.Description != "" {
		input.Description = &req.Description
	}
	if req.ActionType != "" {
		input.ActionType = &req.ActionType
	}
	if len(req.Config) > 0 {
		input.Config = req.Config
	}
	if len(req.Parameters) > 0 {
		input.Parameters = req.Parameters
	}
	input.Enabled = &req.Enabled

	updatedAction, err := h.ActionService.UpdateAction(r.Context(), actionID, input)
	if err != nil {
		if errors.Is(err, services.ErrActionNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, services.ErrVersionConflict) {
			http.Error(w, "Action was modified by another request, please refresh and try again", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, updatedAction)
}

func (h *ActionHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	botID, _, ok := h.authorize(w, r)
	if !ok {
		return
	}

	actionID := r.PathValue("actionId")
	action, err := h.ActionService.GetAction(r.Context(), actionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if action == nil || action.ChatbotID != botID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err = h.ActionService.DeleteAction(r.Context(), actionID); err != nil {
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
	logs, err := h.ActionService.GetActionLogs(r.Context(), botID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"logs":  logs,
		"page":  page,
		"limit": limit,
	})
}
