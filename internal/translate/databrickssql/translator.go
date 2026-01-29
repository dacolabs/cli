// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databrickssql

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed databrickssql.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"last": func(i int, fields []translate.Field) bool {
		return i == len(fields)-1
	},
}).ParseFS(tmplFS, "databrickssql.go.tmpl"))

// Translator translates JSON schemas to Databricks SQL CREATE TABLE statements.
type Translator struct{}

// FileExtension returns the file extension for SQL files.
func (t *Translator) FileExtension() string {
	return ".sql"
}

// Translate converts a JSON schema to a Databricks SQL CREATE TABLE statement with USING DELTA.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	// Inline $ref types as STRUCT<...> by looking up defs
	defMap := make(map[string]*translate.TypeDef, len(data.Defs))
	for i := range data.Defs {
		defMap[data.Defs[i].Name] = &data.Defs[i]
	}
	inlineStruct(data.Root.Fields, defMap)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "databrickssql.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func inlineStruct(fields []translate.Field, defs map[string]*translate.TypeDef) {
	for i := range fields {
		if def, ok := defs[fields[i].Type]; ok {
			inlineStruct(def.Fields, defs)
			fields[i].Type = renderStruct(def.Fields)
		}
	}
}

func renderStruct(fields []translate.Field) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = f.Name + ": " + f.Type
		if f.Tag != "" {
			parts[i] += f.Tag
		}
	}
	return "STRUCT<" + strings.Join(parts, ", ") + ">"
}
