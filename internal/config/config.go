// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package config handles daco project configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SchemaOrganization defines how schemas are organized in the spec.
type SchemaOrganization string

const (
	// SchemaModular stores schemas in separate files under spec/schemas/.
	SchemaModular SchemaOrganization = "modular"
	// SchemaComponents stores schemas in components.schemas section.
	SchemaComponents SchemaOrganization = "components"
	// SchemaInline stores schemas directly in the port definition.
	SchemaInline SchemaOrganization = "inline"
)

// CurrentConfigVersion is the current version of the config file format.
const CurrentConfigVersion = 1

// Config represents the daco.yaml project configuration file.
type Config struct {
	Version int          `yaml:"version"`
	Extends string       `yaml:"extends,omitempty"`
	Path    string       `yaml:"path,omitempty"`
	Schema  SchemaConfig `yaml:"schema,omitempty"`
}

// SchemaConfig contains schema-related settings.
type SchemaConfig struct {
	Organization SchemaOrganization `yaml:"organization,omitempty"`
}

// Load reads a Config from a file path and resolves the extends chain.
// It returns a fully resolved config with all inherited values.
func Load(path string) (*Config, error) {
	return load(path, make(map[string]bool))
}

// Save writes the Config to a file path.
func (c *Config) Save(path string) error {
	f, err := os.Create(path) //nolint:gosec // path is provided by caller
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(c)
}

// loadFile reads a single config file without resolving inheritance.
func loadFile(path string) (*Config, error) {
	f, err := os.Open(path) //nolint:gosec // path is provided by caller
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func load(path string, visited map[string]bool) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	if visited[absPath] {
		return nil, errors.New("circular extends detected")
	}
	visited[absPath] = true

	cfg, err := loadFile(path)
	if err != nil {
		return nil, err
	}

	if cfg.Extends == "" {
		return cfg, nil
	}

	// Resolve extends path relative to the config file's directory
	configDir := filepath.Dir(absPath)
	parentPath := filepath.Join(configDir, cfg.Extends)

	parent, err := load(parentPath, visited)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent config %s: %w", cfg.Extends, err)
	}

	// Merge: child values override parent values
	merged := &Config{
		Version: cfg.Version,
		Extends: cfg.Extends,
		Path:    parent.Path,
		Schema:  parent.Schema,
	}

	if cfg.Path != "" {
		merged.Path = cfg.Path
	}
	if cfg.Schema.Organization != "" {
		merged.Schema.Organization = cfg.Schema.Organization
	}

	return merged, nil
}

// Validate checks the configuration for required fields and valid values.
// For configs with extends, validation is relaxed as values may come from parent.
func (c *Config) Validate() error {
	if c.Version != CurrentConfigVersion {
		return errors.New("unsupported config version")
	}

	// If this config extends another, fields are optional
	if c.Extends != "" {
		// Only validate fields that are set
		if c.Schema.Organization != "" {
			switch c.Schema.Organization {
			case SchemaModular, SchemaComponents, SchemaInline:
				// valid
			default:
				return errors.New("schema organization must be 'modular', 'components', or 'inline'")
			}
		}
		return nil
	}

	// Root config requires all fields
	switch c.Schema.Organization {
	case SchemaModular, SchemaComponents, SchemaInline:
		// valid
	default:
		return errors.New("schema organization must be 'modular', 'components', or 'inline'")
	}
	return nil
}

// ValidateResolved checks a fully resolved config (after inheritance).
func (c *Config) ValidateResolved() error {
	if c.Version != CurrentConfigVersion {
		return errors.New("unsupported config version")
	}
	switch c.Schema.Organization {
	case SchemaModular, SchemaComponents, SchemaInline:
		// valid
	default:
		return errors.New("schema organization must be 'modular', 'components', or 'inline'")
	}
	return nil
}
