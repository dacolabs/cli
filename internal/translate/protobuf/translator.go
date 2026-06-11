// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package protobuf provides Protocol Buffers (proto3) schema translation utilities.
package protobuf

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strconv"
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

	// Assign proto field numbers, attach protovalidate options, and fold constraints +
	// description into a leading comment. Track which imports the file needs.
	needsValidate := false
	applyFieldMeta := func(fields []translate.Field) {
		for j := range fields {
			if opt := protovalidateOption(fields[j].Type, fields[j].Constraints); opt != "" {
				fields[j].Tag = fmt.Sprintf("= %d [%s]", j+1, opt)
				needsValidate = true
			} else {
				fields[j].Tag = fmt.Sprintf("= %d", j+1)
			}
			fields[j].Description = fieldComment(fields[j].Description, fields[j].Constraints)
		}
	}
	for i := range data.Defs {
		applyFieldMeta(data.Defs[i].Fields)
	}
	applyFieldMeta(data.Root.Fields)
	data.Extra["NeedsValidate"] = needsValidate
	data.Extra["NeedsTimestamp"] = usesTimestamp(data.Defs, data.Root)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "protobuf.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// protovalidateOption builds a `(buf.validate.field).<scalar> = {...}` option from the
// field's value constraints, or "" when none apply. multipleOf has no protovalidate
// equivalent and survives only in the leading comment.
func protovalidateOption(typeStr string, c translate.Constraints) string {
	base := strings.TrimPrefix(typeStr, "optional ")
	var rules []string
	switch base {
	case "int32", "int64", "uint32", "uint64", "double", "float":
		if c.Minimum != nil {
			rules = append(rules, "gte: "+protoNum(*c.Minimum))
		}
		if c.Maximum != nil {
			rules = append(rules, "lte: "+protoNum(*c.Maximum))
		}
		if c.ExclusiveMinimum != nil {
			rules = append(rules, "gt: "+protoNum(*c.ExclusiveMinimum))
		}
		if c.ExclusiveMaximum != nil {
			rules = append(rules, "lt: "+protoNum(*c.ExclusiveMaximum))
		}
	case "string":
		if c.Pattern != "" {
			rules = append(rules, "pattern: "+strconv.Quote(c.Pattern))
		}
		if c.MinLength != nil {
			rules = append(rules, fmt.Sprintf("min_len: %d", *c.MinLength))
		}
		if c.MaxLength != nil {
			rules = append(rules, fmt.Sprintf("max_len: %d", *c.MaxLength))
		}
	default:
		return ""
	}
	if len(rules) == 0 {
		return ""
	}
	return fmt.Sprintf("(buf.validate.field).%s = {%s}", base, strings.Join(rules, ", "))
}

// protoNum renders a float bound as a proto numeric literal, dropping the trailing .0
// so 150.0 prints as 150.
func protoNum(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// fieldComment folds the description and the full constraint set into a single comment
// line, so constraints protovalidate can't express are still preserved losslessly.
func fieldComment(desc string, c translate.Constraints) string {
	text := translate.ConstraintsText(c)
	switch {
	case desc != "" && text != "":
		return desc + " (" + text + ")"
	case desc != "":
		return desc
	default:
		return text
	}
}

// usesTimestamp reports whether any field maps to google.protobuf.Timestamp, requiring
// the well-known-types import.
func usesTimestamp(defs []translate.TypeDef, root translate.TypeDef) bool {
	has := func(fields []translate.Field) bool {
		for i := range fields {
			if strings.Contains(fields[i].Type, "google.protobuf.Timestamp") {
				return true
			}
		}
		return false
	}
	for i := range defs {
		if has(defs[i].Fields) {
			return true
		}
	}
	return has(root.Fields)
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
