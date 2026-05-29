// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package avro provides Apache Avro schema translation utilities.
package avro

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dacolabs/daco/internal/translate"
)

// avroSymbolRe matches valid Avro enum symbols: must begin with letter/underscore,
// then any alphanumeric or underscore.
var avroSymbolRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

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

func (r *resolver) MapType(_, valueType string) string {
	return "map:" + valueType
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

func (r *resolver) EnrichField(f *translate.Field) {
	// Promote enum-constrained strings to a native Avro enum type when every
	// value is a string and matches Avro's symbol syntax.
	if f.Type == "string" {
		if symbols := translate.EnumStringSymbols(f.Constraints); symbols != nil {
			if allValidAvroSymbols(symbols) {
				name := translate.ToPascalCase(f.Name)
				f.Type = "enum:" + name + ":" + strings.Join(symbols, ",")
				return
			}
		}
	}

	switch f.Type {
	case "long":
		// Avro has only int (32) and long (64); narrow when bounds fit signed 32-bit.
		if translate.NarrowInteger(f.Constraints) <= translate.Int32 {
			f.Type = "int"
		}
	case "double":
		if kind, shape := translate.NarrowNumber(f.Constraints); kind == translate.NumberDecimal {
			f.Type = fmt.Sprintf("decimal:%d:%d", shape.Precision, shape.Scale)
		}
	}
}

func allValidAvroSymbols(symbols []string) bool {
	for _, s := range symbols {
		if !avroSymbolRe.MatchString(s) {
			return false
		}
	}
	return true
}
