// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package protobuf

import (
	"regexp"
	"strings"

	"github.com/dacolabs/daco/internal/translate"
)

// protoEnumSymbolRe matches valid proto3 enum symbol names.
var protoEnumSymbolRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if schemaType == "string" && format != "" {
		switch format {
		case "date", "date-time":
			return "google.protobuf.Timestamp"
		case "uuid":
			return "string"
		}
	}

	switch schemaType {
	case "string":
		return "string"
	case "integer":
		return "int64"
	case "number":
		return "double"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return "repeated " + elemType
}

func (r *resolver) MapType(keyType, valueType string) string {
	return "map<" + keyType + ", " + valueType + ">"
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
	// Promote string enums to a native proto3 enum marker the translator post-processes.
	if f.Type == "string" {
		if symbols := translate.EnumStringSymbols(f.Constraints); symbols != nil && allValidProtoSymbols(symbols) {
			name := translate.ToPascalCase(f.Name)
			f.Type = "enum:" + name + ":" + strings.Join(symbols, ",")
		}
	}

	if f.Type == "int64" {
		f.Type = protoIntType(translate.NarrowInteger(f.Constraints), translate.IsNonNegative(f.Constraints))
	}

	f.Name = translate.ToSnakeCase(f.Name)
	if f.Nullable {
		f.Type = "optional " + f.Type
	}
}

func allValidProtoSymbols(symbols []string) bool {
	for _, s := range symbols {
		if !protoEnumSymbolRe.MatchString(s) {
			return false
		}
	}
	return true
}

// protoIntType picks a protobuf scalar integer type.
// Protobuf has no 8/16-bit scalars — Int8/Int16 widen to int32/uint32.
func protoIntType(k translate.IntKind, nonNegative bool) string {
	if nonNegative {
		if k == translate.Int64 {
			return "uint64"
		}
		return "uint32"
	}
	if k == translate.Int64 {
		return "int64"
	}
	return "int32"
}
