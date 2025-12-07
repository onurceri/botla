package config

import "testing"

func TestResolveChatbotModel_DefaultsFromEnv(t *testing.T) {
  cfg := &Config{}
  v := ResolveChatbotModel(cfg)
  if v == "" {
    t.Fatalf("expected non-empty model")
  }
}

func TestResolveChatbotModel_OverridesWhenSet(t *testing.T) {
  cfg := &Config{DEFAULT_CHATBOT_MODEL: "gpt-4o"}
  v := ResolveChatbotModel(cfg)
  if v != "gpt-4o" {
    t.Fatalf("expected override, got %q", v)
  }
}
