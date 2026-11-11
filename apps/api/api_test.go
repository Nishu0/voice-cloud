package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"voice-cloud/db"
)

const testToken = "test-admin-token-at-least-thirty-two-characters"

type testAPI struct {
	t       *testing.T
	handler http.Handler
	db      *sql.DB
}

func setup(t *testing.T) *testAPI {
	t.Helper()
	path := filepath.Join(t.TempDir(), "router db.sqlite")
	command := exec.Command("goose", "-env=none", "-dir", "../../db/migrations", "sqlite3", path, "up")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("migrate: %v: %s", err, output)
	}
	database, err := openDatabase(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return &testAPI{t, newHandler(database, testToken), database}
}

func (a *testAPI) call(method, path, body string, status int) *httptest.ResponseRecorder {
	a.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+testToken)
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	a.handler.ServeHTTP(w, r)
	if w.Code != status {
		a.t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, status, w.Body.String())
	}
	return w
}

func jsonObject(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func seedConfig(a *testAPI) {
	a.t.Helper()
	a.call("POST", "/v1/connections", `{"name":"Local","adapter":"whisper","location_kind":"local","base_url":"http://localhost:9000"}`, 201)
	a.call("POST", "/v1/models", `{"name":"large-v3","publisher":"openai","task":"stt"}`, 201)
	a.call("POST", "/v1/deployments", `{"name":"Local Whisper","model_id":1,"connection_id":1,"upstream_model_id":"large-v3"}`, 201)
	a.call("POST", "/v1/policies", `{"name":"Speech","task":"stt"}`, 201)
}

func TestConfigurationWorkflow(t *testing.T) {
	a := setup(t)
	for _, path := range []string{"connections", "models", "deployments", "policies", "api-keys", "requests"} {
		response := a.call("GET", "/v1/"+path, "", 200)
		if strings.TrimSpace(response.Body.String()) != "[]" {
			t.Fatalf("empty list: %s", response.Body.String())
		}
	}
	seedConfig(a)
	for _, path := range []string{"connections", "models", "deployments", "policies"} {
		a.call("GET", "/v1/"+path+"/1", "", 200)
		a.call("GET", "/v1/"+path+"/999", "", 404)
	}
	response := jsonObject(t, a.call("GET", "/v1/connections/1", "", 200))
	if response["enabled"] != true || response["credential_ref"] != nil {
		t.Fatalf("connection JSON: %#v", response)
	}
	response = jsonObject(t, a.call("GET", "/v1/deployments/1", "", 200))
	if _, ok := response["settings_json"].(map[string]any); !ok {
		t.Fatalf("settings must be an object: %#v", response)
	}
	a.call("PUT", "/v1/connections/1", `{"name":"Updated","adapter":"whisper","location_kind":"local","base_url":"http://localhost:9000","enabled":false,"credential_ref":"WHISPER_TOKEN"}`, 200)
	a.call("PUT", "/v1/models/1", `{"name":"Whisper","publisher":"openai","task":"stt","version":"v3"}`, 200)
	a.call("PUT", "/v1/deployments/1", `{"name":"Updated","model_id":1,"connection_id":1,"upstream_model_id":"large-v3","enabled":false,"settings_json":{"beam_size":5}}`, 200)
	a.call("PUT", "/v1/policies/1", `{"name":"Updated","task":"stt","objective":"latency","fallback_enabled":false,"max_attempts":1}`, 200)
	a.call("PUT", "/v1/deployments/1/languages/en-US", "", 200)
	a.call("PUT", "/v1/deployments/1/languages/en-US", "", 200)
	a.call("GET", "/v1/deployments/1/languages", "", 200)
	a.call("PUT", "/v1/policies/1/targets/1", `{"priority":2}`, 200)
	a.call("PUT", "/v1/policies/1/targets/1", `{"priority":1}`, 200)
	a.call("GET", "/v1/policies/1/targets", "", 200)
	a.call("DELETE", "/v1/models/1", "", 409)
	a.call("DELETE", "/v1/connections/1", "", 409)
	a.call("DELETE", "/v1/deployments/1", "", 409)
	a.call("DELETE", "/v1/policies/1/targets/1", "", 204)
	a.call("DELETE", "/v1/deployments/1/languages/en-US", "", 204)
	for _, path := range []string{"policies", "deployments", "models", "connections"} {
		a.call("DELETE", "/v1/"+path+"/1", "", 204)
	}
	a.call("DELETE", "/v1/connections/1", "", 404)
}

func TestAuthorizationAndKeys(t *testing.T) {
	a := setup(t)
	r := httptest.NewRequest("GET", "/v1/connections", nil)
	w := httptest.NewRecorder()
	a.handler.ServeHTTP(w, r)
	if w.Code != 401 || w.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("missing authorization: %d", w.Code)
	}
	a.call("GET", "/healthz", "", 200)
	created := jsonObject(t, a.call("POST", "/v1/api-keys", `{"name":"App"}`, 201))
	token, ok := created["key"].(string)
	if !ok || !strings.HasPrefix(token, "vc_") || len(token) != 67 {
		t.Fatalf("bad token: %#v", created)
	}
	var stored string
	if err := a.db.QueryRow("SELECT key_hash FROM api_keys WHERE id=1").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(token))
	if stored != hex.EncodeToString(hash[:]) {
		t.Fatal("key was not hashed")
	}
	for _, path := range []string{"/v1/api-keys", "/v1/api-keys/1"} {
		body := a.call("GET", path, "", 200).Body.String()
		if strings.Contains(body, token) || strings.Contains(body, stored) || strings.Contains(body, "key_hash") {
			t.Fatal("key material leaked")
		}
	}
	r.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	a.handler.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("router key gained admin access")
	}
	a.call("DELETE", "/v1/api-keys/1", "", 204)
	a.call("DELETE", "/v1/api-keys/1", "", 204)
	var revoked sql.NullInt64
	if err := a.db.QueryRow("SELECT revoked_at FROM api_keys WHERE id=1").Scan(&revoked); err != nil || !revoked.Valid {
		t.Fatalf("key was not revoked: %v", err)
	}
	_, err := db.New(a.db).GetActiveAPIKeyByHash(context.Background(), db.GetActiveAPIKeyByHashParams{KeyHash: stored, NowMs: sql.NullInt64{Int64: revoked.Int64, Valid: true}})
	if err != sql.ErrNoRows {
		t.Fatalf("revoked key remains valid: %v", err)
	}
}

