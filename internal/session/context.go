// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package session provides project context loading for CLI commands.
package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dacolabs/cli/internal/config"
	"github.com/dacolabs/cli/internal/opendpi"
)

var (
	// ErrNotInitialized indicates no daco.yaml was found in the current directory.
	ErrNotInitialized = errors.New("not in a daco project (daco.yaml not found)")

	// ErrInvalidConfig indicates the config file exists but is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrSpecNotFound indicates the spec file referenced by config doesn't exist.
	ErrSpecNotFound = errors.New("spec file not found")

	// ErrInvalidSpec indicates the spec file exists but couldn't be parsed.
	ErrInvalidSpec = errors.New("invalid OpenDPI spec")
)

// ConfigFileName is the name of the daco configuration file.
const ConfigFileName = "daco.yaml"

// contextKey is used to store Context in context.Context.
type contextKey struct{}

// Context holds the resolved project configuration and parsed OpenDPI spec.
type Context struct {
	// Config is the fully resolved configuration (with inheritance applied).
	Config *config.Config

	// Spec is the parsed OpenDPI specification.
	Spec *opendpi.Spec
}

// Load loads the project context from the current working directory and
// returns a new context.Context with the daco Context stored in it.
func Load(ctx context.Context) (context.Context, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, ConfigFileName)
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return nil, ErrNotInitialized
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	if validateErr := cfg.ValidateResolved(); validateErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, validateErr)
	}

	specDir := cfg.Path
	if !filepath.IsAbs(specDir) {
		specDir = filepath.Join(cwd, specDir)
	}

	specPath, parser := findSpecFile(specDir)
	if specPath == "" {
		return nil, fmt.Errorf("%w: no opendpi.yaml or opendpi.json in %s", ErrSpecNotFound, specDir)
	}

	f, err := os.Open(specPath) //nolint:gosec // specPath is derived from config
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSpecNotFound, err)
	}
	defer func() { _ = f.Close() }()

	spec, err := parser.Parse(f, os.DirFS(specDir))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSpec, err)
	}

	dacoCtx := &Context{
		Config: cfg,
		Spec:   spec,
	}

	return context.WithValue(ctx, contextKey{}, dacoCtx), nil
}

// findSpecFile looks for opendpi.yaml or opendpi.json in the given directory.
// Returns the path and parser if found, or empty string if not found.
func findSpecFile(dir string) (string, opendpi.Parser) {
	yamlPath := filepath.Join(dir, "opendpi.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		return yamlPath, opendpi.YAML
	}

	jsonPath := filepath.Join(dir, "opendpi.json")
	if _, err := os.Stat(jsonPath); err == nil {
		return jsonPath, opendpi.JSON
	}

	return "", opendpi.Parser{}
}

// From extracts the daco Context from a context.Context.
// Returns nil if no Context is stored.
func From(ctx context.Context) *Context {
	if dacoCtx, ok := ctx.Value(contextKey{}).(*Context); ok {
		return dacoCtx
	}
	return nil
}
