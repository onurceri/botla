package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/onurceri/botla-app/pkg/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	email := flag.String("email", "", "Email of the user to make admin")
	makeAdmin := flag.Bool("make-admin", true, "Whether to make the user an admin or not")
	flag.Parse()

	if *email == "" {
		flag.Usage()
		return fmt.Errorf("email is required")
	}

	cfg := config.LoadConfig()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
			cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_NAME, cfg.DB_SSLMODE, cfg.DB_SCHEMA)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)", *email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("user with email %s not found", *email)
	}

	_, err = db.Exec("UPDATE users SET is_platform_admin = $1 WHERE email = $2 AND deleted_at IS NULL", *makeAdmin, *email)
	if err != nil {
		return fmt.Errorf("failed to update user admin status: %w", err)
	}

	status := "is now a platform admin"
	if !*makeAdmin {
		status = "is no longer a platform admin"
	}
	fmt.Printf("Success: User %s %s\n", *email, status)
	return nil
}
