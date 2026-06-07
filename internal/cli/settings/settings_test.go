// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package settings

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUser_MissingReturnsSentinel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := LoadUser()
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestLoadUser_CorruptIsWrappedError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, UserDir), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(home, UserDir, UserFile),
		[]byte(":\n  -\n -"),
		0o644,
	))
	_, err := LoadUser()
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrUserNotFound))
}

func TestLoadUser_HappyRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	u := &User{Projects: map[string]string{"acme": "/path/to/acme", "beta": "/path/to/beta"}}
	require.NoError(t, u.Save())
	got, err := LoadUser()
	require.NoError(t, err)
	assert.Equal(t, "/path/to/acme", got.Projects["acme"])
	assert.Equal(t, "/path/to/beta", got.Projects["beta"])
}

func TestUser_Save_CreatesDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoFileExists(t, filepath.Join(home, UserDir, UserFile))
	require.NoError(t, (&User{Projects: map[string]string{}}).Save())
	assert.FileExists(t, filepath.Join(home, UserDir, UserFile))
}

func TestUser_EnsureDefaults_OnLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, UserDir), 0o755))
	// File with no projects key.
	require.NoError(t, os.WriteFile(
		filepath.Join(home, UserDir, UserFile),
		[]byte("\n"),
		0o644,
	))
	got, err := LoadUser()
	require.NoError(t, err)
	assert.NotNil(t, got.Projects)
}

func TestLoadProject_CwdLocal(t *testing.T) {
	dir := t.TempDir()
	writeProjectYAML(t, filepath.Join(dir, ProjectFile), "name: acme\n")
	chdir(t, dir)
	prj, err := LoadProject(&User{Projects: map[string]string{"acme": "/elsewhere"}})
	require.NoError(t, err)
	assert.Equal(t, "acme", prj.Name)
	assert.Equal(t, filepath.Join(dir, ProjectFile), prj.Path)
	// ensureDefaults was applied.
	assert.NotNil(t, prj.Products)
	assert.NotNil(t, prj.Connections)
	assert.NotNil(t, prj.Schemas)
}

func TestLoadProject_RegistryFallback(t *testing.T) {
	projectDir := t.TempDir()
	writeProjectYAML(t, filepath.Join(projectDir, ProjectFile), "name: acme\n")
	subDir := filepath.Join(projectDir, "sub", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	chdir(t, subDir)

	prj, err := LoadProject(&User{Projects: map[string]string{"acme": projectDir}})
	require.NoError(t, err)
	assert.Equal(t, "acme", prj.Name)
}

func TestLoadProject_LongestPrefixWins(t *testing.T) {
	outer := t.TempDir()
	writeProjectYAML(t, filepath.Join(outer, ProjectFile), "name: outer\n")
	inner := filepath.Join(outer, "inner")
	require.NoError(t, os.MkdirAll(inner, 0o755))
	writeProjectYAML(t, filepath.Join(inner, ProjectFile), "name: inner\n")
	cwd := filepath.Join(inner, "more", "deep")
	require.NoError(t, os.MkdirAll(cwd, 0o755))
	chdir(t, cwd)

	prj, err := LoadProject(&User{Projects: map[string]string{"outer": outer, "inner": inner}})
	require.NoError(t, err)
	assert.Equal(t, "inner", prj.Name)
}

func TestLoadProject_NotFound(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	_, err := LoadProject(&User{Projects: map[string]string{}})
	assert.ErrorIs(t, err, ErrProjectNotFound)

	_, err = LoadProject(nil)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestLoadProject_RegistryEntriesPointingNowhere(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	// Registered path doesn't actually contain a daco.yaml file.
	stale := filepath.Join(t.TempDir(), "stale-project")
	require.NoError(t, os.MkdirAll(stale, 0o755))
	_, err := LoadProject(&User{Projects: map[string]string{"stale": stale}})
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProject_Save_RequiresPath(t *testing.T) {
	assert.Error(t, (&Project{Name: "x"}).Save())
}

func TestProject_Save_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ProjectFile)
	prj := &Project{
		Path:        path,
		Name:        "acme",
		Products:    map[string]string{"p": "x.yaml"},
		Connections: map[string]string{"c": "y.yaml"},
		Schemas:     map[string]string{"s": "z.yaml"},
	}
	require.NoError(t, prj.Save())

	chdir(t, dir)
	reloaded, err := LoadProject(&User{Projects: map[string]string{}})
	require.NoError(t, err)
	assert.Equal(t, "acme", reloaded.Name)
	assert.Equal(t, "x.yaml", reloaded.Products["p"])
	assert.Equal(t, "y.yaml", reloaded.Connections["c"])
	assert.Equal(t, "z.yaml", reloaded.Schemas["s"])
}

func TestLoadProjectAt_Happy(t *testing.T) {
	dir := t.TempDir()
	writeProjectYAML(t, filepath.Join(dir, ProjectFile), "name: acme\nproducts:\n  orders: products/orders.yaml\n")
	prj, err := LoadProjectAt(dir)
	require.NoError(t, err)
	assert.Equal(t, "acme", prj.Name)
	assert.Equal(t, filepath.Join(dir, ProjectFile), prj.Path)
	assert.NotNil(t, prj.Products)
	assert.Equal(t, "products/orders.yaml", prj.Products["orders"])
	// ensureDefaults populates the missing maps.
	assert.NotNil(t, prj.Connections)
	assert.NotNil(t, prj.Schemas)
}

func TestLoadProjectAt_MissingFile(t *testing.T) {
	_, err := LoadProjectAt(t.TempDir())
	require.Error(t, err)
}

func TestLoadProjectAt_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	writeProjectYAML(t, filepath.Join(dir, ProjectFile), ":\n  -\n -")
	_, err := LoadProjectAt(dir)
	require.Error(t, err)
}

func TestLoadProject_CorruptCwdFile(t *testing.T) {
	dir := t.TempDir()
	writeProjectYAML(t, filepath.Join(dir, ProjectFile), ":\n  -\n -")
	chdir(t, dir)
	_, err := LoadProject(&User{})
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrProjectNotFound))
}

func writeProjectYAML(t *testing.T, path, body string) {
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