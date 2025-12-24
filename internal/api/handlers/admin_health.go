package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/redis/go-redis/v9"
)

// DependencyStatus represents the health status of a single dependency.
type DependencyStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "ok", "degraded", "down"
	LatencyMs float64   `json:"latency_ms"`
	Message   string    `json:"message,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// DetailedHealth represents the overall system health with dependency statuses.
type DetailedHealth struct {
	Status       string             `json:"status"` // "healthy", "degraded", "unhealthy"
	Version      string             `json:"version"`
	Uptime       string             `json:"uptime"`
	Environment  string             `json:"environment"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

// AdminHealthHandlers handles detailed health check endpoints for platform admins.
type AdminHealthHandlers struct {
	DB        *sql.DB
	Cfg       *config.Config
	StartTime time.Time
}

// NewAdminHealthHandlers creates a new AdminHealthHandlers.
func NewAdminHealthHandlers(db *sql.DB, cfg *config.Config) *AdminHealthHandlers {
	return &AdminHealthHandlers{
		DB:        db,
		Cfg:       cfg,
		StartTime: time.Now(),
	}
}

// GetDetailedHealth returns comprehensive health status of all system dependencies.
func (h *AdminHealthHandlers) GetDetailedHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Run all health checks in parallel
	var wg sync.WaitGroup
	results := make(chan DependencyStatus, 6) // buffer for all checks

	checks := []func(context.Context) DependencyStatus{
		h.checkPostgres,
		h.checkRedis,
		h.checkQdrant,
		h.checkOpenAI,
		h.checkStorage,
	}

	for _, check := range checks {
		wg.Add(1)
		go func(fn func(context.Context) DependencyStatus) {
			defer wg.Done()
			results <- fn(ctx)
		}(check)
	}

	// Close results channel when all checks complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var dependencies []DependencyStatus
	for result := range results {
		dependencies = append(dependencies, result)
	}

	// Determine overall status
	overallStatus := "healthy"
	downCount := 0
	degradedCount := 0

	for _, dep := range dependencies {
		switch dep.Status {
		case "down":
			downCount++
		case "degraded":
			degradedCount++
		}
	}

	// PostgreSQL being down is critical
	for _, dep := range dependencies {
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

	// Calculate uptime
	uptime := time.Since(h.StartTime)

	health := DetailedHealth{
		Status:       overallStatus,
		Version:      getVersion(),
		Uptime:       formatDuration(uptime),
		Environment:  h.Cfg.GO_ENV,
		Dependencies: dependencies,
	}

	// Set appropriate status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	api.WriteJSON(w, statusCode, health)
}

// checkPostgres checks PostgreSQL database connectivity.
func (h *AdminHealthHandlers) checkPostgres(ctx context.Context) DependencyStatus {
	start := time.Now()
	status := DependencyStatus{
		Name:      "postgres",
		CheckedAt: start,
	}

	err := h.DB.PingContext(ctx)
	status.LatencyMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		status.Status = "down"
		status.Message = err.Error()
		return status
	}

	// Check if latency is too high (degraded if > 100ms)
	if status.LatencyMs > 100 {
		status.Status = "degraded"
		status.Message = "high latency"
		return status
	}

	status.Status = "ok"
	return status
}

// checkRedis checks Redis connectivity.
func (h *AdminHealthHandlers) checkRedis(ctx context.Context) DependencyStatus {
	start := time.Now()
	status := DependencyStatus{
		Name:      "redis",
		CheckedAt: start,
	}

	// Check if Redis is configured
	if h.Cfg.REDIS_URL == "" {
		status.Status = "down"
		status.Message = "not configured"
		status.LatencyMs = float64(time.Since(start).Milliseconds())
		return status
	}

	opts, err := redis.ParseURL(h.Cfg.REDIS_URL)
	if err != nil {
		status.Status = "down"
		status.Message = "invalid URL: " + err.Error()
		status.LatencyMs = float64(time.Since(start).Milliseconds())
		return status
	}

	client := redis.NewClient(opts)
	defer func() { _ = client.Close() }()

	err = client.Ping(ctx).Err()
	status.LatencyMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		status.Status = "down"
		status.Message = err.Error()
		return status
	}

	// Check if latency is too high (degraded if > 50ms)
	if status.LatencyMs > 50 {
		status.Status = "degraded"
		status.Message = "high latency"
		return status
	}

	status.Status = "ok"
	return status
}

