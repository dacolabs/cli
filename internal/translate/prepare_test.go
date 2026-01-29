// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import (
	"testing"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubResolver is a minimal TypeResolver for testing Prepare logic.
type stubResolver struct{}

func (s *stubResolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		return format
	}
	return schemaType
}

func (s *stubResolver) ArrayType(elemType string) string {
	return "[]" + elemType
}

func (s *stubResolver) RefType(defName string) string {
	return defName
}

func (s *stubResolver) FormatDefName(defName string) string {
	return defName
}

func (s *stubResolver) FormatRootName(portName string) string {
	return portName + "_root"
}

func (s *stubResolver) EnrichField(_ *Field) {}

func TestPrepare_RootFields(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	}

	data, err := Prepare("users", schema, &stubResolver{})
	require.NoError(t, err)

	assert.Equal(t, "users_root", data.Root.Name)
	assert.Len(t, data.Root.Fields, 2)
	assert.Empty(t, data.Defs)

	// Find fields by name (order is non-deterministic without keyOrder)
	fieldsByName := make(map[string]Field)
	for _, f := range data.Root.Fields {
		fieldsByName[f.Name] = f
	}

	assert.Equal(t, "string", fieldsByName["name"].Type)
	assert.Equal(t, "integer", fieldsByName["age"].Type)
}

func TestPrepare_RequiredFields(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"id":       {Type: "integer"},
			"optional": {Type: "string"},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	fieldsByName := make(map[string]Field)
	for _, f := range data.Root.Fields {
		fieldsByName[f.Name] = f
	}

	assert.False(t, fieldsByName["id"].Nullable)
	assert.True(t, fieldsByName["optional"].Nullable)
}

func TestPrepare_Defs_TopologicalOrder(t *testing.T) {
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

	data, err := Prepare("order", schema, &stubResolver{})
	require.NoError(t, err)

	// Address should come before Customer (dependency order)
	require.Len(t, data.Defs, 2)
	assert.Equal(t, "Address", data.Defs[0].Name)
	assert.Equal(t, "Customer", data.Defs[1].Name)
}

func TestPrepare_RefResolution(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"addr": {Ref: "#/$defs/Address"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"city": {Type: "string"},
				},
			},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	// Root field type should be the ref name
	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "Address", data.Root.Fields[0].Type)
}

func TestPrepare_ArrayType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"tags": {
				Type:  "array",
				Items: &jsonschema.Schema{Type: "string"},
			},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "[]string", data.Root.Fields[0].Type)
}

func TestPrepare_ArrayWithoutItems(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"data": {Type: "array"},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "[]string", data.Root.Fields[0].Type)
}

func TestPrepare_FormatPassedToPrimitiveType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"created": {Type: "string", Format: "date-time"},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "date-time", data.Root.Fields[0].Type)
}

func TestPrepare_InlineObjectExtracted(t *testing.T) {
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

	data, err := Prepare("user", schema, &stubResolver{})
	require.NoError(t, err)

	// Root field should reference the extracted type
	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "Address", data.Root.Fields[0].Type)

	// Extracted type should appear in Defs
	require.Len(t, data.Defs, 1)
	assert.Equal(t, "Address", data.Defs[0].Name)
	assert.Len(t, data.Defs[0].Fields, 2)
}

func TestPrepare_InlineObjectInArray(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"items": {
				Type: "array",
				Items: &jsonschema.Schema{
					Type: "object",
					Properties: map[string]*jsonschema.Schema{
						"id": {Type: "integer"},
					},
				},
			},
		},
	}

	data, err := Prepare("list", schema, &stubResolver{})
	require.NoError(t, err)

	// Root field should be array of extracted type
	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "[]Items", data.Root.Fields[0].Type)

	// Extracted type should appear in Defs
	require.Len(t, data.Defs, 1)
	assert.Equal(t, "Items", data.Defs[0].Name)
}

func TestPrepare_ImplicitObject(t *testing.T) {
	// Object without explicit type but with properties
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"meta": {
				Properties: map[string]*jsonschema.Schema{
					"version": {Type: "string"},
				},
			},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	// Should be extracted as a type
	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "Meta", data.Root.Fields[0].Type)
	require.Len(t, data.Defs, 1)
	assert.Equal(t, "Meta", data.Defs[0].Name)
}

func TestPrepare_EnrichFieldCalled(t *testing.T) {
	// Use a resolver that mutates the field
	enrichResolver := &enrichingResolver{}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	data, err := Prepare("test", schema, enrichResolver)
	require.NoError(t, err)

	require.Len(t, data.Root.Fields, 1)
	assert.Equal(t, "enriched", data.Root.Fields[0].Tag)
}

type enrichingResolver struct{ stubResolver }

func (e *enrichingResolver) EnrichField(f *Field) {
	f.Tag = "enriched"
}

func TestPrepare_ExtraInitialized(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"x": {Type: "string"},
		},
	}

	data, err := Prepare("test", schema, &stubResolver{})
	require.NoError(t, err)

	assert.NotNil(t, data.Extra)
}

func TestPrepare_NoDefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	data, err := Prepare("simple", schema, &stubResolver{})
	require.NoError(t, err)

	assert.Empty(t, data.Defs)
	assert.Equal(t, "simple_root", data.Root.Name)
}

func TestPrepare_ChainedDefs(t *testing.T) {
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

	data, err := Prepare("root", schema, &stubResolver{})
	require.NoError(t, err)

	require.Len(t, data.Defs, 3)
	assert.Equal(t, "A", data.Defs[0].Name)
	assert.Equal(t, "B", data.Defs[1].Name)
	assert.Equal(t, "C", data.Defs[2].Name)
}

func TestExtractDefName(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"#/$defs/Address", "Address"},
		{"#/$defs/User", "User"},
		{"some-other-ref", "some-other-ref"},
		{"", ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, extractDefName(tt.ref), "extractDefName(%q)", tt.ref)
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"address", "Address"},
		{"street_name", "StreetName"},
		{"first-name", "FirstName"},
		{"simple", "Simple"},
		{"a_b_c", "ABC"},
		{"", ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, ToPascalCase(tt.input), "ToPascalCase(%q)", tt.input)
	}
}
