// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databrickssql

import (
	"strings"
	"testing"

	"github.com/dacolabs/cli/internal/translate"
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

	assert.Contains(t, result, "CREATE TABLE users_schema")
	assert.Contains(t, result, "name STRING")
	assert.Contains(t, result, "age BIGINT")
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

	assert.Contains(t, result, "STRING")
	assert.Contains(t, result, "BIGINT")
	assert.Contains(t, result, "DOUBLE")
	assert.Contains(t, result, "BOOLEAN")
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

	assert.Contains(t, result, "TIMESTAMP")
	assert.Contains(t, result, "DATE")
	assert.Contains(t, result, "STRING")
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

	assert.Contains(t, string(output), "ARRAY<STRING>")
}

func TestTranslate_NotNullConstraint(t *testing.T) {
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

	assert.Contains(t, result, "id BIGINT NOT NULL")
	assert.Contains(t, result, "name STRING NOT NULL")
	assert.Contains(t, result, "optional STRING")
	assert.NotContains(t, result, "optional STRING NOT NULL")
}

func TestTranslate_DeltaClause(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("users", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "USING DELTA;")
}

func TestTranslate_FieldComment(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"id":   {Type: "integer", Description: "Unique identifier"},
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("users", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "COMMENT 'Unique identifier'")
	assert.NotContains(t, result, "name STRING COMMENT")
}

func TestTranslate_InlineStruct(t *testing.T) {
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
				},
			},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("user", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "STRUCT<")
	assert.Equal(t, 1, strings.Count(result, "CREATE TABLE"))
}

func TestInlineStruct_CircularRef(t *testing.T) {
	// Construct a circular defMap: A references B, B references A.
	// inlineStruct must not infinite-loop; it should leave the back-edge
	// type name un-inlined.
	defs := map[string]*translate.TypeDef{
		"A": {Name: "A", Fields: []translate.Field{{Name: "b", Type: "B"}}},
		"B": {Name: "B", Fields: []translate.Field{{Name: "a", Type: "A"}}},
	}
	fields := []translate.Field{{Name: "root", Type: "A"}}

	// This must return without stack overflow.
	inlineStruct(fields, defs, make(map[string]bool))

	// A was inlined; its child B was inlined; B's child "A" was skipped (visited).
	assert.Contains(t, fields[0].Type, "STRUCT<")
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

	assert.Contains(t, result, "CREATE TABLE simple_schema")
	assert.Equal(t, 1, strings.Count(result, "CREATE TABLE"))
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".sql", translator.FileExtension())
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

	assert.True(t, strings.HasPrefix(string(output), "-- Code generated by daco; DO NOT EDIT."))
}

func TestTranslate_TableComment(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:        "object",
		Description: "A table of user records",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("users", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.Contains(t, result, "COMMENT 'A table of user records'")
	assert.Contains(t, result, "USING DELTA;")
}

func TestTranslate_NoTableComment(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("users", schema, "schemas")
	require.NoError(t, err)

	result := string(output)

	assert.NotContains(t, result, "COMMENT")
	assert.Contains(t, result, "USING DELTA;")
}
