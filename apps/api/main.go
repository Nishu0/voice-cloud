package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"voice-cloud/db"
)

func run(ctx context.Context) error {
	token := os.Getenv("ADMIN_TOKEN")
	if token != "" && len(strings.TrimSpace(token)) < 32 {
		return errors.New("ADMIN_TOKEN must contain at least 32 characters")
	}
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "data/voice-cloud.db"
	}
	database, err := db.Open(path)
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
