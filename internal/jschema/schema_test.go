// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema_test

import (
	"testing"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/stretchr/testify/assert"
)

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
