package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Open requires an existing database with the current migrations applied.
func Open(path string) (*sql.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	uri := url.URL{Scheme: "file", Path: absolute}
	uri.RawQuery = url.Values{"mode": {"rw"}, "_pragma": {
		"foreign_keys(1)", "busy_timeout(5000)", "journal_mode(WAL)",
	}}.Encode()
	database, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var count int
	if err := database.QueryRowContext(ctx, "SELECT count(*) FROM deployment_health").Scan(&count); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("open migrated database (run make db-up first): %w", err)
	}
	return database, nil
}
