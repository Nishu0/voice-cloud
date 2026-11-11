package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"modernc.org/sqlite"
	"voice-cloud/db"
)

type api struct {
	db      *sql.DB
	queries *db.Queries
	logins  loginLimiter
}

type endpoint func(http.ResponseWriter, *http.Request) error

type httpError struct {
	status  int
	message string
}

func (e httpError) Error() string  { return e.message }
func invalid(message string) error { return httpError{http.StatusBadRequest, message} }

func handle(fn endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if err := fn(w, r); err != nil {
			status, message := http.StatusInternalServerError, "internal server error"
			var problem httpError
			var sqliteError *sqlite.Error
			switch {
			case errors.As(err, &problem):
				status, message = problem.status, problem.message
			case errors.Is(err, sql.ErrNoRows):
				status, message = http.StatusNotFound, "resource not found"
			case errors.As(err, &sqliteError):
				switch sqliteError.Code() & 255 {
				case 19:
					status, message = http.StatusConflict, "database constraint conflict"
				case 5, 6:
					status, message = http.StatusServiceUnavailable, "database is busy; retry later"
				}
			}
			if status == http.StatusInternalServerError {
				slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
			}
			_ = send(w, status, map[string]string{"error": message})
		}
	}
}

func send(w http.ResponseWriter, status int, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(append(body, '\n'))
	return err
}

func result(w http.ResponseWriter, status int, value any, err error) error {
	if err != nil {
		return err
	}
	return send(w, status, value)
}

func removed(w http.ResponseWriter, count int64, err error) error {
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func decode(w http.ResponseWriter, r *http.Request, value any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return httpError{http.StatusUnsupportedMediaType, "Content-Type must be application/json"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	var body json.RawMessage
	if err := decoder.Decode(&body); err != nil {
		return decodeError(err)
	}
	if len(body) == 0 || body[0] != '{' {
		return invalid("body must be a JSON object")
	}
	fields := json.NewDecoder(bytes.NewReader(body))
	fields.DisallowUnknownFields()
	if err := fields.Decode(value); err != nil {
		return decodeError(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err != nil {
			return decodeError(err)
		}
		return invalid("body must contain exactly one JSON object")
	}
	return nil
}

func decodeError(err error) error {
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		return httpError{http.StatusRequestEntityTooLarge, "body exceeds 1 MiB"}
	}
	return invalid("invalid JSON body: " + err.Error())
}

func resourceID(r *http.Request) (int64, error) {
	if r.PathValue("id") == "" {
		return 0, nil
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, invalid("id must be a positive integer")
	}
	return id, nil
}

func required(values ...string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return invalid("required fields must not be blank")
		}
	}
	return nil
}
func taskValid(task string) error {
	if task != "stt" && task != "tts" && task != "llm" {
		return invalid("task must be stt, tts, or llm")
	}
	return nil
}
func object(value *json.RawMessage) error {
	if len(*value) == 0 {
		*value = json.RawMessage(`{}`)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(*value, &fields); err != nil || fields == nil {
		return invalid("JSON settings must be objects")
	}
	return nil
}
func nullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}
func nullableInt(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}
func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
func defaultEnabled(value *bool) int64 { return boolInt(value == nil || *value) }
func stringValue(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}
func intValue(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}
func mapItems[T any](items []T, view func(T) any) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, view(item))
	}
	return result
}

func pagination(r *http.Request) (db.ListRequestsParams, error) {
	params := db.ListRequestsParams{Limit: 50}
	for name, target := range map[string]*int64{"limit": &params.Limit, "offset": &params.Offset} {
		if raw := r.URL.Query().Get(name); raw != "" {
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || value < 0 {
				return params, invalid(fmt.Sprintf("%s must be a non-negative integer", name))
			}
			*target = value
		}
	}
	if params.Limit < 1 || params.Limit > 100 {
		return params, invalid("limit must be between 1 and 100")
	}
	return params, nil
}
