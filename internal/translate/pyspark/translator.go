// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package pyspark provides PySpark schema translation.
package pyspark

import (
	"fmt"
	"strings"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/cli/internal/translate/schema"
	"github.com/google/jsonschema-go/jsonschema"
)

func init() {
	// Auto-register on import
	translate.Register(New())
}

// Translator translates JSON schemas to PySpark StructType definitions.
type Translator struct{}

// New creates a new PySpark translator.
func New() *Translator {
	return &Translator{}
}

// Name returns the translator's identifier.
func (t *Translator) Name() string {
	return "pyspark"
}

// FileExtension returns the file extension for PySpark files.
func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to PySpark Python code.
func (t *Translator) Translate(portName string, s *jsonschema.Schema, rawJSON []byte) ([]byte, error) {
	// Extract property order from raw JSON
	keyOrder := schema.ExtractKeyOrder(rawJSON)
	r := schema.NewResolver(s, keyOrder)

	var sb strings.Builder

	// Add imports
	sb.WriteString(`from pyspark.sql.types import (
    ArrayType,
    BooleanType,
    DateType,
    DoubleType,
    LongType,
    StringType,
    StructField,
    StructType,
    TimestampType,
)

`)

	// Collect used definitions
	usedDefs := r.CollectUsedDefs(s)

	// Generate component schemas from $defs (only used ones, in dependency order)
	if len(usedDefs) > 0 {
		defs, err := generateDefinitions(r, usedDefs)
		if err != nil {
			return nil, err
		}
		if defs != "" {
			sb.WriteString(defs)
			sb.WriteString("\n")
		}
	}

	// Generate main schema as a function
	structDef, err := convertToStructType(s, 1, r, "properties")
	if err != nil {
		return nil, err
	}
	sb.WriteString("def _schema():\n")
	sb.WriteString("    return ")
	sb.WriteString(structDef)
	sb.WriteString("\n")

	return []byte(sb.String()), nil
}

// generateDefinitions generates Python variables for used definitions in dependency order.
func generateDefinitions(r *schema.Resolver, usedDefs map[string]bool) (string, error) {
	if len(usedDefs) == 0 {
		return "", nil
	}

	var result []string

	generating := make(map[string]bool) // Track definitions currently being generated

	var generate func(name string) error
	generate = func(name string) error {
		if r.IsGenerated(name) {
			return nil
		}

		if !usedDefs[name] {
			return nil
		}

		// Skip if already generating (circular dependency) - will be referenced by name
		if generating[name] {
			return nil
		}

		def := r.GetDef(name)
		if def == nil {
			return nil
		}

		generating[name] = true

		// First, generate all dependencies
		deps := r.CollectDependencies(def)
		for _, dep := range deps {
			if usedDefs[dep] {
				if err := generate(dep); err != nil {
					return err
				}
			}
		}

		// Generate this definition
		defPath := "$defs." + name + ".properties"
		structDef, err := convertToStructType(def, 0, r, defPath)
		if err != nil {
			return fmt.Errorf("definition %q: %w", name, err)
		}

		varName := "_" + name
		result = append(result, fmt.Sprintf("%s = %s", varName, structDef))
		r.MarkGenerated(name)

		return nil
	}

	for _, name := range r.GetDefNames() {
		if usedDefs[name] {
			if err := generate(name); err != nil {
				return "", err
			}
		}
	}

	return strings.Join(result, "\n") + "\n", nil
}

// convertToStructType converts a schema to PySpark StructType.
func convertToStructType(s *jsonschema.Schema, indent int, r *schema.Resolver, path string) (string, error) {
	if s == nil {
		return "StructType([])", nil
	}

	// Handle allOf composition
	if len(s.AllOf) > 0 {
		merged := mergeAllOf(s.AllOf, r)
		return convertToStructType(merged, indent, r, path)
	}

	if s.Properties == nil || len(s.Properties) == 0 {
		return "StructType([])", nil
	}

	requiredSet := make(map[string]bool)
	for _, req := range s.Required {
		requiredSet[req] = true
	}

	propNames := r.GetPropertyOrder(path, s.Properties)

	var sb strings.Builder
	baseIndent := strings.Repeat("    ", indent)
	fieldIndent := strings.Repeat("    ", indent+1)

	sb.WriteString("StructType([\n")

	for i, name := range propNames {
		prop := s.Properties[name]
		nullable := !requiredSet[name]

		nestedPath := path + "." + name + ".properties"
		pysparkType, err := convertType(prop, indent+1, r, nestedPath)
		if err != nil {
			return "", fmt.Errorf("property %q: %w", name, err)
		}

		sb.WriteString(fieldIndent)
		sb.WriteString(fmt.Sprintf(`StructField("%s", %s, nullable=%s)`, name, pysparkType, boolToStr(nullable)))

		if i < len(propNames)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(baseIndent)
	sb.WriteString("])")

	return sb.String(), nil
}

// convertType converts a JSON Schema type to PySpark type.
func convertType(s *jsonschema.Schema, indent int, r *schema.Resolver, path string) (string, error) {
	if s == nil {
		return "StringType()", nil
	}

	// Handle $ref
	if s.Ref != "" {
		defName := r.GetRefDefName(s.Ref)
		if defName != "" {
			return "_" + defName, nil
		}
		return "", fmt.Errorf("unresolved $ref: %s", s.Ref)
	}

	// Handle allOf
	if len(s.AllOf) > 0 {
		merged := mergeAllOf(s.AllOf, r)
		return convertToStructType(merged, indent, r, path)
	}

	switch s.Type {
	case "string":
		return convertStringType(s.Format), nil
	case "integer":
		return "LongType()", nil
	case "number":
		return "DoubleType()", nil
	case "boolean":
		return "BooleanType()", nil
	case "array":
		if s.Items == nil {
			return "ArrayType(StringType())", nil
		}
		elementType, err := convertType(s.Items, indent, r, path)
		if err != nil {
			return "", fmt.Errorf("array items: %w", err)
		}
		return fmt.Sprintf("ArrayType(%s)", elementType), nil
	case "object":
		return convertToStructType(s, indent, r, path)
	default:
		// Check if it has properties (implied object)
		if len(s.Properties) > 0 {
			return convertToStructType(s, indent, r, path)
		}
		return "StringType()", nil
	}
}

func convertStringType(format string) string {
	switch format {
	case "date":
		return "DateType()"
	case "date-time":
		return "TimestampType()"
	default:
		return "StringType()"
	}
}

func boolToStr(b bool) string {
	if b {
		return "True"
	}
	return "False"
}

// mergeAllOf merges multiple schemas from an allOf into a single schema.
func mergeAllOf(schemas []*jsonschema.Schema, r *schema.Resolver) *jsonschema.Schema {
	merged := &jsonschema.Schema{
		Type:       "object",
		Properties: make(map[string]*jsonschema.Schema),
		Required:   []string{},
	}

	for _, sub := range schemas {
		resolved := sub
		if sub.Ref != "" {
			if ref := r.ResolveRef(sub.Ref); ref != nil {
				resolved = ref
			}
		}

		for propName, propSchema := range resolved.Properties {
			merged.Properties[propName] = propSchema
		}
		merged.Required = append(merged.Required, resolved.Required...)
	}

	return merged
}
