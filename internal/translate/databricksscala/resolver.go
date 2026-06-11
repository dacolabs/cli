// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package databricksscala provides Databricks Scala StructType schema translation with metadata.
package databricksscala

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dacolabs/daco/internal/translate"
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

func (r *resolver) MapType(keyType, valueType string) string {
	return fmt.Sprintf("MapType(%s, %s)", keyType, valueType)
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
	switch f.Type {
	case "LongType":
		f.Type = scalaSparkIntType(translate.NarrowInteger(f.Constraints))
	case "DoubleType":
		if kind, shape := translate.NarrowNumber(f.Constraints); kind == translate.NumberDecimal {
			f.Type = fmt.Sprintf("DecimalType(%d, %d)", shape.Precision, shape.Scale)
		}
	}

	var entries []string
	if f.Description != "" {
		d, _ := json.Marshal(f.Description)
		entries = append(entries, `"comment": `+string(d))
	}
	entries = append(entries, translate.ConstraintEntriesJSON(f.Constraints)...)
	if len(entries) > 0 {
		f.Tag = `, Metadata.fromJson("""{` + strings.Join(entries, ", ") + `}""")`
	}
}

func scalaSparkIntType(k translate.IntKind) string {
	switch k {
	case translate.Int8:
		return "ByteType"
	case translate.Int16:
		return "ShortType"
	case translate.Int32:
		return "IntegerType"
	default:
		return "LongType"
	}
}
