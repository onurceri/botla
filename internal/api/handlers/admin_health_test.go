package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/config"
)

func TestAdminHealthHandlers_GetDetailedHealth(t *testing.T) {
	// Use the project's testdb pattern for real DB connection
	db := testdb.OpenTestDB(t)

	// Create mock servers for external services
	qdrantSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer qdrantSrv.Close()

	openaiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer openaiSrv.Close()

	cfg := &config.Config{
		QDRANT_URL:      qdrantSrv.URL,
		OPENAI_API_KEY:  "test-key",
		OPENAI_API_BASE: openaiSrv.URL,
		R2_ACCOUNT_ID:   "test-account",
		R2_BUCKET_NAME:  "test-bucket",
		GO_ENV:          "test",
	}

	h := NewAdminHealthHandlers(db, nil, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health/detailed", nil)
	rec := httptest.NewRecorder()

	h.GetDetailedHealth(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 200 or 503, got %d", rec.Code)
	}

	var response DetailedHealth
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status == "" {
		t.Error("expected status to be set")
	}
	if response.Version == "" {
		t.Error("expected version to be set")
	}
	if response.Uptime == "" {
		t.Error("expected uptime to be set")
	}
	if len(response.Dependencies) == 0 {
		t.Error("expected at least one dependency")
	}
}

func TestCheckPostgres_Success(t *testing.T) {
	db := testdb.OpenTestDB(t)
	h := &AdminHealthHandlers{DB: db, Cfg: &config.Config{}}

	result := h.checkPostgres(context.Background())

	if result.Name != "postgres" {
		t.Errorf("expected name 'postgres', got %s", result.Name)
	}
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", result.Status)
	}
	if result.LatencyMs < 0 {
		t.Error("expected positive latency")
	}
}

func TestCheckQdrant_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("expected path /healthz, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			QDRANT_URL: server.URL,
		},
	}

	result := h.checkQdrant(context.Background())

	if result.Name != "qdrant" {
		t.Errorf("expected name 'qdrant', got %s", result.Name)
	}
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %s (%s)", result.Status, result.Message)
	}
}

func TestCheckQdrant_Down(t *testing.T) {
	// Create a mock server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			QDRANT_URL: server.URL,
		},
	}

	result := h.checkQdrant(context.Background())

	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
}

func TestCheckQdrant_WithAPIKey(t *testing.T) {
	// Create a mock server that checks for API key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") != "test-key" {
			t.Error("expected api-key header to be set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			QDRANT_URL:     server.URL,
			QDRANT_API_KEY: "test-key",
		},
	}

	result := h.checkQdrant(context.Background())

	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", result.Status)
	}
}

func TestCheckOpenAI_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("expected path /v1/models, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			OPENAI_API_KEY:  "test-key",
			OPENAI_API_BASE: server.URL,
		},
	}

	result := h.checkOpenAI(context.Background())

	if result.Name != "openai" {
		t.Errorf("expected name 'openai', got %s", result.Name)
	}
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %s (%s)", result.Status, result.Message)
	}
}

func TestCheckOpenAI_NoAPIKey(t *testing.T) {
	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			OPENAI_API_KEY: "",
		},
	}

	result := h.checkOpenAI(context.Background())

	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
	if result.Message != "API key not configured" {
		t.Errorf("expected 'API key not configured', got %s", result.Message)
	}
}

func TestCheckOpenAI_APIError(t *testing.T) {
	// Create a mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			OPENAI_API_KEY:  "invalid-key",
			OPENAI_API_BASE: server.URL,
		},
	}

	result := h.checkOpenAI(context.Background())

	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
}

func TestCheckStorage_Success(t *testing.T) {
	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			R2_ACCOUNT_ID:  "account-id",
			R2_BUCKET_NAME: "bucket-name",
		},
	}

	result := h.checkStorage(context.Background())

	if result.Name != "storage" {
		t.Errorf("expected name 'storage', got %s", result.Name)
	}
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %s", result.Status)
	}
}

func TestCheckStorage_NotConfigured(t *testing.T) {
	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			R2_ACCOUNT_ID:  "",
			R2_BUCKET_NAME: "",
		},
	}

	result := h.checkStorage(context.Background())

	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
	if result.Message != "storage not configured" {
		t.Errorf("expected 'storage not configured', got %s", result.Message)
	}
}

func TestCheckRedis_NotConfigured(t *testing.T) {
	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			REDIS_URL: "",
		},
	}

	result := h.checkRedis(context.Background())

	if result.Name != "redis" {
		t.Errorf("expected name 'redis', got %s", result.Name)
	}
	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
	if result.Message != "not configured" {
		t.Errorf("expected 'not configured', got %s", result.Message)
	}
}

