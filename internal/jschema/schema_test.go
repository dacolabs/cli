// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema_test

import (
	"testing"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/stretchr/testify/assert"
)

func TestFormatFromPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want jschema.Format
	}{
		{"yaml extension", "schema.yaml", jschema.YAML},
		{"yml extension", "schema.yml", jschema.YAML},
		{"json extension", "schema.json", jschema.JSON},
		{"no extension", "schema", jschema.JSON},
		{"path with yaml", "/path/to/schema.yaml", jschema.YAML},
		{"path with json", "/path/to/schema.json", jschema.JSON},
		{"empty string", "", jschema.JSON},
		{"uppercase YAML", "schema.YAML", jschema.JSON}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jschema.FormatFromPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsFileRef(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want bool
	}{
		{"relative file ref", "./other.yaml", true},
		{"parent file ref", "../other.yaml", true},
		{"simple file ref", "other.yaml", true},
		{"internal ref", "#/$defs/address", false},
		{"internal component ref", "#/components/schemas/User", false},
		{"empty string", "", false},
		{"hash only", "#/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jschema.IsFileRef(tt.ref)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsInternalRef(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want bool
	}{
		{"defs ref", "#/$defs/address", true},
		{"components ref", "#/components/schemas/User", true},
		{"definitions ref", "#/definitions/item", true},
		{"file ref", "./other.yaml", false},
		{"simple file ref", "other.yaml", false},
		{"empty string", "", false},
		{"hash only", "#/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jschema.IsInternalRef(tt.ref)
			assert.Equal(t, tt.want, got)
		})
	}
}
