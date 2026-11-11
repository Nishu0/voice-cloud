package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"voice-cloud/db"
)

func (a *api) deploymentHealth(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := a.queries.WithTx(tx)
	deployment, err := queries.GetDeployment(r.Context(), id)
	if err != nil {
		return err
	}
	connection, err := queries.GetConnection(r.Context(), deployment.ConnectionID)
	if err != nil {
		return err
	}
	row, err := queries.GetDeploymentHealth(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	value := map[string]any{
		"deployment_id": id, "probe_kind": "reachability", "status": "unknown",
		"checked_at": nil, "expires_at": nil, "latency_ms": nil, "error_code": "not_checked",
	}
	switch {
	case deployment.Enabled != 1 || connection.Enabled != 1:
		value["error_code"] = "disabled"
	case errors.Is(err, sql.ErrNoRows):
	case row.ConfigHash != db.HealthConfigHash(deployment, connection):
		value["error_code"] = "config_changed"
	default:
		value["checked_at"], value["expires_at"] = row.CheckedAt, row.ExpiresAt
		value["latency_ms"], value["error_code"] = intValue(row.LatencyMs), stringValue(row.ErrorCode)
		if row.ExpiresAt <= time.Now().UnixMilli() {
			value["error_code"] = "stale"
		} else {
			value["status"] = row.Status
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return send(w, http.StatusOK, value)
}
