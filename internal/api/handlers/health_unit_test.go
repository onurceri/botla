package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/config"
)

func TestHealth_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer srv.Close()
	db, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	h := &HealthHandlers{DB: db, Cfg: &config.Config{QDRANT_URL: srv.URL}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.Health(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestHealth_Degraded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer srv.Close()
	db, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	h := &HealthHandlers{DB: db, Cfg: &config.Config{QDRANT_URL: srv.URL}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.Health(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: %d", rr.Code)
	}
}
