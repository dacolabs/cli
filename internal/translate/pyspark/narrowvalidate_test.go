// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowValidate_Metadata(t *testing.T) {
	minimum := 0.0
	maximum := 150.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"age": {Type: "integer", Minimum: &minimum, Maximum: &maximum},
		},
	}

	output, err := (&Translator{}).Translate("person", schema, "schemas")
	require.NoError(t, err)

	result := string(output)
	assert.Contains(t, result, "T.ShortType()")
	assert.Contains(t, result, `metadata={"minimum": 0, "maximum": 150}`)
}
