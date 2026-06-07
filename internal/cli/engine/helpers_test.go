// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dacolabs/daco/internal/cli/settings"
)

func newTestEnv(t *testing.T) (homeDir, projectDir string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	prjDir := t.TempDir()
	chdir(t, prjDir)
	return home, prjDir
}

func newTestUser(t *testing.T) *settings.User {
	t.Helper()
	return &settings.User{Projects: map[string]string{}}
}

func newTestProject(t *testing.T) (*settings.User, *settings.Project, string) {
	t.Helper()
	_, prjDir := newTestEnv(t)
	usr := newTestUser(t)
	require.NoError(t, usr.Save())
	out, err := Init(context.Background(), InitInput{Usr: usr, Cwd: prjDir, Name: "acme"})
	require.NoError(t, err)
	require.True(t, out.Created)
	prj, err := settings.LoadProject(usr)
	require.NoError(t, err)
	return usr, prj, prjDir
}

func mustCreateProduct(t *testing.T, prj *settings.Project, name, path string) {
	t.Helper()
	_, err := ProductsCreate(context.Background(), ProductsCreateInput{Prj: prj, Name: name, Path: path})
	require.NoError(t, err)
}

func mustCreateConnection(t *testing.T, prj *settings.Project, name, path, kind, host string) {
	t.Helper()
	_, err := ConnectionsCreate(context.Background(), ConnectionsCreateInput{
		Prj: prj, Name: name, Path: path, Type: kind, Host: host,
	})
	require.NoError(t, err)
}

func mustCreateSchema(t *testing.T, prj *settings.Project, name, path string) {
	t.Helper()
	_, err := SchemasCreate(context.Background(), SchemasCreateInput{
		Prj: prj, Name: name, Path: path, Type: "object", Title: name,
	})
	require.NoError(t, err)
}

func mustCreatePort(t *testing.T, prj *settings.Project, product, name string) {
	t.Helper()
	_, err := PortsCreate(context.Background(), PortsCreateInput{
		Prj: prj, Product: product, Name: name,
	})
	require.NoError(t, err)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
}