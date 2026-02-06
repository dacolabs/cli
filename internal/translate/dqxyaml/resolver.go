// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package dqxyaml translates JSON Schema constraints to DQX YAML quality checks.
package dqxyaml

import (
	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, _ string) string {
	return schemaType
}

func (r *resolver) ArrayType(_ string) string {
	return "array"
}

func (r *resolver) RefType(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatDefName(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return translate.ToPascalCase(portName)
}

func (r *resolver) EnrichField(_ *translate.Field) {}
