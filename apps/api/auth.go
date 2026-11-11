package main

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"voice-cloud/db"
)

const passwordIterations = 600000
const sessionLifetime = 7 * 24 * time.Hour

type userContextKey struct{}

type loginWindow struct {
	count int
	until time.Time
}
type loginLimiter struct {
	sync.Mutex
	windows map[string]loginWindow
}

func (l *loginLimiter) allow(address string) bool {
	l.Lock()
	defer l.Unlock()
	now := time.Now()
	if l.windows == nil {
		l.windows = make(map[string]loginWindow)
	}
	for key, value := range l.windows {
		if !now.Before(value.until) {
			delete(l.windows, key)
		}
	}
	value := l.windows[address]
	if value.count >= 20 || len(l.windows) >= 4096 && value.count == 0 {
		return false
	}
	if value.count == 0 {
		value.until = now.Add(time.Minute)
	}
	value.count++
	l.windows[address] = value
	return true
}
func (a *api) allowLogin(r *http.Request) error {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !a.logins.allow(ip) {
		return httpError{429, "too many sign-in attempts; retry in a minute"}
	}
	return nil
}

func passwordHash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	if err != nil {
		return "", err
	}
	return "pbkdf2-sha256$600000$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}
func passwordMatches(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" || parts[1] != "600000" {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) != 16 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(want) != 32 {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}
func userJSON(user db.User) any {
	return map[string]any{"id": user.ID, "email": user.Email, "name": user.Name, "role": user.Role, "status": user.Status, "created_at": user.CreatedAt}
}
func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
func newSession(ctx context.Context, queries *db.Queries, user db.User) (any, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	token := "vcs_" + hex.EncodeToString(secret)
	now := time.Now()
	expires := now.Add(sessionLifetime).UnixMilli()
	if err := queries.CreateUserSession(ctx, db.CreateUserSessionParams{TokenHash: tokenHash(token), UserID: user.ID, CreatedAt: now.UnixMilli(), ExpiresAt: expires}); err != nil {
		return nil, err
	}
	return map[string]any{"token": token, "expires_at": expires, "user": userJSON(user)}, nil
}

type signupBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func validateAccount(body *signupBody) error {
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	body.Name = strings.TrimSpace(body.Name)
	address, err := mail.ParseAddress(body.Email)
	if err != nil || address.Address != body.Email || len(body.Email) > 254 {
		return invalid("enter a valid email address")
	}
	if body.Name == "" || len(body.Name) > 100 {
		return invalid("name must contain 1 to 100 characters")
	}
	if len(body.Password) < 12 || len(body.Password) > 1024 {
		return invalid("password must contain 12 to 1024 bytes")
	}
	return nil
}
func (a *api) authStatus(w http.ResponseWriter, r *http.Request) error {
	count, err := a.queries.CountUsers(r.Context())
	if err != nil {
		return err
	}
	return send(w, 200, map[string]bool{"needs_setup": count == 0})
}
func (a *api) signup(w http.ResponseWriter, r *http.Request) error {
	if err := a.allowLogin(r); err != nil {
		return err
	}
	count, err := a.queries.CountUsers(r.Context())
	if err != nil {
		return err
	}
	if count != 0 {
		return httpError{409, "installation is already configured; sign in"}
	}
	var body signupBody
	if err := decode(w, r, &body); err != nil {
		return err
	}
	if err := validateAccount(&body); err != nil {
		return err
	}
	hash, err := passwordHash(body.Password)
	if err != nil {
		return err
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := a.queries.WithTx(tx)
	user, err := queries.CreateInitialUser(r.Context(), db.CreateInitialUserParams{Email: body.Email, Name: body.Name, PasswordHash: hash})
	if errors.Is(err, sql.ErrNoRows) {
		return httpError{409, "installation is already configured; sign in"}
	}
	if err != nil {
		return err
	}
	session, err := newSession(r.Context(), queries, user)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return send(w, 201, session)
}
func (a *api) login(w http.ResponseWriter, r *http.Request) error {
	if err := a.allowLogin(r); err != nil {
		return err
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(w, r, &body); err != nil {
		return err
	}
	if len(body.Email) > 254 || len(body.Password) > 1024 {
		return httpError{401, "invalid email or password"}
	}
	user, err := a.queries.GetUserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(body.Email)))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	// Match the password cost for unknown accounts without revealing whether an email exists.
	encoded := user.PasswordHash
	if encoded == "" {
		encoded = "pbkdf2-sha256$600000$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	}
	matches := passwordMatches(body.Password, encoded)
	if err != nil || !matches || user.Status != "active" {
		return httpError{401, "invalid email or password"}
	}
	session, err := newSession(r.Context(), a.queries, user)
	return result(w, 200, session, err)
}
func (a *api) authenticateUser(adminToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("WWW-Authenticate", "Bearer")
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if adminToken != "" && subtle.ConstantTimeCompare([]byte(tokenHash(token)), []byte(tokenHash(adminToken))) == 1 {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(token, "vcs_") || len(token) != 68 {
			_ = send(w, 401, map[string]string{"error": "sign in required"})
			return
		}
		session, err := a.queries.GetActiveUserSession(r.Context(), db.GetActiveUserSessionParams{TokenHash: tokenHash(token), NowMs: time.Now().UnixMilli()})
		if err != nil {
			handle(func(w http.ResponseWriter, r *http.Request) error {
				if errors.Is(err, sql.ErrNoRows) {
					return httpError{401, "session expired; sign in"}
				}
				return err
			})(w, r)
			return
		}
		user, err := a.queries.GetUser(r.Context(), session.UserID)
		if err != nil {
			handle(func(http.ResponseWriter, *http.Request) error { return err })(w, r)
			return
		}
		if user.Role != "admin" && r.Method != "GET" && r.URL.Path != "/v1/auth/logout" && r.URL.Path != "/v1/auth/me" {
			_ = send(w, 403, map[string]string{"error": "administrator access required"})
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	})
}
func (a *api) currentUser(w http.ResponseWriter, r *http.Request) error {
	user, ok := r.Context().Value(userContextKey{}).(db.User)
	if !ok {
		return httpError{401, "user session required"}
	}
	if r.Method == "PUT" {
		var body struct {
			Name string `json:"name"`
		}
		if err := decode(w, r, &body); err != nil {
			return err
		}
		body.Name = strings.TrimSpace(body.Name)
		if body.Name == "" || len(body.Name) > 100 {
			return invalid("name must contain 1 to 100 characters")
		}
		updated, err := a.queries.UpdateUser(r.Context(), db.UpdateUserParams{ID: user.ID, Name: body.Name, Role: user.Role, UpdatedAt: time.Now().UnixMilli()})
		if err != nil {
			return err
		}
		user = updated
	}
	return send(w, 200, userJSON(user))
}
func (a *api) logout(w http.ResponseWriter, r *http.Request) error {
	_, err := a.queries.RevokeUserSession(r.Context(), db.RevokeUserSessionParams{TokenHash: tokenHash(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")), NowMs: sql.NullInt64{Int64: time.Now().UnixMilli(), Valid: true}})
	if err != nil {
		return err
	}
	w.WriteHeader(204)
	return nil
}
func (a *api) addUser(w http.ResponseWriter, r *http.Request) error {
	var body signupBody
	if err := decode(w, r, &body); err != nil {
		return err
	}
	if err := validateAccount(&body); err != nil {
		return err
	}
	hash, err := passwordHash(body.Password)
	if err != nil {
		return err
	}
	user, err := a.queries.CreateUser(r.Context(), db.CreateUserParams{Email: body.Email, Name: body.Name, PasswordHash: hash, Role: "member"})
	return result(w, 201, userJSON(user), err)
}
