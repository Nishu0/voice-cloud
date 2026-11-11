package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func authCall(t *testing.T, a *testAPI, token, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	a.handler.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, status, w.Body.String())
	}
	return w
}
func TestAccountLifecycle(t *testing.T) {
	a := setup(t)
	a.handler = newHandler(a.db, "")
	authCall(t, a, "", "GET", "/v1/models", "", 401)
	status := jsonObject(t, authCall(t, a, "", "GET", "/v1/auth/status", "", 200))
	if status["needs_setup"] != true {
		t.Fatal(status)
	}
	for _, body := range []string{`{"name":"Admin","email":"invalid","password":"a-long-password"}`, `{"name":"Admin","email":"admin@example.com","password":"short"}`, `{"name":"Admin","email":"admin@example.com","password":"a-long-password","role":"admin"}`} {
		authCall(t, a, "", "POST", "/v1/auth/signup", body, 400)
	}
	created := jsonObject(t, authCall(t, a, "", "POST", "/v1/auth/signup", `{"name":"Admin","email":" Admin@Example.com ","password":"a-long-password"}`, 201))
	token := created["token"].(string)
	user := created["user"].(map[string]any)
	if user["role"] != "admin" || user["email"] != "admin@example.com" {
		t.Fatal(user)
	}
	var password, hash string
	_ = a.db.QueryRow("SELECT password_hash FROM users").Scan(&password)
	_ = a.db.QueryRow("SELECT token_hash FROM user_sessions").Scan(&hash)
	if password == "a-long-password" || !passwordMatches("a-long-password", password) || hash == token || hash != tokenHash(token) {
		t.Fatal("credentials stored incorrectly")
	}
	if strings.Contains(string(mustJSON(user)), "password") {
		t.Fatal("password leaked")
	}
	authCall(t, a, "", "POST", "/v1/auth/signup", `{"name":"Other","email":"other@example.com","password":"a-long-password"}`, 409)
	authCall(t, a, token, "GET", "/v1/auth/me", "", 200)
	authCall(t, a, token, "PUT", "/v1/auth/me", `{"name":"Renamed"}`, 200)
	authCall(t, a, "", "POST", "/v1/auth/login", `{"email":"admin@example.com","password":"wrong"}`, 401)
	authCall(t, a, "", "POST", "/v1/auth/login", `{"email":"unknown@example.com","password":"wrong"}`, 401)
	signedIn := jsonObject(t, authCall(t, a, "", "POST", "/v1/auth/login", `{"email":"ADMIN@example.com","password":"a-long-password"}`, 200))
	second := signedIn["token"].(string)
	key := jsonObject(t, authCall(t, a, token, "POST", "/v1/api-keys", `{"name":"App"}`, 201))["api_key"].(map[string]any)
	if key["created_by_user_id"] != user["id"] {
		t.Fatal(key)
	}
	authCall(t, a, token, "POST", "/v1/users", `{"name":"Member","email":"member@example.com","password":"member-password"}`, 201)
	member := jsonObject(t, authCall(t, a, "", "POST", "/v1/auth/login", `{"email":"member@example.com","password":"member-password"}`, 200))["token"].(string)
	authCall(t, a, member, "GET", "/v1/models", "", 200)
	authCall(t, a, member, "POST", "/v1/models", `{"name":"x","publisher":"x","task":"llm"}`, 403)
	authCall(t, a, member, "POST", "/v1/users", `{}`, 403)
	authCall(t, a, member, "PUT", "/v1/auth/me", `{"name":"Updated"}`, 200)
	authCall(t, a, token, "POST", "/v1/auth/logout", "", 204)
	authCall(t, a, token, "GET", "/v1/auth/me", "", 401)
	// A session survives reopening the handler, but disabled users and expired tokens cannot use it.
	a.handler = newHandler(a.db, "")
	authCall(t, a, second, "GET", "/v1/models", "", 200)
	_, _ = a.db.Exec("UPDATE users SET status='disabled' WHERE email='admin@example.com'")
	authCall(t, a, second, "GET", "/v1/models", "", 401)
	authCall(t, a, "", "POST", "/v1/auth/login", `{"email":"admin@example.com","password":"a-long-password"}`, 401)
	_, _ = a.db.Exec("UPDATE user_sessions SET expires_at=created_at+1 WHERE token_hash=?", tokenHash(member))
	time.Sleep(2 * time.Millisecond)
	authCall(t, a, member, "GET", "/v1/models", "", 401)
}
func mustJSON(value any) []byte { body, _ := json.Marshal(value); return body }
func TestInitialAccountIsAtomic(t *testing.T) {
	a := setup(t)
	a.handler = newHandler(a.db, "")
	results := make(chan int, 2)
	for range 2 {
		go func() {
			r := httptest.NewRequest("POST", "/v1/auth/signup", strings.NewReader(`{"name":"First","email":"first@example.com","password":"a-long-password"}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			a.handler.ServeHTTP(w, r)
			results <- w.Code
		}()
	}
	first, second := <-results, <-results
	if first != 201 && second != 201 {
		t.Fatalf("neither account created: %d %d", first, second)
	}
	var count int
	_ = a.db.QueryRow("SELECT count(*) FROM users WHERE role='admin'").Scan(&count)
	if count != 1 {
		t.Fatal(count)
	}
}
func TestLoginRateLimit(t *testing.T) {
	a := setup(t)
	a.handler = newHandler(a.db, "")
	for range 20 {
		authCall(t, a, "", "POST", "/v1/auth/login", `{}`, 401)
	}
	authCall(t, a, "", "POST", "/v1/auth/login", `{}`, 429)
}
func TestPasswordHashValidation(t *testing.T) {
	hash, err := passwordHash("my-password-is-long")
	if err != nil {
		t.Fatal(err)
	}
	if passwordMatches("wrong", hash) || passwordMatches("x", "bad") || passwordMatches("x", "pbkdf2-sha256$600000$bad$bad") {
		t.Fatal("invalid password accepted")
	}
}
