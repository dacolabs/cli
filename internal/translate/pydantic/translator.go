// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pydantic

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed pydantic.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "pydantic.go.tmpl"))

// Translator translates JSON schemas to Pydantic BaseModel definitions.
type Translator struct{}

// FileExtension returns the file extension for Python files.
func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to Pydantic BaseModel definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	// checks if any field type contains datetime.
	data.Extra["NeedsDatetimeImport"] = false
	for _, def := range data.Defs {
		for _, f := range def.Fields {
			if strings.Contains(f.Type, "datetime.") {
				data.Extra["NeedsDatetimeImport"] = true
			}
		}
	}
	for _, f := range data.Root.Fields {
		if strings.Contains(f.Type, "datetime.") {
			data.Extra["NeedsDatetimeImport"] = true
		}
	}

	// sorts fields so required fields come before optional fields.
	sortFields := func(fields []translate.Field) {
		sort.SliceStable(fields, func(i, j int) bool {
			if fields[i].Nullable != fields[j].Nullable {
				return !fields[i].Nullable
			}
			return false
		})
	}

	for i := range data.Defs {
		sortFields(data.Defs[i].Fields)
	}
	sortFields(data.Root.Fields)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "pydantic.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
