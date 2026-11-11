package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func openDatabase(path string) (*sql.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	uri := url.URL{Scheme: "file", Path: absolute}
	options := url.Values{"mode": {"rw"}, "_pragma": {"foreign_keys(1)", "busy_timeout(5000)", "journal_mode(WAL)"}}
	uri.RawQuery = options.Encode()
	database, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var count int
	if err = database.QueryRowContext(ctx, "SELECT count(*) FROM deployments").Scan(&count); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("open migrated database (run make db-up first): %w", err)
	}
	return database, nil
}

func run(ctx context.Context) error {
	token := os.Getenv("ADMIN_TOKEN")
	if len(strings.TrimSpace(token)) < 32 {
		return errors.New("ADMIN_TOKEN must contain at least 32 characters")
	}
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "data/voice-cloud.db"
	}
	database, err := openDatabase(path)
	if err != nil {
		return err
	}
	defer database.Close()
	address := os.Getenv("API_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	server := &http.Server{Addr: address, Handler: newHandler(database, token), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	stopped := make(chan error, 1)
	go func() { stopped <- server.ListenAndServe() }()
	slog.Info("API listening", "address", address)
	select {
	case err := <-stopped:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}
		if err := <-stopped; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("API stopped", "error", err)
		os.Exit(1)
	}
}
