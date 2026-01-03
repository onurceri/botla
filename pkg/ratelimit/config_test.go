package ratelimit

import (
	"testing"
	"time"
)

func TestNewConfig_Defaults(t *testing.T) {
	// Empty settings should result in default values
	cfg := NewConfig(Settings{})

	// Check defaults
	if cfg.Global.RequestsPerWindow != 60 {
		t.Errorf("expected global requests 60, got %d", cfg.Global.RequestsPerWindow)
	}
	if cfg.Global.WindowSize != 60*time.Second {
		t.Errorf("expected global window 60s, got %v", cfg.Global.WindowSize)
	}
	if cfg.User.RequestsPerWindow != 120 {
		t.Errorf("expected user requests 120, got %d", cfg.User.RequestsPerWindow)
	}
	if cfg.User.WindowSize != 60*time.Second {
		t.Errorf("expected user window 60s, got %v", cfg.User.WindowSize)
	}
}

func TestNewConfig_CustomValues(t *testing.T) {
	s := Settings{
		GlobalRequestsPerMinute: 100,
		GlobalWindowSeconds:     30,
		UserRequestsPerMinute:   200,
		UserWindowSeconds:       45,
	}

	cfg := NewConfig(s)

	if cfg.Global.RequestsPerWindow != 100 {
		t.Errorf("expected global requests 100, got %d", cfg.Global.RequestsPerWindow)
	}
	if cfg.Global.WindowSize != 30*time.Second {
		t.Errorf("expected global window 30s, got %v", cfg.Global.WindowSize)
	}
	if cfg.User.RequestsPerWindow != 200 {
		t.Errorf("expected user requests 200, got %d", cfg.User.RequestsPerWindow)
	}
	if cfg.User.WindowSize != 45*time.Second {
		t.Errorf("expected user window 45s, got %v", cfg.User.WindowSize)
	}
}

func TestNewConfig_EndpointOverrides(t *testing.T) {
	s := Settings{
		EndpointChat:    15,
		EndpointSources: 5,
	}

	cfg := NewConfig(s)

	chatCfg, exists := cfg.EndpointOverrides["/api/v1/chat"]
	if !exists {
		t.Fatal("chat endpoint override not found")
	}
	if chatCfg.RequestsPerWindow != 15 {
		t.Errorf("expected chat requests 15, got %d", chatCfg.RequestsPerWindow)
	}

	sourcesCfg, exists := cfg.EndpointOverrides["/api/v1/sources"]
	if !exists {
		t.Fatal("sources endpoint override not found")
	}
	if sourcesCfg.RequestsPerWindow != 5 {
		t.Errorf("expected sources requests 5, got %d", sourcesCfg.RequestsPerWindow)
	}
}

func TestKey(t *testing.T) {
	tests := []struct {
		tier       Tier
		identifier string
		expected   string
	}{
		{TierGlobal, "192.168.1.1", "ratelimit:global:192.168.1.1"},
		{TierUser, "user-123", "ratelimit:user:user-123"},
		{TierEndpoint, "/api/v1/chat:user:456", "ratelimit:endpoint:/api/v1/chat:user:456"},
	}

	for _, tt := range tests {
		result := Key(tt.tier, tt.identifier)
		if result != tt.expected {
			t.Errorf("Key(%s, %s) = %s; want %s", tt.tier, tt.identifier, result, tt.expected)
		}
	}
}

func TestDefaultConfig_AuthEndpoints(t *testing.T) {
	cfg := DefaultConfig()

	// Test login endpoint default
	loginCfg, exists := cfg.EndpointOverrides["/api/v1/auth/login"]
	if !exists {
		t.Fatal("login endpoint override not found in defaults")
	}
	if loginCfg.RequestsPerWindow != 5 {
		t.Errorf("expected login requests 5, got %d", loginCfg.RequestsPerWindow)
	}
	if loginCfg.WindowSize != 60*time.Second {
		t.Errorf("expected login window 60s, got %v", loginCfg.WindowSize)
	}

	// Test register endpoint default
	registerCfg, exists := cfg.EndpointOverrides["/api/v1/auth/register"]
	if !exists {
		t.Fatal("register endpoint override not found in defaults")
	}
	if registerCfg.RequestsPerWindow != 3 {
		t.Errorf("expected register requests 3, got %d", registerCfg.RequestsPerWindow)
	}

	// Test refresh endpoint default
	refreshCfg, exists := cfg.EndpointOverrides["/api/v1/auth/refresh"]
	if !exists {
		t.Fatal("refresh endpoint override not found in defaults")
	}
	if refreshCfg.RequestsPerWindow != 10 {
		t.Errorf("expected refresh requests 10, got %d", refreshCfg.RequestsPerWindow)
	}
}

func TestNewConfig_AuthEndpointOverrides(t *testing.T) {
	s := Settings{
		AuthLogin:    10,
		AuthRegister: 2,
		AuthRefresh:  20,
	}

	cfg := NewConfig(s)

	// Login should be overridden
	loginCfg, exists := cfg.EndpointOverrides["/api/v1/auth/login"]
	if !exists {
		t.Fatal("login endpoint override not found")
	}
	if loginCfg.RequestsPerWindow != 10 {
		t.Errorf("expected login requests 10, got %d", loginCfg.RequestsPerWindow)
	}

	// Register should be overridden
	registerCfg, exists := cfg.EndpointOverrides["/api/v1/auth/register"]
	if !exists {
		t.Fatal("register endpoint override not found")
	}
	if registerCfg.RequestsPerWindow != 2 {
		t.Errorf("expected register requests 2, got %d", registerCfg.RequestsPerWindow)
	}

	// Refresh should be overridden
	refreshCfg, exists := cfg.EndpointOverrides["/api/v1/auth/refresh"]
	if !exists {
		t.Fatal("refresh endpoint override not found")
	}
	if refreshCfg.RequestsPerWindow != 20 {
		t.Errorf("expected refresh requests 20, got %d", refreshCfg.RequestsPerWindow)
	}
}
