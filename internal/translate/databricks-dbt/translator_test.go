// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package dbtdatabricks

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
				"version: 2",
				"models:",
				"- name: test_port",
				"contract:",
				"enforced: true",
				"columns:",
				"- name: age",
				"data_type: BIGINT",
				"- name: bio",
				"data_type: STRING",
				"- name: name",
				"type: not_null",
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
				"data_type: TIMESTAMP",
				"data_type: DATE",
				"data_type: STRING",
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
				"data_type: ARRAY<STRING>",
			},
		},
		{
			name: "nested object as struct",
			schema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"address": {
						Type: "object",
						Properties: map[string]*jsonschema.Schema{
							"city":   {Type: "string"},
							"street": {Type: "string"},
						},
					},
				},
			},
			wantCode: []string{
				"- name: address",
				"data_type: STRUCT<city: STRING, street: STRING>",
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
				"data_type: DOUBLE",
				"data_type: BIGINT",
				"data_type: BOOLEAN",
			},
		},
		{
			name: "with description",
			schema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"user_id": {Type: "string", Description: "The unique user identifier"},
				},
			},
			wantCode: []string{
				`description: "The unique user identifier"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := New()
			got, err := translator.Translate("test_port", tt.schema)
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
	got, err := translator.Translate("events", schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)
	// Should contain fields from both schemas
	if !strings.Contains(gotStr, "- name: id") {
		t.Error("Should contain id field from base schema")
	}
	if !strings.Contains(gotStr, "- name: timestamp") {
		t.Error("Should contain timestamp field from base schema")
	}
	if !strings.Contains(gotStr, "- name: name") {
		t.Error("Should contain name field from extended schema")
	}
	if !strings.Contains(gotStr, "- name: age") {
		t.Error("Should contain age field from extended schema")
	}
}

func TestTranslator_Name(t *testing.T) {
	translator := New()
	if got := translator.Name(); got != "dbt-databricks" {
		t.Errorf("Name() = %v, want %v", got, "dbt-databricks")
	}
}

func TestTranslator_FileExtension(t *testing.T) {
	translator := New()
	if got := translator.FileExtension(); got != ".yml" {
		t.Errorf("FileExtension() = %v, want %v", got, ".yml")
	}
}

func TestTranslator_ComplexSchema(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"date", "total_orders", "total_revenue"},
		Properties: map[string]*jsonschema.Schema{
			"date": {
				Type:   "string",
				Format: "date",
			},
			"region": {
				Type:        "string",
				Description: "Geographic region",
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
	got, err := translator.Translate("daily_sales", schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)

	// Verify structure
	if !strings.Contains(gotStr, "version: 2") {
		t.Error("Missing version")
	}
	if !strings.Contains(gotStr, "- name: daily_sales") {
		t.Error("Missing model name")
	}
	if !strings.Contains(gotStr, "enforced: true") {
		t.Error("Missing contract enforcement")
	}

	// Verify required field has not_null constraint
	if !strings.Contains(gotStr, "type: not_null") {
		t.Error("Required fields should have not_null constraint")
	}

	// Verify correct types
	if !strings.Contains(gotStr, "data_type: DATE") {
		t.Error("Missing DATE type for date field")
	}
	if !strings.Contains(gotStr, "data_type: BIGINT") {
		t.Error("Missing BIGINT type for integer field")
	}
	if !strings.Contains(gotStr, "data_type: DOUBLE") {
		t.Error("Missing DOUBLE type for number field")
	}
}
