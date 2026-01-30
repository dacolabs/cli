// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package scala

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

	assert.Contains(t, result, "case class UsersSchema(")
	assert.Contains(t, result, "name: Option[String]")
	assert.Contains(t, result, "age: Option[Long]")
}

func TestTranslate_AllPrimitiveTypes(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"str", "int", "num", "bool"},
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

	assert.Contains(t, result, "str: String")
	assert.Contains(t, result, "int: Long")
	assert.Contains(t, result, "num: Double")
	assert.Contains(t, result, "bool: Boolean")
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

	assert.Contains(t, result, "java.time.Instant")
	assert.Contains(t, result, "java.time.LocalDate")
	assert.Contains(t, result, "String")
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

	assert.Contains(t, string(output), "List[String]")
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

	assert.Contains(t, result, "case class Address(")
	assert.Contains(t, result, "address: Option[Address]")
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

	assert.Contains(t, result, "case class Address(")
	assert.Contains(t, result, "address: Option[Address]")
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

	addressIdx := strings.Index(result, "case class Address(")
	customerIdx := strings.Index(result, "case class Customer(")

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

	assert.Contains(t, result, "id: Long")
	assert.Contains(t, result, "name: String")
	assert.Contains(t, result, "optional: Option[String] = None")
}

func TestTranslate_FieldOrder(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"optional_first": {Type: "string"},
			"id":             {Type: "integer"},
			"optional_last":  {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	// Required field (id) should come before optional fields
	idIdx := strings.Index(result, "id: Long")
	opt1Idx := strings.Index(result, "optional_first: Option[String]")
	opt2Idx := strings.Index(result, "optional_last: Option[String]")

	assert.Greater(t, idIdx, 0)
	assert.Greater(t, opt1Idx, 0)
	assert.Greater(t, opt2Idx, 0)
	assert.Less(t, idIdx, opt1Idx)
	assert.Less(t, idIdx, opt2Idx)
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

	aIdx := strings.Index(result, "case class A(")
	bIdx := strings.Index(result, "case class B(")
	cIdx := strings.Index(result, "case class C(")

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

	assert.Contains(t, result, "case class SimpleSchema(")
	assert.Equal(t, 1, strings.Count(result, "case class "))
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

	result := string(output)

	assert.Contains(t, result, "case class Items(")
	assert.Contains(t, result, "List[Items]")
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
}
