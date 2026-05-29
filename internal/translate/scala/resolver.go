// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package scala provides plain Scala case class schema translation.
package scala

import (
	"fmt"

	"github.com/dacolabs/daco/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date":
			return "java.time.LocalDate"
		case "date-time":
			return "java.time.Instant"
		case "uuid":
			return "String"
		}
	}

	switch schemaType {
	case "string":
		return "String"
	case "integer":
		return "Long"
	case "number":
		return "Double"
	case "boolean":
		return "Boolean"
	default:
		return "String"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return fmt.Sprintf("List[%s]", elemType)
}

func (r *resolver) MapType(keyType, valueType string) string {
	return fmt.Sprintf("Map[%s, %s]", keyType, valueType)
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
	switch f.Type {
	case "Long":
		f.Type = scalaIntType(translate.NarrowInteger(f.Constraints))
	case "Double":
		if kind, _ := translate.NarrowNumber(f.Constraints); kind == translate.NumberDecimal {
			f.Type = "BigDecimal"
		}
	}

	if f.Nullable {
		f.Type = fmt.Sprintf("Option[%s]", f.Type)
		f.Tag = " = None"
	}
}

func scalaIntType(k translate.IntKind) string {
	switch k {
	case translate.Int8:
		return "Byte"
	case translate.Int16:
		return "Short"
	case translate.Int32:
		return "Int"
	default:
		return "Long"
	}
}
