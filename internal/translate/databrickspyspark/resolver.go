// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package databrickspyspark provides Databricks PySpark schema translation with metadata comments.
package databrickspyspark

import (
	"fmt"

	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date":
			return "T.DateType()"
		case "date-time":
			return "T.TimestampType()"
		case "uuid":
			return "T.StringType()"
		}
	}

	switch schemaType {
	case "string":
		return "T.StringType()"
	case "integer":
		return "T.LongType()"
	case "number":
		return "T.DoubleType()"
	case "boolean":
		return "T.BooleanType()"
	default:
		return "T.StringType()"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return fmt.Sprintf("T.ArrayType(%s)", elemType)
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
		f.Tag = `, metadata={"comment": "` + f.Description + `"}`
	}
}
