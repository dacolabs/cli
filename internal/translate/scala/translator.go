// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package scala

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed scala.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"last": func(i int, fields []translate.Field) bool {
		return i == len(fields)-1
	},
}).ParseFS(tmplFS, "scala.go.tmpl"))

// Translator translates JSON schemas to Scala case class definitions.
type Translator struct{}

// FileExtension returns the file extension for Scala files.
func (t *Translator) FileExtension() string {
	return ".scala"
}

// Translate converts a JSON schema to Scala case class definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, outputDir string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	data.Extra["Package"] = filepath.Base(outputDir)

	// Sort fields so required fields come before optional fields.
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
	if err := tmpl.ExecuteTemplate(&buf, "scala.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
