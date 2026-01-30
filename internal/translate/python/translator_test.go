// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package python

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

	assert.Contains(t, result, "@dataclasses.dataclass")
	assert.Contains(t, result, "class UsersSchema:")
	assert.Contains(t, result, "name: Optional[str] = None")
	assert.Contains(t, result, "age: Optional[int] = None")
}

func TestTranslate_AllPrimitiveTypes(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"str", "int", "num", "flag"},
		Properties: map[string]*jsonschema.Schema{
			"str":  {Type: "string"},
			"int":  {Type: "integer"},
			"num":  {Type: "number"},
			"flag": {Type: "boolean"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("types", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "str: str")
	assert.Contains(t, result, "int: int")
	assert.Contains(t, result, "num: float")
	assert.Contains(t, result, "flag: bool")
}

func TestTranslate_DateFormats(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"created_at", "birth_date", "uuid"},
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

	assert.Contains(t, result, "created_at: datetime.datetime")
	assert.Contains(t, result, "birth_date: datetime.date")
	assert.Contains(t, result, "uuid: str")
	assert.Contains(t, result, "import datetime")
}

func TestTranslate_NoDatetimeImport(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("test", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.NotContains(t, result, "import datetime")
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

	assert.Contains(t, result, "Optional[list[str]]")
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

	assert.Contains(t, result, "class Address:")
	assert.Contains(t, result, "Optional[Address]")
}

func TestTranslate_WithDefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"address"},
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

	assert.Contains(t, result, "class Address:")
	assert.Contains(t, result, "address: Address")
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

	addressIdx := strings.Index(result, "class Address:")
	customerIdx := strings.Index(result, "class Customer:")

	assert.Greater(t, addressIdx, 0)
	assert.Greater(t, customerIdx, 0)
	assert.Less(t, addressIdx, customerIdx, "Address should come before Customer")
}

func TestTranslate_NullableFields(t *testing.T) {
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

	assert.Contains(t, result, "id: int")
	assert.Contains(t, result, "name: str")
	assert.Contains(t, result, "optional: Optional[str] = None")
}

func TestTranslate_FieldOrder(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"optional": {Type: "string"},
			"id":       {Type: "integer"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("test", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	idIdx := strings.Index(result, "id: int")
	optIdx := strings.Index(result, "optional: Optional[str] = None")

	assert.Greater(t, idIdx, 0)
	assert.Greater(t, optIdx, 0)
	assert.Less(t, idIdx, optIdx, "required fields should come before optional fields")
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

	assert.Contains(t, result, "class SimpleSchema:")
	classCount := strings.Count(result, "class ")
	assert.Equal(t, 1, classCount, "Should only have one class definition")
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

	assert.Contains(t, result, "class Items:")
	assert.Contains(t, result, "Optional[list[Items]]")
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

	aIdx := strings.Index(result, "class A:")
	bIdx := strings.Index(result, "class B:")
	cIdx := strings.Index(result, "class C:")

	assert.Greater(t, aIdx, 0)
	assert.Greater(t, bIdx, 0)
	assert.Greater(t, cIdx, 0)
	assert.Less(t, aIdx, bIdx, "A should come before B")
	assert.Less(t, bIdx, cIdx, "B should come before C")
}

func TestTranslate_GeneratedHeader(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"x": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("test", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "# Code generated by daco; DO NOT EDIT.")
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".py", translator.FileExtension())
}
