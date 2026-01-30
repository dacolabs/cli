// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package avro

import (
	"encoding/json"
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	assert.Equal(t, "record", result["type"])
	assert.Equal(t, "UsersSchema", result["name"])
	assert.Equal(t, "schemas", result["namespace"])

	fields := result["fields"].([]any)
	fieldNames := extractFieldNames(fields)
	assert.Contains(t, fieldNames, "name")
	assert.Contains(t, fieldNames, "age")
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))
	assert.Equal(t, "string", fieldTypes["str"])
	assert.Equal(t, "long", fieldTypes["int"])
	assert.Equal(t, "double", fieldTypes["num"])
	assert.Equal(t, "boolean", fieldTypes["flag"])
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))

	// Logical types are maps
	createdAt := fieldTypes["created_at"].(map[string]any)
	assert.Equal(t, "long", createdAt["type"])
	assert.Equal(t, "timestamp-millis", createdAt["logicalType"])

	birthDate := fieldTypes["birth_date"].(map[string]any)
	assert.Equal(t, "int", birthDate["type"])
	assert.Equal(t, "date", birthDate["logicalType"])

	uuid := fieldTypes["uuid"].(map[string]any)
	assert.Equal(t, "string", uuid["type"])
	assert.Equal(t, "uuid", uuid["logicalType"])
}

func TestTranslate_ArrayType(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"tags"},
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))
	tagsType := fieldTypes["tags"].(map[string]any)
	assert.Equal(t, "array", tagsType["type"])
	assert.Equal(t, "string", tagsType["items"])
}

func TestTranslate_NestedObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"address"},
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))
	addressType := fieldTypes["address"].(map[string]any)
	assert.Equal(t, "record", addressType["type"])
	assert.Equal(t, "Address", addressType["name"])
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))
	addressType := fieldTypes["address"].(map[string]any)
	assert.Equal(t, "record", addressType["type"])
	assert.Equal(t, "Address", addressType["name"])
}

func TestTranslate_TopologicalOrder(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"customer"},
		Properties: map[string]*jsonschema.Schema{
			"customer": {Ref: "#/$defs/Customer"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Customer": {
				Type:     "object",
				Required: []string{"name", "address"},
				Properties: map[string]*jsonschema.Schema{
					"name":    {Type: "string"},
					"address": {Ref: "#/$defs/Address"},
				},
			},
			"Address": {
				Type:     "object",
				Required: []string{"street"},
				Properties: map[string]*jsonschema.Schema{
					"street": {Type: "string"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("order", schema, "schemas")
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	// Customer should be inlined in root, Address inlined inside Customer
	fields := result["fields"].([]any)
	customerField := findField(fields, "customer")
	require.NotNil(t, customerField, "customer field not found in root")
	customerType := customerField["type"].(map[string]any)
	assert.Equal(t, "Customer", customerType["name"])

	// Address should be inlined inside Customer
	customerFields := customerType["fields"].([]any)
	addressField := findField(customerFields, "address")
	require.NotNil(t, addressField, "address field not found in Customer")
	addressType := addressField["type"].(map[string]any)
	assert.Equal(t, "Address", addressType["name"])
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))

	// Required fields are plain types
	assert.Equal(t, "long", fieldTypes["id"])
	assert.Equal(t, "string", fieldTypes["name"])

	// Nullable field is a union
	optionalType := fieldTypes["optional"].([]any)
	assert.Equal(t, "null", optionalType[0])
	assert.Equal(t, "string", optionalType[1])
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	assert.Equal(t, "SimpleSchema", result["name"])
	fields := result["fields"].([]any)
	assert.Len(t, fields, 1)
}

func TestTranslate_ArrayOfObjects(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"items"},
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

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	fieldTypes := extractFieldTypes(result["fields"].([]any))
	itemsType := fieldTypes["items"].(map[string]any)
	assert.Equal(t, "array", itemsType["type"])

	itemsItems := itemsType["items"].(map[string]any)
	assert.Equal(t, "record", itemsItems["type"])
	assert.Equal(t, "Items", itemsItems["name"])
}

func TestTranslate_ChainedDeps(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"c"},
		Properties: map[string]*jsonschema.Schema{
			"c": {Ref: "#/$defs/C"},
		},
		Defs: map[string]*jsonschema.Schema{
			"C": {
				Type:     "object",
				Required: []string{"b"},
				Properties: map[string]*jsonschema.Schema{
					"b": {Ref: "#/$defs/B"},
				},
			},
			"B": {
				Type:     "object",
				Required: []string{"a"},
				Properties: map[string]*jsonschema.Schema{
					"a": {Ref: "#/$defs/A"},
				},
			},
			"A": {
				Type:     "object",
				Required: []string{"value"},
				Properties: map[string]*jsonschema.Schema{
					"value": {Type: "string"},
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("root", schema, "schemas")
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(output, &result))

	// C inlined in root, B inlined in C, A inlined in B
	fields := result["fields"].([]any)
	cField := findField(fields, "c")
	cType := cField["type"].(map[string]any)
	assert.Equal(t, "C", cType["name"])

	cFields := cType["fields"].([]any)
	bField := findField(cFields, "b")
	bType := bField["type"].(map[string]any)
	assert.Equal(t, "B", bType["name"])

	bFields := bType["fields"].([]any)
	aField := findField(bFields, "a")
	aType := aField["type"].(map[string]any)
	assert.Equal(t, "A", aType["name"])
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".avsc", translator.FileExtension())
}

// Test helpers

func extractFieldNames(fields []any) []string {
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		field := f.(map[string]any)
		names = append(names, field["name"].(string))
	}
	return names
}

func extractFieldTypes(fields []any) map[string]any {
	result := make(map[string]any)
	for _, f := range fields {
		field := f.(map[string]any)
		result[field["name"].(string)] = field["type"]
	}
	return result
}

func findField(fields []any, name string) map[string]any {
	for _, f := range fields {
		field := f.(map[string]any)
		if field["name"].(string) == name {
			return field
		}
	}
	return nil
}
