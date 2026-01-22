// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

// helper to serialize schema to JSON for tests
func mustMarshal(t *testing.T, schema *jsonschema.Schema) []byte {
	t.Helper()
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	return data
}

func TestTranslator_Translate(t *testing.T) {
	tests := []struct {
		name     string
		schema   *jsonschema.Schema
		wantCode []string // Expected code snippets
	}{
		{
			name: "simple object with primitives",
			schema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"name", "age"},
				Properties: map[string]*jsonschema.Schema{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
					"bio":  {Type: "string"},
				},
			},
			wantCode: []string{
				"from pyspark.sql.types import",
				"StructType([",
				`StructField("age", LongType(), nullable=False)`,
				`StructField("bio", StringType(), nullable=True)`,
				`StructField("name", StringType(), nullable=False)`,
			},
		},
		{
			name: "object with date formats",
			schema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"created_at": {Type: "string", Format: "date-time"},
					"birth_date": {Type: "string", Format: "date"},
					"user_id":    {Type: "string", Format: "uuid"},
				},
			},
			wantCode: []string{
				"TimestampType()",
				"DateType()",
				"StringType()",
			},
		},
		{
			name: "object with array",
			schema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"tags": {
						Type:  "array",
						Items: &jsonschema.Schema{Type: "string"},
					},
				},
			},
			wantCode: []string{
				"ArrayType(StringType())",
			},
		},
		{
			name: "nested object",
			schema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"user": {
						Type: "object",
						Properties: map[string]*jsonschema.Schema{
							"name": {Type: "string"},
							"age":  {Type: "integer"},
						},
					},
				},
			},
			wantCode: []string{
				"StructType([",
				`StructField("user", StructType([`,
				`StructField("age", LongType()`,
				`StructField("name", StringType()`,
			},
		},
		{
			name: "number types",
			schema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"price":    {Type: "number"},
					"quantity": {Type: "integer"},
					"active":   {Type: "boolean"},
				},
			},
			wantCode: []string{
				"DoubleType()",
				"LongType()",
				"BooleanType()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := New()
			rawJSON := mustMarshal(t, tt.schema)
			got, err := translator.Translate("test_port", tt.schema, rawJSON)
			if err != nil {
				t.Fatalf("Translate() error = %v", err)
			}

			gotStr := string(got)
			for _, want := range tt.wantCode {
				if !strings.Contains(gotStr, want) {
					t.Errorf("Translate() missing expected code snippet:\nwant: %q\ngot:\n%s", want, gotStr)
				}
			}
		})
	}
}

func TestTranslator_SchemaWithRefs(t *testing.T) {
	// Test with a schema that has $ref references
	addressSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"street": {Type: "string"},
			"city":   {Type: "string"},
			"zip":    {Type: "string"},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
			"address": {
				Ref: "#/$defs/Address",
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": addressSchema,
		},
	}

	translator := New()
	rawJSON := mustMarshal(t, schema)
	got, err := translator.Translate("users", schema, rawJSON)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Should contain the _Address variable reference
	if !strings.Contains(gotStr, "_Address") {
		t.Error("Should contain _Address variable for referenced schema")
	}
}

