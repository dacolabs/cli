// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package databricksscala provides Databricks Scala StructType schema translation with metadata.
package databricksscala

import (
	"fmt"
	"strings"

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

func (r *resolver) EnrichField(f *translate.Field) {
	if f.Description != "" {
		escaped := strings.ReplaceAll(f.Description, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		f.Tag = `.withComment("` + escaped + `")`
	}
}
