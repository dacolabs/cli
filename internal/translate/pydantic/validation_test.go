// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pydantic

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_IntegerByteRangeEmitsField(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerByteRange(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "from pydantic import BaseModel, Field")
	assert.Contains(t, result, "count: int = Field(ge=0, le=127)")
}

func TestValidation_StringMaxLengthEmitsField(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringMaxLength(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "code: str = Field(max_length=32)")
}

func TestValidation_PatternEmitsField(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringPattern(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "pattern=")
	assert.Contains(t, result, "email: str = Field(pattern=")
}

func TestValidation_MultipleOfEmitsField(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.DecimalFromMultipleOf(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "price: float = Field(multiple_of=0.01)")
}

func TestValidation_NumberRangeEmitsField(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.NumberRange(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "ge=-273.15")
	assert.Contains(t, result, "le=1000")
}

func TestValidation_NoConstraintsNoFieldImport(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"name"},
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}
	out, err := (&Translator{}).Translate("t", schema, "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "from pydantic import BaseModel")
	assert.NotContains(t, result, "Field")
}