func TestCheckRedis_InvalidURL(t *testing.T) {
	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			REDIS_URL: "not-a-valid-url",
		},
	}

	result := h.checkRedis(context.Background())

	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
	if result.Message == "" {
		t.Error("expected error message for invalid URL")
	}
}

func TestCheckRedis_ConnectionError(t *testing.T) {
	// Use a URL that will fail to connect (non-existent port)
	h := &AdminHealthHandlers{
		Cfg: &config.Config{
			REDIS_URL: "redis://localhost:63999", // Non-existent port
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := h.checkRedis(ctx)

	if result.Status != "down" {
		t.Errorf("expected status 'down', got %s", result.Status)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "seconds only",
			duration: 5 * time.Second,
			want:     "5s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 30*time.Second,
			want:     "2m 30s",
		},
		{
			name:     "hours minutes seconds",
			duration: 3*time.Hour + 15*time.Minute + 45*time.Second,
			want:     "3h 15m 45s",
		},
		{
			name:     "days hours minutes",
			duration: 2*24*time.Hour + 5*time.Hour + 30*time.Minute,
			want:     "2d 5h 30m",
		},
		{
			name:     "zero duration",
			duration: 0,
			want:     "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	version := getVersion()
	if version != "dev" {
		t.Errorf("expected 'dev', got %s", version)
	}

	// Set env variable
	t.Setenv("APP_VERSION", "1.0.0")
	version = getVersion()
	if version != "1.0.0" {
		t.Errorf("expected '1.0.0', got %s", version)
	}
}

func TestDependencyStatus_JSON(t *testing.T) {
	status := DependencyStatus{
		Name:      "test",
		Status:    "ok",
		LatencyMs: 10.5,
		CheckedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DependencyStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != status.Name {
		t.Errorf("expected name %s, got %s", status.Name, decoded.Name)
	}
	if decoded.Status != status.Status {
		t.Errorf("expected status %s, got %s", status.Status, decoded.Status)
	}
}

func TestDetailedHealth_JSON(t *testing.T) {
	health := DetailedHealth{
		Status:      "healthy",
		Version:     "1.0.0",
		Uptime:      "1h 30m",
		Environment: "production",
		Dependencies: []DependencyStatus{
			{Name: "postgres", Status: "ok", LatencyMs: 5},
			{Name: "qdrant", Status: "ok", LatencyMs: 10},
		},
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DetailedHealth
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Status != health.Status {
		t.Errorf("expected status %s, got %s", health.Status, decoded.Status)
	}
	if len(decoded.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(decoded.Dependencies))
	}
}

func TestOverallStatusCalculation(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []DependencyStatus
		wantStatus   string
	}{
		{
			name: "all healthy",
			dependencies: []DependencyStatus{
				{Name: "postgres", Status: "ok"},
				{Name: "qdrant", Status: "ok"},
				{Name: "openai", Status: "ok"},
			},
			wantStatus: "healthy",
		},
		{
			name: "postgres down is unhealthy",
			dependencies: []DependencyStatus{
				{Name: "postgres", Status: "down"},
				{Name: "qdrant", Status: "ok"},
				{Name: "openai", Status: "ok"},
			},
			wantStatus: "unhealthy",
		},
		{
			name: "non-critical down is degraded",
			dependencies: []DependencyStatus{
				{Name: "postgres", Status: "ok"},
				{Name: "qdrant", Status: "down"},
				{Name: "openai", Status: "ok"},
			},
			wantStatus: "degraded",
		},
		{
			name: "degraded status",
			dependencies: []DependencyStatus{
				{Name: "postgres", Status: "ok"},
				{Name: "qdrant", Status: "degraded"},
				{Name: "openai", Status: "ok"},
			},
			wantStatus: "degraded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overallStatus := "healthy"
			downCount := 0
			degradedCount := 0

			for _, dep := range tt.dependencies {
				switch dep.Status {
				case "down":
					downCount++
				case "degraded":
					degradedCount++
				}
			}

			// PostgreSQL being down is critical
			for _, dep := range tt.dependencies {
				if dep.Name == "postgres" && dep.Status == "down" {
					overallStatus = "unhealthy"
					break
				}
			}

			if overallStatus != "unhealthy" {
				if downCount > 0 {
					overallStatus = "degraded"
				} else if degradedCount > 0 {
					overallStatus = "degraded"
				}
			}

			if overallStatus != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, overallStatus)
			}
		})
	}
}
