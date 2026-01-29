// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package translate provides schema translation utilities.
package translate

import (
	"fmt"

	"github.com/dacolabs/jsonschema-go/jsonschema"
)

// Translator defines the interface all format translators must implement.
type Translator interface {
	// Translate converts a JSON schema to the target format.
	// portName is used to name the output schema variable (e.g., "users" -> "users_schema").
	// outputDir is the output directory path, used to derive package names where applicable.
	Translate(portName string, schema *jsonschema.Schema, outputDir string) ([]byte, error)

	// FileExtension returns the appropriate file extension (e.g., ".py", ".sql")
	FileExtension() string
}

// Register maps format names to their translators.
type Register map[string]Translator

// Get retrieves a translator by name.
func (r Register) Get(name string) (Translator, error) {
	t, ok := r[name]
	if !ok {
		return nil, fmt.Errorf("unknown translator: %s", name)
	}
	return t, nil
}

// Available returns all registered translator names.
func (r Register) Available() []string {
	names := make([]string, 0, len(r))
	for name := range r {
		names = append(names, name)
	}
	return names
}
