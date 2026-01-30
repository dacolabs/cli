// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package sparkscala

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed sparkscala.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "sparkscala.go.tmpl"))

// Translator translates JSON schemas to Spark Scala StructType definitions.
type Translator struct{}

// FileExtension returns the file extension for Scala files.
func (t *Translator) FileExtension() string {
	return ".scala"
}

// Translate converts a JSON schema to Spark Scala StructType definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	data.Extra["ObjectName"] = translate.ToPascalCase(portName)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sparkscala.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
