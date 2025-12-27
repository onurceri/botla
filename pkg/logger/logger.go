package logger

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/onurceri/botla-co/pkg/requestid"
)

type Logger struct {
	level string
}

func New(level string) *Logger {
	if level == "" {
		level = "INFO"
	}
	return &Logger{level: strings.ToUpper(level)}
}

func (l *Logger) allowed(level string) bool {
	order := map[string]int{"DEBUG": 0, "INFO": 1, "WARN": 2, "ERROR": 3}
	return order[strings.ToUpper(level)] >= order[l.level]
}

func (l *Logger) write(level string, msg string, fields map[string]any) {
	if !l.allowed(level) {
		return
	}
	entry := map[string]any{
		"ts":     time.Now().UTC().Format(time.RFC3339Nano),
		"level":  level,
		"msg":    msg,
		"fields": fields,
	}
	b, _ := json.Marshal(entry)
	_, _ = os.Stdout.Write(append(b, '\n'))
}

func (l *Logger) Debug(msg string, fields map[string]any) { l.write("DEBUG", msg, fields) }
func (l *Logger) Info(msg string, fields map[string]any)  { l.write("INFO", msg, fields) }
func (l *Logger) Warn(msg string, fields map[string]any)  { l.write("WARN", msg, fields) }
func (l *Logger) Error(msg string, fields map[string]any) { l.write("ERROR", msg, fields) }

// InfoCtx logs with request ID from context
func (l *Logger) InfoCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := extractRequestID(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("INFO", msg, fields)
}

// ErrorCtx logs with request ID from context
func (l *Logger) ErrorCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := extractRequestID(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("ERROR", msg, fields)
}

// WarnCtx logs with request ID from context
func (l *Logger) WarnCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := extractRequestID(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("WARN", msg, fields)
}

// DebugCtx logs with request ID from context
func (l *Logger) DebugCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := extractRequestID(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("DEBUG", msg, fields)
}

// extractRequestID extracts request ID from context using the shared requestid package
func extractRequestID(ctx context.Context) string {
	return requestid.FromContext(ctx)
}
