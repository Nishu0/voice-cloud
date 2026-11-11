package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"voice-cloud/db"
)

func run(ctx context.Context) error {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "data/voice-cloud.db"
	}
	database, err := db.Open(path)
	if err != nil {
		return err
	}
	defer database.Close()
	slog.Info("worker started")
	newWorker(database).run(ctx)
	slog.Info("worker stopped")
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
