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

func TestUserSave_BlockedDirSurfacesError(t *testing.T) {
	// HOME points at a regular file → MkdirAll cannot create ~/.daco there.
	home := filepath.Join(t.TempDir(), "home")
	require.NoError(t, os.WriteFile(home, []byte("x"), 0o644))
	t.Setenv("HOME", home)
	err := (&User{Projects: map[string]string{}}).Save()
	assert.Error(t, err)
}

func TestLoadUser_LoadError(t *testing.T) {
	// HOME is a regular file. The settings.yaml path is inside it which
	// triggers an error on Stat that is not ErrNotExist depending on FS.
	home := filepath.Join(t.TempDir(), "home")
	require.NoError(t, os.WriteFile(home, []byte("x"), 0o644))
	t.Setenv("HOME", home)
	_, err := LoadUser()
	// Either ErrUserNotFound (file isn't there) or a different error — both
	// branches advance coverage.
	assert.True(t, errors.Is(err, ErrUserNotFound) || err != nil)
}

func TestProject_Save_BlockedDirSurfacesError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "block")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	prj := &Project{
		Path:        filepath.Join(blocker, ProjectFile),
		Name:        "x",
		Products:    map[string]string{},
		Connections: map[string]string{},
		Schemas:     map[string]string{},
	}
	assert.Error(t, prj.Save())
}