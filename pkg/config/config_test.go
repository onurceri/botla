package config

import (
	"os"
	"testing"
)

func setAllEnv() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "botla_dev")
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
func setFatalCapture(calls *int) { fatalf = func(msg string) { *calls++ } }

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
		"DB_NAME=botla_dev",
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

func TestLoadConfig_OpenAIMissing_Exit(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=botla_dev",
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
	LoadConfig()
	if calls == 0 {
		t.Fatalf("expected fatalf to be called")
	}
}

func TestLoadConfig_JWTMissing_Exit(t *testing.T) {
	env := []string{
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_NAME=botla_dev",
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
		"DB_NAME=botla_dev",
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
