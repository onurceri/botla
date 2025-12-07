package handlers

import (
	"os"
	"testing"
	"time"
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

