// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package protobuf

import (
	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, _ string) string {
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
		f.Type = "optional " + f.Type
	}
}
