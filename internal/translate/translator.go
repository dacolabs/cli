// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package translate provides schema translation utilities.
package translate

import (
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

// Translator defines the interface all format translators must implement.
type Translator interface {
	// Name returns the translator's identifier (e.g., "pyspark", "dbt-sql")
	Name() string

	// Translate converts a JSON schema to the target format
	// portName is used to name the output schema variable (e.g., "users" -> "users_schema")
	Translate(portName string, schema *jsonschema.Schema) ([]byte, error)

	// FileExtension returns the appropriate file extension (e.g., ".py", ".sql")
	FileExtension() string
}

var translators = make(map[string]Translator)

// Register adds a translator to the registry.
func Register(t Translator) {
	translators[t.Name()] = t
}

// Get retrieves a translator by name.
func Get(name string) (Translator, error) {
	t, ok := translators[name]
	if !ok {
		return nil, fmt.Errorf("unknown translator: %s", name)
	}
	return t, nil
}

// Available returns all registered translator names.
func Available() []string {
	names := make([]string, 0, len(translators))
	for name := range translators {
		names = append(names, name)
	}
	return names
}
