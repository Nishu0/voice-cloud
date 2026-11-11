package main

import (
	"encoding/json"
	"net/http"

	"voice-cloud/db"
)

func requestJSON(row db.Request) any {
	return struct {
		db.Request
		LanguageTag        any             `json:"language_tag"`
		PolicyID           any             `json:"policy_id"`
		RoutingSnapshot    json.RawMessage `json:"routing_snapshot_json"`
		FinishedAt         any             `json:"finished_at"`
		ErrorCode          any             `json:"error_code"`
		TraceID            any             `json:"trace_id"`
		IdempotencyKey     any             `json:"idempotency_key"`
		RequestFingerprint any             `json:"request_fingerprint"`
	}{row, stringValue(row.LanguageTag), intValue(row.PolicyID), json.RawMessage(row.RoutingSnapshotJson), intValue(row.FinishedAt), stringValue(row.ErrorCode), stringValue(row.TraceID), stringValue(row.IdempotencyKey), stringValue(row.RequestFingerprint)}
}

func attemptJSON(row db.Attempt) any {
	return struct {
		db.Attempt
		ProviderRequestID any `json:"provider_request_id"`
		FirstOutputAt     any `json:"first_output_at"`
		FinishedAt        any `json:"finished_at"`
		ErrorCode         any `json:"error_code"`
	}{row, stringValue(row.ProviderRequestID), intValue(row.FirstOutputAt), intValue(row.FinishedAt), stringValue(row.ErrorCode)}
}

func usageJSON(row db.AttemptUsage) any {
	return struct {
		db.AttemptUsage
		RateSnapshot json.RawMessage `json:"rate_snapshot_json"`
		CostNanos    any             `json:"cost_nanos"`
	}{row, json.RawMessage(row.RateSnapshotJson), intValue(row.CostNanos)}
}

func (a *api) requests(w http.ResponseWriter, r *http.Request) error {
	if id := r.PathValue("request"); id != "" {
		row, err := a.queries.GetRequest(r.Context(), id)
		return result(w, 200, requestJSON(row), err)
	}
	params, err := pagination(r)
	if err != nil {
		return err
	}
	rows, err := a.queries.ListRequests(r.Context(), params)
	return result(w, 200, mapItems(rows, requestJSON), err)
}
func (a *api) requestAttempts(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("request")
	if _, err := a.queries.GetRequest(r.Context(), id); err != nil {
		return err
	}
	rows, err := a.queries.ListRequestAttempts(r.Context(), id)
	return result(w, 200, mapItems(rows, attemptJSON), err)
}
func (a *api) requestUsage(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("request")
	if _, err := a.queries.GetRequest(r.Context(), id); err != nil {
		return err
	}
	rows, err := a.queries.ListRequestUsage(r.Context(), id)
	return result(w, 200, mapItems(rows, usageJSON), err)
}