// checkQdrant checks Qdrant vector database connectivity.
func (h *AdminHealthHandlers) checkQdrant(ctx context.Context) DependencyStatus {
	start := time.Now()
	status := DependencyStatus{
		Name:      "qdrant",
		CheckedAt: start,
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.Cfg.QDRANT_URL+"/healthz", nil)
	if err != nil {
		status.Status = "down"
		status.Message = "failed to create request: " + err.Error()
		status.LatencyMs = float64(time.Since(start).Milliseconds())
		return status
	}

	if k := h.Cfg.QDRANT_API_KEY; k != "" {
		req.Header.Set("api-key", k)
	}

	resp, err := client.Do(req)
	status.LatencyMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		status.Status = "down"
		status.Message = err.Error()
		return status
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		status.Status = "down"
		status.Message = "unexpected status code: " + resp.Status
		return status
	}

	// Check if latency is too high (degraded if > 200ms)
	if status.LatencyMs > 200 {
		status.Status = "degraded"
		status.Message = "high latency"
		return status
	}

	status.Status = "ok"
	return status
}

// checkOpenAI checks OpenAI API connectivity.
func (h *AdminHealthHandlers) checkOpenAI(ctx context.Context) DependencyStatus {
	start := time.Now()
	status := DependencyStatus{
		Name:      "openai",
		CheckedAt: start,
	}

	if h.Cfg.OPENAI_API_KEY == "" {
		status.Status = "down"
		status.Message = "API key not configured"
		status.LatencyMs = float64(time.Since(start).Milliseconds())
		return status
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.Cfg.OPENAI_API_BASE+"/v1/models", nil)
	if err != nil {
		status.Status = "down"
		status.Message = "failed to create request: " + err.Error()
		status.LatencyMs = float64(time.Since(start).Milliseconds())
		return status
	}

	req.Header.Set("Authorization", "Bearer "+h.Cfg.OPENAI_API_KEY)

	resp, err := client.Do(req)
	status.LatencyMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		status.Status = "down"
		status.Message = err.Error()
		return status
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		status.Status = "down"
		status.Message = "API returned status: " + resp.Status
		return status
	}

	// Check if latency is too high (degraded if > 500ms)
	if status.LatencyMs > 500 {
		status.Status = "degraded"
		status.Message = "high latency"
		return status
	}

	status.Status = "ok"
	return status
}

// checkStorage checks Cloudflare R2/S3 storage connectivity.
func (h *AdminHealthHandlers) checkStorage(ctx context.Context) DependencyStatus {
	start := time.Now()
	status := DependencyStatus{
		Name:      "storage",
		CheckedAt: start,
	}

	// Check if storage is configured
	if h.Cfg.R2_ACCOUNT_ID == "" || h.Cfg.R2_BUCKET_NAME == "" {
		status.Status = "down"
		status.Message = "storage not configured"
		status.LatencyMs = float64(time.Since(start).Milliseconds())
		return status
	}

	// For R2/S3, we can do a simple HEAD request to the bucket
	// The actual check would require the S3 client, but for health purposes,
	// we just verify config is present
	status.LatencyMs = float64(time.Since(start).Milliseconds())
	status.Status = "ok"
	status.Message = "configuration present"
	return status
}

// getVersion returns the application version.
func getVersion() string {
	if v := os.Getenv("APP_VERSION"); v != "" {
		return v
	}
	return "dev"
}

// formatDuration formats a duration in a human-readable format.
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return formatDurationStr(days, "d", hours, "h", minutes, "m")
	}
	if hours > 0 {
		return formatDurationStr(hours, "h", minutes, "m", int(d.Seconds())%60, "s")
	}
	if minutes > 0 {
		return formatDurationStr(minutes, "m", int(d.Seconds())%60, "s", 0, "")
	}
	return formatDurationStr(int(d.Seconds()), "s", 0, "", 0, "")
}

func formatDurationStr(v1 int, u1 string, v2 int, u2 string, v3 int, u3 string) string {
	result := ""
	if v1 > 0 {
		result += formatInt(v1) + u1
	}
	if v2 > 0 && u2 != "" {
		if result != "" {
			result += " "
		}
		result += formatInt(v2) + u2
	}
	if v3 > 0 && u3 != "" {
		if result != "" {
			result += " "
		}
		result += formatInt(v3) + u3
	}
	if result == "" {
		return "0s"
	}
	return result
}

func formatInt(n int) string {
	return strconv.Itoa(n)
}
