// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package sparksql provides Spark DDL schema translation.
package sparksql

import (
	"fmt"

	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date":
			return "DATE"
		case "date-time":
			return "TIMESTAMP"
		case "uuid":
			return "STRING"
		}
	}

	switch schemaType {
	case "string":
		return "STRING"
	case "integer":
		return "BIGINT"
	case "number":
		return "DOUBLE"
	case "boolean":
		return "BOOLEAN"
	default:
		return "STRING"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return fmt.Sprintf("ARRAY<%s>", elemType)
}

func (r *resolver) RefType(defName string) string {
	return defName
}

func (r *resolver) FormatDefName(defName string) string {
	return defName
}

func (r *resolver) FormatRootName(portName string) string {
	return portName + "_schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	if !f.Nullable {
		f.Tag = " NOT NULL"
	}
}
