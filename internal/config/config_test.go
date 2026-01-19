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
		Schema: SchemaConfig{
			Organization: SchemaModular,
		},
	}

	err := cfg.Save(cfgPath)
	require.NoError(t, err)

	loaded, err := Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, cfg.Version, loaded.Version)
	assert.Equal(t, cfg.Path, loaded.Path)
	assert.Equal(t, cfg.Schema.Organization, loaded.Schema.Organization)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid config with modular",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: SchemaModular,
				},
			},
			wantErr: "",
		},
		{
			name: "valid config with components",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: SchemaComponents,
				},
			},
			wantErr: "",
		},
		{
			name: "valid config with inline",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: SchemaInline,
				},
			},
			wantErr: "",
		},
		{
			name: "invalid schema organization",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: "invalid",
				},
			},
			wantErr: "schema organization must be",
		},
		{
			name: "unsupported version",
			cfg: Config{
				Version: 99,
				Schema: SchemaConfig{
					Organization: SchemaModular,
				},
			},
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
		Schema: SchemaConfig{
			Organization: SchemaModular,
		},
	}

	err := cfg.Save(cfgPath)
	require.NoError(t, err)

	content, err := os.ReadFile(cfgPath) //nolint:gosec // test file path
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "version: 1")
	assert.Contains(t, output, "path: spec/spec.yaml")
	assert.Contains(t, output, "organization: modular")
}

func TestConfig_Load(t *testing.T) {
	cfg, err := Load("testdata/valid.yaml")
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "spec/spec.yaml", cfg.Path)
	assert.Equal(t, SchemaModular, cfg.Schema.Organization)
}

func TestConfig_Load_WithExtends(t *testing.T) {
	cfg, err := Load("testdata/child.yaml")
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "spec/spec.yaml", cfg.Path)
	assert.Equal(t, SchemaModular, cfg.Schema.Organization)
}

func TestConfig_Load_WithExtendsOverride(t *testing.T) {
	cfg, err := Load("testdata/child-override.yaml")
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "custom/spec.yaml", cfg.Path)
	assert.Equal(t, SchemaInline, cfg.Schema.Organization)
}

func TestConfig_Load_NotFound(t *testing.T) {
	_, err := Load("testdata/nonexistent.yaml")
	assert.Error(t, err)
}

func TestConfig_Load_Invalid(t *testing.T) {
	_, err := Load("testdata/invalid.yaml")
	assert.Error(t, err)
}

func TestConfig_Save(t *testing.T) {
	cfg := Config{
		Version: 1,
		Path:    "spec/spec.yaml",
		Schema: SchemaConfig{
			Organization: SchemaComponents,
		},
	}

	tmpFile := t.TempDir() + "/test-config.yaml"

	err := cfg.Save(tmpFile)
	require.NoError(t, err)

	loaded, err := Load(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, cfg.Version, loaded.Version)
	assert.Equal(t, cfg.Path, loaded.Path)
	assert.Equal(t, cfg.Schema.Organization, loaded.Schema.Organization)
}

