// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema

import (
	"testing"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsFileRef_ExternalRefs(t *testing.T) {
	tests := []struct {
		ref      string
		expected bool
	}{
		{"./schema.yaml", true},
		{"../schemas/address.yaml", true},
		{"schemas/user.json", true},
		{"external.yaml", true},
		{"/absolute/path.yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsFileRef(tt.ref))
		})
	}
}

func TestIsFileRef_InternalRefs(t *testing.T) {
	tests := []struct {
		ref      string
		expected bool
	}{
		{"#/$defs/Address", false},
		{"#/components/schemas/User", false},
		{"#/definitions/Order", false},
		{"#/properties/name", false},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsFileRef(tt.ref))
		})
	}
}

func TestIsFileRef_EmptyRef(t *testing.T) {
	assert.False(t, IsFileRef(""))
}

func TestExtractKeyOrder_SimpleObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
			"city": {Type: "string"},
		},
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Should have a "properties" key
	assert.Contains(t, keyOrder, "properties")
	// Should contain all property names
	props := keyOrder["properties"]
	assert.Contains(t, props, "name")
	assert.Contains(t, props, "age")
	assert.Contains(t, props, "city")
}

func TestExtractKeyOrder_NestedObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"user": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"firstName": {Type: "string"},
					"lastName":  {Type: "string"},
				},
			},
		},
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Should have root properties
	assert.Contains(t, keyOrder, "properties")
	// Should have nested properties
	assert.Contains(t, keyOrder, "properties.user.properties")
	nestedProps := keyOrder["properties.user.properties"]
	assert.Contains(t, nestedProps, "firstName")
	assert.Contains(t, nestedProps, "lastName")
}

func TestExtractKeyOrder_WithDefs(t *testing.T) {
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
					"city":   {Type: "string"},
					"zip":    {Type: "string"},
				},
			},
		},
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Should have root properties
	assert.Contains(t, keyOrder, "properties")
	// Should have $defs/Address properties
	assert.Contains(t, keyOrder, "$defs.Address.properties")
	defProps := keyOrder["$defs.Address.properties"]
	assert.Contains(t, defProps, "street")
	assert.Contains(t, defProps, "city")
	assert.Contains(t, defProps, "zip")
}

func TestExtractKeyOrder_EmptySchema(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Empty schema has no properties key
	assert.NotContains(t, keyOrder, "properties")
}

func TestExtractKeyOrder_ArraySchema(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "array",
		Items: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id":   {Type: "integer"},
				"name": {Type: "string"},
			},
		},
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Should have items properties
	assert.Contains(t, keyOrder, "items.properties")
	itemProps := keyOrder["items.properties"]
	assert.Contains(t, itemProps, "id")
	assert.Contains(t, itemProps, "name")
}

func TestExtractKeyOrder_AllOf(t *testing.T) {
	schema := &jsonschema.Schema{
		AllOf: []*jsonschema.Schema{
			{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"base": {Type: "string"},
				},
			},
			{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"extended": {Type: "string"},
				},
			},
		},
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Should have properties from allOf schemas
	// Note: the exact path depends on how allOf is serialized
	assert.NotEmpty(t, keyOrder)
}

func TestExtractKeyOrder_PrimitiveSchema(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "string",
	}

	keyOrder, err := ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Primitive schema has no properties
	assert.Empty(t, keyOrder)
}

func TestExtractKeyOrderFromJSON_PreservesOrder(t *testing.T) {
	// JSON with intentionally non-alphabetical property order
	data := []byte(`{
		"type": "object",
		"properties": {
			"zeta": {"type": "string"},
			"alpha": {"type": "integer"},
			"mu": {"type": "boolean"}
		}
	}`)

	keyOrder, err := ExtractKeyOrderFromJSON(data)
	require.NoError(t, err)

	assert.Contains(t, keyOrder, "properties")
	assert.Equal(t, []string{"zeta", "alpha", "mu"}, keyOrder["properties"])
}

func TestExtractKeyOrderFromJSON_NestedObject(t *testing.T) {
	data := []byte(`{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"lastName": {"type": "string"},
					"firstName": {"type": "string"}
				}
			}
		}
	}`)

	keyOrder, err := ExtractKeyOrderFromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, []string{"user"}, keyOrder["properties"])
	assert.Equal(t, []string{"lastName", "firstName"}, keyOrder["properties.user.properties"])
}

