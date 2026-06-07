// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dacolabs/daco/internal/cli/settings"
)

func TestInit_Valid_FreshProject(t *testing.T) {
	_, prjDir := newTestEnv(t)
	usr := newTestUser(t)
	require.NoError(t, usr.Save())

	out, err := Init(context.Background(), InitInput{Usr: usr, Cwd: prjDir, Name: "acme"})
	require.NoError(t, err)
	assert.True(t, out.Created)
	assert.Equal(t, "acme", out.Name)
	assert.Equal(t, filepath.Join(prjDir, settings.ProjectFile), out.ProjectPath)

	// User registry now contains the project.
	reload, err := settings.LoadUser()
	require.NoError(t, err)
	assert.Equal(t, prjDir, reload.Projects["acme"])

	// daco.yaml exists on disk.
	assert.FileExists(t, out.ProjectPath)
}

func TestInit_Valid_ReRegister(t *testing.T) {
	usr, _, prjDir := newTestProject(t)
	// Mutate the registry so the re-register can be observed.
	usr.Projects = map[string]string{}
	require.NoError(t, usr.Save())

	out, err := Init(context.Background(), InitInput{Usr: usr, Cwd: prjDir})
	require.NoError(t, err)
	assert.False(t, out.Created)
	assert.Equal(t, "acme", out.Name)

	reload, err := settings.LoadUser()
	require.NoError(t, err)
	assert.Equal(t, prjDir, reload.Projects["acme"])
}

func TestInit_NameRejectedWhenProjectExists(t *testing.T) {
	usr, _, prjDir := newTestProject(t)
	_, err := Init(context.Background(), InitInput{Usr: usr, Cwd: prjDir, Name: "other"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInit_NoNameOnFreshErrors(t *testing.T) {
	_, prjDir := newTestEnv(t)
	usr := newTestUser(t)
	require.NoError(t, usr.Save())
	_, err := Init(context.Background(), InitInput{Usr: usr, Cwd: prjDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestInit_ValidationAggregates(t *testing.T) {
	_, err := Init(context.Background(), InitInput{})
	require.Error(t, err)
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Contains(t, errs.Error(), "cwd: required")
	assert.Contains(t, errs.Error(), "usr: user settings not loaded")
}

func TestInit_Valid_PartialValidation(t *testing.T) {
	t.Run("missing usr only", func(t *testing.T) {
		_, err := Init(context.Background(), InitInput{Cwd: t.TempDir()})
		var errs Errors
		require.True(t, errors.As(err, &errs))
		assert.Equal(t, Errors{"usr: user settings not loaded"}, errs)
	})
	t.Run("missing cwd only", func(t *testing.T) {
		_, err := Init(context.Background(), InitInput{Usr: newTestUser(t)})
		var errs Errors
		require.True(t, errors.As(err, &errs))
		assert.Equal(t, Errors{"cwd: required"}, errs)
	})
}