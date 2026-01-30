// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package avro provides Apache Avro schema translation utilities.
package avro

import (
	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if schemaType == "string" && format != "" {
		switch format {
		case "date":
			return "date"
		case "date-time":
			return "timestamp-millis"
		case "uuid":
			return "uuid"
		}
	}

	switch schemaType {
	case "string":
		return "string"
	case "integer":
		return "long"
	case "number":
		return "double"
	case "boolean":
		return "boolean"
	default:
		return "string"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return "array:" + elemType
}

func (r *resolver) RefType(defName string) string {
	return "ref:" + defName
}

func (r *resolver) FormatDefName(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return translate.ToPascalCase(portName) + "Schema"
}

func (r *resolver) EnrichField(_ *translate.Field) {}
