package config

import (
	"os"
	"testing"
)

func setAllEnv() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "botla_test")
	os.Setenv("DB_USER", "botla")
	os.Setenv("DB_PASSWORD", "botla")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("OPENAI_API_KEY", "k")
	os.Setenv("JWT_SECRET", "s")
	os.Setenv("PORT", "8080")
}

func TestLoadConfig_Success_DefaultCORS(t *testing.T) {
	setAllEnv()
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	c := LoadConfig()
	if c == nil || c.CORS_ALLOWED_ORIGINS == "" {
		t.Fatalf("config load failed")
	}
}

// Override fatalf in tests to avoid os.Exit and capture calls.
func setFatalCapture(calls *int) { fatalf = func(_ string) { *calls++ } }

func TestLoadConfig_DBEnvMissing_Exit(t *testing.T) {
	env := []string{
		"DB_HOST=",
		"DB_PORT=",
		"DB_NAME=",
		"DB_USER=",
		"DB_PASSWORD=",
		"QDRANT_URL=http://localhost:6333",
		"OPENAI_API_KEY=k",
		"JWT_SECRET=s",
		"PORT=8080",
	}
	var calls int
	setFatalCapture(&calls)
	for _, kv := range env {
		k, v := splitKV(kv)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	LoadConfig()
	if calls == 0 {
		t.Fatalf("expected fatalf to be called")
	}
}

func TestLoadConfig_QdrantMissing_Exit(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=botla_test",
		"DB_USER=botla",
		"DB_PASSWORD=botla",
		"QDRANT_URL=",
		"OPENAI_API_KEY=k",
		"JWT_SECRET=s",
		"PORT=8080",
	}
	var calls int
	setFatalCapture(&calls)
	for _, kv := range env {
		k, v := splitKV(kv)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	LoadConfig()
	if calls == 0 {
		t.Fatalf("expected fatalf to be called")
	}
}

func TestLoadConfig_OpenAIMissing_Warns(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=botla_test",
		"DB_USER=botla",
		"DB_PASSWORD=botla",
		"QDRANT_URL=http://localhost:6333",
		"OPENAI_API_KEY=",
		"JWT_SECRET=s",
		"PORT=8080",
	}
	var calls int
	setFatalCapture(&calls)
	for _, kv := range env {
		k, v := splitKV(kv)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	cfg := LoadConfig()
	// OPENAI_API_KEY is optional (logs warning only), so fatalf should NOT be called
	if calls != 0 {
		t.Fatalf("expected fatalf NOT to be called for missing OPENAI_API_KEY")
	}
	if cfg == nil {
		t.Fatalf("expected config to be returned")
	}
}

func TestLoadConfig_JWTMissing_Exit(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=botla_test",
		"DB_USER=botla",
		"DB_PASSWORD=botla",
		"QDRANT_URL=http://localhost:6333",
		"OPENAI_API_KEY=k",
		"JWT_SECRET=",
		"PORT=8080",
	}
	var calls int
	setFatalCapture(&calls)
	for _, kv := range env {
		k, v := splitKV(kv)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	LoadConfig()
	if calls == 0 {
		t.Fatalf("expected fatalf to be called")
	}
}

func TestLoadConfig_PortMissing_Exit(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=botla_test",
		"DB_USER=botla",
		"DB_PASSWORD=botla",
		"QDRANT_URL=http://localhost:6333",
		"OPENAI_API_KEY=k",
		"JWT_SECRET=s",
		"PORT=",
	}
	var calls int
	setFatalCapture(&calls)
	for _, kv := range env {
		k, v := splitKV(kv)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	LoadConfig()
	if calls == 0 {
		t.Fatalf("expected fatalf to be called")
	}
}

// splitKV splits "KEY=VALUE" into key and value.
func splitKV(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

func TestLoadConfig_CookieSecure_Development(t *testing.T) {
	setAllEnv()
	os.Unsetenv("GO_ENV")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.CookieSecure != false {
		t.Errorf("expected CookieSecure to be false in development, got %v", c.CookieSecure)
	}
}

func TestLoadConfig_CookieSecure_Production(t *testing.T) {
	setAllEnv()
	os.Setenv("GO_ENV", "production")
	defer os.Unsetenv("GO_ENV")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.CookieSecure != true {
		t.Errorf("expected CookieSecure to be true in production, got %v", c.CookieSecure)
	}
}

func TestLoadConfig_AnalyticsWorkerCount_Default(t *testing.T) {
	setAllEnv()
	os.Unsetenv("ANALYTICS_WORKER_COUNT")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.ANALYTICS_WORKER_COUNT != 10 {
		t.Errorf("expected ANALYTICS_WORKER_COUNT default to be 10, got %d", c.ANALYTICS_WORKER_COUNT)
	}
}

func TestLoadConfig_AnalyticsWorkerCount_Custom(t *testing.T) {
	setAllEnv()
	os.Setenv("ANALYTICS_WORKER_COUNT", "20")
	defer os.Unsetenv("ANALYTICS_WORKER_COUNT")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.ANALYTICS_WORKER_COUNT != 20 {
		t.Errorf("expected ANALYTICS_WORKER_COUNT to be 20, got %d", c.ANALYTICS_WORKER_COUNT)
	}
}

func TestLoadConfig_AnalyticsWorkerCount_Invalid(t *testing.T) {
	setAllEnv()
	os.Setenv("ANALYTICS_WORKER_COUNT", "invalid")
	defer os.Unsetenv("ANALYTICS_WORKER_COUNT")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.ANALYTICS_WORKER_COUNT != 10 {
		t.Errorf("expected ANALYTICS_WORKER_COUNT to default to 10 for invalid value, got %d", c.ANALYTICS_WORKER_COUNT)
	}
}

func TestLoadConfig_AnalyticsWorkerCount_Zero(t *testing.T) {
	setAllEnv()
	os.Setenv("ANALYTICS_WORKER_COUNT", "0")
	defer os.Unsetenv("ANALYTICS_WORKER_COUNT")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.ANALYTICS_WORKER_COUNT != 10 {
		t.Errorf("expected ANALYTICS_WORKER_COUNT to default to 10 for zero value, got %d", c.ANALYTICS_WORKER_COUNT)
	}
}

func TestLoadConfig_AnalyticsWorkerCount_Negative(t *testing.T) {
	setAllEnv()
	os.Setenv("ANALYTICS_WORKER_COUNT", "-5")
	defer os.Unsetenv("ANALYTICS_WORKER_COUNT")
	c := LoadConfig()
	if c == nil {
		t.Fatalf("config load failed")
	}
	if c.ANALYTICS_WORKER_COUNT != 10 {
		t.Errorf("expected ANALYTICS_WORKER_COUNT to default to 10 for negative value, got %d", c.ANALYTICS_WORKER_COUNT)
	}
}
