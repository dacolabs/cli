// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databrickssql

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowValidate_RangeAndMultipleOfChecks(t *testing.T) {
	minimum := 0.0
	maximum := 150.0
	multipleOf := 0.01
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"age":   {Type: "integer", Minimum: &minimum, Maximum: &maximum},
			"price": {Type: "number", MultipleOf: &multipleOf},
		},
	}

	output, err := (&Translator{}).Translate("orders", schema, "schemas")
	require.NoError(t, err)

	result := string(output)
	// Range CHECK survives even though the type narrowed (TINYINT/SMALLINT approximates).
	assert.Contains(t, result, "`age` BETWEEN 0 AND 150")
	// multipleOf becomes a modulo CHECK.
	assert.Contains(t, result, "`price` % 0.01 = 0")
}
