package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"voice-cloud/db"
)

func probeURL(deployment db.Deployment, connection db.Connection) (string, string) {
	var settings struct {
		Healthcheck *struct {
			URL string `json:"url"`
		} `json:"healthcheck"`
	}
	if err := json.Unmarshal([]byte(deployment.SettingsJson), &settings); err != nil {
		return "", "invalid_config"
	}
	if settings.Healthcheck == nil {
		return "", "not_configured"
	}
	if connection.LocationKind != "local" && connection.Adapter != "custom" {
		return "", "unsupported_probe"
	}
	endpoint, err := url.Parse(settings.Healthcheck.URL)
	if err != nil {
		return "", "invalid_config"
	}
	base, err := url.Parse(connection.BaseUrl)
	if err != nil || base.User != nil || base.Hostname() == "" {
		return "", "invalid_config"
	}
	if (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Hostname() == "" ||
		endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" ||
		endpoint.Scheme != base.Scheme || !strings.EqualFold(endpoint.Host, base.Host) {
		return "", "invalid_config"
	}
	return endpoint.String(), ""
}

func (w *worker) probe(ctx context.Context, id int64) error {
	deployment, err := w.queries.GetDeployment(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	connection, err := w.queries.GetConnection(ctx, deployment.ConnectionID)
	if err != nil {
		return err
	}
	if deployment.Enabled != 1 || connection.Enabled != 1 {
		return nil
	}
	hash := db.HealthConfigHash(deployment, connection)
	endpoint, code := probeURL(deployment, connection)
	status := "unknown"
	latency := sql.NullInt64{}
	if code == "" {
		probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		request, err := http.NewRequestWithContext(probeCtx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		started := time.Now()
		response, err := w.client.Do(request)
		if ctx.Err() != nil {
			if response != nil {
				_ = response.Body.Close()
			}
			return ctx.Err()
		}
		status = "unhealthy"
		latency = sql.NullInt64{Int64: time.Since(started).Milliseconds(), Valid: true}
		if err != nil {
			code = "network_error"
			var networkError interface{ Timeout() bool }
			if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &networkError) && networkError.Timeout() {
				code = "timeout"
			}
		} else {
			_ = response.Body.Close()
			if response.StatusCode >= 200 && response.StatusCode < 300 {
				status = "healthy"
			} else {
				code = fmt.Sprintf("http_status_%d", response.StatusCode)
			}
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	now := time.Now().UnixMilli()
	result := db.SetDeploymentHealthParams{
		DeploymentID: id, ProbeKind: "reachability", Status: status, CheckedAt: now,
		ExpiresAt: now + healthLifetime.Milliseconds(), LatencyMs: latency,
		ErrorCode: sql.NullString{String: code, Valid: code != ""}, ConfigHash: hash,
	}
	// Recheck configuration in a short transaction after releasing the network request.
	tx, err := w.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := w.queries.WithTx(tx)
	current, err := queries.GetDeployment(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	currentConnection, err := queries.GetConnection(ctx, current.ConnectionID)
	if err != nil {
		return err
	}
	if current.Enabled != 1 || currentConnection.Enabled != 1 || db.HealthConfigHash(current, currentConnection) != hash {
		return nil
	}
	if _, err := queries.SetDeploymentHealth(ctx, result); err != nil {
		return err
	}
	return tx.Commit()
}
