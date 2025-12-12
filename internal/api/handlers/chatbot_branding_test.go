package handlers

import (
	"testing"

	"github.com/onurceri/botla-co/internal/models"
)

func TestApplyChatbotUpdates_ClearsCustomBrandingWhenHiddenOff(t *testing.T) {
	c := &models.Chatbot{HideBranding: true, CustomBranding: &models.CustomBranding{Text: "ACME"}}
	req := createChatbotRequest{HideBranding: boolPtr(false), CustomBranding: nil}
	applyChatbotUpdates(c, req)
	if c.HideBranding != false {
		t.Fatalf("expected HideBranding=false, got %v", c.HideBranding)
	}
	if c.CustomBranding != nil {
		t.Fatalf("expected CustomBranding=nil when hiding is off")
	}
}

func boolPtr(b bool) *bool { return &b }
