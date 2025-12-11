package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/logger"
)

// HandoffService handles human agent handoff logic
type HandoffService struct {
	DB  *sql.DB
	Log *logger.Logger
}

// NewHandoffService creates a new HandoffService instance
func NewHandoffService(dbPool *sql.DB, log *logger.Logger) *HandoffService {
	return &HandoffService{
		DB:  dbPool,
		Log: log,
	}
}

// HandoffResult contains the result of a handoff request
type HandoffResult struct {
	RequestID    string `json:"request_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	EmailSent    bool   `json:"email_sent,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// RequestHandoff creates a handoff request and notifies operators
func (s *HandoffService) RequestHandoff(ctx context.Context, bot *models.Chatbot, conversationID, notes string) (*HandoffResult, error) {
	lc := strings.TrimSpace(bot.LanguageCode)
	if i := strings.Index(lc, "-"); i > 0 {
		lc = lc[:i]
	}
	cfg := langconfig.Get(lc)
	// Validate handoff is enabled
	if !bot.HandoffEnabled {
		return nil, errors.New(cfg.ResponseTemplates.Errors["HANDOFF_NOT_ENABLED"])
	}

	// Create handoff request in database
	// Check for existing active handoff request
	exists, err := db.HasActiveHandoffRequest(ctx, s.DB, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing handoff: %w", err)
	}
	if exists {
		// Return specific error that can be handled by caller
		return nil, errors.New(cfg.ResponseTemplates.Errors["HANDOFF_ALREADY_EXISTS"])
	}

	req := &models.HandoffRequest{
		ChatbotID:      bot.ID,
		ConversationID: conversationID,
		Notes:          &notes,
	}

	requestID, err := db.CreateHandoffRequest(ctx, s.DB, req)
	if err != nil {
		msg := cfg.ResponseTemplates.Errors["HANDOFF_CREATE_FAILED"]
		if msg == "" {
			msg = "failed to create handoff request"
		}
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	result := &HandoffResult{
		RequestID: requestID,
		Status:    models.HandoffStatusPending,
	}

	// Handle based on handoff type
	switch bot.HandoffType {
	case string(models.HandoffTypeEmail):
		err = s.handleEmailHandoff(ctx, bot, requestID, conversationID, notes)
		if err != nil {
			result.ErrorMessage = err.Error()
			if s.Log != nil {
				s.Log.Warn("handoff_email_failed", map[string]any{
					"request_id": requestID,
					"error":      err.Error(),
				})
			}
		} else {
			result.EmailSent = true
		}
	default:
		// Log unsupported type but don't fail - request is still created
		if s.Log != nil {
			s.Log.Warn("unsupported_handoff_type", map[string]any{
				"type": bot.HandoffType,
			})
		}
	}

	// Set user-friendly message
	if bot.FallbackMessages != nil && bot.FallbackMessages.HandoffMessage != "" {
		result.Message = bot.FallbackMessages.HandoffMessage
	} else {
		hm := cfg.ResponseTemplates.Errors["HANDOFF_RECEIVED"]
		if hm == "" {
			hm = "Talebiniz alındı. En kısa sürede bir temsilcimiz sizinle iletişime geçecektir."
		}
		result.Message = hm
	}

	return result, nil
}

// handleEmailHandoff sends an email notification for handoff
func (s *HandoffService) handleEmailHandoff(ctx context.Context, bot *models.Chatbot, requestID, conversationID, notes string) error {
	lc := strings.TrimSpace(bot.LanguageCode)
	if i := strings.Index(lc, "-"); i > 0 {
		lc = lc[:i]
	}
	cfg := langconfig.Get(lc)
	if bot.HandoffConfig == nil || bot.HandoffConfig.EmailTo == "" {
		return errors.New(cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_NOT_CONFIGURED"])
	}

	// Load conversation messages (last 50 messages should be enough for handoff)
	messages, err := db.ListRecentMessages(ctx, s.DB, conversationID, 50)
	if err != nil {
		msg := cfg.ResponseTemplates.Errors["HANDOFF_CONVERSATION_LOAD_FAILED"]
		if msg == "" {
			msg = "failed to load conversation"
		}
		return fmt.Errorf("%s: %w", msg, err)
	}

	// Build email body
	emailBody := s.buildHandoffEmailBody(bot.Name, requestID, messages, notes, cfg)

	// Get email subject
	subject := bot.HandoffConfig.EmailSubject
	if subject == "" {
		tmpl := cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_SUBJECT"]
		if tmpl == "" {
			tmpl = "[Botla] New Support Request - %s"
		}
		subject = fmt.Sprintf(tmpl, bot.Name)
	}

	// Log the handoff request (email sending to be implemented with SMTP service)
	if s.Log != nil {
		s.Log.Info("handoff_email_request", map[string]any{
			"request_id":      requestID,
			"chatbot_id":      bot.ID,
			"conversation_id": conversationID,
			"email_to":        bot.HandoffConfig.EmailTo,
			"subject":         subject,
			"body_length":     len(emailBody),
		})
	}

	// Update analytics asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// isHandoff=true, tokens=0, responseTime=0 (not applicable for handoff event itself)
		if err := db.IncrementAnalytics(bgCtx, s.DB, bot.ID, time.Now(), false, 0, true, 0); err != nil {
			if s.Log != nil {
				s.Log.Error("handoff_analytics_failed", map[string]any{"error": err.Error()})
			}
		}
	}()

	// TODO: Implement actual email sending when SMTP service is available
	// For now, we just log the request and consider it successful
	// This allows the feature to work (creates handoff request) without email infra

	return nil
}

// buildHandoffEmailBody creates the email body with conversation transcript
func (s *HandoffService) buildHandoffEmailBody(botName, requestID string, messages []models.Message, notes string, cfg langconfig.LanguageConfig) string {
	var sb strings.Builder

	sb.WriteString(cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_BODY_HEADER"])
	sb.WriteString(fmt.Sprintf("Bot: %s\n", botName))
	sb.WriteString(fmt.Sprintf("%s: %s\n", cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_LABEL_REQUEST_ID"], requestID))
	sb.WriteString(fmt.Sprintf("%s: %s\n", cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_LABEL_DATE"], time.Now().Format("2006-01-02 15:04:05")))

	if notes != "" {
		sb.WriteString(fmt.Sprintf("\n%s:\n%s\n", cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_LABEL_USER_NOTE"], notes))
	}

	sb.WriteString("\n--- Konuşma Dökümü ---\n\n")

	for _, msg := range messages {
		role := cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_LABEL_USER"]
		if msg.Role == "assistant" {
			role = cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_LABEL_BOT"]
		}
		sb.WriteString(fmt.Sprintf("[%s] %s:\n%s\n\n", msg.CreatedAt.Format("15:04"), role, msg.Content))
	}

	sb.WriteString("---\n")
	sb.WriteString(cfg.ResponseTemplates.Errors["HANDOFF_EMAIL_BODY_FOOTER"])

	return sb.String()
}

// GetHandoffRequests returns all handoff requests for a chatbot
func (s *HandoffService) GetHandoffRequests(ctx context.Context, chatbotID string) ([]*models.HandoffRequest, error) {
	return db.GetHandoffRequestsByBotID(ctx, s.DB, chatbotID)
}

// UpdateHandoffStatus updates the status of a handoff request
func (s *HandoffService) UpdateHandoffStatus(ctx context.Context, requestID, status string, assignedTo *string) error {
	// Validate status
	validStatuses := map[string]bool{
		models.HandoffStatusPending:  true,
		models.HandoffStatusAssigned: true,
		models.HandoffStatusResolved: true,
	}
	if !validStatuses[status] {
		lc := "tr"
		msg := langconfig.Get(lc).ResponseTemplates.Errors["ERR_INVALID_STATUS"]
		formatted := fmt.Sprintf(msg, status)
		return errors.New(formatted)
	}

	return db.UpdateHandoffRequestStatus(ctx, s.DB, requestID, status, assignedTo)
}
