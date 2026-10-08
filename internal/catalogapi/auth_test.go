package catalogapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dacolabs/cli/internal/catalogapi"
)

func TestNewAuthenticatedClientListDatasetsSendsBearer(t *testing.T) {
	t.Parallel()

	var gotAuth, gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"datasets": []any{},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := catalogapi.NewAuthenticatedClient(srv.URL, "test-token", srv.Client())
	if err != nil {
		t.Fatalf("NewAuthenticatedClient: %v", err)
	}
	pageSize := int32(5)
	res, err := client.ListDatasetsWithResponse(context.Background(), &catalogapi.ListDatasetsParams{
		PageSize: &pageSize,
	})
	if err != nil {
		t.Fatalf("ListDatasetsWithResponse: %v", err)
	}
	if res.StatusCode() != http.StatusOK {
		body, _ := io.ReadAll(res.HTTPResponse.Body)
		t.Fatalf("status %d body %s", res.StatusCode(), body)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want Bearer test-token", gotAuth)
	}
	if gotPath != "/api/catalog/datasets" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "pageSize=5" {
		t.Fatalf("query = %q", gotQuery)
	}
	if res.JSON200 == nil {
		t.Fatal("expected JSON200 payload")
	}
}
