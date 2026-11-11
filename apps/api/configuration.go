package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"voice-cloud/db"
)

type connectionBody struct {
	Name          string  `json:"name"`
	Adapter       string  `json:"adapter"`
	LocationKind  string  `json:"location_kind"`
	BaseURL       string  `json:"base_url"`
	CredentialRef *string `json:"credential_ref"`
	Enabled       *bool   `json:"enabled"`
}

func connectionJSON(row db.Connection) any {
	return struct {
		db.Connection
		CredentialRef any  `json:"credential_ref"`
		Enabled       bool `json:"enabled"`
	}{row, stringValue(row.CredentialRef), row.Enabled == 1}
}

func (a *api) connections(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		if id == 0 {
			rows, err := a.queries.ListConnections(ctx)
			return result(w, 200, mapItems(rows, connectionJSON), err)
		}
		row, err := a.queries.GetConnection(ctx, id)
		return result(w, 200, connectionJSON(row), err)
	case http.MethodDelete:
		count, err := a.queries.DeleteConnection(ctx, id)
		return removed(w, count, err)
	default:
		var body connectionBody
		if err := decode(w, r, &body); err != nil {
			return err
		}
		if err := required(body.Name, body.Adapter, body.BaseURL); err != nil {
			return err
		}
		if body.LocationKind != "local" && body.LocationKind != "cloud" {
			return invalid("location_kind must be local or cloud")
		}
		parsed, err := url.Parse(body.BaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
			return invalid("base_url must be an HTTP(S) URL without credentials or a fragment")
		}
		if body.CredentialRef != nil {
			if err := required(*body.CredentialRef); err != nil {
				return err
			}
		}
		if id == 0 {
			row, err := a.queries.CreateConnection(ctx, db.CreateConnectionParams{Name: body.Name, Adapter: body.Adapter, LocationKind: body.LocationKind, BaseUrl: body.BaseURL, CredentialRef: nullableString(body.CredentialRef), Enabled: defaultEnabled(body.Enabled)})
			return result(w, 201, connectionJSON(row), err)
		}
		row, err := a.queries.UpdateConnection(ctx, db.UpdateConnectionParams{ID: id, Name: body.Name, Adapter: body.Adapter, LocationKind: body.LocationKind, BaseUrl: body.BaseURL, CredentialRef: nullableString(body.CredentialRef), Enabled: defaultEnabled(body.Enabled), UpdatedAt: time.Now().UnixMilli()})
		return result(w, 200, connectionJSON(row), err)
	}
}

type modelBody struct {
	Name      string `json:"name"`
	Publisher string `json:"publisher"`
	Task      string `json:"task"`
	Version   string `json:"version"`
}

func (a *api) models(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		if id == 0 {
			rows, err := a.queries.ListModels(ctx)
			if rows == nil {
				rows = []db.Model{}
			}
			return result(w, 200, rows, err)
		}
		row, err := a.queries.GetModel(ctx, id)
		return result(w, 200, row, err)
	case http.MethodDelete:
		count, err := a.queries.DeleteModel(ctx, id)
		return removed(w, count, err)
	default:
		var body modelBody
		if err := decode(w, r, &body); err != nil {
			return err
		}
		if err := required(body.Name, body.Publisher); err != nil {
			return err
		}
		if err := taskValid(body.Task); err != nil {
			return err
		}
		if id == 0 {
			row, err := a.queries.CreateModel(ctx, db.CreateModelParams{Name: body.Name, Publisher: body.Publisher, Task: body.Task, Version: body.Version})
			return result(w, 201, row, err)
		}
		previous, err := a.queries.GetModel(ctx, id)
		if err != nil {
			return err
		}
		if previous.Task != body.Task {
			return httpError{409, "model task is immutable; create a new model"}
		}
		row, err := a.queries.UpdateModel(ctx, db.UpdateModelParams{ID: id, Name: body.Name, Publisher: body.Publisher, Task: body.Task, Version: body.Version, UpdatedAt: time.Now().UnixMilli()})
		return result(w, 200, row, err)
	}
}

type deploymentBody struct {
	Name            string          `json:"name"`
	ModelID         int64           `json:"model_id"`
	ConnectionID    int64           `json:"connection_id"`
	UpstreamModelID string          `json:"upstream_model_id"`
	Enabled         *bool           `json:"enabled"`
	Capabilities    json.RawMessage `json:"capabilities_json"`
	Pricing         json.RawMessage `json:"pricing_json"`
	Settings        json.RawMessage `json:"settings_json"`
}

func deploymentJSON(row db.Deployment) any {
	return struct {
		db.Deployment
		Enabled      bool            `json:"enabled"`
		Capabilities json.RawMessage `json:"capabilities_json"`
		Pricing      json.RawMessage `json:"pricing_json"`
		Settings     json.RawMessage `json:"settings_json"`
	}{row, row.Enabled == 1, json.RawMessage(row.CapabilitiesJson), json.RawMessage(row.PricingJson), json.RawMessage(row.SettingsJson)}
}

