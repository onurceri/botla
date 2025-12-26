package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/config"
)

func buildDSN(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	base := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_NAME, cfg.DB_SSLMODE)
	if cfg.DB_SCHEMA != "" && cfg.DB_SCHEMA != "public" {
		return fmt.Sprintf("%s&options=-c%%20search_path%%3D%s", base, cfg.DB_SCHEMA)
	}
	return base
}

func New(cfg *config.Config) (*sql.DB, error) {
	dsn := buildDSN(cfg)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return conn, nil
}
