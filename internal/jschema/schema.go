// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package jschema provides JSON Schema loading, parsing, and traversal utilities.
package jschema

import (
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// Schema is the JSON Schema type.
type Schema = jsonschema.Schema

// Format represents the serialization format for schemas.
type Format int

const (
	// JSON format.
	JSON Format = iota
	// YAML format.
	YAML
)

// FormatFromPath determines the format from a file path extension.
// Returns YAML for .yaml and .yml extensions, JSON otherwise.
func FormatFromPath(path string) Format {
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return YAML
	}
	return JSON
}

// IsFileRef returns true if ref is an external file reference.
// File refs do not start with "#/".
func IsFileRef(ref string) bool {
	return ref != "" && !strings.HasPrefix(ref, "#/")
}

// IsInternalRef returns true if ref is an internal reference (starts with "#/").
func IsInternalRef(ref string) bool {
	return strings.HasPrefix(ref, "#/")
}