func (a *api) deployments(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		if id == 0 {
			rows, err := a.queries.ListDeployments(ctx)
			return result(w, 200, mapItems(rows, deploymentJSON), err)
		}
		row, err := a.queries.GetDeployment(ctx, id)
		return result(w, 200, deploymentJSON(row), err)
	case http.MethodDelete:
		count, err := a.queries.DeleteDeployment(ctx, id)
		return removed(w, count, err)
	default:
		var body deploymentBody
		if err := decode(w, r, &body); err != nil {
			return err
		}
		if err := required(body.Name, body.UpstreamModelID); err != nil {
			return err
		}
		if body.ModelID < 1 || body.ConnectionID < 1 {
			return invalid("model_id and connection_id must be positive")
		}
		for _, value := range []*json.RawMessage{&body.Capabilities, &body.Pricing, &body.Settings} {
			if err := object(value); err != nil {
				return err
			}
		}
		if id == 0 {
			row, err := a.queries.CreateDeployment(ctx, db.CreateDeploymentParams{Name: body.Name, ModelID: body.ModelID, ConnectionID: body.ConnectionID, UpstreamModelID: body.UpstreamModelID, Enabled: defaultEnabled(body.Enabled), CapabilitiesJson: string(body.Capabilities), PricingJson: string(body.Pricing), SettingsJson: string(body.Settings)})
			return result(w, 201, deploymentJSON(row), err)
		}
		previous, err := a.queries.GetDeployment(ctx, id)
		if err != nil {
			return err
		}
		if previous.ModelID != body.ModelID || previous.ConnectionID != body.ConnectionID || previous.UpstreamModelID != body.UpstreamModelID {
			return httpError{409, "deployment model and connection are immutable; create a new deployment"}
		}
		row, err := a.queries.UpdateDeployment(ctx, db.UpdateDeploymentParams{ID: id, Name: body.Name, ModelID: body.ModelID, ConnectionID: body.ConnectionID, UpstreamModelID: body.UpstreamModelID, Enabled: defaultEnabled(body.Enabled), CapabilitiesJson: string(body.Capabilities), PricingJson: string(body.Pricing), SettingsJson: string(body.Settings), UpdatedAt: time.Now().UnixMilli()})
		return result(w, 200, deploymentJSON(row), err)
	}
}

type policyBody struct {
	Name            string          `json:"name"`
	Task            string          `json:"task"`
	Objective       string          `json:"objective"`
	FallbackEnabled *bool           `json:"fallback_enabled"`
	MaxAttempts     *int64          `json:"max_attempts"`
	Constraints     json.RawMessage `json:"constraints_json"`
}

func policyJSON(row db.RoutingPolicy) any {
	return struct {
		db.RoutingPolicy
		FallbackEnabled bool            `json:"fallback_enabled"`
		Constraints     json.RawMessage `json:"constraints_json"`
	}{row, row.FallbackEnabled == 1, json.RawMessage(row.ConstraintsJson)}
}

func (a *api) policies(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		if id == 0 {
			rows, err := a.queries.ListRoutingPolicies(ctx)
			return result(w, 200, mapItems(rows, policyJSON), err)
		}
		row, err := a.queries.GetRoutingPolicy(ctx, id)
		return result(w, 200, policyJSON(row), err)
	case http.MethodDelete:
		count, err := a.queries.DeleteRoutingPolicy(ctx, id)
		return removed(w, count, err)
	default:
		var body policyBody
		if err := decode(w, r, &body); err != nil {
			return err
		}
		if err := required(body.Name); err != nil {
			return err
		}
		if err := taskValid(body.Task); err != nil {
			return err
		}
		if body.Objective == "" {
			body.Objective = "balanced"
		}
		switch body.Objective {
		case "balanced", "latency", "cost", "quality":
		default:
			return invalid("invalid routing objective")
		}
		attempts := int64(2)
		if body.MaxAttempts != nil {
			attempts = *body.MaxAttempts
		}
		if attempts < 1 {
			return invalid("max_attempts must be positive")
		}
		if err := object(&body.Constraints); err != nil {
			return err
		}
		if id == 0 {
			row, err := a.queries.CreateRoutingPolicy(ctx, db.CreateRoutingPolicyParams{Name: body.Name, Task: body.Task, Objective: body.Objective, FallbackEnabled: defaultEnabled(body.FallbackEnabled), MaxAttempts: attempts, ConstraintsJson: string(body.Constraints)})
			return result(w, 201, policyJSON(row), err)
		}
		previous, err := a.queries.GetRoutingPolicy(ctx, id)
		if err != nil {
			return err
		}
		if previous.Task != body.Task {
			return httpError{409, "policy task is immutable; create a new policy"}
		}
		row, err := a.queries.UpdateRoutingPolicy(ctx, db.UpdateRoutingPolicyParams{ID: id, Name: body.Name, Task: body.Task, Objective: body.Objective, FallbackEnabled: defaultEnabled(body.FallbackEnabled), MaxAttempts: attempts, ConstraintsJson: string(body.Constraints), UpdatedAt: time.Now().UnixMilli()})
		return result(w, 200, policyJSON(row), err)
	}
}
