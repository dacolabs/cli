// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package protobuf provides Protocol Buffers (proto3) schema translation utilities.
package protobuf

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
)

//go:embed protobuf.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "protobuf.go.tmpl"))

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
