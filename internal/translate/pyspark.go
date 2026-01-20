// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package translate provides schema translation utilities.
package translate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// PySparkTranslator translates JSON schemas to PySpark StructType definitions.
type PySparkTranslator struct {
	defs    map[string]*jsonschema.Schema
	visited map[*jsonschema.Schema]bool
}

// NewPySparkTranslator creates a new PySpark translator.
func NewPySparkTranslator() *PySparkTranslator {
	return &PySparkTranslator{
		visited: make(map[*jsonschema.Schema]bool),
	}
}

// Translate converts a JSON schema to PySpark Python code.
func (t *PySparkTranslator) Translate(schema *jsonschema.Schema) (string, error) {
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

	// Generate schema definition
	sb.WriteString("schema = ")
	if err := t.translateSchema(&sb, schema, 0); err != nil {
		return "", err
	}
	sb.WriteString("\n")

	return sb.String(), nil
}

func (t *PySparkTranslator) translateSchema(sb *strings.Builder, schema *jsonschema.Schema, depth int) error {
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
			sb.WriteString(fmt.Sprintf(", nullable=%s)", nullableStr))

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

func (t *PySparkTranslator) translateType(sb *strings.Builder, schema *jsonschema.Schema, depth int) error {
	// Handle $ref by resolving from $defs
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
