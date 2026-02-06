// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package dqxyaml

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed dqxyaml.go.tmpl
var tmplFS embed.FS

var tmpl = template.Must(
	template.New("").Funcs(template.FuncMap{
		"yamlScalar": yamlScalar,
	}).ParseFS(tmplFS, "dqxyaml.go.tmpl"),
)

// Translator translates JSON Schema constraints to DQX YAML quality check definitions.
type Translator struct{}

// FileExtension returns the file extension for DQX YAML files.
func (t *Translator) FileExtension() string {
	return ".yaml"
}

// dqxCheck represents a single DQX quality check for template rendering.
type dqxCheck struct {
	Function string
	Args     []dqxArg
}

// dqxArg represents a single argument in a quality check.
type dqxArg struct {
	Key   string
	Value any   // original typed value (empty if List is set)
	List  []any // for list values like enum allowed
}

// Translate converts a JSON schema's constraints into DQX YAML quality checks.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	// Build def lookup: name → *TypeDef
	defsByName := make(map[string]*translate.TypeDef, len(data.Defs))
	for i := range data.Defs {
		defsByName[data.Defs[i].Name] = &data.Defs[i]
	}

	// Flatten TypeDef graph into checks with dot-notation column paths
	var checks []dqxCheck
	flattenFields(data.Root.Fields, "", defsByName, &checks, make(map[string]bool))

	data.Extra["Checks"] = checks

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "dqxyaml.go.tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// flattenFields walks the TypeDef graph recursively, building dot-notation
// column paths and collecting quality checks from field constraints.
func flattenFields(fields []translate.Field, prefix string, defs map[string]*translate.TypeDef, checks *[]dqxCheck, visited map[string]bool) {
	for i := range fields {
		colName := fields[i].Name
		if prefix != "" {
			colName = prefix + "." + fields[i].Name
		}

		// If field references a def, recurse into its fields
		if def, ok := defs[fields[i].Type]; ok {
			if visited[fields[i].Type] {
				continue
			}
			visited[fields[i].Type] = true
			flattenFields(def.Fields, colName, defs, checks, visited)
			delete(visited, fields[i].Type)
			continue
		}

		// Emit checks from field constraints
		if !fields[i].Nullable {
			*checks = append(*checks, newCheck("is_not_null", colName))
		}
		collectChecks(colName, &fields[i].Constraints, checks)
	}
}

// collectChecks inspects a field's constraints and appends DQX checks.
func collectChecks(colName string, c *translate.Constraints, checks *[]dqxCheck) {
	// Format checks
	collectFormatChecks(colName, c.Format, checks)

	// Enum
	if len(c.Enum) > 0 {
		*checks = append(*checks, dqxCheck{
			Function: "is_in_list",
			Args: []dqxArg{
				{Key: "column", Value: colName},
				{Key: "allowed", List: c.Enum},
			},
		})
	}

	// Const
	if c.Const != nil {
		*checks = append(*checks, newCheckWithArgs("is_equal_to",
			dqxArg{Key: "column", Value: colName},
			dqxArg{Key: "value", Value: *c.Const},
		))
	}

	// Numeric range checks
	collectNumericChecks(colName, c, checks)

	// String length checks
	if c.MinLength != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("length(%s) >= %d", quoteSQL(colName), *c.MinLength),
			fmt.Sprintf("%s must have minimum length of %d", colName, *c.MinLength),
		))
	}
	if c.MaxLength != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("length(%s) <= %d", quoteSQL(colName), *c.MaxLength),
			fmt.Sprintf("%s must have maximum length of %d", colName, *c.MaxLength),
		))
	}

	// Pattern
	if c.Pattern != "" {
		*checks = append(*checks, newCheckWithArgs("regex_match",
			dqxArg{Key: "column", Value: colName},
			dqxArg{Key: "regex", Value: c.Pattern},
		))
	}

	// MultipleOf
	if c.MultipleOf != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("%s %% %v = 0", quoteSQL(colName), *c.MultipleOf),
			fmt.Sprintf("%s must be a multiple of %v", colName, *c.MultipleOf),
		))
	}

	// Array constraints
	if c.MinItems != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("size(%s) >= %d", quoteSQL(colName), *c.MinItems),
			fmt.Sprintf("%s must have at least %d items", colName, *c.MinItems),
		))
	}
	if c.MaxItems != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("size(%s) <= %d", quoteSQL(colName), *c.MaxItems),
			fmt.Sprintf("%s must have at most %d items", colName, *c.MaxItems),
		))
	}
}

