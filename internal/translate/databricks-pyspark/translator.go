// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package databricks provides Databricks PySpark schema translation with metadata comments.
package databricks

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

// Translator translates JSON schemas to Databricks PySpark StructType definitions with metadata.
type Translator struct {
	defs            map[string]*jsonschema.Schema
	visited         map[*jsonschema.Schema]bool
	components      map[string]string // Maps schema name to variable name
	currentDefining string            // Name of the component currently being defined
}

// New creates a new Databricks PySpark translator.
func New() *Translator {
	return &Translator{
		visited:    make(map[*jsonschema.Schema]bool),
		components: make(map[string]string),
	}
}

// Name returns the translator's identifier.
func (t *Translator) Name() string {
	return "databricks-pyspark"
}

// FileExtension returns the file extension for PySpark files.
func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to Databricks PySpark Python code with metadata comments.
// portName is used to name the output schema variable (e.g., "users" -> "users_schema")
func (t *Translator) Translate(portName string, schema *jsonschema.Schema) ([]byte, error) {
	var sb strings.Builder

	// Store the defs for reference resolution
	t.defs = schema.Defs
	if t.defs == nil {
		t.defs = make(map[string]*jsonschema.Schema)
	}

	// Add imports
	sb.WriteString("from pyspark.sql.types import (\n")
	sb.WriteString("    StructType,\n")
	sb.WriteString("    StructField,\n")
	sb.WriteString("    StringType,\n")
	sb.WriteString("    IntegerType,\n")
	sb.WriteString("    LongType,\n")
	sb.WriteString("    DoubleType,\n")
	sb.WriteString("    BooleanType,\n")
	sb.WriteString("    TimestampType,\n")
	sb.WriteString("    DateType,\n")
	sb.WriteString("    ArrayType,\n")
	sb.WriteString(")\n\n")

	// Generate component schemas from $defs
	if len(t.defs) > 0 {
		sb.WriteString("# Component schemas\n")

		// Sort component names for consistent output
		componentNames := make([]string, 0, len(t.defs))
		for name := range t.defs {
			componentNames = append(componentNames, name)
		}
		sort.Strings(componentNames)

		for _, name := range componentNames {
			componentSchema := t.defs[name]
			varName := "_" + toSnakeCase(name)
			t.components[name] = varName

			t.currentDefining = name // Track which component we're defining
			sb.WriteString(varName + " = ")
			if err := t.translateSchema(&sb, componentSchema, 0); err != nil {
				return nil, err
			}
			sb.WriteString("\n\n")
			t.currentDefining = "" // Clear after definition
		}
	}

	// Generate main schema definition with port name
	sb.WriteString("# Main schema\n")
	schemaVarName := toSnakeCase(portName) + "_schema"
	sb.WriteString(schemaVarName + " = ")
	if err := t.translateSchema(&sb, schema, 0); err != nil {
		return nil, err
	}
	sb.WriteString("\n")

	return []byte(sb.String()), nil
}

func (t *Translator) translateSchema(sb *strings.Builder, schema *jsonschema.Schema, depth int) error {
	indent := strings.Repeat("    ", depth)

	if schema.Type == "object" || (schema.Type == "" && len(schema.Properties) > 0) {
		sb.WriteString("StructType([\n")

		// Get property names sorted for consistent output
		propNames := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)

		for i, propName := range propNames {
			propSchema := schema.Properties[propName]
			nullable := !contains(schema.Required, propName)

			sb.WriteString(indent + "    StructField(")
			sb.WriteString(fmt.Sprintf("%q, ", propName))

			if err := t.translateType(sb, propSchema, depth+1); err != nil {
				return err
			}

			nullableStr := "True"
			if !nullable {
				nullableStr = "False"
			}
			sb.WriteString(fmt.Sprintf(", nullable=%s", nullableStr))

			// Add metadata with comment if description exists
			if propSchema.Description != "" {
				// Escape any quotes in the description
				escapedDesc := strings.ReplaceAll(propSchema.Description, `"`, `\"`)
				sb.WriteString(fmt.Sprintf(`, metadata={"comment": "%s"}`, escapedDesc))
			}

			sb.WriteString(")")

			if i < len(propNames)-1 {
				sb.WriteString(",\n")
			} else {
				sb.WriteString("\n")
			}
		}

		sb.WriteString(indent + "])")
		return nil
	}

	return t.translateType(sb, schema, depth)
}

