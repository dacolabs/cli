// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databricks

import (
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestTranslator_Translate_WithDescriptions(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id", "email"},
		Properties: map[string]*jsonschema.Schema{
			"id": {
				Type:        "string",
				Format:      "uuid",
				Description: "Unique identifier for the user",
			},
			"email": {
				Type:        "string",
				Description: "User's email address",
			},
			"name": {
				Type:        "string",
				Description: "Full name of the user",
			},
			"age": {
				Type: "integer",
				// No description - should not have metadata
			},
		},
	}

	translator := New()
	got, err := translator.Translate("users", schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)

	// Verify metadata comments are present for fields with descriptions
	if !strings.Contains(gotStr, `metadata={"comment": "Unique identifier for the user"}`) {
		t.Error("Missing metadata comment for 'id' field")
	}
	if !strings.Contains(gotStr, `metadata={"comment": "User's email address"}`) {
		t.Error("Missing metadata comment for 'email' field")
	}
	if !strings.Contains(gotStr, `metadata={"comment": "Full name of the user"}`) {
		t.Error("Missing metadata comment for 'name' field")
	}

	// Verify that field without description doesn't have metadata
	// The age field should have nullable but no metadata
	if strings.Contains(gotStr, `"age"`) && strings.Contains(gotStr, `"age", LongType(), nullable=True, metadata`) {
		t.Error("Field 'age' without description should not have metadata")
	}

	// Verify the schema variable name uses port name
	if !strings.Contains(gotStr, "users_schema = StructType([") {
		t.Error("Missing users_schema variable declaration")
	}
}

func TestTranslator_Translate_EscapedQuotes(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"notes": {
				Type:        "string",
				Description: `User's "special" notes with quotes`,
			},
		},
	}

	translator := New()
	got, err := translator.Translate("data", schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)

	// Verify quotes are escaped in the metadata
	if !strings.Contains(gotStr, `metadata={"comment": "User's \"special\" notes with quotes"}`) {
		t.Errorf("Quotes should be escaped in metadata comment, got:\n%s", gotStr)
	}
}

func TestTranslator_Name(t *testing.T) {
	translator := New()
	if got := translator.Name(); got != "databricks-pyspark" {
		t.Errorf("Name() = %v, want %v", got, "databricks-pyspark")
	}
}

func TestTranslator_FileExtension(t *testing.T) {
	translator := New()
	if got := translator.FileExtension(); got != ".py" {
		t.Errorf("FileExtension() = %v, want %v", got, ".py")
	}
}

func TestTranslator_ComponentsWithDescriptions(t *testing.T) {
	addressSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"street": {
				Type:        "string",
				Description: "Street address",
			},
			"city": {
				Type:        "string",
				Description: "City name",
			},
		},
	}

	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {
				Type:        "string",
				Description: "User name",
			},
			"address": {
				Ref: "#/$defs/Address",
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": addressSchema,
		},
	}

	translator := New()
	got, err := translator.Translate("user_profiles", schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)

	// Verify component has underscore prefix
	if !strings.Contains(gotStr, "_address = StructType([") {
		t.Error("Missing _address component variable")
	}

	// Verify descriptions in component schema
	if !strings.Contains(gotStr, `metadata={"comment": "Street address"}`) {
		t.Error("Missing metadata comment for 'street' field in component")
	}
	if !strings.Contains(gotStr, `metadata={"comment": "City name"}`) {
		t.Error("Missing metadata comment for 'city' field in component")
	}

	// Verify main schema uses port name
	if !strings.Contains(gotStr, "user_profiles_schema = StructType([") {
		t.Error("Missing user_profiles_schema variable declaration")
	}

	// Verify description in main schema
	if !strings.Contains(gotStr, `metadata={"comment": "User name"}`) {
		t.Error("Missing metadata comment for 'name' field in main schema")
	}
}

func TestTranslator_AllOfWithDescriptions(t *testing.T) {
	baseSchema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"id": {
				Type:        "string",
				Description: "Unique event identifier",
			},
			"timestamp": {
				Type:        "string",
				Format:      "date-time",
				Description: "When the event occurred",
			},
		},
	}

	schema := &jsonschema.Schema{
		AllOf: []*jsonschema.Schema{
			{Ref: "#/$defs/Base"},
			{
				Type:     "object",
				Required: []string{"user_id"},
				Properties: map[string]*jsonschema.Schema{
					"user_id": {
						Type:        "string",
						Format:      "uuid",
						Description: "ID of the user who triggered the event",
					},
					"action": {
						Type:        "string",
						Description: "Action performed",
					},
				},
			},
		},
		Defs: map[string]*jsonschema.Schema{
			"Base": baseSchema,
		},
	}

	translator := New()
	got, err := translator.Translate("activity_events", schema)
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	gotStr := string(got)

	// Verify descriptions from both base and extended schemas
	if !strings.Contains(gotStr, `metadata={"comment": "Unique event identifier"}`) {
		t.Error("Missing metadata comment for 'id' field from base schema")
	}
	if !strings.Contains(gotStr, `metadata={"comment": "When the event occurred"}`) {
		t.Error("Missing metadata comment for 'timestamp' field from base schema")
	}
	if !strings.Contains(gotStr, `metadata={"comment": "ID of the user who triggered the event"}`) {
		t.Error("Missing metadata comment for 'user_id' field")
	}
	if !strings.Contains(gotStr, `metadata={"comment": "Action performed"}`) {
		t.Error("Missing metadata comment for 'action' field")
	}
}
