package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/logger"
)

type recordingEmailSender struct {
	err  error
	sent []EmailMessage
}

func (r *recordingEmailSender) Send(_ context.Context, msg EmailMessage) error {
	r.sent = append(r.sent, msg)
	return r.err
}

func TestNoopEmailSender_ReportsNotConfigured(t *testing.T) {
	for _, s := range []NoopEmailSender{{}, {Log: logger.New("test")}} {
		err := s.Send(context.Background(), EmailMessage{To: []string{"ops@example.com"}, Subject: "s", Body: "b"})
		if !errors.Is(err, ErrEmailNotConfigured) {
			t.Fatalf("expected ErrEmailNotConfigured, got %v", err)
		}
	}
}

func TestNewHandoffService_DefaultsToNoopSender(t *testing.T) {
	svc := NewHandoffService(nil, nil, nil, nil)
	if _, ok := svc.EmailSender.(NoopEmailSender); !ok {
		t.Fatalf("expected NoopEmailSender default, got %T", svc.EmailSender)
	}
}

func newEmailHandoffService(sender EmailSender) *HandoffService {
	handoffRepo := repository.NewMockHandoffRepo()
	handoffRepo.CreateHandoffRequestFunc = func(context.Context, *models.HandoffRequest) (string, error) {
		return "req-1", nil
	}
	convRepo := repository.NewMockConversationRepo()
	convRepo.ListRecentMessagesFunc = func(context.Context, string, int) ([]models.Message, error) {
		return []models.Message{{Role: "user", Content: "help me"}}, nil
	}
	svc := NewHandoffService(handoffRepo, convRepo, repository.NewMockAnalyticsRepo(), nil)
	svc.EmailSender = sender
	return svc
}

func emailHandoffBot() *models.Chatbot {
	return &models.Chatbot{
		ID:             "bot-1",
		Name:           "Support Bot",
		LanguageCode:   "en-US",
		HandoffEnabled: true,
		HandoffType:    string(models.HandoffTypeEmail),
		HandoffConfig:  &models.HandoffConfig{EmailTo: "ops@example.com"},
	}
}

func TestRequestHandoff_EmailDelivery(t *testing.T) {
	tests := []struct {
		name          string
		sender        EmailSender
		wantSent      bool
		wantErrorText string
	}{
		{name: "delivered", sender: &recordingEmailSender{}, wantSent: true},
		{name: "no provider configured", sender: &recordingEmailSender{err: ErrEmailNotConfigured}, wantSent: false},
		{name: "nil sender falls back to noop", sender: nil, wantSent: false},
		{name: "transport failure", sender: &recordingEmailSender{err: errors.New("smtp: 421 try later")}, wantSent: false, wantErrorText: errHandoffEmailSendFailed.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newEmailHandoffService(tt.sender)
			res, err := svc.RequestHandoff(context.Background(), emailHandoffBot(), "conv-1", "please call me")
			if err != nil {
				t.Fatalf("RequestHandoff must not fail on email problems: %v", err)
			}
			if res.RequestID != "req-1" {
				t.Fatalf("unexpected request id %q", res.RequestID)
			}
			if res.EmailSent != tt.wantSent {
				t.Fatalf("EmailSent = %v, want %v", res.EmailSent, tt.wantSent)
			}
			if res.ErrorMessage != tt.wantErrorText {
				t.Fatalf("ErrorMessage = %q, want %q", res.ErrorMessage, tt.wantErrorText)
			}
			if strings.Contains(res.ErrorMessage, "smtp") {
				t.Fatalf("transport details leaked to caller: %q", res.ErrorMessage)
			}
		})
	}
}

func TestRequestHandoff_EmailMessageContent(t *testing.T) {
	sender := &recordingEmailSender{}
	svc := newEmailHandoffService(sender)
	bot := emailHandoffBot()
	bot.HandoffConfig.EmailSubject = "Custom subject"

	if _, err := svc.RequestHandoff(context.Background(), bot, "conv-1", "please call me"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(sender.sent))
	}
	msg := sender.sent[0]
	if len(msg.To) != 1 || msg.To[0] != "ops@example.com" {
		t.Fatalf("unexpected recipients %v", msg.To)
	}
	if msg.Subject != "Custom subject" {
		t.Fatalf("unexpected subject %q", msg.Subject)
	}
	for _, want := range []string{"Support Bot", "req-1", "please call me", "help me"} {
		if !strings.Contains(msg.Body, want) {
			t.Fatalf("body missing %q:\n%s", want, msg.Body)
		}
	}
}
