// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "daco.yaml")

	cfg := Config{
		Version: 1,
		Path:    "spec/spec.yaml",
	}

	err := cfg.Save(cfgPath)
	require.NoError(t, err)

	loaded, err := Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, cfg.Version, loaded.Version)
	assert.Equal(t, cfg.Path, loaded.Path)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "valid config",
			cfg:     Config{Version: 1},
			wantErr: "",
		},
		{
			name:    "unsupported version",
			cfg:     Config{Version: 99},
			wantErr: "unsupported config version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestConfig_SaveFormat(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "daco.yaml")

	cfg := Config{
		Version: 1,
		Path:    "spec/spec.yaml",
	}

	err := cfg.Save(cfgPath)
	require.NoError(t, err)

	content, err := os.ReadFile(cfgPath) //nolint:gosec // test file path
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "version: 1")
	assert.Contains(t, output, "path: spec/spec.yaml")
}

func TestConfig_Load(t *testing.T) {
	cfg, err := Load("testdata/valid.yaml")
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "spec/spec.yaml", cfg.Path)
}

func TestConfig_Load_NotFound(t *testing.T) {
	_, err := Load("testdata/nonexistent.yaml")
	assert.Error(t, err)
}

func TestConfig_Load_Invalid(t *testing.T) {
	_, err := Load("testdata/invalid.yaml")
	assert.Error(t, err)
}

func TestConfig_Save_InvalidPath(t *testing.T) {
	cfg := Config{Version: 1}

	err := cfg.Save("/nonexistent/directory/config.yaml")
	assert.Error(t, err)
}

func TestConfig_Load_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.yaml")
	require.NoError(t, os.WriteFile(emptyFile, []byte(""), 0o600))

	_, err := Load(emptyFile)
	assert.Error(t, err)
}
