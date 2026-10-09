package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/langconfig"
	"github.com/onurceri/botla-app/pkg/logger"
)

// HandoffService handles human agent handoff logic
type HandoffService struct {
	HandoffRepo      repository.HandoffRepository
	ConversationRepo repository.ConversationRepository
	AnalyticsRepo    repository.AnalyticsRepository
	// EmailSender delivers handoff notifications. Defaults to NoopEmailSender
	// (log only) because no mail provider is integrated yet.
	EmailSender EmailSender
	Log         *logger.Logger
}

// NewHandoffService creates a new HandoffService instance
func NewHandoffService(handoffRepo repository.HandoffRepository, conversationRepo repository.ConversationRepository, analyticsRepo repository.AnalyticsRepository, log *logger.Logger) *HandoffService {
	return &HandoffService{
		HandoffRepo:      handoffRepo,
		ConversationRepo: conversationRepo,
		AnalyticsRepo:    analyticsRepo,
		EmailSender:      NoopEmailSender{Log: log},
		Log:              log,
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

type InvalidHandoffStatusError struct {
	Status string
}

func (e InvalidHandoffStatusError) Error() string {
	return fmt.Sprintf("invalid status: %s", e.Status)
}

// RequestHandoff creates a handoff request and notifies operators
func (s *HandoffService) RequestHandoff(ctx context.Context, bot *models.Chatbot, conversationID, notes string) (*HandoffResult, error) {
	// Validate handoff is enabled
	if !bot.HandoffEnabled {
		return nil, ErrHandoffNotEnabled
	}

	lc := strings.TrimSpace(bot.LanguageCode)
	if i := strings.Index(lc, "-"); i > 0 {
		lc = lc[:i]
	}
	cfg := langconfig.Get(lc)

	// Create handoff request in database
	// Check for existing active handoff request
	exists, err := s.HandoffRepo.HasActiveHandoffRequest(ctx, conversationID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to check existing handoff")
	}
	if exists {
		return nil, ErrHandoffExists
	}

	req := &models.HandoffRequest{
		ChatbotID:      bot.ID,
		ConversationID: conversationID,
		Notes:          &notes,
	}

	requestID, err := s.HandoffRepo.CreateHandoffRequest(ctx, req)
	if err != nil {
		msg := cfg.UserMessages.Errors["HANDOFF_CREATE_FAILED"]
		if msg == "" {
			msg = "failed to create handoff request"
		}
		return nil, pkgerrors.Wrapf(err, "%s", msg)
	}

	result := &HandoffResult{
		RequestID: requestID,
		Status:    models.HandoffStatusPending,
	}

	// Handle based on handoff type
	switch bot.HandoffType {
	case string(models.HandoffTypeEmail):
		sent, emailErr := s.handleEmailHandoff(ctx, bot, requestID, conversationID, notes)
		if emailErr != nil {
			result.ErrorMessage = emailErr.Error()
			if s.Log != nil {
				s.Log.Warn("handoff_email_failed", map[string]any{
					"request_id": requestID,
					"error":      emailErr.Error(),
				})
			}
		}
		result.EmailSent = sent
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
		hm := cfg.UserMessages.Errors["HANDOFF_RECEIVED"]
		if hm == "" {
			hm = "Talebiniz alındı. En kısa sürede bir temsilcimiz sizinle iletişime geçecektir."
		}
		result.Message = hm
	}

	return result, nil
}

// errHandoffEmailSendFailed is the user-facing error when the transport fails;
// the underlying error is logged, not returned to the widget.
var errHandoffEmailSendFailed = errors.New("failed to send handoff email")

// handleEmailHandoff notifies the chatbot's operator address about a handoff.
// It reports whether an email was actually delivered: with no mail provider
// configured (ErrEmailNotConfigured) the request still succeeds but sent=false.
func (s *HandoffService) handleEmailHandoff(ctx context.Context, bot *models.Chatbot, requestID, conversationID, notes string) (bool, error) {
	lc := strings.TrimSpace(bot.LanguageCode)
	if i := strings.Index(lc, "-"); i > 0 {
		lc = lc[:i]
	}
	cfg := langconfig.Get(lc)
	if bot.HandoffConfig == nil || bot.HandoffConfig.EmailTo == "" {
		return false, errors.New(cfg.UserMessages.Errors["HANDOFF_EMAIL_NOT_CONFIGURED"])
	}

	// Load conversation messages (last 50 messages should be enough for handoff)
	messages, err := s.ConversationRepo.ListRecentMessages(ctx, conversationID, 50)
	if err != nil {
		msg := cfg.UserMessages.Errors["HANDOFF_CONVERSATION_LOAD_FAILED"]
		if msg == "" {
			msg = "failed to load conversation"
		}
		return false, pkgerrors.Wrapf(err, "%s", msg)
	}

	// Build email body
	emailBody := s.buildHandoffEmailBody(bot.Name, requestID, messages, notes, cfg)

	// Get email subject
	subject := bot.HandoffConfig.EmailSubject
	if subject == "" {
		tmpl := cfg.UserMessages.Errors["HANDOFF_EMAIL_SUBJECT"]
		if tmpl == "" {
			tmpl = "[Botla] New Support Request - %s"
		}
		subject = fmt.Sprintf(tmpl, bot.Name)
	}

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
	// LI-001: Wrap in panic recovery to prevent crash
	go func() {
		defer func() {
			if r := recover(); r != nil && s.Log != nil {
				s.Log.Error("handoff_analytics_panic", map[string]any{"panic": r})
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// isHandoff=true, tokens=0, responseTime=0 (not applicable for handoff event itself)
		if err := s.AnalyticsRepo.IncrementAnalytics(bgCtx, bot.ID, false, 0, true, 0); err != nil {
			if s.Log != nil {
				s.Log.Error("handoff_analytics_failed", map[string]any{"error": err.Error()})
			}
		}
	}()

	sender := s.EmailSender
	if sender == nil {
		sender = NoopEmailSender{Log: s.Log}
	}
	err = sender.Send(ctx, EmailMessage{
		To:      []string{bot.HandoffConfig.EmailTo},
		Subject: subject,
		Body:    emailBody,
	})
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrEmailNotConfigured):
		// No mail provider integrated: the request is stored and visible in the
		// dashboard, it just is not emailed.
		return false, nil
	default:
		if s.Log != nil {
			s.Log.Error("handoff_email_send_failed", map[string]any{
				"request_id": requestID,
				"chatbot_id": bot.ID,
				"error":      err.Error(),
			})
		}
		return false, errHandoffEmailSendFailed
	}
}

// buildHandoffEmailBody creates the email body with conversation transcript
func (s *HandoffService) buildHandoffEmailBody(botName, requestID string, messages []models.Message, notes string, cfg langconfig.LanguageConfig) string {
	var sb strings.Builder

	sb.WriteString(cfg.UserMessages.Errors["HANDOFF_EMAIL_BODY_HEADER"])
	sb.WriteString(fmt.Sprintf("Bot: %s\n", botName))
	sb.WriteString(fmt.Sprintf("%s: %s\n", cfg.UserMessages.Errors["HANDOFF_EMAIL_LABEL_REQUEST_ID"], requestID))
	sb.WriteString(fmt.Sprintf("%s: %s\n", cfg.UserMessages.Errors["HANDOFF_EMAIL_LABEL_DATE"], time.Now().Format("2006-01-02 15:04:05")))

	if notes != "" {
		sb.WriteString(fmt.Sprintf("\n%s:\n%s\n", cfg.UserMessages.Errors["HANDOFF_EMAIL_LABEL_USER_NOTE"], notes))
	}

	sb.WriteString("\n--- Konuşma Dökümü ---\n\n")

	for _, msg := range messages {
		role := cfg.UserMessages.Errors["HANDOFF_EMAIL_LABEL_USER"]
		if msg.Role == "assistant" {
			role = cfg.UserMessages.Errors["HANDOFF_EMAIL_LABEL_BOT"]
		}
		sb.WriteString(fmt.Sprintf("[%s] %s:\n%s\n\n", msg.CreatedAt.Format("15:04"), role, msg.Content))
	}

	sb.WriteString("---\n")
	sb.WriteString(cfg.UserMessages.Errors["HANDOFF_EMAIL_BODY_FOOTER"])

	return sb.String()
}

// Handoff list paging bounds.
const (
	DefaultHandoffListLimit = 50
	MaxHandoffListLimit     = 100
)

// HandoffRequestPage is one page of a chatbot's handoff requests.
type HandoffRequestPage struct {
	Requests []*models.HandoffRequest `json:"requests"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	Limit    int                      `json:"limit"`
}

// ListHandoffRequests returns page (1-based) of a chatbot's handoff requests,
// newest first. An empty status lists all statuses; an unknown status yields
// InvalidHandoffStatusError. page < 1 and limit outside 1..MaxHandoffListLimit
// are rejected with ErrInvalidRequest.
func (s *HandoffService) ListHandoffRequests(ctx context.Context, chatbotID, status string, page, limit int) (*HandoffRequestPage, error) {
	if status != "" && !models.IsValidHandoffStatus(status) {
		return nil, InvalidHandoffStatusError{Status: status}
	}
	if page < 1 || limit < 1 || limit > MaxHandoffListLimit {
		return nil, ErrInvalidRequest
	}

	requests, total, err := s.HandoffRepo.ListHandoffRequestsByBotID(ctx, chatbotID, models.HandoffListFilter{
		Status: status,
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "list handoff requests")
	}
	if requests == nil {
		requests = []*models.HandoffRequest{}
	}
	return &HandoffRequestPage{Requests: requests, Total: total, Page: page, Limit: limit}, nil
}

// UpdateHandoffStatus updates the status of a handoff request
func (s *HandoffService) UpdateHandoffStatus(ctx context.Context, requestID, status string, assignedTo *string) error {
	if !models.IsValidHandoffStatus(status) {
		return InvalidHandoffStatusError{Status: status}
	}

	if err := s.HandoffRepo.UpdateHandoffRequestStatus(ctx, requestID, status, assignedTo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrHandoffNotFound
		}
		return pkgerrors.Wrapf(err, "update handoff status")
	}
	return nil
}

// maxContactEmailLength is the RFC 5321 limit for a forward-path address.
const maxContactEmailLength = 254

// NormalizeContactEmail trims and validates a visitor-supplied contact email.
// It accepts only a bare address (no display name such as "Bob <bob@x.io>"),
// requires a dotted domain and enforces the 254-character limit.
func NormalizeContactEmail(raw string) (string, error) {
	email := strings.TrimSpace(raw)
	if email == "" || len(email) > maxContactEmailLength {
		return "", ErrInvalidContactEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Name != "" || addr.Address != email {
		return "", ErrInvalidContactEmail
	}
	at := strings.LastIndex(email, "@")
	domain := email[at+1:]
	if at <= 0 || !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return "", ErrInvalidContactEmail
	}
	return email, nil
}

// SubmitContactEmail records the visitor's contact email on a handoff request of
// chatbotID and stamps contacted_at. Returns ErrInvalidContactEmail for a bad
// address, ErrHandoffNotFound when the request does not exist or belongs to
// another chatbot, and ErrHandoffClosed when it is already resolved.
func (s *HandoffService) SubmitContactEmail(ctx context.Context, chatbotID, requestID, email string) error {
	normalized, err := NormalizeContactEmail(email)
	if err != nil {
		return err
	}

	req, err := s.HandoffRepo.GetHandoffRequestByID(ctx, requestID)
	if err != nil {
		return pkgerrors.Wrapf(err, "get handoff request")
	}
	if req == nil || req.ChatbotID != chatbotID {
		return ErrHandoffNotFound
	}
	if req.Status == models.HandoffStatusResolved {
		return ErrHandoffClosed
	}

	if err := s.HandoffRepo.SetHandoffContactEmail(ctx, requestID, chatbotID, normalized); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrHandoffNotFound
		}
		return pkgerrors.Wrapf(err, "set handoff contact email")
	}
	return nil
}
