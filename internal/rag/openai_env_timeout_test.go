package rag

import (
	"testing"
)

func TestNewOpenAIClientFromEnv_TimeoutOverride(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "k")
	t.Setenv("OPENAI_TIMEOUT_MS", "2500")
	c, err := NewOpenAIClientFromEnv()
	if err != nil {
		t.Fatalf("client err: %v", err)
	}
	if c.http.Timeout.Milliseconds() != 2500 {
		t.Fatalf("timeout not applied")
	}
}