func TestInvalidInputAndRelationships(t *testing.T) {
	a := setup(t)
	seedConfig(a)
	for _, test := range []struct {
		path, body string
		status     int
	}{
		{"connections", `{}`, 400},
		{"connections", `null`, 400},
		{"connections", `[]`, 400},
		{"connections", `{"name":"Bad","unknown":true}`, 400},
		{"connections", `{} {}`, 400},
		{"connections", `{"name":"Bad","adapter":"a","location_kind":"local","base_url":"file:///tmp"}`, 400},
		{"connections", `{"name":"Bad","adapter":"a","location_kind":"local","base_url":"https://user:secret@example.com"}`, 400},
		{"models", `{"name":"Bad","publisher":"p","task":"invalid"}`, 400},
		{"deployments", `{"name":"Bad","model_id":1,"connection_id":1,"upstream_model_id":"x","settings_json":[]}`, 400},
		{"deployments", `{"name":"Missing","model_id":999,"connection_id":1,"upstream_model_id":"x"}`, 409},
		{"policies", `{"name":"Bad","task":"stt","max_attempts":0}`, 400},
		{"api-keys", `{"name":"Bad","expires_at":1}`, 400},
	} {
		a.call("POST", "/v1/"+test.path, test.body, test.status)
	}
	a.call("GET", "/v1/models/0", "", 400)
	a.call("GET", "/v1/models/nope", "", 400)
	a.call("GET", "/v1/requests?limit=101", "", 400)
	a.call("GET", "/v1/requests?offset=-1", "", 400)
	a.call("PUT", "/v1/deployments/1/languages/invalid_tag", "", 400)
	a.call("GET", "/v1/deployments/999/languages", "", 404)
	a.call("PUT", "/v1/models/1", `{"name":"Whisper","publisher":"openai","task":"tts"}`, 409)
	a.call("PUT", "/v1/policies/1", `{"name":"Speech","task":"tts"}`, 409)
	a.call("PUT", "/v1/deployments/1", `{"name":"Local","model_id":1,"connection_id":1,"upstream_model_id":"different"}`, 409)
	a.call("POST", "/v1/models", `{"name":"TTS","publisher":"other","task":"tts"}`, 201)
	a.call("POST", "/v1/deployments", `{"name":"TTS","model_id":2,"connection_id":1,"upstream_model_id":"tts"}`, 201)
	a.call("PUT", "/v1/policies/1/targets/2", `{"priority":0}`, 400)
	a.call("PUT", "/v1/policies/1/targets/1", `null`, 400)
	if body := a.call("GET", "/v1/policies/1/targets", "", 200).Body.String(); strings.TrimSpace(body) != "[]" {
		t.Fatalf("invalid target was persisted: %s", body)
	}
	r := httptest.NewRequest("POST", "/v1/models", strings.NewReader(`{}`))
	r.Header.Set("Authorization", "Bearer "+testToken)
	w := httptest.NewRecorder()
	a.handler.ServeHTTP(w, r)
	if w.Code != 415 {
		t.Fatalf("missing content type: %d", w.Code)
	}
	a.call("POST", "/v1/models", fmt.Sprintf(`{"name":%q}`, strings.Repeat("x", (1<<20)+1)), 413)
}

