// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package dqxyaml

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectChecks_MultipleOf(t *testing.T) {
	multipleOf := 0.5
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"score"},
		Properties: map[string]*jsonschema.Schema{
			"score": {Type: "number", MultipleOf: &multipleOf},
		},
	}

	output, err := (&Translator{}).Translate("game", schema, "schemas")
	require.NoError(t, err)

	assert.Contains(t, string(output), "`score` % 0.5 = 0")
}
