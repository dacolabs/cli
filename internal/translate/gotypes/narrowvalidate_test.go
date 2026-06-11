// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package gotypes

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowValidate_StructTag(t *testing.T) {
	minimum := 0.0
	maximum := 150.0
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"age"},
		Properties: map[string]*jsonschema.Schema{
			"age": {Type: "integer", Minimum: &minimum, Maximum: &maximum},
		},
	}

	output, err := (&Translator{}).Translate("person", schema, "schemas")
	require.NoError(t, err)

	result := string(output)
	// Narrowed (non-negative) plus a go-playground validate tag alongside json.
	assert.Contains(t, result, "uint16")
	assert.Contains(t, result, "`json:\"age\" validate:\"min=0,max=150\"`")
}

func TestNarrowValidate_EnumOneof(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"color"},
		Properties: map[string]*jsonschema.Schema{
			"color": {Type: "string", Enum: []any{"red", "green", "blue"}},
		},
	}

	output, err := (&Translator{}).Translate("paint", schema, "schemas")
	require.NoError(t, err)

	assert.Contains(t, string(output), `validate:"oneof=red green blue"`)
}
