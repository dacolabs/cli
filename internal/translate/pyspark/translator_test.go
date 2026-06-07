// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslate_SimpleObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("users", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Check imports
	assert.Contains(t, result, "import pyspark.sql.types as T")

	// Check main schema
	assert.Contains(t, result, "users_schema = T.StructType")
	assert.Contains(t, result, `"name"`)
	assert.Contains(t, result, `"age"`)
	assert.Contains(t, result, "T.StringType()")
	assert.Contains(t, result, "T.LongType()")
}

func TestTranslate_AllPrimitiveTypes(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"str":  {Type: "string"},
			"int":  {Type: "integer"},
			"num":  {Type: "number"},
			"bool": {Type: "boolean"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("types", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "T.StringType()")
	assert.Contains(t, result, "T.LongType()")
	assert.Contains(t, result, "T.DoubleType()")
	assert.Contains(t, result, "T.BooleanType()")
}

func TestTranslate_DateFormats(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"created_at": {Type: "string", Format: "date-time"},
			"birth_date": {Type: "string", Format: "date"},
			"uuid":       {Type: "string", Format: "uuid"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("dates", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "T.TimestampType()")
	assert.Contains(t, result, "T.DateType()")
	// UUID is still StringType
	assert.Contains(t, result, "T.StringType()")
}

func TestTranslate_ArrayType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"tags": {
				Type:  "array",
				Items: &jsonschema.Schema{Type: "string"},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("items", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "T.ArrayType(T.StringType())")
}

func TestTranslate_NestedObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"address": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"street": {Type: "string"},
					"city":   {Type: "string"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Inline object should be extracted as a separate variable
	assert.Contains(t, result, "_address = T.StructType")
	assert.Contains(t, result, `"address", _address`)
}

func TestTranslate_WithDefs_VariableReference(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"address": {Ref: "#/$defs/Address"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"street": {Type: "string"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Should have Address as a separate variable (prefixed with _)
	assert.Contains(t, result, "_address = T.StructType")
	// Main schema should reference the variable, not inline the schema
	assert.Contains(t, result, `"address", _address`)
}

func TestTranslate_TopologicalOrder(t *testing.T) {
	// Customer depends on Address, so Address should come first
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"customer": {Ref: "#/$defs/Customer"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Customer": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"name":    {Type: "string"},
					"address": {Ref: "#/$defs/Address"},
				},
			},
			"Address": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"street": {Type: "string"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("order", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Address should appear before Customer in the output
	addressIdx := strings.Index(result, "_address = T.StructType")
	customerIdx := strings.Index(result, "_customer = T.StructType")

	assert.Greater(t, addressIdx, 0, "Address definition should exist")
	assert.Greater(t, customerIdx, 0, "Customer definition should exist")
	assert.Less(t, addressIdx, customerIdx, "Address should come before Customer")
}

func TestTranslate_RequiredFields(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id", "name"},
		Properties: map[string]*jsonschema.Schema{
			"id":       {Type: "integer"},
			"name":     {Type: "string"},
			"optional": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Required fields should have nullable=False
	assert.Contains(t, result, `"id", T.LongType(), nullable=False`)
	assert.Contains(t, result, `"name", T.StringType(), nullable=False`)
	// Optional field should have nullable=True
	assert.Contains(t, result, `"optional", T.StringType(), nullable=True`)
}

func TestTranslate_ChainedDeps(t *testing.T) {
	// C -> B -> A (C depends on B, B depends on A)
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"c": {Ref: "#/$defs/C"},
		},
		Defs: map[string]*jsonschema.Schema{
			"C": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"b": {Ref: "#/$defs/B"},
				},
			},
			"B": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"a": {Ref: "#/$defs/A"},
				},
			},
			"A": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"value": {Type: "string"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("root", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Order should be A, B, C
	aIdx := strings.Index(result, "_a = T.StructType")
	bIdx := strings.Index(result, "_b = T.StructType")
	cIdx := strings.Index(result, "_c = T.StructType")

	assert.Greater(t, aIdx, 0)
	assert.Greater(t, bIdx, 0)
	assert.Greater(t, cIdx, 0)
	assert.Less(t, aIdx, bIdx, "A should come before B")
	assert.Less(t, bIdx, cIdx, "B should come before C")
}

func TestTranslate_NoDefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("simple", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Should just have the main schema
	assert.Contains(t, result, "simple_schema = T.StructType")
	// Count occurrences of " = T.StructType" (excluding imports)
	lines := strings.Split(result, "\n")
	schemaCount := 0
	for _, line := range lines {
		if strings.Contains(line, " = T.StructType") {
			schemaCount++
		}
	}
	assert.Equal(t, 1, schemaCount, "Should only have one schema definition")
}

func TestTranslate_ArrayOfObjects(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"items": {
				Type: "array",
				Items: &jsonschema.Schema{
					Type: "object",
					Properties: map[string]*jsonschema.Schema{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
					},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("list", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Array element object should be extracted as a separate variable
	assert.Contains(t, result, "_items = T.StructType")
	assert.Contains(t, result, "T.ArrayType(_items)")
}

func TestTranslate_RefToComponentsRewritten(t *testing.T) {
	// After parse.go processes, component refs are rewritten to $defs
	// This test verifies the translator works with properly rewritten refs
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"user": {Ref: "#/$defs/User"},
		},
		Defs: map[string]*jsonschema.Schema{
			"User": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"id": {Type: "integer"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Should reference User variable
	assert.Contains(t, result, "_user = T.StructType")
	assert.Contains(t, result, `"user", _user`)
}

func TestTranslate_DecimalType(t *testing.T) {
	multipleOf := 0.01
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"price": {Type: "number", MultipleOf: &multipleOf},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("order", schema, "schemas")
	require.NoError(t, err)

	result := string(output)
	assert.Contains(t, result, "T.DecimalType(38, 2)")
}

func TestTranslate_DecimalTypeVaryingScale(t *testing.T) {
	tests := []struct {
		multipleOf float64
		expected   string
	}{
		{0.1, "T.DecimalType(38, 1)"},
		{0.01, "T.DecimalType(38, 2)"},
		{0.001, "T.DecimalType(38, 3)"},
		{0.0001, "T.DecimalType(38, 4)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			multipleOf := tt.multipleOf
			schema := &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"value": {Type: "number", MultipleOf: &multipleOf},
				},
			}

			translator := &Translator{}
			output, err := translator.Translate("data", schema, "schemas")
			require.NoError(t, err)
			assert.Contains(t, string(output), tt.expected)
		})
	}
}

func TestTranslate_DecimalTypeRequired(t *testing.T) {
	multipleOf := 0.01
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"price"},
		Properties: map[string]*jsonschema.Schema{
			"price": {Type: "number", MultipleOf: &multipleOf},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("order", schema, "schemas")
	require.NoError(t, err)

	result := string(output)
	assert.Contains(t, result, `"price", T.DecimalType(38, 2), nullable=False`)
}

func TestTranslate_NumberWithoutConstraints_DoubleType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"value": {Type: "number"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.DoubleType()")
}

func TestTranslate_NumberWithBoundsIsDouble(t *testing.T) {
	min := -1000.0
	max := 1000.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"temperature": {Type: "number", Minimum: &min, Maximum: &max},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("sensor", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.DoubleType()")
}

func TestTranslate_IntegerType(t *testing.T) {
	min := 0.0
	max := 2147483647.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"count": {Type: "integer", Minimum: &min, Maximum: &max},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.IntegerType()")
}

func TestTranslate_ShortType(t *testing.T) {
	min := -32768.0
	max := 32767.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"small": {Type: "integer", Minimum: &min, Maximum: &max},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.ShortType()")
}

func TestTranslate_ByteType(t *testing.T) {
	min := 0.0
	max := 127.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"flags": {Type: "integer", Minimum: &min, Maximum: &max},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.ByteType()")
}

func TestTranslate_IntegerWithoutConstraints_LongType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"big": {Type: "integer"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.LongType()")
}

func TestTranslate_DecimalTypeWithMaximum(t *testing.T) {
	multipleOf := 0.01
	max := 99999.99
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"price": {Type: "number", MultipleOf: &multipleOf, Maximum: &max},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("order", schema, "schemas")
	require.NoError(t, err)

	// 99999 = 5 integer digits + 2 scale = precision 7
	assert.Contains(t, string(output), "T.DecimalType(7, 2)")
}

func TestTranslate_DecimalTypeWithSmallMaximum(t *testing.T) {
	multipleOf := 0.001
	max := 0.999
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"ratio": {Type: "number", MultipleOf: &multipleOf, Maximum: &max},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)

	// max < 1, so precision = scale = 3
	assert.Contains(t, string(output), "T.DecimalType(3, 3)")
}

func TestTranslate_MultipleOfWholeNumber_NoDecimal(t *testing.T) {
	multipleOf := 1.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"qty": {Type: "number", MultipleOf: &multipleOf},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	// multipleOf >= 1 should not trigger DecimalType
	assert.Contains(t, string(output), "T.DoubleType()")
}

func TestTranslate_MapType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"tags": {
				Type:                 "object",
				AdditionalProperties: &jsonschema.Schema{Type: "string"},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.MapType(T.StringType(), T.StringType())")
}

func TestTranslate_MapTypeWithIntegerValues(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"counts": {
				Type:                 "object",
				AdditionalProperties: &jsonschema.Schema{Type: "integer"},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("data", schema, "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(output), "T.MapType(T.StringType(), T.LongType())")
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".py", translator.FileExtension())
}
