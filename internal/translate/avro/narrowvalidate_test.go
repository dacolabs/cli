// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package avro

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowValidate_FieldProps(t *testing.T) {
	minimum := 0.0
	maximum := 150.0
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"age"},
		Properties: map[string]*jsonschema.Schema{
			"age": {Type: "integer", Minimum: &minimum, Maximum: &maximum, Description: "Age in years"},
		},
	}

	output, err := (&Translator{}).Translate("person", schema, "schemas")
	require.NoError(t, err)

	result := string(output)
	// Narrowed to int, with description -> doc and constraints as custom props.
	assert.Contains(t, result, `"type": "int"`)
	assert.Contains(t, result, `"doc": "Age in years"`)
	assert.Contains(t, result, `"minimum": 0`)
	assert.Contains(t, result, `"maximum": 150`)
}
