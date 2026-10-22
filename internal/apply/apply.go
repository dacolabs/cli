package apply

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dacolabs/cli/internal/catalogapi"
)

type Action string

const (
	ActionCreated   Action = "created"
	ActionPatched   Action = "patched"
	ActionUnchanged Action = "unchanged"
	ActionError     Action = "error"
)

type Result struct {
	Action  Action
	Urn     string
	Version string
	Message string
}

// CatalogClient is the subset of the generated client used by apply.
type CatalogClient interface {
	GetDatasetVersionWithResponse(ctx context.Context, urn catalogapi.DatasetURN, version catalogapi.VersionPathParameter, reqEditors ...catalogapi.RequestEditorFn) (*catalogapi.GetDatasetVersionResponse, error)
	CreateDatasetWithResponse(ctx context.Context, body catalogapi.CreateDatasetJSONRequestBody, reqEditors ...catalogapi.RequestEditorFn) (*catalogapi.CreateDatasetResponse, error)
	PatchDatasetVersionWithApplicationMergePatchPlusJSONBodyWithResponse(ctx context.Context, urn catalogapi.DatasetURN, version catalogapi.VersionPathParameter, params *catalogapi.PatchDatasetVersionParams, body catalogapi.PatchDatasetVersionApplicationMergePatchPlusJSONRequestBody, reqEditors ...catalogapi.RequestEditorFn) (*catalogapi.PatchDatasetVersionResponse, error)
}

// Run reconciles Dataset units against Catalog. dryRun plans without POST/PATCH.
func Run(ctx context.Context, client CatalogClient, units []Unit, dryRun bool) ([]Result, error) {
	out := make([]Result, 0, len(units))
	for _, unit := range units {
		out = append(out, reconcileOne(ctx, client, unit, dryRun))
	}
	return out, nil
}

func reconcileOne(ctx context.Context, client CatalogClient, unit Unit, dryRun bool) Result {
	base := Result{Urn: unit.Input.Urn, Version: string(unit.Input.Version)}
	get, err := client.GetDatasetVersionWithResponse(ctx, unit.Input.Urn, unit.Input.Version)
	if err != nil {
		base.Action = ActionError
		base.Message = err.Error()
		return base
	}
	switch get.StatusCode() {
	case http.StatusNotFound:
		if dryRun {
			base.Action = ActionCreated
			base.Message = "dry-run"
			return base
		}
		created, err := client.CreateDatasetWithResponse(ctx, unit.Input)
		if err != nil {
			base.Action = ActionError
			base.Message = err.Error()
			return base
		}
		if created.StatusCode() < 200 || created.StatusCode() >= 300 {
			base.Action = ActionError
			base.Message = fmt.Sprintf("%s: %s", created.Status(), bytes.TrimSpace(created.Body))
			return base
		}
		base.Action = ActionCreated
		return base
	case http.StatusOK:
		if get.JSON200 == nil {
			base.Action = ActionError
			base.Message = "empty get response"
			return base
		}
		existing := get.JSON200
		sameContract, err := contractsEqual(unit.Input.Contract, existing.Contract)
		if err != nil {
			base.Action = ActionError
			base.Message = err.Error()
			return base
		}
		if !sameContract {
			base.Action = ActionError
			base.Message = fmt.Sprintf("contract differs for %s@%s; bump version in YAML to publish a new Dataset Version", unit.Input.Urn, unit.Input.Version)
			return base
		}
		patch := mutablePatch(unit.Input, *existing)
		if patch == nil {
			base.Action = ActionUnchanged
			return base
		}
		if dryRun {
			base.Action = ActionPatched
			base.Message = "dry-run"
			return base
		}
		patched, err := client.PatchDatasetVersionWithApplicationMergePatchPlusJSONBodyWithResponse(ctx, unit.Input.Urn, unit.Input.Version, nil, *patch)
		if err != nil {
			base.Action = ActionError
			base.Message = err.Error()
			return base
		}
		if patched.StatusCode() < 200 || patched.StatusCode() >= 300 {
			base.Action = ActionError
			base.Message = fmt.Sprintf("%s: %s", patched.Status(), bytes.TrimSpace(patched.Body))
			return base
		}
		base.Action = ActionPatched
		return base
	default:
		base.Action = ActionError
		base.Message = fmt.Sprintf("%s: %s", get.Status(), bytes.TrimSpace(get.Body))
		return base
	}
}

func contractsEqual(desired catalogapi.ContractInput, existing catalogapi.Contract) (bool, error) {
	want, err := canonicalJSON(desired)
	if err != nil {
		return false, err
	}
	// Normalize existing Contract into the same JSON shape as ContractInput.
	asInput := catalogapi.ContractInput{Schema: existing.Schema, Metadata: existing.Metadata}
	got, err := canonicalJSON(asInput)
	if err != nil {
		return false, err
	}
	return bytes.Equal(want, got), nil
}

func mutablePatch(desired catalogapi.DatasetInput, existing catalogapi.Dataset) *catalogapi.DatasetVersionPatch {
	var patch catalogapi.DatasetVersionPatch
	changed := false
	if desired.Title != existing.Title {
		t := desired.Title
		patch.Title = &t
		changed = true
	}
	if desired.Description != existing.Description {
		d := desired.Description
		patch.Description = &d
		changed = true
	}
	metaEqual, err := metadataEqual(desired.Metadata, existing.Metadata)
	if err != nil || !metaEqual {
		m := desired.Metadata
		patch.Metadata = &m
		changed = true
	}
	if !changed {
		return nil
	}
	return &patch
}

func metadataEqual(a, b catalogapi.JSONMetadata) (bool, error) {
	ab, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	var an, bn any
	if err := json.Unmarshal(ab, &an); err != nil {
		return false, err
	}
	if err := json.Unmarshal(bb, &bn); err != nil {
		return false, err
	}
	ab2, err := json.Marshal(an)
	if err != nil {
		return false, err
	}
	bb2, err := json.Marshal(bn)
	if err != nil {
		return false, err
	}
	return bytes.Equal(ab2, bb2), nil
}
