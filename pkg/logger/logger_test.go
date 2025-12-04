package logger

import "testing"

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
