// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

// TypeResolver converts JSON Schema types to target-language type strings and naming conventions.
// Each translator implements this interface to control how schemas map to its output format.
type TypeResolver interface {
	// PrimitiveType maps a JSON Schema type and format to a target type string.
	// Format is checked first, allowing "date-time" to override "string".
	PrimitiveType(schemaType, format string) string

	// ArrayType wraps an element type string in an array type.
	ArrayType(elemType string) string

	// RefType returns the type string for a $ref reference.
	RefType(defName string) string

	// FormatDefName formats a $def name for the target language.
	FormatDefName(defName string) string

	// FormatRootName formats the root schema name from the port name.
	FormatRootName(portName string) string

	// EnrichField applies language-specific post-processing to a resolved field.
	// It may mutate any combination of the field's properties:
	//   - Name: rename for target conventions (e.g. snake_case → PascalCase for Go)
	//   - Type: wrap for nullability (e.g. Optional[T] for Python, *T for Go)
	//   - Tag:  set annotations (e.g. json struct tags for Go, " = None" for Python)
	// Called once per field after type resolution, before template execution.
	EnrichField(f *Field)
}
