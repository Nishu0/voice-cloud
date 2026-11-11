package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"voice-cloud/db"
)

func keyJSON(row db.ApiKey) any {
	return map[string]any{"id": row.ID, "name": row.Name, "key_prefix": row.KeyPrefix,
		"created_by_user_id": intValue(row.CreatedByUserID),
		"default_policy_id":  intValue(row.DefaultPolicyID), "expires_at": intValue(row.ExpiresAt),
		"revoked_at": intValue(row.RevokedAt), "created_at": row.CreatedAt, "updated_at": row.UpdatedAt}
}

func (a *api) keys(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		if id == 0 {
			rows, err := a.queries.ListAPIKeys(ctx)
			return result(w, 200, mapItems(rows, keyJSON), err)
		}
		row, err := a.queries.GetAPIKey(ctx, id)
		return result(w, 200, keyJSON(row), err)
	case http.MethodDelete:
		if _, err := a.queries.GetAPIKey(ctx, id); err != nil {
			return err
		}
		_, err := a.queries.RevokeAPIKey(ctx, db.RevokeAPIKeyParams{ID: id, NowMs: sql.NullInt64{Int64: time.Now().UnixMilli(), Valid: true}})
		if err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	default:
		var body struct {
			Name            string `json:"name"`
			DefaultPolicyID *int64 `json:"default_policy_id"`
			ExpiresAt       *int64 `json:"expires_at"`
		}
		if err := decode(w, r, &body); err != nil {
			return err
		}
		if err := required(body.Name); err != nil {
			return err
		}
		if body.DefaultPolicyID != nil && *body.DefaultPolicyID < 1 {
			return invalid("default_policy_id must be positive")
		}
		if body.ExpiresAt != nil && *body.ExpiresAt <= time.Now().UnixMilli() {
			return invalid("expires_at must be a future Unix millisecond timestamp")
		}
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return err
		}
		token := "vc_" + hex.EncodeToString(secret)
		hash := sha256.Sum256([]byte(token))
		row, err := a.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{Name: body.Name, KeyHash: hex.EncodeToString(hash[:]), KeyPrefix: token[:11], DefaultPolicyID: nullableInt(body.DefaultPolicyID), ExpiresAt: nullableInt(body.ExpiresAt)})
		return result(w, 201, map[string]any{"key": token, "api_key": keyJSON(row)}, err)
	}
}
