package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/onurceri/botla-co/pkg/config"
)

type HealthHandlers struct {
	DB  *sql.DB
	Cfg *config.Config
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
	status := http.StatusOK
	if dep["db"] != "ok" || dep["qdrant"] != "ok" {
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "dependencies": dep})
}

func qdrantHealthy(ctx context.Context, cfg *config.Config) error {
	client := &http.Client{Timeout: 2 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.QDRANT_URL+"/healthz", nil)
	if k := os.Getenv("QDRANT_API_KEY"); k != "" {
		req.Header.Set("api-key", k)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return http.ErrHandlerTimeout
	}
	return nil
}
