// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package python

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/dacolabs/daco/internal/translate"
	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed python.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(template.ParseFS(tmplFS, "python.go.tmpl"))

// Translator translates JSON schemas to Python dataclass definitions.
type Translator struct{}

// FileExtension returns the file extension for Python files.
func (t *Translator) FileExtension() string {
	return ".py"
}

// Translate converts a JSON schema to Python dataclass definitions.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	// checks if any field type uses datetime, Literal, uuid, or Decimal.
	data.Extra["NeedsDatetimeImport"] = false
	data.Extra["NeedsLiteralImport"] = false
	data.Extra["NeedsUUIDImport"] = false
	data.Extra["NeedsDecimalImport"] = false
	scan := func(f translate.Field) {
		if strings.Contains(f.Type, "datetime.") {
			data.Extra["NeedsDatetimeImport"] = true
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

	// Dataclasses don't validate, so synthesise a __post_init__ body per class from the
	// field constraints, keyed by class name for the template to look up.
	postInit := make(map[string][]string)
	needsRe := false
	build := func(td translate.TypeDef) {
		var lines []string
		for i := range td.Fields {
			checks, usedRe := validationChecks(td.Fields[i])
			lines = append(lines, checks...)
			needsRe = needsRe || usedRe
		}
		if len(lines) > 0 {
			postInit[td.Name] = lines
		}
	}
	for i := range data.Defs {
		build(data.Defs[i])
	}
	build(data.Root)
	data.Extra["PostInit"] = postInit
	data.Extra["NeedsReImport"] = needsRe

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "python.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// validationChecks renders __post_init__ guard statements for a field's constraints,
// returning the body lines (8-space indented) and whether any used the re module.
// Optional fields are guarded with `is not None` so None passes.
func validationChecks(f translate.Field) ([]string, bool) {
	expr := "self." + f.Name
	c := f.Constraints
	usedRe := false

	type check struct{ cond, msg string }
	var checks []check
	add := func(cond, msg string) { checks = append(checks, check{cond, msg}) }

	if c.Minimum != nil {
		add(fmt.Sprintf("%s >= %s", expr, pyNum(*c.Minimum)), fmt.Sprintf("%s must be >= %s", f.Name, pyNum(*c.Minimum)))
	}
	if c.Maximum != nil {
		add(fmt.Sprintf("%s <= %s", expr, pyNum(*c.Maximum)), fmt.Sprintf("%s must be <= %s", f.Name, pyNum(*c.Maximum)))
	}
	if c.ExclusiveMinimum != nil {
		add(fmt.Sprintf("%s > %s", expr, pyNum(*c.ExclusiveMinimum)), fmt.Sprintf("%s must be > %s", f.Name, pyNum(*c.ExclusiveMinimum)))
	}
	if c.ExclusiveMaximum != nil {
		add(fmt.Sprintf("%s < %s", expr, pyNum(*c.ExclusiveMaximum)), fmt.Sprintf("%s must be < %s", f.Name, pyNum(*c.ExclusiveMaximum)))
	}
	if c.MultipleOf != nil {
		add(fmt.Sprintf("%s %% %s == 0", expr, pyNum(*c.MultipleOf)), fmt.Sprintf("%s must be a multiple of %s", f.Name, pyNum(*c.MultipleOf)))
	}
	if c.MinLength != nil {
		add(fmt.Sprintf("len(%s) >= %d", expr, *c.MinLength), fmt.Sprintf("%s must have length >= %d", f.Name, *c.MinLength))
	}
	if c.MaxLength != nil {
		add(fmt.Sprintf("len(%s) <= %d", expr, *c.MaxLength), fmt.Sprintf("%s must have length <= %d", f.Name, *c.MaxLength))
	}
	if c.Pattern != "" {
		add(fmt.Sprintf("re.fullmatch(%s, %s) is not None", strconv.Quote(c.Pattern), expr), fmt.Sprintf("%s must match %s", f.Name, c.Pattern))
		usedRe = true
	}
	if len(c.Enum) > 0 {
		parts := make([]string, len(c.Enum))
		for i, v := range c.Enum {
			parts[i] = translate.FormatLiteral(v)
		}
		add(fmt.Sprintf("%s in (%s)", expr, strings.Join(parts, ", ")), fmt.Sprintf("%s must be one of (%s)", f.Name, strings.Join(parts, ", ")))
	}
	if c.Const != nil {
		add(fmt.Sprintf("%s == %s", expr, translate.FormatLiteral(*c.Const)), fmt.Sprintf("%s must equal %s", f.Name, translate.FormatLiteral(*c.Const)))
	}

	guard := ""
	if f.Nullable {
		guard = expr + " is not None and "
	}
	lines := make([]string, 0, len(checks)*2)
	for _, ck := range checks {
		lines = append(lines, fmt.Sprintf("        if %snot (%s):", guard, ck.cond))
		lines = append(lines, fmt.Sprintf("            raise ValueError(%s)", strconv.Quote(ck.msg)))
	}
	return lines, usedRe
}

// pyNum renders a float bound as a Python numeric literal, dropping the trailing .0.
func pyNum(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}
