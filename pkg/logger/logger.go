package logger

import (
	"encoding/json"
	"os"
	"strings"
	"time"
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