// collectFormatChecks emits checks for the JSON Schema "format" keyword.
func collectFormatChecks(colName, format string, checks *[]dqxCheck) {
	switch format {
	case "date":
		*checks = append(*checks, newCheck("is_valid_date", colName))
	case "date-time":
		*checks = append(*checks, newCheck("is_valid_timestamp", colName))
	case "ipv4":
		*checks = append(*checks, newCheck("is_valid_ipv4_address", colName))
	case "ipv6":
		*checks = append(*checks, newCheck("is_valid_ipv6_address", colName))
	case "uuid":
		*checks = append(*checks, newCheckWithArgs("regex_match",
			dqxArg{Key: "column", Value: colName},
			dqxArg{Key: "regex", Value: `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`},
		))
	case "email":
		*checks = append(*checks, newCheckWithArgs("regex_match",
			dqxArg{Key: "column", Value: colName},
			dqxArg{Key: "regex", Value: `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`},
		))
	}
}

// collectNumericChecks emits checks for minimum, maximum, exclusiveMinimum, and exclusiveMaximum.
func collectNumericChecks(colName string, c *translate.Constraints, checks *[]dqxCheck) {
	hasMin := c.Minimum != nil
	hasMax := c.Maximum != nil

	switch {
	case hasMin && hasMax:
		*checks = append(*checks, dqxCheck{
			Function: "is_in_range",
			Args: []dqxArg{
				{Key: "column", Value: colName},
				{Key: "min_limit", Value: *c.Minimum},
				{Key: "max_limit", Value: *c.Maximum},
			},
		})
	case hasMin:
		*checks = append(*checks, newCheckWithArgs("is_not_less_than",
			dqxArg{Key: "column", Value: colName},
			dqxArg{Key: "limit", Value: *c.Minimum},
		))
	case hasMax:
		*checks = append(*checks, newCheckWithArgs("is_not_greater_than",
			dqxArg{Key: "column", Value: colName},
			dqxArg{Key: "limit", Value: *c.Maximum},
		))
	}

	if c.ExclusiveMinimum != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("%s > %v", quoteSQL(colName), *c.ExclusiveMinimum),
			fmt.Sprintf("%s must be greater than %v", colName, *c.ExclusiveMinimum),
		))
	}
	if c.ExclusiveMaximum != nil {
		*checks = append(*checks, newSQLCheck(
			fmt.Sprintf("%s < %v", quoteSQL(colName), *c.ExclusiveMaximum),
			fmt.Sprintf("%s must be less than %v", colName, *c.ExclusiveMaximum),
		))
	}
}

// newCheck creates a check with a single column argument.
func newCheck(function, colName string) dqxCheck {
	return dqxCheck{
		Function: function,
		Args:     []dqxArg{{Key: "column", Value: colName}},
	}
}

// newCheckWithArgs creates a check with explicit arguments.
func newCheckWithArgs(function string, args ...dqxArg) dqxCheck {
	return dqxCheck{Function: function, Args: args}
}

// newSQLCheck creates a sql_expression check.
func newSQLCheck(expression, msg string) dqxCheck {
	return dqxCheck{
		Function: "sql_expression",
		Args: []dqxArg{
			{Key: "expression", Value: expression},
			{Key: "msg", Value: msg},
		},
	}
}

// yamlScalar formats any Go value as a YAML scalar string suitable for
// inline embedding in a YAML template. It delegates to yaml.Marshal to
// guarantee correct quoting for all YAML special characters, boolean
// literals, numeric strings, etc.
func yamlScalar(v any) string {
	if v == nil {
		return "null"
	}
	// Render whole-number floats as integers (18.0 → 18).
	if f, ok := v.(float64); ok {
		if f == float64(int64(f)) {
			v = int64(f)
		}
	}
	out, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return strings.TrimRight(string(out), "\n")
}

// quoteSQL backtick-quotes a column name for use in Spark SQL expressions.
// Dot-notation paths are split and each segment is quoted individually:
//
//	"address.street" → "`address`.`street`"
func quoteSQL(colName string) string {
	parts := strings.Split(colName, ".")
	for i, p := range parts {
		parts[i] = "`" + strings.ReplaceAll(p, "`", "``") + "`"
	}
	return strings.Join(parts, ".")
}

