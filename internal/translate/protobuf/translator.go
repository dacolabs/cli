// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package protobuf provides Protocol Buffers (proto3) schema translation utilities.
package protobuf

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/dacolabs/daco/internal/translate"
	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed protobuf.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"inc":            func(i int) int { return i + 1 },
	"screamingSnake": screamingSnake,
}).ParseFS(tmplFS, "protobuf.go.tmpl"))

// screamingSnake converts PascalCase to SCREAMING_SNAKE_CASE for enum prefixes.
func screamingSnake(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			sb.WriteRune('_')
		}
		if r >= 'a' && r <= 'z' {
			sb.WriteRune(r - 32)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Translator translates JSON schemas to Protocol Buffers (proto3) message definitions.
type Translator struct{}

// FileExtension returns the file extension for Protocol Buffers files.
func (t *Translator) FileExtension() string {
	return ".proto"
}

// Translate converts a JSON schema to proto3 message definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, outputDir string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	data.Extra["Package"] = filepath.Base(outputDir)

	// Extract enum:Name:sym1,sym2 markers from field types; replace with the enum name
	// and collect top-level enum declarations.
	enums := extractEnums(data.Defs, &data.Root)
	data.Extra["Enums"] = enums

	// sets sequential proto field numbers (= 1, = 2, ...) on each message.
	for i := range data.Defs {
		for j := range data.Defs[i].Fields {
			data.Defs[i].Fields[j].Tag = fmt.Sprintf("= %d", j+1)
		}
	}
	for j := range data.Root.Fields {
		data.Root.Fields[j].Tag = fmt.Sprintf("= %d", j+1)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "protobuf.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// protoEnum describes a top-level enum declaration in the generated .proto file.
type protoEnum struct {
	Name    string
	Symbols []string
}

// extractEnums walks every field, rewriting `enum:Name:sym1,sym2` markers to the
// enum name and accumulating declarations to render at the top of the file.
// Enums with identical (name, symbols) are deduplicated.
func extractEnums(defs []translate.TypeDef, root *translate.TypeDef) []protoEnum {
	seen := make(map[string]bool)
	var enums []protoEnum

	process := func(fields []translate.Field) {
		for i := range fields {
			rest, ok := strings.CutPrefix(fields[i].Type, "enum:")
			if !ok {
				continue
			}
			parts := strings.SplitN(rest, ":", 2)
			if len(parts) != 2 {
				continue
			}
			name := parts[0]
			symbols := strings.Split(parts[1], ",")
			fields[i].Type = name

			key := name + "|" + parts[1]
			if seen[key] {
				continue
			}
			seen[key] = true
			enums = append(enums, protoEnum{Name: name, Symbols: symbols})
		}
	}

	for i := range defs {
		process(defs[i].Fields)
	}
	process(root.Fields)
	return enums
}