func TestConfig_Save_InvalidPath(t *testing.T) {
	cfg := Config{
		Version: 1,
		Schema: SchemaConfig{
			Organization: SchemaModular,
		},
	}

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

func TestConfig_Validate_EmptyOrganization(t *testing.T) {
	cfg := Config{
		Version: 1,
		Schema: SchemaConfig{
			Organization: "",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema organization must be")
}

func TestSchemaOrganization_Values(t *testing.T) {
	assert.Equal(t, SchemaOrganization("modular"), SchemaModular)
	assert.Equal(t, SchemaOrganization("components"), SchemaComponents)
	assert.Equal(t, SchemaOrganization("inline"), SchemaInline)
}

func TestConfig_LoadWithInheritance(t *testing.T) {
	tmpDir := t.TempDir()

	// Create parent config
	parentCfg := Config{
		Version: 1,
		Path:    "spec/spec.yaml",
		Schema: SchemaConfig{
			Organization: SchemaModular,
		},
	}
	parentPath := filepath.Join(tmpDir, "daco.yaml")
	require.NoError(t, parentCfg.Save(parentPath))

	// Create child directory and config
	childDir := filepath.Join(tmpDir, "child")
	require.NoError(t, os.MkdirAll(childDir, 0o750))

	childCfg := Config{
		Version: 1,
		Extends: "../daco.yaml",
	}
	childPath := filepath.Join(childDir, "daco.yaml")
	require.NoError(t, childCfg.Save(childPath))

	// Load child config with inheritance
	loaded, err := Load(childPath)
	require.NoError(t, err)

	// Should inherit parent's path and schema
	assert.Equal(t, "spec/spec.yaml", loaded.Path)
	assert.Equal(t, SchemaModular, loaded.Schema.Organization)
}

func TestConfig_LoadWithInheritance_Override(t *testing.T) {
	tmpDir := t.TempDir()

	// Create parent config
	parentCfg := Config{
		Version: 1,
		Path:    "spec/spec.yaml",
		Schema: SchemaConfig{
			Organization: SchemaModular,
		},
	}
	parentPath := filepath.Join(tmpDir, "daco.yaml")
	require.NoError(t, parentCfg.Save(parentPath))

	// Create child directory and config that overrides path
	childDir := filepath.Join(tmpDir, "child")
	require.NoError(t, os.MkdirAll(childDir, 0o750))

	childCfg := Config{
		Version: 1,
		Extends: "../daco.yaml",
		Path:    "local/spec.yaml",
		Schema: SchemaConfig{
			Organization: SchemaInline,
		},
	}
	childPath := filepath.Join(childDir, "daco.yaml")
	require.NoError(t, childCfg.Save(childPath))

	// Load child config with inheritance
	loaded, err := Load(childPath)
	require.NoError(t, err)

	// Should have child's overridden values
	assert.Equal(t, "local/spec.yaml", loaded.Path)
	assert.Equal(t, SchemaInline, loaded.Schema.Organization)
}

func TestConfig_LoadWithInheritance_CircularDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config that extends itself
	cfgPath := filepath.Join(tmpDir, "daco.yaml")
	cfg := Config{
		Version: 1,
		Extends: "daco.yaml",
	}
	require.NoError(t, cfg.Save(cfgPath))

	_, err := Load(cfgPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular extends detected")
}

func TestConfig_LoadWithInheritance_ParentNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Version: 1,
		Extends: "nonexistent.yaml",
	}
	cfgPath := filepath.Join(tmpDir, "daco.yaml")
	require.NoError(t, cfg.Save(cfgPath))

	_, err := Load(cfgPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load parent config")
}

func TestConfig_ValidateResolved(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid resolved config",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: SchemaModular,
				},
			},
			wantErr: "",
		},
		{
			name: "unsupported version",
			cfg: Config{
				Version: 99,
				Schema: SchemaConfig{
					Organization: SchemaModular,
				},
			},
			wantErr: "unsupported config version",
		},
		{
			name: "invalid schema organization",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: "invalid",
				},
			},
			wantErr: "schema organization must be",
		},
		{
			name: "empty schema organization",
			cfg: Config{
				Version: 1,
				Schema: SchemaConfig{
					Organization: "",
				},
			},
			wantErr: "schema organization must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateResolved()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_WithExtends(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid child config with extends",
			cfg: Config{
				Version: 1,
				Extends: "../parent.yaml",
			},
			wantErr: "",
		},
		{
			name: "child config with valid organization",
			cfg: Config{
				Version: 1,
				Extends: "../parent.yaml",
				Schema: SchemaConfig{
					Organization: SchemaInline,
				},
			},
			wantErr: "",
		},
		{
			name: "child config with invalid organization",
			cfg: Config{
				Version: 1,
				Extends: "../parent.yaml",
				Schema: SchemaConfig{
					Organization: "invalid",
				},
			},
			wantErr: "schema organization must be",
		},
		{
			name: "child config with unsupported version",
			cfg: Config{
				Version: 99,
				Extends: "../parent.yaml",
			},
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
