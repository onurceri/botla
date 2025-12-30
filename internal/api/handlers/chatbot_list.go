package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// ListOrCreate handles GET (list) and POST (create) for chatbots
func (h *ChatbotHandlers) ListOrCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listChatbots(w, r, userID)
	case http.MethodPost:
		h.createChatbot(w, r, userID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// listChatbots returns all chatbots for a user or workspace
func (h *ChatbotHandlers) listChatbots(w http.ResponseWriter, r *http.Request, userID string) {
	var bots []models.Chatbot
	var err error

	// Check for workspace context from header
	if wsID, ok := middleware.WorkspaceIDFromContext(r.Context()); ok && wsID != "" {
		// Verify workspace membership using Helper
		var wsInfo *models.Workspace
		wsInfo, err = h.WorkspaceService.CheckAccess(r.Context(), userID, wsID)
		if err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, "workspace_check_error")
			return
		}
		if wsInfo == nil {
			api.WriteErrorCode(w, http.StatusForbidden, "not_workspace_member")
			return
		}

		bots, err = db.GetChatbotsByWorkspace(r.Context(), h.DB, wsID)
	} else {
		// Fallback to user-based query
		bots, err = db.GetChatbotsByUserID(r.Context(), h.DB, userID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.WriteJSON(w, http.StatusOK, bots)
}

// createChatbot creates a new chatbot
func (h *ChatbotHandlers) createChatbot(w http.ResponseWriter, r *http.Request, userID string) {
	var req createChatbotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	langCodeBCP := normalizeLocale(defaultString(req.Language, "tr-TR"))
	baseLang := baseLangCode(langCodeBCP)
	langCfg := langconfig.Get(baseLang)

	// Check plan limits
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, "get_plan_error: "+err.Error())
		return
	}

	if plan != nil && plan.Config.MaxChatbots > 0 {
		count, countErr := db.CountChatbotsByUserID(r.Context(), h.DB, userID)
		if countErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if count >= plan.Config.MaxChatbots {
			api.WriteErrorCode(w, http.StatusForbidden, api.ErrMaxChatbotsExceeded)
			return
		}
	}

	// Get workspace/org context from headers with security checks
	var wsID, orgID *string
	if ws, ok := middleware.WorkspaceIDFromContext(r.Context()); ok && ws != "" {
		// Verify workspace membership using Helper
		var wsInfo *models.Workspace
		wsInfo, err = h.WorkspaceService.CheckAccess(r.Context(), userID, ws)
		if err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, "workspace_check_error: "+err.Error())
			return
		}
		if wsInfo == nil {
			api.WriteErrorCode(w, http.StatusForbidden, "not_workspace_member")
			return
		}

		wsID = &ws
		// Ensure orgID matches the workspace's orgID
		orgIDStr := wsInfo.OrganizationID
		orgID = &orgIDStr
	} else if org, ok := middleware.OrgIDFromContext(r.Context()); ok && org != "" {
		// Verify org membership
		var memInfo *models.Membership
		memInfo, err = h.OrgService.CheckMembership(r.Context(), userID, org)
		if err != nil {
			api.WriteErrorCode(w, http.StatusInternalServerError, "membership_check_error: "+err.Error())
			return
		}
		if memInfo == nil {
			api.WriteErrorCode(w, http.StatusForbidden, "not_org_member")
			return
		}
		orgID = &org
	}

	bot := h.buildNewChatbot(userID, wsID, orgID, req, langCodeBCP, langCfg)

	newID, err := db.CreateChatbot(r.Context(), h.DB, bot)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, "create_bot_error: "+err.Error())
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, newID)
	if err != nil || c == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusCreated, c)
}

// buildNewChatbot constructs a new chatbot with defaults
func (h *ChatbotHandlers) buildNewChatbot(userID string, wsID, orgID *string, req createChatbotRequest, langCode string, langCfg langconfig.LanguageConfig) *models.Chatbot {
	return &models.Chatbot{
		UserID:               userID,
		WorkspaceID:          wsID,
		OrganizationID:       orgID,
		Name:                 req.Name,
		Description:          req.Description,
		CustomInstruction:    defaultString(req.CustomInstruction, ""),
		LanguageCode:         langCode,
		Model:                defaultString(req.Model, config.DefaultChatbotModel()),
		Temperature:          defaultFloat32(req.Temperature, 0.7),
		MaxTokens:            defaultInt(req.MaxTokens, 4096),
		ThemeColor:           defaultString(req.ThemeColor, "rgba(255, 174, 0, 1)"),
		WelcomeMessage:       defaultString(req.WelcomeMessage, langCfg.UserMessages.WelcomeMessage),
		Position:             defaultString(req.Position, "bottom-right"),
		BotMessageColor:      defaultString(req.BotMessageColor, "rgba(252, 252, 253, 1)"),
		UserMessageColor:     defaultString(req.UserMessageColor, "rgba(250, 171, 0, 0.91)"),
		BotMessageTextColor:  defaultString(req.BotMessageTextColor, "rgba(0, 0, 0, 1)"),
		UserMessageTextColor: defaultString(req.UserMessageTextColor, "rgba(255, 255, 255, 1)"),
		ChatFontFamily:       defaultString(req.ChatFontFamily, "Inter, sans-serif"),
		ChatHeaderColor:      defaultString(req.ChatHeaderColor, "rgba(242, 167, 36, 1)"),
		ChatHeaderTextColor:  defaultString(req.ChatHeaderTextColor, "rgba(247, 241, 241, 1)"),
		ChatBackgroundColor:  defaultString(req.ChatBackgroundColor, "rgba(255, 245, 230, 1)"),
		BubbleRadius:         defaultString(req.BubbleRadius, "22px"),
		InputBackgroundColor: defaultString(req.InputBackgroundColor, "rgba(255, 255, 255, 0.5)"),
		InputTextColor:       defaultString(req.InputTextColor, "rgba(28, 28, 30, 1)"),
		SendButtonColor:      defaultString(req.SendButtonColor, "rgba(246, 140, 0, 1)"),
		BotIcon:              req.BotIcon,
		BotDisplayName:       req.BotDisplayName,
		SecureEmbedEnabled:   boolValue(req.SecureEmbedEnabled, false),
		AllowedDomains:       req.AllowedDomains,
		EmbedSecret:          req.EmbedSecret,
		SuggestedQuestions:   suggestionsValue(req.SuggestedQuestions),
		SuggestionsEnabled:   boolValue(req.SuggestionsEnabled, false),
		IncludePaths:         pathsValue(req.IncludePaths),
		ExcludePaths:         pathsValue(req.ExcludePaths),
		SelectorWhitelist:    selectorsValue(req.SelectorWhitelist),
		DiscoveryMode:        defaultString(req.DiscoveryMode, "auto"),
		RefreshPolicy:        defaultString(req.RefreshPolicy, "manual"),
		RefreshFrequency:     req.RefreshFrequency,
		HideBranding:         boolValue(req.HideBranding, false),
		CustomBranding:       req.CustomBranding,
		ConfidenceThreshold:  defaultFloat64(req.ConfidenceThreshold, 0.7),
		ThresholdConfig:      req.ThresholdConfig,
		FallbackMessages:     req.FallbackMessages,
		TopicRestrictions:    req.TopicRestrictions,
		HandoffEnabled:       boolValue(req.HandoffEnabled, false),
		HandoffType:          defaultString(req.HandoffType, ""),
		HandoffConfig:        req.HandoffConfig,
	}
}