func TestTranslator_AllOf(t *testing.T) {
	// Test with allOf composition
	baseSchema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"id":        {Type: "string"},
			"timestamp": {Type: "string", Format: "date-time"},
		},
	}

	schema := &jsonschema.Schema{
		AllOf: []*jsonschema.Schema{
			{Ref: "#/$defs/Base"},
			{
				Type:     "object",
				Required: []string{"name"},
				Properties: map[string]*jsonschema.Schema{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Base": baseSchema,
		},
	}

	translator := New()
	rawJSON := mustMarshal(t, schema)
	got, err := translator.Translate("events", schema, rawJSON)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Should contain fields from both schemas
	if !strings.Contains(gotStr, `"id"`) {
		t.Error("Should contain id field from base schema")
	}
	if !strings.Contains(gotStr, `"timestamp"`) {
		t.Error("Should contain timestamp field from base schema")
	}
	if !strings.Contains(gotStr, `"name"`) {
		t.Error("Should contain name field from extended schema")
	}
	if !strings.Contains(gotStr, `"age"`) {
		t.Error("Should contain age field from extended schema")
	}
}

func TestTranslator_CircularRefs(t *testing.T) {
	// Test with circular references - should work since refs are emitted by name
	nodeSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
			"parent": {
				Ref: "#/$defs/Node",
			},
			"children": {
				Type: "array",
				Items: &jsonschema.Schema{
					Ref: "#/$defs/Node",
				},
			},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"root": {
				Ref: "#/$defs/Node",
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Node": nodeSchema,
		},
	}

	translator := New()
	rawJSON := mustMarshal(t, schema)
	got, err := translator.Translate("nodes", schema, rawJSON)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Node should reference itself by variable name
	if !strings.Contains(gotStr, "_Node = StructType([") {
		t.Error("Missing _Node variable declaration")
	}
	if !strings.Contains(gotStr, `StructField("parent", _Node`) {
		t.Error("Should reference _Node for parent field")
	}
	if !strings.Contains(gotStr, "ArrayType(_Node)") {
		t.Error("Should reference _Node in children array")
	}
}

func TestTranslator_ComplexSchema(t *testing.T) {
	// Test with a schema similar to the multi-port example
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"date", "total_orders", "total_revenue"},
		Properties: map[string]*jsonschema.Schema{
			"date": {
				Type:   "string",
				Format: "date",
			},
			"region": {
				Type: "string",
			},
			"total_orders": {
				Type: "integer",
			},
			"total_revenue": {
				Type: "number",
			},
			"avg_order_value": {
				Type: "number",
			},
		},
	}

	translator := New()
	rawJSON := mustMarshal(t, schema)
	got, err := translator.Translate("daily_sales", schema, rawJSON)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Verify imports
	if !strings.Contains(gotStr, "from pyspark.sql.types import") {
		t.Error("Missing imports")
	}

	// Verify correct types
	if !strings.Contains(gotStr, "DateType()") {
		t.Error("Missing DateType for date field")
	}
	if !strings.Contains(gotStr, "LongType()") {
		t.Error("Missing LongType for integer field")
	}
	if !strings.Contains(gotStr, "DoubleType()") {
		t.Error("Missing DoubleType for number field")
	}
}

func TestTranslator_Name(t *testing.T) {
	translator := New()
	if got := translator.Name(); got != "pyspark" {
		t.Errorf("Name() = %v, want %v", got, "pyspark")
	}
}

func TestTranslator_FileExtension(t *testing.T) {
	translator := New()
	if got := translator.FileExtension(); got != ".py" {
		t.Errorf("FileExtension() = %v, want %v", got, ".py")
	}
}

func TestTranslator_ComponentExtraction(t *testing.T) {
	// Test that components are extracted as variables
	nodeSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"root": {
				Ref: "#/$defs/Node",
			},
			"children": {
				Type: "array",
				Items: &jsonschema.Schema{
					Ref: "#/$defs/Node",
				},
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Node": nodeSchema,
		},
	}

	translator := New()
	rawJSON := mustMarshal(t, schema)
	got, err := translator.Translate("tree", schema, rawJSON)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)

	// Verify component schema is extracted as a variable with underscore prefix
	if !strings.Contains(gotStr, "_Node = StructType([") {
		t.Error("Missing _Node variable declaration")
	}

	// Verify main schema references the component variable
	if !strings.Contains(gotStr, `StructField("root", _Node`) {
		t.Error("Should reference _Node variable for root field")
	}
	if !strings.Contains(gotStr, "ArrayType(_Node)") {
		t.Error("Should reference _Node variable in array")
	}

	// Verify the component is defined before it's used
	componentIdx := strings.Index(gotStr, "_Node = ")
	mainSchemaIdx := strings.Index(gotStr, "def _schema():")
	if componentIdx == -1 || mainSchemaIdx == -1 {
		t.Error("Missing component or main schema sections")
	} else if componentIdx > mainSchemaIdx {
		t.Error("Component schema should be defined before main schema")
	}
}
