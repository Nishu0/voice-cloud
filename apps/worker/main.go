package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// shortcut: no background jobs exist yet; add job processing when the first task is implemented.
	slog.Info("worker started; no background jobs configured")
	<-ctx.Done()
	slog.Info("worker stopped")
}
