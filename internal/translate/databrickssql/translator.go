// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databrickssql

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/dacolabs/daco/internal/translate"
)

//go:embed databrickssql.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"last": func(i int, fields []translate.Field) bool {
		return i == len(fields)-1
	},
	"lastConstraint": func(i int, constraints []constraintClause) bool {
		return i == len(constraints)-1
	},
	"sqlEscape": func(s string) string {
		return strings.ReplaceAll(s, "'", "''")
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
	inlineStruct(data.Root.Fields, defMap, make(map[string]bool))

	// Collect Delta CHECK constraints from field constraints (pattern/enum/const/length).
	data.Extra["Constraints"] = collectConstraints(data.Root.Fields)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "databrickssql.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// constraintClause holds a single Delta CONSTRAINT name/expression pair.
type constraintClause struct {
	Name string
	Expr string
}

// collectConstraints walks fields and emits Delta CHECK constraints for the constraints
// that aren't already encoded as the column's NOT NULL / VARCHAR(n) attributes.
func collectConstraints(fields []translate.Field) []constraintClause {
	var out []constraintClause
	for _, f := range fields {
		col := "`" + f.Name + "`"

		if syms := translate.EnumStringSymbols(f.Constraints); syms != nil {
			parts := make([]string, len(syms))
			for i, s := range syms {
				parts[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
			}
			out = append(out, constraintClause{
				Name: f.Name + "_enum",
				Expr: fmt.Sprintf("%s IN (%s)", col, strings.Join(parts, ", ")),
			})
		}

		if f.Constraints.Const != nil {
			if s, ok := (*f.Constraints.Const).(string); ok {
				out = append(out, constraintClause{
					Name: f.Name + "_const",
					Expr: fmt.Sprintf("%s = '%s'", col, strings.ReplaceAll(s, "'", "''")),
				})
			} else {
				out = append(out, constraintClause{
					Name: f.Name + "_const",
					Expr: fmt.Sprintf("%s = %v", col, *f.Constraints.Const),
				})
			}
		}

		if f.Constraints.Pattern != "" {
			out = append(out, constraintClause{
				Name: f.Name + "_pattern",
				Expr: fmt.Sprintf("%s RLIKE '%s'", col, strings.ReplaceAll(f.Constraints.Pattern, "'", "''")),
			})
		}

		if f.Constraints.MinLength != nil {
			out = append(out, constraintClause{
				Name: f.Name + "_minlen",
				Expr: fmt.Sprintf("length(%s) >= %d", col, *f.Constraints.MinLength),
			})
		}
	}
	return out
}

func inlineStruct(fields []translate.Field, defs map[string]*translate.TypeDef, visited map[string]bool) {
	for i := range fields {
		name := fields[i].Type
		if def, ok := defs[name]; ok {
			if visited[name] {
				continue
			}
			visited[name] = true
			inlineStruct(def.Fields, defs, visited)
			fields[i].Type = renderStruct(def.Fields)
			delete(visited, name)
		}
	}
}

func renderStruct(fields []translate.Field) string {
	parts := make([]string, len(fields))
	for i := range fields {
		parts[i] = fields[i].Name + ": " + fields[i].Type
		if fields[i].Tag != "" {
			parts[i] += fields[i].Tag
		}
	}
	return "STRUCT<" + strings.Join(parts, ", ") + ">"
}
