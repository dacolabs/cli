// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

// SchemaData is the complete input passed to a translator template.
type SchemaData struct {
	Defs        []TypeDef      // $defs in topological order
	Root        TypeDef        // the root/main schema
	Description string         // root schema description, if any
	Extra       map[string]any // translator-specific template data
}

// TypeDef represents a named type definition (a $def or the root schema).
type TypeDef struct {
	Name   string  // formatted name, e.g. "_Address" (PySpark) or "Address" (Go)
	Fields []Field // ordered fields
}

// Field represents a single property within a type definition.
type Field struct {
	Name        string      // property name (may be mutated by EnrichField)
	Type        string      // fully resolved target type string
	Nullable    bool        // true if not in schema.Required
	Tag         string      // language-specific annotation, e.g. `json:"name,omitempty"`
	Description string      // schema description, if any
	Constraints Constraints // JSON Schema validation constraints
}

// Constraints holds JSON Schema validation constraints for a field.
type Constraints struct {
	Enum             []any
	Const            *any
	Pattern          string
	Format           string
	MinLength        *int
	MaxLength        *int
	Minimum          *float64
	Maximum          *float64
	ExclusiveMinimum *float64
	ExclusiveMaximum *float64
	MultipleOf       *float64
	MinItems         *int
	MaxItems         *int
}
