// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package gotypes provides Go struct type schema translation utilities.
package gotypes

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed gotypes.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "gotypes.go.tmpl"))

// Translator translates JSON schemas to Go struct type definitions.
type Translator struct{}

// FileExtension returns the file extension for Go source files.
func (t *Translator) FileExtension() string {
	return ".go"
}

// Translate converts a JSON schema to Go struct definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, outputDir string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	data.Extra["Package"] = filepath.Base(outputDir)

	// checks if any field type contains time.Time.
	data.Extra["NeedsTimeImport"] = false
	for _, def := range data.Defs {
		for i := range def.Fields {
			if strings.Contains(def.Fields[i].Type, "time.Time") {
				data.Extra["NeedsTimeImport"] = true
			}
		}
	}
	for i := range data.Root.Fields {
		if strings.Contains(data.Root.Fields[i].Type, "time.Time") {
			data.Extra["NeedsTimeImport"] = true
		}
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "gotypes.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
