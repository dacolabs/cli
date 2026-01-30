// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package sparkscala

import (
	"strings"
	"testing"

	"github.com/dacolabs/jsonschema-go/jsonschema"
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

	assert.Contains(t, result, "import org.apache.spark.sql.types._")
	assert.Contains(t, result, "object Users extends Serializable {")
	assert.Contains(t, result, "lazy val users_schema: StructType = StructType(Array(")
	assert.Contains(t, result, `"name", StringType`)
	assert.Contains(t, result, `"age", LongType`)
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

	assert.Contains(t, result, "StringType")
	assert.Contains(t, result, "LongType")
	assert.Contains(t, result, "DoubleType")
	assert.Contains(t, result, "BooleanType")
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

	assert.Contains(t, result, "TimestampType")
	assert.Contains(t, result, "DateType")
	assert.Contains(t, result, "StringType")
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

	assert.Contains(t, string(output), "ArrayType(StringType)")
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

	assert.Contains(t, result, "lazy val _Address: StructType = StructType(Array(")
	assert.Contains(t, result, `"address", _Address`)
}

func TestTranslate_WithDefs(t *testing.T) {
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

	assert.Contains(t, result, "lazy val _Address: StructType = StructType(Array(")
	assert.Contains(t, result, `"address", _Address`)
}

func TestTranslate_TopologicalOrder(t *testing.T) {
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

	addressIdx := strings.Index(result, "lazy val _Address: StructType = StructType")
	customerIdx := strings.Index(result, "lazy val _Customer: StructType = StructType")

	assert.Greater(t, addressIdx, 0)
	assert.Greater(t, customerIdx, 0)
	assert.Less(t, addressIdx, customerIdx)
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

	assert.Contains(t, result, `"id", LongType, nullable = false`)
	assert.Contains(t, result, `"name", StringType, nullable = false`)
	assert.Contains(t, result, `"optional", StringType, nullable = true`)
}

func TestTranslate_ChainedDeps(t *testing.T) {
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

	aIdx := strings.Index(result, "lazy val _A: StructType = StructType")
	bIdx := strings.Index(result, "lazy val _B: StructType = StructType")
	cIdx := strings.Index(result, "lazy val _C: StructType = StructType")

	assert.Greater(t, aIdx, 0)
	assert.Greater(t, bIdx, 0)
	assert.Greater(t, cIdx, 0)
	assert.Less(t, aIdx, bIdx)
	assert.Less(t, bIdx, cIdx)
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

	assert.Contains(t, result, "lazy val simple_schema: StructType = StructType(Array(")
	lines := strings.Split(result, "\n")
	schemaCount := 0
	for _, line := range lines {
		if strings.Contains(line, "lazy val ") && strings.Contains(line, " = StructType") {
			schemaCount++
		}
	}
	assert.Equal(t, 1, schemaCount)
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

	assert.Contains(t, result, "lazy val _Items: StructType = StructType(Array(")
	assert.Contains(t, result, "ArrayType(_Items)")
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".scala", translator.FileExtension())
}

func TestTranslate_GeneratedHeader(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("test", schema, "schemas")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(string(output), "// Code generated by daco; DO NOT EDIT."))
	assert.Contains(t, string(output), "object Test extends Serializable {")
}
