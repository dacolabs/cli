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

	"github.com/dacolabs/daco/internal/translate"
	"github.com/google/jsonschema-go/jsonschema"
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

	// checks if any field type contains datetime, Literal, or any tag uses Field(...).
	data.Extra["NeedsDatetimeImport"] = false
	data.Extra["NeedsFieldImport"] = false
	data.Extra["NeedsLiteralImport"] = false
	data.Extra["NeedsUUIDImport"] = false
	data.Extra["NeedsDecimalImport"] = false
	scan := func(f translate.Field) {
		if strings.Contains(f.Type, "datetime.") {
			data.Extra["NeedsDatetimeImport"] = true
		}
		if strings.Contains(f.Tag, "Field(") {
			data.Extra["NeedsFieldImport"] = true
		}
		if strings.Contains(f.Type, "Literal[") {
			data.Extra["NeedsLiteralImport"] = true
		}
		if strings.Contains(f.Type, "uuid.UUID") {
			data.Extra["NeedsUUIDImport"] = true
		}
		if strings.Contains(f.Type, "Decimal") {
			data.Extra["NeedsDecimalImport"] = true
		}
	}
	for _, def := range data.Defs {
		for i := range def.Fields {
			scan(def.Fields[i])
		}
	}
	for i := range data.Root.Fields {
		scan(data.Root.Fields[i])
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
