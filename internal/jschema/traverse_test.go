// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema

import (
	"os"
	"testing"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func loadYAML(t *testing.T, data []byte) *jsonschema.Schema {
	t.Helper()
	var schema jsonschema.Schema
	require.NoError(t, yaml.Unmarshal(data, &schema))
	return &schema
}

func TestTraverse_SimpleSchema(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("simple.yaml")
	require.NoError(t, err)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 2 properties (name, age)
	assert.Equal(t, 3, count)
}

func TestTraverse_WithProperties(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("simple.yaml")
	require.NoError(t, err)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "object")
	assert.Contains(t, types, "string")
	assert.Contains(t, types, "integer")
}

func TestTraverse_WithAllOf(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("complex/allof.yaml")
	require.NoError(t, err)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 2 allOf schemas + 2 properties (base, extended)
	assert.Equal(t, 5, count)
}

func TestTraverse_WithAnyOf(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("complex/anyof.yaml")
	require.NoError(t, err)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 3 anyOf schemas + 1 property (value)
	assert.Equal(t, 5, count)
}

func TestTraverse_WithDefs(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("with-defs.yaml")
	require.NoError(t, err)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + address property (ref) + $defs/address + its 2 properties (street, city)
	assert.Equal(t, 5, count)
}

func TestTraverse_WithItems(t *testing.T) {
	data := []byte(`
type: array
items:
  type: object
  properties:
    item:
      type: string
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + items schema + 1 property (item)
	assert.Equal(t, 3, count)
}

func TestTraverse_CircularRefs(t *testing.T) {
	// Create a schema that references itself
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"self": nil, // Will be set to schema itself
		},
	}
	schema.Properties["self"] = schema

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Should handle cycle without infinite loop
	// Root + self (same as root, but visited once)
	assert.Equal(t, 1, count)
}

func TestTraverse_NilResolver(t *testing.T) {
	data := []byte(`
type: object
properties:
  ref:
    $ref: "#/$defs/other"
$defs:
  other:
    type: string
`)
	schema := loadYAML(t, data)

	var refs []string
	for s := range Traverse(schema, nil) {
		if s.Ref != "" {
			refs = append(refs, s.Ref)
		}
	}

	// Without resolver, we still see the ref schema but don't follow it
	assert.Contains(t, refs, "#/$defs/other")
}

func TestTraverse_WithResolver(t *testing.T) {
	targetSchema := &jsonschema.Schema{
		Type: "string",
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"ref": {Ref: "#/$defs/target"},
		},
	}

	resolver := func(ref string) *jsonschema.Schema {
		if ref == "#/$defs/target" {
			return targetSchema
		}
		return nil
	}

	var types []string
	for s := range Traverse(schema, resolver) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	// With resolver, we follow the ref and see the target schema
	assert.Contains(t, types, "object")
	assert.Contains(t, types, "string")
}

func TestTraverse_WithOneOf(t *testing.T) {
	data := []byte(`
oneOf:
  - type: string
  - type: integer
  - type: boolean
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 3 oneOf schemas
	assert.Equal(t, 4, count)
}

func TestTraverse_WithNot(t *testing.T) {
	data := []byte(`
type: object
not:
  type: array
`)
	schema := loadYAML(t, data)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "object")
	assert.Contains(t, types, "array")
}

func TestTraverse_WithConditional(t *testing.T) {
	data := []byte(`
type: object
if:
  properties:
    kind:
      const: "a"
then:
  properties:
    a_field:
      type: string
else:
  properties:
    b_field:
      type: integer
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + if + if/properties/kind + then + then/properties/a_field + else + else/properties/b_field
	assert.Equal(t, 7, count)
}

func TestTraverse_WithPatternProperties(t *testing.T) {
	data := []byte(`
type: object
patternProperties:
  "^S_":
    type: string
  "^I_":
    type: integer
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 2 pattern properties
	assert.Equal(t, 3, count)
}

func TestTraverse_WithAdditionalProperties(t *testing.T) {
	data := []byte(`
type: object
additionalProperties:
  type: string
`)
	schema := loadYAML(t, data)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "object")
	assert.Contains(t, types, "string")
}

