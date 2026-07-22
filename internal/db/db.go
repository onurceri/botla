package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-app/pkg/config"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

func buildDSN(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	// Use simple protocol to avoid prepared statement name collisions in connection pools.
	// NOTE: the pgx v5 connection-string parameter is "default_query_exec_mode" (not
	// "default_exec_mode"); an unrecognized key is forwarded to the server as a runtime
	// parameter and rejected with: unrecognized configuration parameter.
	base := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&default_query_exec_mode=simple_protocol", cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_NAME, cfg.DB_SSLMODE)
	if cfg.DB_SCHEMA != "" && cfg.DB_SCHEMA != "public" {
		return fmt.Sprintf("%s&options=-c%%20search_path%%3D%s", base, cfg.DB_SCHEMA)
	}
	return base
}

func New(cfg *config.Config) (*sql.DB, error) {
	dsn := buildDSN(cfg)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "open db")
	}
	// Keep MaxIdleConns close to MaxOpenConns so bursty load doesn't churn TCP
	// connections (open 25, immediately close 20 when traffic dips). The lifetime
	// settings recycle connections so stale ones silently dropped by load
	// balancers/firewalls don't accumulate.
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(15)
	conn.SetConnMaxLifetime(30 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, pkgerrors.Wrapf(err, "ping db")
	}
	return conn, nil
}
