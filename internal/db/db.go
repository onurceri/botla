package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/pkg/config"
)

func New(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_NAME)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
