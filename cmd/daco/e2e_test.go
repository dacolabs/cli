// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "daco-e2e-")
	if err != nil {
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "daco")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func runDaco(t *testing.T, home, cwd string, args ...string) (stdout, stderr string, exit int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "HOME="+home)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	exit = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			t.Fatalf("run daco: %v", err)
		}
	}
	return stdoutBuf.String(), stderrBuf.String(), exit
}

func newE2EProject(t *testing.T) (home, project string) {
	t.Helper()
	home = t.TempDir()
	project = t.TempDir()
	stdout, stderr, exit := runDaco(t, home, project, "init", "--name", "acme")
	require.Equal(t, 0, exit, "init failed: %s %s", stdout, stderr)
	return home, project
}

func TestE2E_Init_Happy(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	stdout, _, exit := runDaco(t, home, project, "init", "--name", "acme")
	assert.Equal(t, 0, exit)
	assert.Contains(t, stdout, "Created")
	assert.FileExists(t, filepath.Join(project, "daco.yaml"))
	assert.FileExists(t, filepath.Join(home, ".daco", "settings.yaml"))
}

func TestE2E_Init_NoName_JSONError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	_, stderr, exit := runDaco(t, home, project, "init")
	assert.Equal(t, 1, exit)
	stderr = strings.TrimSpace(stderr)
	assert.True(t, strings.HasPrefix(stderr, "["), "expected JSON array, got: %s", stderr)
	assert.Contains(t, stderr, "name is required")
}

func TestE2E_AggregatedValidationErrors(t *testing.T) {
	home, project := newE2EProject(t)
	_, stderr, exit := runDaco(t, home, project, "schemas", "create")
	assert.Equal(t, 1, exit)
	assert.Contains(t, stderr, "name: required")
	assert.Contains(t, stderr, "path: required")
}

func TestE2E_Help_AllGroups(t *testing.T) {
	groups := []string{"init", "products", "connections", "schemas", "ports"}
	home := t.TempDir()
	project := t.TempDir()
	for _, g := range groups {
		t.Run(g, func(t *testing.T) {
			_, _, exit := runDaco(t, home, project, g, "--help")
			assert.Equal(t, 0, exit)
		})
	}
}

func TestE2E_ProductsCreate_Roundtrip(t *testing.T) {
	home, project := newE2EProject(t)
	stdout, _, exit := runDaco(t, home, project, "products", "create",
		"--name", "orders", "--path", "products/orders.yaml")
	require.Equal(t, 0, exit)
	assert.Contains(t, stdout, "Scaffolded")
	assert.FileExists(t, filepath.Join(project, "products", "orders.yaml"))

	stdout, _, exit = runDaco(t, home, project, "products", "list")
	require.Equal(t, 0, exit)
	assert.Contains(t, stdout, "orders")
}

func TestE2E_ConnectionsCreate(t *testing.T) {
	home, project := newE2EProject(t)
	_, _, exit := runDaco(t, home, project, "connections", "create",
		"--name", "db", "--path", "connections/db.yaml",
		"--type", "postgresql", "--host", "localhost:5432")
	require.Equal(t, 0, exit)
	assert.FileExists(t, filepath.Join(project, "connections", "db.yaml"))
}

func TestE2E_SchemasCreate(t *testing.T) {
	home, project := newE2EProject(t)
	_, _, exit := runDaco(t, home, project, "schemas", "create",
		"--name", "user", "--path", "schemas/user.yaml",
		"--title", "User")
	require.Equal(t, 0, exit)
	assert.FileExists(t, filepath.Join(project, "schemas", "user.yaml"))
}

func TestE2E_PortsCreate(t *testing.T) {
	home, project := newE2EProject(t)
	_, _, exit := runDaco(t, home, project, "products", "create",
		"--name", "orders", "--path", "products/orders.yaml")
	require.Equal(t, 0, exit)
	_, _, exit = runDaco(t, home, project, "ports", "create",
		"--product", "orders", "--name", "daily")
	require.Equal(t, 0, exit)
}

func TestE2E_SchemasTranslate(t *testing.T) {
	home, project := newE2EProject(t)
	require.NoError(t, os.MkdirAll(filepath.Join(project, "schemas"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(project, "schemas", "user.yaml"), []byte(`type: object
title: User
properties:
  id: {type: integer}
  email: {type: string}
`), 0o644))
	_, _, exit := runDaco(t, home, project, "schemas", "create",
		"--name", "user", "--path", "schemas/user.yaml")
	require.Equal(t, 0, exit)

	stdout, stderr, exit := runDaco(t, home, project, "schemas", "translate",
		"--name", "user", "--format", "pydantic", "--output", "out")
	require.Equal(t, 0, exit, "exit=%d stdout=%s stderr=%s", exit, stdout, stderr)
	assert.Contains(t, stdout, "Wrote")
	assert.FileExists(t, filepath.Join(project, "out", "user.py"))
}

func TestE2E_PortsTranslate(t *testing.T) {
	home, project := newE2EProject(t)
	require.NoError(t, os.MkdirAll(filepath.Join(project, "products"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(project, "products", "orders.yaml"), []byte(`opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections: {}
ports:
  daily:
    schema:
      type: object
      title: Daily
      properties:
        amount: {type: number}
`), 0o644))
	_, _, exit := runDaco(t, home, project, "products", "create",
		"--name", "orders", "--path", "products/orders.yaml")
	require.Equal(t, 0, exit)

	stdout, stderr, exit := runDaco(t, home, project, "ports", "translate",
		"--product", "orders", "--port", "daily", "--format", "pydantic", "--output", "models")
	require.Equal(t, 0, exit, "exit=%d stdout=%s stderr=%s", exit, stdout, stderr)
	assert.Contains(t, stdout, "Wrote")
	assert.FileExists(t, filepath.Join(project, "models", "daily.py"))
}

func TestE2E_UnknownCommandRejected(t *testing.T) {
	home, project := newE2EProject(t)
	_, _, exit := runDaco(t, home, project, "products", "describe", `{"name":"x"}`)
	assert.Equal(t, 1, exit)
}