// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databrickspyspark

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed databrickspyspark.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "databrickspyspark.go.tmpl"))

// Translator translates JSON schemas to Databricks PySpark StructType definitions with metadata.
type Translator struct{}

// FileExtension returns the file extension for Python files.
func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to Databricks PySpark StructType definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "databrickspyspark.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
