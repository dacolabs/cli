// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package dbtdatabricks provides dbt-databricks schema translation.
package dbtdatabricks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/google/jsonschema-go/jsonschema"
)

func init() {
	// Auto-register on import
	translate.Register(New())
}

// Translator translates JSON schemas to dbt-databricks contract YAML definitions.
type Translator struct {
	defs map[string]*jsonschema.Schema
}

// New creates a new dbt-databricks translator.
func New() *Translator {
	return &Translator{}
}

// Name returns the translator's identifier.
func (t *Translator) Name() string {
	return "dbt-databricks"
}

// FileExtension returns the file extension for dbt schema files.
func (t *Translator) FileExtension() string {
	return ".yml"
}

// Translate converts a JSON schema to dbt-databricks contract YAML.
// portName is used to name the model (e.g., "users" -> model name "users")
func (t *Translator) Translate(portName string, schema *jsonschema.Schema) ([]byte, error) {
	var sb strings.Builder

	// Store the defs for reference resolution
	t.defs = schema.Defs
	if t.defs == nil {
		t.defs = make(map[string]*jsonschema.Schema)
	}

	// Write dbt schema YAML header
	sb.WriteString("version: 2\n\n")
	sb.WriteString("models:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s\n", portName))
	sb.WriteString("    config:\n")
	sb.WriteString("      contract:\n")
	sb.WriteString("        enforced: true\n")
	sb.WriteString("    columns:\n")

	// Get the columns from the schema
	columns, err := t.extractColumns(schema)
	if err != nil {
		return nil, err
	}

	for _, col := range columns {
		sb.WriteString(fmt.Sprintf("      - name: %s\n", col.name))
		sb.WriteString(fmt.Sprintf("        data_type: %s\n", col.dataType))
		if col.description != "" {
			// Escape quotes and handle multiline descriptions
			escapedDesc := strings.ReplaceAll(col.description, `"`, `\"`)
			sb.WriteString(fmt.Sprintf("        description: \"%s\"\n", escapedDesc))
		}
		if !col.nullable {
			sb.WriteString("        constraints:\n")
			sb.WriteString("          - type: not_null\n")
		}
	}

	return []byte(sb.String()), nil
}

type column struct {
	name        string
	dataType    string
	description string
	nullable    bool
}

func (t *Translator) extractColumns(schema *jsonschema.Schema) ([]column, error) {
	// Handle allOf composition by merging all schemas
	if len(schema.AllOf) > 0 {
		merged := &jsonschema.Schema{
			Type:       "object",
			Properties: make(map[string]*jsonschema.Schema),
			Required:   []string{},
		}

		for _, subSchema := range schema.AllOf {
			resolvedSchema := subSchema
			if subSchema.Ref != "" {
				refName := t.extractRefName(subSchema.Ref)
				if refSchema, ok := t.defs[refName]; ok {
					resolvedSchema = refSchema
				}
			}

			for propName, propSchema := range resolvedSchema.Properties {
				merged.Properties[propName] = propSchema
			}
			merged.Required = append(merged.Required, resolvedSchema.Required...)
		}

		return t.extractColumns(merged)
	}

	if schema.Type != "object" && len(schema.Properties) == 0 {
		return nil, fmt.Errorf("schema must be an object type")
	}

	// Get property names sorted for consistent output
	propNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	columns := make([]column, 0, len(propNames))
	for _, propName := range propNames {
		propSchema := schema.Properties[propName]
		nullable := !contains(schema.Required, propName)

		dataType, err := t.translateType(propSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to translate type for %s: %w", propName, err)
		}

		columns = append(columns, column{
			name:        propName,
			dataType:    dataType,
			description: propSchema.Description,
			nullable:    nullable,
		})
	}

	return columns, nil
}

func (t *Translator) translateType(schema *jsonschema.Schema) (string, error) {
	// Handle $ref
	if schema.Ref != "" {
		refName := t.extractRefName(schema.Ref)
		if refSchema, ok := t.defs[refName]; ok {
			return t.translateType(refSchema)
		}
		return "STRING", nil
	}

	// Handle arrays
	if schema.Type == "array" {
		if schema.Items != nil {
			itemType, err := t.translateType(schema.Items)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("ARRAY<%s>", itemType), nil
		}
		return "ARRAY<STRING>", nil
	}

	// Handle nested objects as STRUCT
	if schema.Type == "object" || (schema.Type == "" && len(schema.Properties) > 0) {
		return t.translateStruct(schema)
	}

	// Handle primitive types with format
	if schema.Format != "" {
		switch schema.Format {
		case "date":
			return "DATE", nil
		case "date-time":
			return "TIMESTAMP", nil
		case "uuid":
			return "STRING", nil
		}
	}

	// Handle primitive types
	switch schema.Type {
	case "string":
		return "STRING", nil
	case "integer":
		return "BIGINT", nil
	case "number":
		return "DOUBLE", nil
	case "boolean":
		return "BOOLEAN", nil
	default:
		return "STRING", nil
	}
}

func (t *Translator) translateStruct(schema *jsonschema.Schema) (string, error) {
	// Get property names sorted for consistent output
	propNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	fields := make([]string, 0, len(propNames))
	for _, propName := range propNames {
		propSchema := schema.Properties[propName]
		dataType, err := t.translateType(propSchema)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("%s: %s", propName, dataType))
	}

	return fmt.Sprintf("STRUCT<%s>", strings.Join(fields, ", ")), nil
}

func (t *Translator) extractRefName(ref string) string {
	if strings.HasPrefix(ref, "#/components/schemas/") {
		return strings.TrimPrefix(ref, "#/components/schemas/")
	} else if strings.HasPrefix(ref, "#/$defs/") {
		return strings.TrimPrefix(ref, "#/$defs/")
	} else if strings.HasPrefix(ref, "#/definitions/") {
		return strings.TrimPrefix(ref, "#/definitions/")
	}
	return ref
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
