// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package python provides Python dataclass schema translation utilities.
package python

import (
	"strings"

	"github.com/dacolabs/daco/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if schemaType == "string" && format != "" {
		switch format {
		case "date":
			return "datetime.date"
		case "date-time":
			return "datetime.datetime"
		}
	}

	switch schemaType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	default:
		return "str"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return "list[" + elemType + "]"
}

func (r *resolver) MapType(keyType, valueType string) string {
	return "dict[" + keyType + ", " + valueType + "]"
}

func (r *resolver) RefType(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatDefName(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return translate.ToPascalCase(portName) + "Schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	if lit := buildLiteralType(f.Constraints); lit != "" {
		f.Type = lit
	}

	if f.Nullable {
		f.Type = "Optional[" + f.Type + "]"
		f.Tag = " = None"
	}
}

// buildLiteralType returns a typing.Literal[...] type when the field is constrained
// to a fixed value (const) or a closed set (enum).
func buildLiteralType(c translate.Constraints) string {
	if c.Const != nil {
		return "Literal[" + translate.FormatLiteral(*c.Const) + "]"
	}
	if len(c.Enum) > 0 {
		parts := make([]string, len(c.Enum))
		for i, v := range c.Enum {
			parts[i] = translate.FormatLiteral(v)
		}
		return "Literal[" + strings.Join(parts, ", ") + "]"
	}
	return ""
}
