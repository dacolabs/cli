// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package pyspark provides PySpark schema translation utilities.
package pyspark

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed pyspark.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "pyspark.go.tmpl"))

// Translator translates JSON schemas to PySpark StructType definitions.
type Translator struct{}

// FileExtension returns the file extension for PySpark schemas.
func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to PySpark Python code.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "pyspark.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
