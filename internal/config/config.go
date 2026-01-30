// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package config handles daco project configuration.
package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// CurrentConfigVersion is the current version of the config file format.
const CurrentConfigVersion = 1

// Config represents the daco.yaml project configuration file.
type Config struct {
	Version int    `yaml:"version"`
	Path    string `yaml:"path,omitempty"`
}

// Load reads a Config from a file path.
func Load(path string) (*Config, error) {
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

// Validate checks the configuration for required fields and valid values.
func (c *Config) Validate() error {
	if c.Version != CurrentConfigVersion {
		return errors.New("unsupported config version")
	}
	return nil
}
