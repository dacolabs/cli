package apply_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dacolabs/cli/internal/apply"
	"github.com/dacolabs/cli/internal/catalogapi"
)

func TestApplyCreatesMissingVersion(t *testing.T) {
	t.Parallel()
	var created bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/versions/"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"about:blank","title":"Not Found","status":404}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/catalog/datasets":
			created = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"urn": "urn:daco:dataset:orders", "version": "1.0.0",
				"title": "Orders", "description": "d", "metadata": map[string]any{},
				"contract": map[string]any{"schema": map[string]any{"type": "object"}, "metadata": map[string]any{}},
				"revision": 1, "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
			})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := catalogapi.NewAuthenticatedClient(srv.URL, "tok", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	units := []apply.Unit{{
		Source: "test",
		Input: catalogapi.DatasetInput{
			Urn: "urn:daco:dataset:orders", Version: "1.0.0", Title: "Orders", Description: "d",
			Metadata: catalogapi.JSONMetadata{},
			Contract: catalogapi.ContractInput{Schema: map[string]any{"type": "object"}, Metadata: catalogapi.JSONMetadata{}},
		},
	}}
	res, err := apply.Run(context.Background(), client, units, false)
	if err != nil {
		t.Fatal(err)
	}
	if !created || len(res) != 1 || res[0].Action != apply.ActionCreated {
		t.Fatalf("created=%v res=%+v", created, res)
	}
}

func TestApplyPatchesMutableFields(t *testing.T) {
	t.Parallel()
	existing := map[string]any{
		"urn": "urn:daco:dataset:orders", "version": "1.0.0",
		"title": "Old", "description": "d", "metadata": map[string]any{"keep": true, "drop": 1},
		"contract": map[string]any{"schema": map[string]any{"type": "object"}, "metadata": map[string]any{}},
		"revision": 1, "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	var patched bool
	var gotIfMatch string
	var patchBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("ETag", `"rev-1"`)
			_ = json.NewEncoder(w).Encode(existing)
		case http.MethodPatch:
			gotIfMatch = r.Header.Get("If-Match")
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &patchBody); err != nil {
				t.Errorf("patch body: %v", err)
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}
			if !strings.HasSuffix(r.URL.Path, "/versions/1.0.0") {
				t.Errorf("path=%s", r.URL.Path)
				http.Error(w, "bad path", http.StatusBadRequest)
				return
			}
			patched = true
			existing["title"] = "Orders"
			existing["metadata"] = map[string]any{"keep": true}
			_ = json.NewEncoder(w).Encode(existing)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)
	client, err := catalogapi.NewAuthenticatedClient(srv.URL, "tok", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	units := []apply.Unit{{
		Source: "test",
		Input: catalogapi.DatasetInput{
			Urn: "urn:daco:dataset:orders", Version: "1.0.0", Title: "Orders", Description: "d",
			Metadata: catalogapi.JSONMetadata{"keep": true},
			Contract: catalogapi.ContractInput{Schema: map[string]any{"type": "object"}, Metadata: catalogapi.JSONMetadata{}},
		},
	}}
	res, err := apply.Run(context.Background(), client, units, false)
	if err != nil {
		t.Fatal(err)
	}
	if !patched || res[0].Action != apply.ActionPatched {
		t.Fatalf("patched=%v res=%+v", patched, res)
	}
	if gotIfMatch != `"rev-1"` {
		t.Fatalf("If-Match=%q", gotIfMatch)
	}
	if patchBody["title"] != "Orders" {
		t.Fatalf("patch title=%v", patchBody["title"])
	}
	meta, _ := patchBody["metadata"].(map[string]any)
	drop, dropOK := meta["drop"]
	if meta["keep"] != true || !dropOK || drop != nil {
		t.Fatalf("metadata patch=%v (want keep=true and drop present as null)", meta)
	}
}

func TestApplyFailsWhenContractDiffers(t *testing.T) {
	t.Parallel()
	existing := map[string]any{
		"urn": "urn:daco:dataset:orders", "version": "1.0.0",
		"title": "Orders", "description": "d", "metadata": map[string]any{},
		"contract": map[string]any{"schema": map[string]any{"type": "string"}, "metadata": map[string]any{}},
		"revision": 1, "createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected mutate %s", r.Method)
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(existing)
	}))
	t.Cleanup(srv.Close)
	client, err := catalogapi.NewAuthenticatedClient(srv.URL, "tok", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	units := []apply.Unit{{
		Source: "test",
		Input: catalogapi.DatasetInput{
			Urn: "urn:daco:dataset:orders", Version: "1.0.0", Title: "Orders", Description: "d",
			Metadata: catalogapi.JSONMetadata{},
			Contract: catalogapi.ContractInput{Schema: map[string]any{"type": "object"}, Metadata: catalogapi.JSONMetadata{}},
		},
	}}
	res, err := apply.Run(context.Background(), client, units, false)
	if err != nil {
		t.Fatal(err)
	}
	if res[0].Action != apply.ActionError || !strings.Contains(res[0].Message, "bump") {
		t.Fatalf("res=%+v", res[0])
	}
}

func TestDryRunDoesNotMutate(t *testing.T) {
	t.Parallel()
	var mutated bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			mutated = true
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	client, err := catalogapi.NewAuthenticatedClient(srv.URL, "tok", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	units := []apply.Unit{{
		Source: "test",
		Input: catalogapi.DatasetInput{
			Urn: "urn:daco:dataset:orders", Version: "1.0.0", Title: "Orders", Description: "d",
			Metadata: catalogapi.JSONMetadata{},
			Contract: catalogapi.ContractInput{Schema: map[string]any{"type": "object"}, Metadata: catalogapi.JSONMetadata{}},
		},
	}}
	res, err := apply.Run(context.Background(), client, units, true)
	if err != nil {
		t.Fatal(err)
	}
	if mutated || res[0].Action != apply.ActionCreated {
		t.Fatalf("mutated=%v res=%+v", mutated, res)
	}
}
