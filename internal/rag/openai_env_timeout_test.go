package rag

import (
	"testing"

	"github.com/onurceri/botla-co/pkg/config"
)

func TestNewOpenAIClient_TimeoutOverride(t *testing.T) {
	cfg := &config.Config{
		OPENAI_API_KEY:    "k",
		OPENAI_TIMEOUT_MS: 2500,
	}
	c, err := NewOpenAIClient(cfg)
	if err != nil {
		t.Fatalf("client err: %v", err)
	}
	if c.http.Timeout.Milliseconds() != 2500 {
		t.Fatalf("timeout not applied")
	}
}
