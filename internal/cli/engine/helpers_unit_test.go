// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dacolabs/daco/internal/cli/settings"
)

func TestResolveProjectPath(t *testing.T) {
	prj := &settings.Project{Path: "/home/me/proj/daco.yaml"}
	assert.Equal(t, "/home/me/proj/products/x.yaml", resolveProjectPath(prj, "products/x.yaml"))
	assert.Equal(t, "/abs/already.yaml", resolveProjectPath(prj, "/abs/already.yaml"))
}

type fakeTranslator struct{ extOut string }

func (f fakeTranslator) Translate(string, *jsonschema.Schema, string) ([]byte, error) {
	return []byte("x"), nil
}
func (f fakeTranslator) FileExtension() string { return f.extOut }

func TestExt_AcceptsBothFormats(t *testing.T) {
	assert.Equal(t, ".py", ext(fakeTranslator{extOut: ".py"}))
	assert.Equal(t, ".sql", ext(fakeTranslator{extOut: "sql"}))
	assert.Equal(t, "", ext(fakeTranslator{extOut: ""}))
}

// Save errors are forced by setting the project Path to a file inside a path
// where a regular file blocks MkdirAll from creating the parent.
func blockSaveDir(t *testing.T, prj *settings.Project, prjDir string) {
	t.Helper()
	blocker := filepath.Join(prjDir, "block")
	require.NoError(t, os.WriteFile(blocker, []byte("blocker"), 0o644))
	prj.Path = filepath.Join(blocker, "daco.yaml")
}

func TestProductsDelete_SaveErrorSurfaces(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	blockSaveDir(t, prj, prjDir)
	_, err := ProductsDelete(context.Background(), ProductsDeleteInput{Prj: prj, Name: "orders"})
	require.Error(t, err)
}

func TestSchemasDelete_SaveErrorSurfaces(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	blockSaveDir(t, prj, prjDir)
	_, err := SchemasDelete(context.Background(), SchemasDeleteInput{Prj: prj, Name: "user"})
	require.Error(t, err)
}

func TestConnectionsDelete_SaveErrorSurfaces(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "h:5432")
	blockSaveDir(t, prj, prjDir)
	_, err := ConnectionsDelete(context.Background(), ConnectionsDeleteInput{Prj: prj, Name: "db"})
	require.Error(t, err)
}

func TestRunTranslator_UnknownFormat(t *testing.T) {
	_, err := runTranslator("x", &jsonschema.Schema{}, "nope", t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope")
}

func TestRunTranslator_MkdirError(t *testing.T) {
	// Blocker file at the output path prevents MkdirAll.
	dir := t.TempDir()
	blocker := filepath.Join(dir, "out")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	_, err := runTranslator("x", &jsonschema.Schema{}, "pydantic", blocker)
	require.Error(t, err)
}