func TestExtractKeyOrderFromJSON_WithDefs(t *testing.T) {
	data := []byte(`{
		"type": "object",
		"$defs": {
			"Address": {
				"type": "object",
				"properties": {
					"zip": {"type": "string"},
					"street": {"type": "string"},
					"city": {"type": "string"}
				}
			}
		},
		"properties": {
			"address": {"$ref": "#/$defs/Address"}
		}
	}`)

	keyOrder, err := ExtractKeyOrderFromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, []string{"zip", "street", "city"}, keyOrder["$defs.Address.properties"])
	assert.Equal(t, []string{"address"}, keyOrder["properties"])
}

func TestExtractKeyOrderFromYAML_PreservesOrder(t *testing.T) {
	data := []byte(`type: object
properties:
  zeta:
    type: string
  alpha:
    type: integer
  mu:
    type: boolean
`)

	keyOrder, err := ExtractKeyOrderFromYAML(data)
	require.NoError(t, err)

	assert.Contains(t, keyOrder, "properties")
	assert.Equal(t, []string{"zeta", "alpha", "mu"}, keyOrder["properties"])
}

func TestExtractKeyOrderFromYAML_NestedObject(t *testing.T) {
	data := []byte(`type: object
properties:
  user:
    type: object
    properties:
      lastName:
        type: string
      firstName:
        type: string
`)

	keyOrder, err := ExtractKeyOrderFromYAML(data)
	require.NoError(t, err)

	assert.Equal(t, []string{"user"}, keyOrder["properties"])
	assert.Equal(t, []string{"lastName", "firstName"}, keyOrder["properties.user.properties"])
}

func TestExtractKeyOrderFromYAML_WithDefs(t *testing.T) {
	data := []byte(`type: object
$defs:
  Address:
    type: object
    properties:
      zip:
        type: string
      street:
        type: string
      city:
        type: string
properties:
  address:
    $ref: "#/$defs/Address"
`)

	keyOrder, err := ExtractKeyOrderFromYAML(data)
	require.NoError(t, err)

	assert.Equal(t, []string{"zip", "street", "city"}, keyOrder["$defs.Address.properties"])
	assert.Equal(t, []string{"address"}, keyOrder["properties"])
}

func TestSetPropertyOrder_Simple(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"zeta":  {Type: "string"},
			"alpha": {Type: "integer"},
			"mu":    {Type: "boolean"},
		},
	}

	keyOrder := map[string][]string{
		"properties": {"zeta", "alpha", "mu"},
	}

	SetPropertyOrder(schema, keyOrder)

	assert.Equal(t, []string{"zeta", "alpha", "mu"}, schema.PropertyOrder)
}

func TestSetPropertyOrder_Nested(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"user": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"lastName":  {Type: "string"},
					"firstName": {Type: "string"},
				},
			},
		},
	}

	keyOrder := map[string][]string{
		"properties":                 {"user"},
		"properties.user.properties": {"lastName", "firstName"},
	}

	SetPropertyOrder(schema, keyOrder)

	assert.Equal(t, []string{"user"}, schema.PropertyOrder)
	assert.Equal(t, []string{"lastName", "firstName"}, schema.Properties["user"].PropertyOrder)
}

func TestSetPropertyOrder_WithDefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"address": {Ref: "#/$defs/Address"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"zip":    {Type: "string"},
					"street": {Type: "string"},
					"city":   {Type: "string"},
				},
			},
		},
	}

	keyOrder := map[string][]string{
		"properties":               {"address"},
		"$defs.Address.properties": {"zip", "street", "city"},
	}

	SetPropertyOrder(schema, keyOrder)

	assert.Equal(t, []string{"address"}, schema.PropertyOrder)
	assert.Equal(t, []string{"zip", "street", "city"}, schema.Defs["Address"].PropertyOrder)
}

func TestSetPropertyOrder_FiltersNonExistentKeys(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	keyOrder := map[string][]string{
		"properties": {"name", "removed_field"},
	}

	SetPropertyOrder(schema, keyOrder)

	assert.Equal(t, []string{"name"}, schema.PropertyOrder)
}
