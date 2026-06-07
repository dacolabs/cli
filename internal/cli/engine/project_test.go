// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectFormat_RewritesAllArtifacts(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")

	// Corrupt the YAML formatting (extra indentation, trailing spaces) so we
	// can verify Format normalized it.
	schemaPath := filepath.Join(prjDir, "schemas/user.yaml")
	require.NoError(t, os.WriteFile(schemaPath,
		[]byte("type:    object\ntitle:    user   \n"),
		0o644))

	out, err := ProjectFormat(context.Background(), ProjectFormatInput{Prj: prj})
	require.NoError(t, err)
	assert.Len(t, out.Files, 3)
	for _, f := range out.Files {
		assert.True(t, filepath.IsAbs(f), "expected absolute path, got %s", f)
	}

	// After format the schema file no longer has the multiple-space artifacts.
	body, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "   ", "format should have collapsed runs of spaces")
}

func TestProjectFormat_Validation(t *testing.T) {
	_, err := ProjectFormat(context.Background(), ProjectFormatInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Contains(t, errs.Error(), "prj: project not loaded")
}

func TestProjectFormat_SurfacesLoadError(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	// Make the schema file unreadable as YAML.
	require.NoError(t, os.WriteFile(
		filepath.Join(prjDir, "schemas/user.yaml"),
		[]byte(":\n  -\n -"),
		0o644,
	))
	_, err := ProjectFormat(context.Background(), ProjectFormatInput{Prj: prj})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "format schema")
}

func TestProjectLint_EmptyProjectHasNoIssues(t *testing.T) {
	_, prj, _ := newTestProject(t)
	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	assert.Empty(t, out.Issues)
}

func TestProjectLint_Validation(t *testing.T) {
	_, err := ProjectLint(context.Background(), ProjectLintInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Contains(t, errs.Error(), "prj: project not loaded")
}

func TestProjectLint_ReportsMissingFiles(t *testing.T) {
	_, prj, _ := newTestProject(t)
	prj.Schemas["ghost"] = "schemas/ghost.yaml"
	prj.Connections["missing"] = "connections/missing.yaml"
	prj.Products["lost"] = "products/lost.yaml"
	require.NoError(t, prj.Save())

	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Issues, 3)
	// Sorted by path so order is deterministic.
	for _, issue := range out.Issues {
		assert.Contains(t, issue.Message, "no such file")
	}
}

func TestProjectLint_ReportsConnectionMissingTypeOrHost(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	// Empty out the connection so it's neither a $ref nor has type/host.
	writeFile(t, filepath.Join(prjDir, "connections/warehouse.yaml"), "{}\n")
	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Issues, 1)
	assert.Contains(t, out.Issues[0].Message, "missing type or host")
}

func TestProjectLint_ReportsConnectionRefThatDoesNotResolve(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	writeFile(t, filepath.Join(prjDir, "connections/warehouse.yaml"), "$ref: ./does-not-exist.yaml\n")
	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Issues, 1)
	assert.Contains(t, out.Issues[0].Message, "$ref")
	assert.Contains(t, out.Issues[0].Message, "does not resolve")
}

func TestProjectLint_ReportsProductWithUnknownConnectionAndMissingLocation(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	// Hand-roll a product file with a port bound to an unknown connection and a
	// connection with empty binding.
	writeFile(t, filepath.Join(prjDir, "products/orders.yaml"), `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections: {}
ports:
  daily:
    connections:
      - connection: ""
        location: nowhere
      - connection: ghost
        location: ""
`)
	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(out.Issues), 3)

	var (
		emptyBinding, unknownConnection, missingLocation bool
	)
	for _, issue := range out.Issues {
		switch {
		case contains(issue.Message, "empty name"):
			emptyBinding = true
		case contains(issue.Message, "unknown connection"):
			unknownConnection = true
		case contains(issue.Message, "missing location"):
			missingLocation = true
		}
	}
	assert.True(t, emptyBinding, "expected empty-binding issue")
	assert.True(t, unknownConnection, "expected unknown-connection issue")
	assert.True(t, missingLocation, "expected missing-location issue")
}

func TestProjectLint_ReportsPortSchemaRefDoesNotResolve(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	writeFile(t, filepath.Join(prjDir, "products/orders.yaml"), `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections: {}
ports:
  daily:
    schema:
      $ref: ../schemas/ghost.yaml
`)
	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Issues, 1)
	assert.Contains(t, out.Issues[0].Message, "schema $ref")
	assert.Contains(t, out.Issues[0].Message, "does not resolve")
}

func TestProjectLint_ReportsProductConnectionInlineMissingTypeHost(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	writeFile(t, filepath.Join(prjDir, "products/orders.yaml"), `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections:
  warehouse: {}
ports: {}
`)
	out, err := ProjectLint(context.Background(), ProjectLintInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Issues, 1)
	assert.Contains(t, out.Issues[0].Message, "no $ref")
	assert.Contains(t, out.Issues[0].Message, "missing type/host")
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
