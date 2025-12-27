package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/onurceri/botla-co/pkg/requestid"
)

func TestNew_DefaultLevel(t *testing.T) {
	l := New("")
	l.Debug("msg", nil)
	l.Info("msg", map[string]any{"k": 1})
	l.Warn("msg", nil)
	l.Error("msg", nil)
}

func TestNew_UppercaseLevel(t *testing.T) {
	l := New("debug")
	l.Debug("d", nil)
	l.Info("i", nil)
}

func TestLoggerCtx_WithRequestID(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New("INFO")
	
	// Create context with request ID using the shared package
	ctx := requestid.ToContext(context.Background(), "test-req-123")
	
	// Log with context
	log.InfoCtx(ctx, "test_message", map[string]any{"key": "value"})
	
	// Restore stdout
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	// Parse JSON output
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	
	// Verify request_id is in fields
	fields, ok := entry["fields"].(map[string]any)
	if !ok {
		t.Fatal("expected fields to be a map")
	}
	
	if reqID, ok := fields["request_id"].(string); !ok || reqID != "test-req-123" {
		t.Errorf("expected request_id to be 'test-req-123', got %v", fields["request_id"])
	}
}

func TestLoggerCtx_WithoutRequestID(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log := New("INFO")
	
	// Create context without request ID
	ctx := context.Background()
	
	// Log with context
	log.InfoCtx(ctx, "test_message", map[string]any{"key": "value"})
	
	// Restore stdout
	w.Close()
	os.Stdout = old
	
	var buf bytes.Buffer
	io.Copy(&buf, r)
	
	// Parse JSON output
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	
	// Verify request_id is NOT in fields
	fields, ok := entry["fields"].(map[string]any)
	if !ok {
		t.Fatal("expected fields to be a map")
	}
	
	if _, exists := fields["request_id"]; exists {
		t.Error("expected request_id to not be present when not in context")
	}
}
