package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/config"
)

type HealthHandlers struct {
	DB         *sql.DB
	Cfg        *config.Config
	Queue      *processing.SourceQueue
	LLMFactory *rag.ClientFactory
}

func (h *HealthHandlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	dep := map[string]any{}
	if err := h.DB.PingContext(ctx); err != nil {
		dep["db"] = "down"
	} else {
		dep["db"] = "ok"
	}
	if err := qdrantHealthy(ctx, h.Cfg); err != nil {
		dep["qdrant"] = "down"
	} else {
		dep["qdrant"] = "ok"
	}
	
	queueStats := map[string]any{
		"status":  "ok",
		"workers": 0,
		"pending": 0,
	}
	if h.Queue != nil {
		queueStats["workers"] = h.Queue.WorkerCount()
		queueStats["pending"] = h.Queue.QueueLength()
	}
	dep["queue"] = queueStats

	// Add LLM circuit breaker status
	if h.LLMFactory != nil {
		llmStatus := h.LLMFactory.GetCircuitBreakerStatus()
		if len(llmStatus) > 0 {
			dep["llm"] = llmStatus
		}
	}

	status := http.StatusOK
	if dep["db"] != "ok" || dep["qdrant"] != "ok" {
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{"status": "ok", "dependencies": dep}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func qdrantHealthy(ctx context.Context, cfg *config.Config) error {
	client := &http.Client{Timeout: 2 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.QDRANT_URL+"/healthz", nil)
	if k := os.Getenv("QDRANT_API_KEY"); k != "" {
		req.Header.Set("api-key", k)
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant health check request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return http.ErrHandlerTimeout
	}
	return nil
}
