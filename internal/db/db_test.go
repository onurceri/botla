package db

import (
	"testing"

	"github.com/onurceri/botla-co/pkg/config"
)

func TestBuildDSN(t *testing.T) {
	cfg := &config.Config{
		DB_USER:     "u",
		DB_PASSWORD: "p",
		DB_HOST:     "h",
		DB_PORT:     "5432",
		DB_NAME:     "n",
	}
	dsn := buildDSN(cfg)
	expected := "postgres://u:p@h:5432/n?sslmode=disable"
	if dsn != expected {
		t.Fatalf("dsn mismatch: got %q want %q", dsn, expected)
	}
}
