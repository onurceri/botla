package services

import (
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/langconfig"
)

// TestBuildHandoffEmailBody verifies HND-004 (Turkish template) and HND-005 (Conversation history)
func TestBuildHandoffEmailBody(t *testing.T) {
	svc := &HandoffService{}
	botName := "Test Bot"
	reqID := "req-123"
	notes := "User needs help"

	// Mock messages for HND-005 (History)
	msgs := []models.Message{
		{
			Role:      "user",
			Content:   "Hello, I have a problem",
			CreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			Role:      "assistant",
			Content:   "Hi, how can I help?",
			CreatedAt: time.Date(2023, 1, 1, 10, 0, 5, 0, time.UTC),
		},
	}

	// 1. Test English (Default/Fallback)
	cfgEn := langconfig.Get("en")
	bodyEn := svc.buildHandoffEmailBody(botName, reqID, msgs, notes, cfgEn)

	// Check content
	if !strings.Contains(bodyEn, "Test Bot") {
		t.Error("EN: missing bot name")
	}
	if !strings.Contains(bodyEn, "req-123") {
		t.Error("EN: missing request ID")
	}
	if !strings.Contains(bodyEn, "User needs help") {
		t.Error("EN: missing notes")
	}
	if !strings.Contains(bodyEn, "Hello, I have a problem") {
		t.Error("EN: missing user message (HND-005)")
	}
	if !strings.Contains(bodyEn, "Hi, how can I help?") {
		t.Error("EN: missing bot message (HND-005)")
	}

	// 2. Test Turkish (HND-004)
	cfgTr := langconfig.Get("tr")
	bodyTr := svc.buildHandoffEmailBody(botName, reqID, msgs, notes, cfgTr)

	// Check Turkish specific strings
	// "Yeni bir destek talebi alındı"
	if !strings.Contains(bodyTr, "Yeni bir destek talebi alındı") {
		t.Error("TR: missing Turkish header (HND-004)")
	}
	// "Talep ID"
	if !strings.Contains(bodyTr, "Talep ID") {
		t.Error("TR: missing Turkish label 'Talep ID'")
	}
	// "Kullanıcı Notu"
	if !strings.Contains(bodyTr, "Kullanıcı Notu") {
		t.Error("TR: missing Turkish label 'Kullanıcı Notu'")
	}
	// "Konuşma Dökümü"
	if !strings.Contains(bodyTr, "Konuşma Dökümü") {
		t.Error("TR: missing Turkish transcript header")
	}
}
