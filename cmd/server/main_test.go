package main

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/logger"
)

func TestServerStartAndShutdown(t *testing.T) {
	log := logger.New("INFO")
	srv := newHTTPServer("0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	startServerAsync(srv, log, "0")
	time.Sleep(100 * time.Millisecond)
	db, err := sql.Open("pgx", "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	shutdownServer(srv, log, db)
}
