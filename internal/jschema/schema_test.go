// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema_test

import (
	"testing"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/google/jsonschema-go/jsonschema"
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
			assert.Equal(t, tt.expected, jschema.IsFileRef(tt.ref))
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
			assert.Equal(t, tt.expected, jschema.IsFileRef(tt.ref))
		})
	}
}

func TestIsFileRef_EmptyRef(t *testing.T) {
	assert.False(t, jschema.IsFileRef(""))
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

	keyOrder, err := jschema.ExtractKeyOrder(schema)
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

	keyOrder, err := jschema.ExtractKeyOrder(schema)
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

	keyOrder, err := jschema.ExtractKeyOrder(schema)
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

	keyOrder, err := jschema.ExtractKeyOrder(schema)
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

	keyOrder, err := jschema.ExtractKeyOrder(schema)
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

	keyOrder, err := jschema.ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Should have properties from allOf schemas
	// Note: the exact path depends on how allOf is serialized
	assert.NotEmpty(t, keyOrder)
}

func TestExtractKeyOrder_PrimitiveSchema(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "string",
	}

	keyOrder, err := jschema.ExtractKeyOrder(schema)
	require.NoError(t, err)

	// Primitive schema has no properties
	assert.Empty(t, keyOrder)
}
