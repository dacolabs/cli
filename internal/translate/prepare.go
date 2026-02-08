// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import (
	"fmt"
	"strings"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

// prepareContext holds mutable state during schema preparation.
type prepareContext struct {
	resolver  TypeResolver
	keyOrder  map[string][]string
	extracted []TypeDef // inline objects extracted as named types
}

// Prepare converts a JSON Schema into a SchemaData ready for template execution.
// It walks the schema tree, resolves types using the provided TypeResolver,
// and returns the data in topological order.
// Inline nested objects are automatically extracted as named type definitions.
func Prepare(portName string, schema *jsonschema.Schema, resolver TypeResolver) (*SchemaData, error) {
	keyOrder, err := jschema.ExtractKeyOrder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to extract key order: %w", err)
	}

	ctx := &prepareContext{
		resolver: resolver,
		keyOrder: keyOrder,
	}

	data := &SchemaData{
		Description: schema.Description,
		Extra:       make(map[string]any),
	}

	// Process $defs in topological order
	for defName, defSchema := range jschema.TraverseDefs(schema) {
		defPath := "$defs." + defName
		fields := ctx.resolveFields(defSchema, defPath)
		data.Defs = append(data.Defs, TypeDef{
			Name:   resolver.FormatDefName(defName),
			Fields: fields,
		})
	}

	// Process root schema
	data.Root = TypeDef{
		Name:   resolver.FormatRootName(portName),
		Fields: ctx.resolveFields(schema, ""),
	}

	// Append inline objects extracted during resolution
	data.Defs = append(data.Defs, ctx.extracted...)

	return data, nil
}

func (c *prepareContext) resolveFields(schema *jsonschema.Schema, path string) []Field {
	propNames := getOrderedKeys(schema, c.keyOrder, path)

	fields := make([]Field, 0, len(propNames))
	for _, propName := range propNames {
		propSchema := schema.Properties[propName]

		nullable := true
		for _, req := range schema.Required {
			if req == propName {
				nullable = false
				break
			}
		}

		var propPath string
		if path == "" {
			propPath = "properties." + propName
		} else {
			propPath = path + ".properties." + propName
		}

		typeStr := c.resolveType(propSchema, propName, propPath)

		f := Field{
			Name:        propName,
			Type:        typeStr,
			Nullable:    nullable,
			Description: propSchema.Description,
			Constraints: extractConstraints(propSchema),
		}
		c.resolver.EnrichField(&f)
		fields = append(fields, f)
	}

	return fields
}

func (c *prepareContext) resolveType(schema *jsonschema.Schema, fieldName, path string) string {
	// Handle $ref
	if schema.Ref != "" {
		defName := extractDefName(schema.Ref)
		return c.resolver.RefType(defName)
	}

	// Handle arrays
	if schema.Type == "array" {
		if schema.Items != nil {
			elemType := c.resolveType(schema.Items, fieldName, path)
			return c.resolver.ArrayType(elemType)
		}
		return c.resolver.ArrayType(c.resolver.PrimitiveType("string", ""))
	}

	// Handle inline objects — extract as a named type definition
	if schema.Type == "object" || (schema.Type == "" && len(schema.Properties) > 0) {
		defName := ToPascalCase(fieldName)
		fields := c.resolveFields(schema, path)
		c.extracted = append(c.extracted, TypeDef{
			Name:   c.resolver.FormatDefName(defName),
			Fields: fields,
		})
		return c.resolver.RefType(defName)
	}

	// Handle primitives (format checked inside PrimitiveType)
	return c.resolver.PrimitiveType(schema.Type, schema.Format)
}

// extractConstraints copies JSON Schema validation constraints from a property schema.
func extractConstraints(s *jsonschema.Schema) Constraints {
	return Constraints{
		Enum:             s.Enum,
		Const:            s.Const,
		Pattern:          s.Pattern,
		Format:           s.Format,
		MinLength:        s.MinLength,
		MaxLength:        s.MaxLength,
		Minimum:          s.Minimum,
		Maximum:          s.Maximum,
		ExclusiveMinimum: s.ExclusiveMinimum,
		ExclusiveMaximum: s.ExclusiveMaximum,
		MultipleOf:       s.MultipleOf,
		MinItems:         s.MinItems,
		MaxItems:         s.MaxItems,
	}
}

// extractDefName extracts the definition name from a $ref string.
func extractDefName(ref string) string {
	if name, ok := strings.CutPrefix(ref, "#/$defs/"); ok {
		return name
	}
	return ref
}

// getOrderedKeys returns property names in their original JSON order.
func getOrderedKeys(schema *jsonschema.Schema, keyOrder map[string][]string, path string) []string {
	// Prefer PropertyOrder when set (populated during schema loading from raw bytes)
	if len(schema.PropertyOrder) > 0 {
		var result []string
		for _, key := range schema.PropertyOrder {
			if _, exists := schema.Properties[key]; exists {
				result = append(result, key)
			}
		}
		return result
	}

	// Fall back to keyOrder map (from ExtractKeyOrder)
	orderPath := "properties"
	if path != "" {
		orderPath = path + ".properties"
	}

	if order, ok := keyOrder[orderPath]; ok {
		var result []string
		for _, key := range order {
			if _, exists := schema.Properties[key]; exists {
				result = append(result, key)
			}
		}
		return result
	}

	// Last resort: map iteration order (non-deterministic)
	keys := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		keys = append(keys, name)
	}
	return keys
}

// ToSnakeCase converts a string to a valid snake_case identifier.
// It splits on non-alphanumeric characters, lowercases each part,
// and prefixes with underscore if the result starts with a digit.
func ToSnakeCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9')
	})

	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}

	result := strings.Join(parts, "_")
	if result != "" && result[0] >= '0' && result[0] <= '9' {
		result = "_" + result
	}
	return result
}

// ToPascalCase converts a snake_case string to PascalCase for type name generation.
func ToPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})

	var sb strings.Builder
	for _, part := range parts {
		if part != "" {
			sb.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}

	return sb.String()
}
