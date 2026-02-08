// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package markdown

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed markdown.go.tmpl
var tmplFS embed.FS

var funcMap = template.FuncMap{
	"formatConstraints": formatConstraints,
}

var tmpl = template.Must(template.New("markdown.go.tmpl").Funcs(funcMap).ParseFS(tmplFS, "markdown.go.tmpl"))

// Translator translates JSON schemas to markdown documentation.
type Translator struct{}

// FileExtension returns the file extension for markdown files.
func (t *Translator) FileExtension() string {
	return ".md"
}

// Translate converts a JSON schema to markdown documentation.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "markdown.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// formatConstraints formats the constraints for a field as a human-readable string.
func formatConstraints(c translate.Constraints) string {
	var parts []string

	if len(c.Enum) > 0 {
		enumVals := make([]string, len(c.Enum))
		for i, v := range c.Enum {
			enumVals[i] = fmt.Sprintf("`%v`", v)
		}
		parts = append(parts, "enum: "+strings.Join(enumVals, ", "))
	}

	if c.Const != nil {
		parts = append(parts, fmt.Sprintf("const: `%v`", *c.Const))
	}

	if c.Pattern != "" {
		parts = append(parts, fmt.Sprintf("pattern: `%s`", c.Pattern))
	}

	if c.Format != "" {
		parts = append(parts, fmt.Sprintf("format: %s", c.Format))
	}

	if c.MinLength != nil {
		parts = append(parts, fmt.Sprintf("minLength: %d", *c.MinLength))
	}

	if c.MaxLength != nil {
		parts = append(parts, fmt.Sprintf("maxLength: %d", *c.MaxLength))
	}

	if c.Minimum != nil {
		parts = append(parts, fmt.Sprintf("minimum: %v", *c.Minimum))
	}

	if c.Maximum != nil {
		parts = append(parts, fmt.Sprintf("maximum: %v", *c.Maximum))
	}

	if c.ExclusiveMinimum != nil {
		parts = append(parts, fmt.Sprintf("exclusiveMinimum: %v", *c.ExclusiveMinimum))
	}

	if c.ExclusiveMaximum != nil {
		parts = append(parts, fmt.Sprintf("exclusiveMaximum: %v", *c.ExclusiveMaximum))
	}

	if c.MultipleOf != nil {
		parts = append(parts, fmt.Sprintf("multipleOf: %v", *c.MultipleOf))
	}

	if c.MinItems != nil {
		parts = append(parts, fmt.Sprintf("minItems: %d", *c.MinItems))
	}

	if c.MaxItems != nil {
		parts = append(parts, fmt.Sprintf("maxItems: %d", *c.MaxItems))
	}

	return strings.Join(parts, ", ")
}