func TestActivityAndDatabaseSettings(t *testing.T) {
	a := setup(t)
	seedConfig(a)
	queries := db.New(a.db)
	ctx := context.Background()
	key, err := queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{Name: "app", KeyHash: "hash", KeyPrefix: "prefix"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateRequest(ctx, db.CreateRequestParams{ID: "r1", ApiKeyID: key.ID, Task: "stt", Transport: "websocket", RoutingSnapshotJson: `{"mode":"explicit"}`, StartedAt: 1000})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAttempt(ctx, db.CreateAttemptParams{ID: "a1", RequestID: "r1", DeploymentID: 1, AttemptNumber: 1, StartedAt: 1000, SelectionReason: "explicit"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.CreateAttemptUsage(ctx, db.CreateAttemptUsageParams{AttemptID: "a1", Metric: "input_audio", Quantity: 500, Unit: "audio_ms", RateSnapshotJson: `{}`, Currency: "USD", CostStatus: "unknown"})
	if err != nil {
		t.Fatal(err)
	}
	response := jsonObject(t, a.call("GET", "/v1/requests/r1", "", 200))
	if response["finished_at"] != nil || response["policy_id"] != nil {
		t.Fatalf("nullable fields: %#v", response)
	}
	if _, ok := response["routing_snapshot_json"].(map[string]any); !ok {
		t.Fatal("snapshot is not a JSON object")
	}
	a.call("GET", "/v1/requests?limit=1&offset=0", "", 200)
	a.call("GET", "/v1/requests/r1/attempts", "", 200)
	body := a.call("GET", "/v1/requests/r1/usage", "", 200).Body.String()
	if !strings.Contains(body, `"cost_nanos":null`) {
		t.Fatalf("unknown cost must be null: %s", body)
	}
	a.call("GET", "/v1/requests/missing/attempts", "", 404)
	a.call("GET", "/v1/requests/missing/usage", "", 404)
	a.call("POST", "/v1/requests", `{}`, 405)
	a.call("DELETE", "/v1/connections/1", "", 409)
	var foreignKeys, timeout int
	var mode string
	for pragma, target := range map[string]any{"foreign_keys": &foreignKeys, "busy_timeout": &timeout, "journal_mode": &mode} {
		if err := a.db.QueryRow("PRAGMA " + pragma).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if foreignKeys != 1 || timeout != 5000 || mode != "wal" {
		t.Fatalf("pragmas: %d, %d, %s", foreignKeys, timeout, mode)
	}
	// Reopening the physical connection must preserve its PRAGMA configuration.
	a.db.SetMaxIdleConns(0)
	if err := a.db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil || foreignKeys != 1 {
		t.Fatalf("foreign keys lost on reconnect: %v", err)
	}
}

func TestStartupRejectsMissingDatabase(t *testing.T) {
	if database, err := openDatabase(filepath.Join(t.TempDir(), "missing.db")); err == nil {
		_ = database.Close()
		t.Fatal("missing database should fail")
	}
	t.Setenv("ADMIN_TOKEN", "")
	if err := run(context.Background()); err == nil {
		t.Fatal("missing admin token should fail")
	}
}

func TestCustomAndLocalConnections(t *testing.T) {
	a := setup(t)
	for index, provider := range []struct{ adapter, location, url string }{
		{"custom", "cloud", "https://models.example/v1"},
		{"local", "local", "http://127.0.0.1:8081/v1"},
	} {
		t.Run(provider.adapter, func(t *testing.T) {
			connection := jsonObject(t, a.call("POST", "/v1/connections", fmt.Sprintf(
				`{"name":%q,"adapter":%q,"location_kind":%q,"base_url":%q,"enabled":false}`,
				"My "+provider.adapter+" endpoint", provider.adapter, provider.location, provider.url), 201))
			model := jsonObject(t, a.call("POST", "/v1/models", fmt.Sprintf(
				`{"name":%q,"publisher":"self","task":"llm"}`, "my-model-"+provider.adapter), 201))
			deployment := jsonObject(t, a.call("POST", "/v1/deployments", fmt.Sprintf(
				`{"name":%q,"model_id":%.0f,"connection_id":%.0f,"upstream_model_id":"my-model","enabled":false}`,
				"My "+provider.adapter+" deployment", model["id"], connection["id"]), 201))
			if deployment["connection_id"] != float64(index+1) || deployment["enabled"] != false {
				t.Fatalf("custom/local deployment was not saved: %#v", deployment)
			}
		})
	}
}
