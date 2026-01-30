// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package sparkscala provides Spark Scala StructType schema translation utilities.
package sparkscala

import (
	"fmt"

	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date":
			return "DateType"
		case "date-time":
			return "TimestampType"
		case "uuid":
			return "StringType"
		}
	}

	switch schemaType {
	case "string":
		return "StringType"
	case "integer":
		return "LongType"
	case "number":
		return "DoubleType"
	case "boolean":
		return "BooleanType"
	default:
		return "StringType"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return fmt.Sprintf("ArrayType(%s)", elemType)
}

func (r *resolver) RefType(defName string) string {
	return "_" + defName
}

func (r *resolver) FormatDefName(defName string) string {
	return "_" + defName
}

func (r *resolver) FormatRootName(portName string) string {
	return portName + "_schema"
}

func (r *resolver) EnrichField(_ *translate.Field) {}
