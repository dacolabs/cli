// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package sparksql

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed sparksql.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"last": func(i int, fields []translate.Field) bool {
		return i == len(fields)-1
	},
}).ParseFS(tmplFS, "sparksql.go.tmpl"))

// Translator translates JSON schemas to Spark DDL CREATE TABLE statements.
type Translator struct{}

// FileExtension returns the file extension for SQL files.
func (t *Translator) FileExtension() string {
	return ".sql"
}

// Translate converts a JSON schema to a Spark DDL CREATE TABLE statement.
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
	if err := inlineStruct(data.Root.Fields, defMap, make(map[string]bool)); err != nil {
		return nil, fmt.Errorf("failed to inline struct definitions: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sparksql.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// inlineStruct replaces ref-type fields with inline STRUCT<...> definitions.
// It uses a visited set to detect circular type references and prevent infinite recursion.
func inlineStruct(fields []translate.Field, defs map[string]*translate.TypeDef, visited map[string]bool) error {
	for i := range fields {
		typeName := fields[i].Type
		if def, ok := defs[typeName]; ok {
			if visited[typeName] {
				return fmt.Errorf("circular type reference detected: %s", typeName)
			}
			visited[typeName] = true
			if err := inlineStruct(def.Fields, defs, visited); err != nil {
				return err
			}
			delete(visited, typeName)
			fields[i].Type = renderStruct(def.Fields)
		}
	}
	return nil
}

// renderStruct renders fields as STRUCT<field1: TYPE, field2: TYPE>.
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
