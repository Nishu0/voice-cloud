package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"voice-cloud/db"
)

func newHandler(database *sql.DB, adminToken string) http.Handler {
	a := &api{db: database, queries: db.New(database)}
	admin := http.NewServeMux()
	for path, fn := range map[string]endpoint{
		"connections": a.connections, "models": a.models,
		"deployments": a.deployments, "policies": a.policies,
	} {
		for _, method := range []string{"GET", "POST"} {
			admin.HandleFunc(method+" /v1/"+path, handle(fn))
		}
		for _, method := range []string{"GET", "PUT", "DELETE"} {
			admin.HandleFunc(method+" /v1/"+path+"/{id}", handle(fn))
		}
	}
	admin.HandleFunc("GET /v1/deployments/{id}/languages", handle(a.languages))
	admin.HandleFunc("GET /v1/deployments/{id}/health", handle(a.deploymentHealth))
	admin.HandleFunc("PUT /v1/deployments/{id}/languages/{language}", handle(a.languages))
	admin.HandleFunc("DELETE /v1/deployments/{id}/languages/{language}", handle(a.languages))
	admin.HandleFunc("GET /v1/policies/{id}/targets", handle(a.targets))
	admin.HandleFunc("PUT /v1/policies/{id}/targets/{deployment}", handle(a.targets))
	admin.HandleFunc("DELETE /v1/policies/{id}/targets/{deployment}", handle(a.targets))
	admin.HandleFunc("GET /v1/api-keys", handle(a.keys))
	admin.HandleFunc("POST /v1/api-keys", handle(a.keys))
	admin.HandleFunc("GET /v1/api-keys/{id}", handle(a.keys))
	admin.HandleFunc("DELETE /v1/api-keys/{id}", handle(a.keys))
	admin.HandleFunc("GET /v1/requests", handle(a.requests))
	admin.HandleFunc("GET /v1/requests/{request}", handle(a.requests))
	admin.HandleFunc("GET /v1/requests/{request}/attempts", handle(a.requestAttempts))
	admin.HandleFunc("GET /v1/requests/{request}/usage", handle(a.requestUsage))
	admin.HandleFunc("GET /v1/auth/me", handle(a.currentUser))
	admin.HandleFunc("PUT /v1/auth/me", handle(a.currentUser))
	admin.HandleFunc("POST /v1/auth/logout", handle(a.logout))
	admin.HandleFunc("POST /v1/users", handle(a.addUser))
	root := http.NewServeMux()
	root.HandleFunc("GET /v1/auth/status", handle(a.authStatus))
	root.HandleFunc("POST /v1/auth/signup", handle(a.signup))
	root.HandleFunc("POST /v1/auth/login", handle(a.login))
	root.HandleFunc("GET /healthz", handle(func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil {
			return httpError{http.StatusServiceUnavailable, "database unavailable"}
		}
		return send(w, http.StatusOK, map[string]string{"status": "ok"})
	}))
	root.Handle("/", a.authenticateUser(adminToken, admin))
	return root
}
