package main

import (
	"net/http"
	"regexp"
	"strconv"

	"voice-cloud/db"
)

var languageTag = regexp.MustCompile(`^[A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*$`)

func (a *api) languages(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	if _, err := a.queries.GetDeployment(r.Context(), id); err != nil {
		return err
	}
	if r.Method == http.MethodGet {
		rows, err := a.queries.ListDeploymentLanguages(r.Context(), id)
		if rows == nil {
			rows = []db.DeploymentLanguage{}
		}
		return result(w, 200, rows, err)
	}
	language := r.PathValue("language")
	if !languageTag.MatchString(language) {
		return invalid("invalid language tag")
	}
	if r.Method == http.MethodDelete {
		count, err := a.queries.RemoveDeploymentLanguage(r.Context(), db.RemoveDeploymentLanguageParams{DeploymentID: id, LanguageTag: language})
		return removed(w, count, err)
	}
	if err := a.queries.AddDeploymentLanguage(r.Context(), db.AddDeploymentLanguageParams{DeploymentID: id, LanguageTag: language}); err != nil {
		return err
	}
	return send(w, 200, db.DeploymentLanguage{DeploymentID: id, LanguageTag: language})
}

func (a *api) targets(w http.ResponseWriter, r *http.Request) error {
	id, err := resourceID(r)
	if err != nil {
		return err
	}
	ctx := r.Context()
	if _, err := a.queries.GetRoutingPolicy(ctx, id); err != nil {
		return err
	}
	if r.Method == http.MethodGet {
		rows, err := a.queries.ListPolicyTargets(ctx, id)
		if rows == nil {
			rows = []db.PolicyTarget{}
		}
		return result(w, 200, rows, err)
	}
	deploymentID, err := strconv.ParseInt(r.PathValue("deployment"), 10, 64)
	if err != nil || deploymentID < 1 {
		return invalid("deployment id must be positive")
	}
	if r.Method == http.MethodDelete {
		count, err := a.queries.RemovePolicyTarget(ctx, db.RemovePolicyTargetParams{PolicyID: id, DeploymentID: deploymentID})
		return removed(w, count, err)
	}
	var body struct {
		Priority int64 `json:"priority"`
	}
	if err := decode(w, r, &body); err != nil {
		return err
	}
	if body.Priority < 0 {
		return invalid("priority must be non-negative")
	}
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := a.queries.WithTx(tx)
	policy, err := queries.GetRoutingPolicy(ctx, id)
	if err != nil {
		return err
	}
	deployment, err := queries.GetDeployment(ctx, deploymentID)
	if err != nil {
		return err
	}
	model, err := queries.GetModel(ctx, deployment.ModelID)
	if err != nil {
		return err
	}
	if policy.Task != model.Task {
		return invalid("deployment task must match policy task")
	}
	if err := queries.SetPolicyTarget(ctx, db.SetPolicyTargetParams{PolicyID: id, DeploymentID: deploymentID, Priority: body.Priority}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return send(w, 200, db.PolicyTarget{PolicyID: id, DeploymentID: deploymentID, Priority: body.Priority})
}
