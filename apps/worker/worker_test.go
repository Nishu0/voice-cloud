package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"voice-cloud/db"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func setup(t *testing.T) *worker {
	t.Helper()
	path := filepath.Join(t.TempDir(), "worker db.sqlite")
	command := exec.Command("goose", "-env=none", "-dir", "../../db/migrations", "sqlite3", path, "up")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("migrate: %v: %s", err, output)
	}
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return newWorker(database)
}

func deployment(t *testing.T, w *worker, name, adapter, settings string) int64 {
	t.Helper()
	ctx := context.Background()
	connection, err := w.queries.CreateConnection(ctx, db.CreateConnectionParams{
		Name: name, Adapter: adapter, LocationKind: "cloud", BaseUrl: "http://localhost:9000",
		CredentialRef: sql.NullString{String: "PROVIDER_SECRET", Valid: true}, Enabled: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	model, err := w.queries.CreateModel(ctx, db.CreateModelParams{Name: name, Publisher: "test", Task: "stt"})
	if err != nil {
		t.Fatal(err)
	}
	row, err := w.queries.CreateDeployment(ctx, db.CreateDeploymentParams{
		Name: name, ModelID: model.ID, ConnectionID: connection.ID, UpstreamModelID: "test", Enabled: 1,
		SettingsJson: settings, PricingJson: "{}", CapabilitiesJson: "{}",
	})
	if err != nil {
		t.Fatal(err)
	}
	return row.ID
}

func response(status int) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}
}

const healthSettings = `{"healthcheck":{"url":"http://localhost:9000/healthz"}}`

func TestCleanupPreservesValidCredentialsAndRouterData(t *testing.T) {
	w := setup(t)
	now := time.Now().UnixMilli()
	ctx := context.Background()
	user, err := w.queries.CreateUser(ctx, db.CreateUserParams{Email: "test@example.com", Name: "Test", PasswordHash: "encoded-hash", Role: "member"})
	if err != nil {
		t.Fatal(err)
	}
	for index, expires := range []int64{now - 1000, now + 3600000} {
		token := strings.Repeat(fmt.Sprint(index), 64)
		if err := w.queries.CreateUserSession(ctx, db.CreateUserSessionParams{TokenHash: token, UserID: user.ID, CreatedAt: now - 3600000, ExpiresAt: expires}); err != nil {
			t.Fatal(err)
		}
		if err := w.queries.CreateUserToken(ctx, db.CreateUserTokenParams{TokenHash: token, UserID: user.ID, Purpose: "password_reset", CreatedAt: now - 3600000, ExpiresAt: expires}); err != nil {
			t.Fatal(err)
		}
	}
	key, err := w.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{Name: "app", KeyHash: "hash", KeyPrefix: "prefix"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.queries.CreateRequest(ctx, db.CreateRequestParams{ID: "r1", ApiKeyID: key.ID, Task: "stt", Transport: "http", RoutingSnapshotJson: "{}", StartedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := w.cleanup(ctx); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"users", "user_sessions", "user_tokens", "api_keys", "requests"} {
		var count int
		if err := w.database.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("%s: count %d, err %v", table, count, err)
		}
	}
	if err := w.cleanup(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestProbeResultsAndNoCredentialDisclosure(t *testing.T) {
	for _, test := range []struct {
		name, adapter, settings, status, code string
		httpStatus                            int
	}{
		{"success", "custom", healthSettings, "healthy", "", 200},
		{"http_failure", "custom", healthSettings, "unhealthy", "http_status_503", 503},
		{"redirect", "custom", healthSettings, "unhealthy", "http_status_302", 302},
		{"network_failure", "custom", healthSettings, "unhealthy", "network_error", 0},
		{"timeout", "custom", healthSettings, "unhealthy", "timeout", -1},
		{"unconfigured", "custom", `{}`, "unknown", "not_configured", 0},
		{"unsupported", "openai", healthSettings, "unknown", "unsupported_probe", 0},
		{"cross_origin", "custom", `{"healthcheck":{"url":"http://other.example/health"}}`, "unknown", "invalid_config", 0},
		{"credentials", "custom", `{"healthcheck":{"url":"http://user:secret@localhost:9000/health"}}`, "unknown", "invalid_config", 0},
		{"query_secret", "custom", `{"healthcheck":{"url":"http://localhost:9000/health?key=secret"}}`, "unknown", "invalid_config", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := setup(t)
			t.Setenv("PROVIDER_SECRET", "do-not-send")
			id := deployment(t, w, "test", test.adapter, test.settings)
			calls := 0
			w.client.Timeout = 20 * time.Millisecond
			w.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Error("credentials were sent")
				}
				if test.httpStatus == -1 {
					<-r.Context().Done()
					return nil, r.Context().Err()
				}
				if test.httpStatus == 0 {
					return nil, errors.New("private upstream error")
				}
				result := response(test.httpStatus)
				if test.httpStatus == 302 {
					result.Header.Set("Location", "http://other.example/health")
				}
				return result, nil
			})
			if err := w.probe(context.Background(), id); err != nil {
				t.Fatal(err)
			}
			row, err := w.queries.GetDeploymentHealth(context.Background(), id)
			if err != nil {
				t.Fatal(err)
			}
			if row.Status != test.status || row.ErrorCode.String != test.code {
				t.Fatalf("result: %+v", row)
			}
			if row.ExpiresAt-row.CheckedAt != healthLifetime.Milliseconds() {
				t.Fatal("unexpected expiry")
			}
			if test.status == "unknown" && calls != 0 {
				t.Fatal("unconfigured/unsupported probe performed HTTP")
			}
			if test.status != "unknown" && calls != 1 {
				t.Fatal("probe followed redirect or retried")
			}
		})
	}
}