func TestTraverse_WithPrefixItems(t *testing.T) {
	data := []byte(`
type: array
prefixItems:
  - type: string
  - type: integer
  - type: boolean
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 3 prefixItems
	assert.Equal(t, 4, count)
}

func TestTraverse_WithContains(t *testing.T) {
	data := []byte(`
type: array
contains:
  type: number
  minimum: 5
`)
	schema := loadYAML(t, data)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "array")
	assert.Contains(t, types, "number")
}

func TestTraverse_WithDefinitions(t *testing.T) {
	// definitions is the draft-07 equivalent of $defs
	data := []byte(`
type: object
definitions:
  name:
    type: string
  age:
    type: integer
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + 2 definitions
	assert.Equal(t, 3, count)
}

func TestTraverse_WithDependentSchemas(t *testing.T) {
	data := []byte(`
type: object
dependentSchemas:
  credit_card:
    properties:
      billing_address:
        type: string
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + dependentSchemas/credit_card + billing_address property
	assert.Equal(t, 3, count)
}

func TestTraverse_EarlyTermination(t *testing.T) {
	// Test that iterator can be stopped early
	data := []byte(`
type: object
properties:
  first:
    type: string
  second:
    type: integer
  third:
    type: boolean
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
		if count == 2 {
			break // Stop after 2 schemas
		}
	}

	assert.Equal(t, 2, count)
}

func TestTraverse_WithPropertyNames(t *testing.T) {
	data := []byte(`
type: object
propertyNames:
  pattern: "^[a-z]+$"
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + propertyNames schema
	assert.Equal(t, 2, count)
}

func TestTraverse_WithAdditionalItems(t *testing.T) {
	data := []byte(`
type: array
items:
  - type: string
  - type: integer
additionalItems:
  type: boolean
`)
	schema := loadYAML(t, data)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "array")
	assert.Contains(t, types, "boolean")
}

func TestTraverse_WithContentSchema(t *testing.T) {
	data := []byte(`
type: string
contentMediaType: application/json
contentSchema:
  type: object
  properties:
    data:
      type: string
`)
	schema := loadYAML(t, data)

	var count int
	for range Traverse(schema, nil) {
		count++
	}

	// Root + contentSchema + data property
	assert.Equal(t, 3, count)
}

func TestTraverse_WithUnevaluatedProperties(t *testing.T) {
	data := []byte(`
type: object
unevaluatedProperties:
  type: string
`)
	schema := loadYAML(t, data)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "object")
	assert.Contains(t, types, "string")
}

func TestTraverse_WithUnevaluatedItems(t *testing.T) {
	data := []byte(`
type: array
unevaluatedItems:
  type: number
`)
	schema := loadYAML(t, data)

	var types []string
	for s := range Traverse(schema, nil) {
		if s.Type != "" {
			types = append(types, s.Type)
		}
	}

	assert.Contains(t, types, "array")
	assert.Contains(t, types, "number")
}

func TestTraverse_EarlyTerminationAtVariousPoints(t *testing.T) {
	// Schema with many different keywords to test early termination at each point
	data := []byte(`
type: object
properties:
  prop1:
    type: string
patternProperties:
  "^x_":
    type: integer
additionalProperties:
  type: boolean
propertyNames:
  pattern: "^[a-z]"
unevaluatedProperties:
  type: number
allOf:
  - type: object
anyOf:
  - type: object
oneOf:
  - type: object
not:
  type: array
if:
  type: object
then:
  type: object
else:
  type: object
dependentSchemas:
  foo:
    type: object
contentSchema:
  type: object
$defs:
  def1:
    type: string
definitions:
  def2:
    type: integer
`)
	schema := loadYAML(t, data)

	// Test stopping at different iteration counts
	// Schema has 17 total nodes
	for stopAt := 1; stopAt <= 17; stopAt++ {
		var count int
		for range Traverse(schema, nil) {
			count++
			if count >= stopAt {
				break
			}
		}
		assert.Equal(t, stopAt, count, "should stop at %d", stopAt)
	}
}

func TestTraverse_EarlyTerminationWithArrayKeywords(t *testing.T) {
	data := []byte(`
type: array
items:
  type: string
prefixItems:
  - type: integer
  - type: boolean
additionalItems:
  type: number
contains:
  type: object
unevaluatedItems:
  type: array
`)
	schema := loadYAML(t, data)

	// Schema has 7 total nodes
	for stopAt := 1; stopAt <= 7; stopAt++ {
		var count int
		for range Traverse(schema, nil) {
			count++
			if count >= stopAt {
				break
			}
		}
		assert.Equal(t, stopAt, count)
	}
}

func TestTraverse_EarlyTerminationWithResolver(t *testing.T) {
	targetSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"nested": {Type: "string"},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"ref": {Ref: "#/$defs/target"},
		},
	}

	resolver := func(ref string) *jsonschema.Schema {
		if ref == "#/$defs/target" {
			return targetSchema
		}
		return nil
	}

	// Stop after visiting the ref but before following it
	var count int
	for range Traverse(schema, resolver) {
		count++
		if count >= 2 {
			break
		}
	}
	assert.Equal(t, 2, count)
}

// TraverseDefs tests

func TestTraverseDefs_Empty(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
	}

	var names []string //nolint:prealloc
	for name := range TraverseDefs(schema) {
		names = append(names, name)
	}

	assert.Empty(t, names)
}

func TestTraverseDefs_SingleDef(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Defs: map[string]*jsonschema.Schema{
			"Address": {Type: "object"},
		},
	}

	names := make([]string, 0, 1)
	for name := range TraverseDefs(schema) {
		names = append(names, name)
	}

	assert.Equal(t, []string{"Address"}, names)
}

func TestTraverseDefs_TopologicalOrder_Simple(t *testing.T) {
	// Address has no deps, Customer depends on Address
	schema := &jsonschema.Schema{
		Type: "object",
		Defs: map[string]*jsonschema.Schema{
			"Customer": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
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

	names := make([]string, 0, 2)
	for name := range TraverseDefs(schema) {
		names = append(names, name)
	}

	// Address should come before Customer (dependency first)
	assert.Equal(t, 2, len(names))
	addressIdx := -1
	customerIdx := -1
	for i, name := range names {
		if name == "Address" {
			addressIdx = i
		}
		if name == "Customer" {
			customerIdx = i
		}
	}
	assert.True(t, addressIdx < customerIdx, "Address should come before Customer, got order: %v", names)
}

func TestTraverseDefs_TopologicalOrder_Chain(t *testing.T) {
	// C depends on B, B depends on A
	schema := &jsonschema.Schema{
		Type: "object",
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

	names := make([]string, 0, 3)
	for name := range TraverseDefs(schema) {
		names = append(names, name)
	}

	// A should come first, then B, then C
	assert.Equal(t, 3, len(names))
	indices := make(map[string]int)
	for i, name := range names {
		indices[name] = i
	}
	assert.True(t, indices["A"] < indices["B"], "A should come before B")
	assert.True(t, indices["B"] < indices["C"], "B should come before C")
}

func TestTraverseDefs_TopologicalOrder_MultipleRefs(t *testing.T) {
	// Order depends on both Address and Phone
	schema := &jsonschema.Schema{
		Type: "object",
		Defs: map[string]*jsonschema.Schema{
			"Order": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"shipping": {Ref: "#/$defs/Address"},
					"contact":  {Ref: "#/$defs/Phone"},
				},
			},
			"Address": {Type: "object"},
			"Phone":   {Type: "object"},
		},
	}

	names := make([]string, 0, 3)
	for name := range TraverseDefs(schema) {
		names = append(names, name)
	}

	// Address and Phone should come before Order
	assert.Equal(t, 3, len(names))
	indices := make(map[string]int)
	for i, name := range names {
		indices[name] = i
	}
	assert.True(t, indices["Address"] < indices["Order"], "Address should come before Order")
	assert.True(t, indices["Phone"] < indices["Order"], "Phone should come before Order")
}

func TestTraverseDefs_NoDeps(t *testing.T) {
	// All independent schemas
	schema := &jsonschema.Schema{
		Type: "object",
		Defs: map[string]*jsonschema.Schema{
			"A": {Type: "string"},
			"B": {Type: "integer"},
			"C": {Type: "boolean"},
		},
	}

	names := make([]string, 0, 3)
	for name := range TraverseDefs(schema) {
		names = append(names, name)
	}

	// All 3 should be present (order doesn't matter when no deps)
	assert.Equal(t, 3, len(names))
	assert.Contains(t, names, "A")
	assert.Contains(t, names, "B")
	assert.Contains(t, names, "C")
}

func TestTraverseDefs_EarlyTermination(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Defs: map[string]*jsonschema.Schema{
			"A": {Type: "string"},
			"B": {Type: "integer"},
			"C": {Type: "boolean"},
		},
	}

	var count int
	for range TraverseDefs(schema) {
		count++
		if count >= 2 {
			break
		}
	}

	assert.Equal(t, 2, count)
}

func TestTraverseDefs_YieldsSchema(t *testing.T) {
	addressSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"street": {Type: "string"},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Defs: map[string]*jsonschema.Schema{
			"Address": addressSchema,
		},
	}

	for name, s := range TraverseDefs(schema) {
		assert.Equal(t, "Address", name)
		assert.Equal(t, addressSchema, s)
	}
}

// RewriteRefs tests

func TestRewriteRefs_ComponentsSchemas(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"address": {Ref: "#/components/schemas/Address"},
			"contact": {Ref: "#/components/schemas/Contact"},
		},
	}

	RewriteRefs(schema)

	assert.Equal(t, "#/$defs/Address", schema.Properties["address"].Ref)
	assert.Equal(t, "#/$defs/Contact", schema.Properties["contact"].Ref)
}

func TestRewriteRefs_Definitions(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"user": {Ref: "#/definitions/User"},
		},
	}

	RewriteRefs(schema)

	assert.Equal(t, "#/$defs/User", schema.Properties["user"].Ref)
}

func TestRewriteRefs_NestedRefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"order": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"customer": {Ref: "#/components/schemas/Customer"},
				},
			},
		},
	}

	RewriteRefs(schema)

	assert.Equal(t, "#/$defs/Customer", schema.Properties["order"].Properties["customer"].Ref)
}

func TestRewriteRefs_MixedRefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"component":  {Ref: "#/components/schemas/Component"},
			"definition": {Ref: "#/definitions/Definition"},
			"def":        {Ref: "#/$defs/AlreadyDef"},
		},
	}

	RewriteRefs(schema)

	assert.Equal(t, "#/$defs/Component", schema.Properties["component"].Ref)
	assert.Equal(t, "#/$defs/Definition", schema.Properties["definition"].Ref)
	assert.Equal(t, "#/$defs/AlreadyDef", schema.Properties["def"].Ref) // Unchanged
}

func TestRewriteRefs_PreservesDefsRefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"address": {Ref: "#/$defs/Address"},
		},
	}

	RewriteRefs(schema)

	// Should remain unchanged
	assert.Equal(t, "#/$defs/Address", schema.Properties["address"].Ref)
}

func TestRewriteRefs_InArrayItems(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "array",
		Items: &jsonschema.Schema{
			Ref: "#/components/schemas/Item",
		},
	}

	RewriteRefs(schema)

	assert.Equal(t, "#/$defs/Item", schema.Items.Ref)
}

func TestRewriteRefs_InAllOf(t *testing.T) {
	schema := &jsonschema.Schema{
		AllOf: []*jsonschema.Schema{
			{Ref: "#/components/schemas/Base"},
			{Ref: "#/definitions/Extended"},
		},
	}

	RewriteRefs(schema)

	assert.Equal(t, "#/$defs/Base", schema.AllOf[0].Ref)
	assert.Equal(t, "#/$defs/Extended", schema.AllOf[1].Ref)
}

func TestRewriteRefs_NoRefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	}

	// Should not panic and leave schema unchanged
	RewriteRefs(schema)

	assert.Equal(t, "", schema.Properties["name"].Ref)
	assert.Equal(t, "", schema.Properties["age"].Ref)
}
