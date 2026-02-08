// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package markdown provides markdown schema documentation utilities.
package markdown

import (
	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	return schemaType
}

func (r *resolver) ArrayType(elemType string) string {
	return "array(" + elemType + ")"
}

func (r *resolver) RefType(defName string) string {
	return "[" + translate.ToPascalCase(defName) + "](#" + translate.ToPascalCase(defName) + ")"
}

func (r *resolver) FormatDefName(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return translate.ToPascalCase(portName)
}

func (r *resolver) EnrichField(f *translate.Field) {}
