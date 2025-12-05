package handlers

import (
	"os"
	"testing"
	"time"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

func TestChatTimeoutEnvParsing(t *testing.T) {
	old := os.Getenv("CHAT_TIMEOUT_MS")
	defer func() { _ = os.Setenv("CHAT_TIMEOUT_MS", old) }()
	_ = os.Unsetenv("CHAT_TIMEOUT_MS")
	if chatTimeout() != 20*time.Second {
		t.Fatalf("default timeout mismatch")
	}
	_ = os.Setenv("CHAT_TIMEOUT_MS", "1500")
	if chatTimeout() != 1500*time.Millisecond {
		t.Fatalf("env timeout mismatch")
	}
}

func TestDefaultLang(t *testing.T) {
	if defaultLang("") != "tr" {
		t.Fatal("empty must default to tr")
	}
	if defaultLang("en") != "en" {
		t.Fatal("non-empty must be preserved")
	}
}

func TestSystemPromptFallback(t *testing.T) {
	cfg := langconfig.Get("tr")
	if systemPrompt("", cfg) != cfg.ResponseTemplates.DefaultSystemPrompt {
		t.Fatalf("expected default system prompt fallback")
	}
	if systemPrompt("x", cfg) != "x" {
		t.Fatalf("expected provided system prompt")
	}
}