func (t *Translator) translateType(sb *strings.Builder, schema *jsonschema.Schema, depth int) error {
	// Handle $ref by using the component variable
	if schema.Ref != "" {
		// Extract the reference name from "#/components/schemas/Name" or "$defs/Name"
		refName := schema.Ref
		if strings.HasPrefix(refName, "#/components/schemas/") {
			refName = strings.TrimPrefix(refName, "#/components/schemas/")
		} else if strings.HasPrefix(refName, "#/$defs/") {
			refName = strings.TrimPrefix(refName, "#/$defs/")
		} else if strings.HasPrefix(refName, "#/definitions/") {
			refName = strings.TrimPrefix(refName, "#/definitions/")
		}

		// If we're currently defining this component (self-reference), use StringType to avoid circular definition
		if t.currentDefining == refName {
			sb.WriteString("StringType()")
			return nil
		}

		// If we have a component variable for this ref, use it
		if varName, ok := t.components[refName]; ok {
			sb.WriteString(varName)
			return nil
		}

		// Look up the referenced schema
		if refSchema, ok := t.defs[refName]; ok {
			// Prevent infinite recursion for circular refs
			if t.visited[refSchema] {
				// For circular refs, just use StringType as a fallback
				sb.WriteString("StringType()")
				return nil
			}
			t.visited[refSchema] = true
			defer func() { delete(t.visited, refSchema) }()

			return t.translateType(sb, refSchema, depth)
		}
		// If not found in defs, treat as StringType
		sb.WriteString("StringType()")
		return nil
	}

	// Handle allOf composition by merging all schemas
	if len(schema.AllOf) > 0 {
		// Merge all schemas in allOf into a single schema
		merged := &jsonschema.Schema{
			Type:       "object",
			Properties: make(map[string]*jsonschema.Schema),
			Required:   []string{},
		}

		for _, subSchema := range schema.AllOf {
			// Resolve references in allOf
			resolvedSchema := subSchema
			if subSchema.Ref != "" {
				refName := subSchema.Ref
				if strings.HasPrefix(refName, "#/components/schemas/") {
					refName = strings.TrimPrefix(refName, "#/components/schemas/")
				} else if strings.HasPrefix(refName, "#/$defs/") {
					refName = strings.TrimPrefix(refName, "#/$defs/")
				}
				if refSchema, ok := t.defs[refName]; ok {
					resolvedSchema = refSchema
				}
			}

			// Merge properties
			for propName, propSchema := range resolvedSchema.Properties {
				merged.Properties[propName] = propSchema
			}
			// Merge required fields
			merged.Required = append(merged.Required, resolvedSchema.Required...)
		}

		return t.translateSchema(sb, merged, depth)
	}

	// Handle arrays
	if schema.Type == "array" {
		sb.WriteString("ArrayType(")
		if schema.Items != nil {
			if err := t.translateType(sb, schema.Items, depth); err != nil {
				return err
			}
		} else {
			sb.WriteString("StringType()")
		}
		sb.WriteString(")")
		return nil
	}

	// Handle objects
	if schema.Type == "object" || (schema.Type == "" && len(schema.Properties) > 0) {
		return t.translateSchema(sb, schema, depth)
	}

	// Handle primitive types with format
	if schema.Format != "" {
		switch schema.Format {
		case "date":
			sb.WriteString("DateType()")
			return nil
		case "date-time":
			sb.WriteString("TimestampType()")
			return nil
		case "uuid":
			sb.WriteString("StringType()")
			return nil
		}
	}

	// Handle primitive types
	switch schema.Type {
	case "string":
		sb.WriteString("StringType()")
	case "integer":
		sb.WriteString("LongType()")
	case "number":
		sb.WriteString("DoubleType()")
	case "boolean":
		sb.WriteString("BooleanType()")
	default:
		// If no type is specified, default to StringType
		sb.WriteString("StringType()")
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// toSnakeCase converts a string to snake_case.
// E.g., "MySchema" -> "my_schema", "HTTPResponse" -> "http_response"
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Add underscore before uppercase letters (except at start)
			if i > 0 {
				prevRune := rune(s[i-1])
				// Don't add underscore if previous char was also uppercase or underscore
				if !(prevRune >= 'A' && prevRune <= 'Z') && prevRune != '_' {
					result.WriteRune('_')
				}
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