func TestProbeDiscardsChangedDisabledAndDeletedConfigurations(t *testing.T) {
	for _, change := range []string{"settings", "disabled", "deleted", "connection"} {
		t.Run(change, func(t *testing.T) {
			w := setup(t)
			id := deployment(t, w, "test", "custom", healthSettings)
			w.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				var err error
				switch change {
				case "settings":
					_, err = w.database.Exec(`UPDATE deployments SET settings_json = '{}' WHERE id = ?`, id)
				case "disabled":
					_, err = w.database.Exec(`UPDATE deployments SET enabled = 0 WHERE id = ?`, id)
				case "deleted":
					_, err = w.queries.DeleteDeployment(context.Background(), id)
				case "connection":
					_, err = w.database.Exec(`UPDATE connections SET base_url = 'http://changed.example'`)
				}
				if err != nil {
					t.Error(err)
				}
				return response(200), nil
			})
			if err := w.probe(context.Background(), id); err != nil {
				t.Fatal(err)
			}
			if _, err := w.queries.GetDeploymentHealth(context.Background(), id); !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("outdated sample persisted: %v", err)
			}
		})
	}
}

func TestProbePoolIsBoundedAndSkipsDisabledTargets(t *testing.T) {
	w := setup(t)
	for index := range 12 {
		deployment(t, w, fmt.Sprint(index), "custom", healthSettings)
	}
	disabled := deployment(t, w, "disabled", "custom", healthSettings)
	if _, err := w.database.Exec("UPDATE deployments SET enabled = 0 WHERE id = ?", disabled); err != nil {
		t.Fatal(err)
	}
	disabledConnection := deployment(t, w, "disabled-connection", "custom", healthSettings)
	if _, err := w.database.Exec("UPDATE connections SET enabled = 0 WHERE id = (SELECT connection_id FROM deployments WHERE id = ?)", disabledConnection); err != nil {
		t.Fatal(err)
	}
	gate := make(chan struct{})
	started := make(chan struct{}, 12)
	var calls, active, peak atomic.Int32
	w.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		current := active.Add(1)
		defer active.Add(-1)
		for maximum := peak.Load(); current > maximum; maximum = peak.Load() {
			if peak.CompareAndSwap(maximum, current) {
				break
			}
		}
		started <- struct{}{}
		select {
		case <-gate:
			return response(200), nil
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.probeAll(ctx) }()
	for range probeConcurrency {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("pool did not start four probes")
		}
	}
	select {
	case <-started:
		t.Fatal("pool exceeded four concurrent probes")
	case <-time.After(20 * time.Millisecond):
	}
	close(gate)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("pool did not finish")
	}
	if peak.Load() != 4 || calls.Load() != 12 {
		t.Fatalf("peak %d, calls %d", peak.Load(), calls.Load())
	}
}

func TestCancellationStopsInflightProbeWithoutWritingFailure(t *testing.T) {
	w := setup(t)
	id := deployment(t, w, "test", "custom", healthSettings)
	started := make(chan struct{})
	w.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		close(started)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.probeAll(ctx) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("probe did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not cancel probe")
	}
	if _, err := w.queries.GetDeploymentHealth(context.Background(), id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("shutdown became a health failure")
	}
}

func TestScheduleStartsImmediatelyAndNeverOverlaps(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, stopped := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	go func() {
		defer close(stopped)
		schedule(ctx, time.Millisecond, func(ctx context.Context) error {
			if calls.Add(1) == 1 {
				close(started)
			}
			<-ctx.Done()
			return ctx.Err()
		})
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not run immediately")
	}
	time.Sleep(20 * time.Millisecond)
	if calls.Load() != 1 {
		t.Fatal("scheduled runs overlapped")
	}
	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop")
	}
}

func TestScheduleRetriesAfterJobFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan struct{})
	var calls atomic.Int32
	go func() {
		defer close(finished)
		schedule(ctx, time.Millisecond, func(context.Context) error {
			if calls.Add(1) == 2 {
				cancel()
			}
			return errors.New("temporary database failure")
		})
	}()
	select {
	case <-finished:
		if calls.Load() != 2 {
			t.Fatal("scheduler did not retry the failed job")
		}
	case <-time.After(time.Second):
		t.Fatal("scheduler did not retry or shut down")
	}
}
