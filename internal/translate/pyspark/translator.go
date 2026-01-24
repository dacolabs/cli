// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package pyspark provides PySpark schema translation utilities.
package pyspark

import (
	"fmt"
	"strings"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/google/jsonschema-go/jsonschema"
)

// Translator translates JSON schemas to PySpark StructType definitions.
type Translator struct{}

func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to PySpark Python code.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema) ([]byte, error) {
	var sb strings.Builder

	// Get property key order for preserving original ordering
	keyOrder, err := jschema.ExtractKeyOrder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to extract key order: %w", err)
	}

	sb.WriteString("import pyspark.sql.types as T\n")
	sb.WriteString("\n")

	// Generate variables for defs in topological order (dependencies first)
	for defName, defSchema := range jschema.TraverseDefs(schema) {
		sb.WriteString("_" + defName + " = ")
		if err := translateSchema(&sb, defSchema, 0, keyOrder, ""); err != nil {
			return nil, err
		}
		sb.WriteString("\n\n")
	}

	// Generate main schema
	sb.WriteString(portName + "_schema = ")
	if err := translateSchema(&sb, schema, 0, keyOrder, ""); err != nil {
		return nil, err
	}
	sb.WriteString("\n")

	return []byte(sb.String()), nil
}

func translateSchema(sb *strings.Builder, schema *jsonschema.Schema, depth int, keyOrder map[string][]string, path string) error {
	indent := strings.Repeat("    ", depth)

	if schema.Type == "object" || (schema.Type == "" && len(schema.Properties) > 0) {
		sb.WriteString("T.StructType([\n")

		// Get property names in original order from keyOrder, or fall back to map iteration
		propNames := getOrderedKeys(schema, keyOrder, path)

		for i, propName := range propNames {
			propSchema := schema.Properties[propName]

			nullable := true
			for _, s := range schema.Required {
				if s == propName {
					nullable = false
				}
			}

			sb.WriteString(indent + "    T.StructField(")
			sb.WriteString(fmt.Sprintf("%q, ", propName))

			// Build path for nested properties
			var propPath string
			if path == "" {
				propPath = "properties." + propName
			} else {
				propPath = path + ".properties." + propName
			}

			if err := translateType(sb, propSchema, depth+1, keyOrder, propPath); err != nil {
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

	return translateType(sb, schema, depth, keyOrder, path)
}

func translateType(sb *strings.Builder, schema *jsonschema.Schema, depth int, keyOrder map[string][]string, path string) error {
	// Handle $ref by emitting the variable name
	if schema.Ref != "" {
		refName := extractRefName(schema.Ref)
		sb.WriteString(refName)
		return nil
	}

	// Handle arrays
	if schema.Type == "array" {
		sb.WriteString("T.ArrayType(")
		if schema.Items != nil {
			if err := translateType(sb, schema.Items, depth, keyOrder, path); err != nil {
				return err
			}
		} else {
			sb.WriteString("T.StringType()")
		}
		sb.WriteString(")")
		return nil
	}

	// Handle objects
	if schema.Type == "object" || (schema.Type == "" && len(schema.Properties) > 0) {
		return translateSchema(sb, schema, depth, keyOrder, path)
	}

	// Handle primitive types with format
	if schema.Format != "" {
		switch schema.Format {
		case "date":
			sb.WriteString("T.DateType()")
			return nil
		case "date-time":
			sb.WriteString("T.TimestampType()")
			return nil
		case "uuid":
			sb.WriteString("T.StringType()")
			return nil
		}
	}

	// Handle primitive types
	switch schema.Type {
	case "string":
		sb.WriteString("T.StringType()")
	case "integer":
		sb.WriteString("T.LongType()")
	case "number":
		sb.WriteString("T.DoubleType()")
	case "boolean":
		sb.WriteString("T.BooleanType()")
	default:
		sb.WriteString("T.StringType()")
	}

	return nil
}

// extractRefName extracts the schema name from a $ref string.
func extractRefName(ref string) string {
	if name, ok := strings.CutPrefix(ref, "#/$defs/"); ok {
		return "_" + name
	}
	return ref
}

// getOrderedKeys returns property names in their original order from keyOrder.
func getOrderedKeys(schema *jsonschema.Schema, keyOrder map[string][]string, path string) []string {
	// Try to find the order for this path
	orderPath := "properties"
	if path != "" {
		orderPath = path + ".properties"
	}

	if order, ok := keyOrder[orderPath]; ok {
		// Filter to only include keys that exist in the schema
		var result []string
		for _, key := range order {
			if _, exists := schema.Properties[key]; exists {
				result = append(result, key)
			}
		}
		return result
	}

	// Fall back to collecting keys from the map
	var keys []string
	for name := range schema.Properties {
		keys = append(keys, name)
	}
	return keys
}
