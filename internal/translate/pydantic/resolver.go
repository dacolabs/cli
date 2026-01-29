// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package pydantic provides Pydantic BaseModel schema translation utilities.
package pydantic

import (
	"github.com/dacolabs/cli/internal/translate"
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
	if f.Nullable {
		f.Type = "Optional[" + f.Type + "]"
		f.Tag = " = None"
	}
}
