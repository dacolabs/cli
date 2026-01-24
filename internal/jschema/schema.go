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

// IsFileRef returns true if ref is an external file reference.
// File refs do not start with "#/".
func IsFileRef(ref string) bool {
	return ref != "" && !strings.HasPrefix(ref, "#/")
}
