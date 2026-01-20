// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

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
			got, err := translator.Translate(tt.schema)
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
				Ref: "#/components/schemas/Address",
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": addressSchema,
		},
	}

	translator := New()
	got, err := translator.Translate(schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Should contain the address fields
	if !strings.Contains(gotStr, `"street"`) {
		t.Error("Should contain street field from referenced schema")
	}
	if !strings.Contains(gotStr, `"city"`) {
		t.Error("Should contain city field from referenced schema")
	}
	if !strings.Contains(gotStr, `"zip"`) {
		t.Error("Should contain zip field from referenced schema")
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
			{Ref: "#/components/schemas/Base"},
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
	got, err := translator.Translate(schema)
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
	// Test with circular references (should not infinite loop)
	nodeSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
			"parent": {
				Ref: "#/components/schemas/Node",
			},
			"children": {
				Type: "array",
				Items: &jsonschema.Schema{
					Ref: "#/components/schemas/Node",
				},
			},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"root": {
				Ref: "#/components/schemas/Node",
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Node": nodeSchema,
		},
	}

	translator := New()
	got, err := translator.Translate(schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Should not infinite loop and should generate something
	if !strings.Contains(gotStr, "StructType") {
		t.Error("Should generate StructType")
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
	got, err := translator.Translate(schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Verify imports
	if !strings.Contains(gotStr, "from pyspark.sql.types import") {
		t.Error("Missing imports")
	}

	// Verify required fields are not nullable
	if !strings.Contains(gotStr, `StructField("date", DateType(), nullable=False)`) {
		t.Error("Required field 'date' should be nullable=False")
	}
	if !strings.Contains(gotStr, `StructField("total_orders", LongType(), nullable=False)`) {
		t.Error("Required field 'total_orders' should be nullable=False")
	}

	// Verify optional field is nullable
	if !strings.Contains(gotStr, `StructField("region", StringType(), nullable=True)`) {
		t.Error("Optional field 'region' should be nullable=True")
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
