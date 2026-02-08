// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package markdown

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

	assert.Contains(t, result, "# Users")
	assert.Contains(t, result, "| `name` | string | No |")
	assert.Contains(t, result, "| `age` | integer | No |")
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

	assert.Contains(t, result, "| `str` | string | Yes |")
	assert.Contains(t, result, "| `int` | integer | Yes |")
	assert.Contains(t, result, "| `num` | number | Yes |")
	assert.Contains(t, result, "| `flag` | boolean | Yes |")
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

	assert.Contains(t, result, "| `created_at` | string | Yes |")
	assert.Contains(t, result, "format: date-time")
	assert.Contains(t, result, "| `birth_date` | string | Yes |")
	assert.Contains(t, result, "format: date")
	assert.Contains(t, result, "| `uuid` | string | Yes |")
	assert.Contains(t, result, "format: uuid")
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

	assert.Contains(t, result, "array(string)")
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

	assert.Contains(t, result, "## Address")
	assert.Contains(t, result, "[Address](#Address)")
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

	assert.Contains(t, result, "## Address")
	assert.Contains(t, result, "[Address](#Address)")
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

	addressIdx := strings.Index(result, "## Address")
	customerIdx := strings.Index(result, "## Customer")

	assert.Greater(t, addressIdx, 0)
	assert.Greater(t, customerIdx, 0)
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

	assert.Contains(t, result, "| `id` | integer | Yes |")
	assert.Contains(t, result, "| `name` | string | Yes |")
	assert.Contains(t, result, "| `optional` | string | No |")
}

func TestTranslate_WithDescription(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:        "object",
		Description: "This is a user schema",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string", Description: "The user's full name"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "This is a user schema")
	assert.Contains(t, result, "The user's full name")
}

func TestTranslate_WithConstraints(t *testing.T) {
	minLen := 1
	maxLen := 100
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"email": {
				Type:      "string",
				Format:    "email",
				MinLength: &minLen,
				MaxLength: &maxLen,
			},
			"status": {
				Type: "string",
				Enum: []any{"active", "inactive", "pending"},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "format: email")
	assert.Contains(t, result, "minLength: 1")
	assert.Contains(t, result, "maxLength: 100")
	assert.Contains(t, result, "`active`")
	assert.Contains(t, result, "`inactive`")
	assert.Contains(t, result, "`pending`")
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

	assert.Contains(t, result, "## Items")
	assert.Contains(t, result, "array([Items](#Items))")
}

func TestTranslate_Title(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"x": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("test_schema", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "# TestSchema")
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".md", translator.FileExtension())
}
