// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package dqxyaml

import (
	"fmt"
	"testing"

	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// testCheck mirrors the DQX YAML structure for test assertions.
type testCheck struct {
	Criticality string `yaml:"criticality"`
	Check       struct {
		Function  string         `yaml:"function"`
		Arguments map[string]any `yaml:"arguments"`
	} `yaml:"check"`
}

func TestTranslate_RequiredFields(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id", "name"},
		Properties: map[string]*jsonschema.Schema{
			"id":       {Type: "integer"},
			"name":     {Type: "string"},
			"optional": {Type: "string"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	assertCheckExists(t, checks, "is_not_null", "id")
	assertCheckExists(t, checks, "is_not_null", "name")
}

func TestTranslate_EnumConstraint(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"status": {
				Type: "string",
				Enum: []any{"active", "inactive", "deleted"},
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_in_list", "status")
	allowed := checks[0].Check.Arguments["allowed"].([]any)
	assert.Len(t, allowed, 3)
	assert.Equal(t, "active", allowed[0])
	assert.Equal(t, "inactive", allowed[1])
	assert.Equal(t, "deleted", allowed[2])
}

func TestTranslate_ConstConstraint(t *testing.T) {
	constVal := any("fixed")
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"version": {
				Type:  "string",
				Const: &constVal,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_equal_to", "version")
	assert.Equal(t, "fixed", checks[0].Check.Arguments["value"])
}

func TestTranslate_PatternConstraint(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"country": {
				Type:    "string",
				Pattern: `^[A-Z]{2}$`,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "regex_match", "country")
	assert.Equal(t, `^[A-Z]{2}$`, checks[0].Check.Arguments["regex"])
}

func TestTranslate_NumericRange(t *testing.T) {
	min := 0.0
	max := 100.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"score": {
				Type:    "number",
				Minimum: &min,
				Maximum: &max,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_in_range", "score")
	assert.EqualValues(t, 0, checks[0].Check.Arguments["min_limit"])
	assert.EqualValues(t, 100, checks[0].Check.Arguments["max_limit"])
}

func TestTranslate_MinimumOnly(t *testing.T) {
	min := 18.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"age": {
				Type:    "integer",
				Minimum: &min,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_not_less_than", "age")
	assert.EqualValues(t, 18, checks[0].Check.Arguments["limit"])
}

func TestTranslate_MaximumOnly(t *testing.T) {
	max := 200.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"weight": {
				Type:    "number",
				Maximum: &max,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_not_greater_than", "weight")
	assert.EqualValues(t, 200, checks[0].Check.Arguments["limit"])
}

func TestTranslate_ExclusiveBounds(t *testing.T) {
	exMin := 0.0
	exMax := 100.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"rate": {
				Type:             "number",
				ExclusiveMinimum: &exMin,
				ExclusiveMaximum: &exMax,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	assertCheck(t, checks[0], "sql_expression", "")
	assert.Equal(t, "`rate` > 0", checks[0].Check.Arguments["expression"])
	assertCheck(t, checks[1], "sql_expression", "")
	assert.Equal(t, "`rate` < 100", checks[1].Check.Arguments["expression"])
}

func TestTranslate_StringLength(t *testing.T) {
	minLen := 3
	maxLen := 50
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"username": {
				Type:      "string",
				MinLength: &minLen,
				MaxLength: &maxLen,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	assertCheck(t, checks[0], "sql_expression", "")
	assert.Equal(t, "length(`username`) >= 3", checks[0].Check.Arguments["expression"])
	assertCheck(t, checks[1], "sql_expression", "")
	assert.Equal(t, "length(`username`) <= 50", checks[1].Check.Arguments["expression"])
}

func TestTranslate_FormatDate(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"birth_date": {Type: "string", Format: "date"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_valid_date", "birth_date")
}

func TestTranslate_FormatDateTime(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"created_at": {Type: "string", Format: "date-time"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_valid_timestamp", "created_at")
}

func TestTranslate_FormatUUID(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id": {Type: "string", Format: "uuid"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "regex_match", "id")
	regex := checks[0].Check.Arguments["regex"].(string)
	assert.Contains(t, regex, "[0-9a-fA-F]")
}

func TestTranslate_FormatEmail(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"email": {Type: "string", Format: "email"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "regex_match", "email")
	regex := checks[0].Check.Arguments["regex"].(string)
	assert.Contains(t, regex, "@")
}

func TestTranslate_FormatIPv4(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"ip": {Type: "string", Format: "ipv4"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_valid_ipv4_address", "ip")
}

func TestTranslate_FormatIPv6(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"ip": {Type: "string", Format: "ipv6"},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_valid_ipv6_address", "ip")
}

func TestTranslate_NestedObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"address": {
				Type:     "object",
				Required: []string{"street"},
				Properties: map[string]*jsonschema.Schema{
					"street":  {Type: "string"},
					"country": {Type: "string", Pattern: `^[A-Z]{2}$`},
				},
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	nullCheck := findCheckByFunction(checks, "is_not_null")
	require.NotNil(t, nullCheck)
	assert.Equal(t, "address.street", nullCheck.Check.Arguments["column"])

	regexCheck := findCheckByFunction(checks, "regex_match")
	require.NotNil(t, regexCheck)
	assert.Equal(t, "address.country", regexCheck.Check.Arguments["column"])
	assert.Equal(t, `^[A-Z]{2}$`, regexCheck.Check.Arguments["regex"])
}

func TestTranslate_WithDefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"address"},
		Properties: map[string]*jsonschema.Schema{
			"address": {Ref: "#/$defs/Address"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Address": {
				Type:     "object",
				Required: []string{"street"},
				Properties: map[string]*jsonschema.Schema{
					"street": {Type: "string"},
					"city":   {Type: "string"},
				},
			},
		},
	}

	checks := translateSchema(t, schema)

	// address.$ref resolves to object → recurse, so no is_not_null for address itself
	// but street is required inside Address
	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_not_null", "address.street")
}

func TestTranslate_ArrayConstraints(t *testing.T) {
	minItems := 1
	maxItems := 10
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"tags": {
				Type:     "array",
				MinItems: &minItems,
				MaxItems: &maxItems,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	assertCheck(t, checks[0], "sql_expression", "")
	assert.Equal(t, "size(`tags`) >= 1", checks[0].Check.Arguments["expression"])
	assertCheck(t, checks[1], "sql_expression", "")
	assert.Equal(t, "size(`tags`) <= 10", checks[1].Check.Arguments["expression"])
}

func TestTranslate_MultipleOf(t *testing.T) {
	multipleOf := 5.0
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"quantity": {
				Type:       "integer",
				MultipleOf: &multipleOf,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "sql_expression", "")
	assert.Equal(t, "`quantity` % 5 = 0", checks[0].Check.Arguments["expression"])
}

func TestTranslate_MultipleConstraints(t *testing.T) {
	min := 0.0
	max := 150.0
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"age"},
		Properties: map[string]*jsonschema.Schema{
			"age": {
				Type:    "integer",
				Minimum: &min,
				Maximum: &max,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	assertCheck(t, checks[0], "is_not_null", "age")
	assertCheck(t, checks[1], "is_in_range", "age")
}

func TestTranslate_NoConstraints(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	}

	translator := &Translator{}
	output, err := translator.Translate("test", schema, "")
	require.NoError(t, err)
	assert.Equal(t, "[]\n", string(output))
}

func TestTranslate_AllCriticalities(t *testing.T) {
	schema := &jsonschema.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"id": {Type: "integer"},
		},
	}

	checks := translateSchema(t, schema)

	for _, c := range checks {
		assert.Equal(t, "error", c.Criticality)
	}
}

func TestTranslate_ChainedDefs(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"customer": {Ref: "#/$defs/Customer"},
		},
		Defs: map[string]*jsonschema.Schema{
			"Customer": {
				Type:     "object",
				Required: []string{"name"},
				Properties: map[string]*jsonschema.Schema{
					"name":    {Type: "string"},
					"address": {Ref: "#/$defs/Address"},
				},
			},
			"Address": {
				Type:     "object",
				Required: []string{"city"},
				Properties: map[string]*jsonschema.Schema{
					"city": {Type: "string"},
				},
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 2)
	columns := []string{
		checks[0].Check.Arguments["column"].(string),
		checks[1].Check.Arguments["column"].(string),
	}
	assert.Contains(t, columns, "customer.name")
	assert.Contains(t, columns, "customer.address.city")
}

func TestFlattenFields_CircularRef(t *testing.T) {
	// Construct a circular defMap: A references B, B references A.
	// flattenFields must not infinite-loop; it should skip the back-edge.
	defs := map[string]*translate.TypeDef{
		"A": {Name: "A", Fields: []translate.Field{
			{Name: "name", Type: "string", Nullable: false},
			{Name: "b", Type: "B"},
		}},
		"B": {Name: "B", Fields: []translate.Field{
			{Name: "a", Type: "A"},
		}},
	}
	fields := []translate.Field{{Name: "root", Type: "A"}}

	var checks []dqxCheck
	flattenFields(fields, "", defs, &checks, make(map[string]bool))

	// A was expanded; B was expanded; B's child "A" was skipped (visited).
	// Only the leaf field "root.name" (non-nullable) should produce a check.
	require.Len(t, checks, 1)
	assert.Equal(t, "is_not_null", checks[0].Function)
	assert.Equal(t, "root.name", checks[0].Args[0].Value)
}

func TestQuoteSQL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"username", "`username`"},
		{"address.street", "`address`.`street`"},
		{"customer.address.city", "`customer`.`address`.`city`"},
		{"user`name", "`user``name`"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, quoteSQL(tt.input))
	}
}

func TestYamlScalar(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		// Safe plain string scalars — returned unchanged.
		{"hello", "hello"},
		{`^[A-Z]{2}$`, `^[A-Z]{2}$`},
		{"simple column", "simple column"},
		{"it's", "it's"},

		// Strings that look like numbers — must be quoted to stay strings.
		{"42", `"42"`},
		{"18.5", `"18.5"`},

		// Leading YAML indicator characters — must be quoted.
		{"`rate` > 0", "'`rate` > 0'"},
		{"@directive", "'@directive'"},
		{"#comment", "'#comment'"},
		{"*alias", "'*alias'"},
		{": mapping", "': mapping'"},
		{"[a-z]+", "'[a-z]+'"},
		{"{foo}", "'{foo}'"},

		// Inline comment marker.
		{"foo # bar", "'foo # bar'"},

		// Mapping-value indicator in the middle.
		{"key: value", "'key: value'"},

		// Trailing colon.
		{"foo:", "'foo:'"},

		// YAML 1.1 booleans and null — must be quoted to stay strings.
		{"off", `"off"`},
		{"OFF", `"OFF"`},
		{"true", `"true"`},
		{"True", `"True"`},
		{"yes", `"yes"`},
		{"Yes", `"Yes"`},
		{"no", `"no"`},
		{"null", `"null"`},
		{"y", `"y"`},
		{"n", `"n"`},

		// Empty string.
		{"", `""`},

		// Value that needs quoting AND contains a single quote.
		{"#it's", "'#it''s'"},

		// Numeric types — rendered as plain YAML numbers.
		{float64(42), "42"},
		{float64(18.5), "18.5"},
		{float64(0), "0"},
		{float64(100), "100"},
		{int64(7), "7"},

		// Booleans — rendered as plain YAML booleans.
		{true, "true"},
		{false, "false"},

		// Nil — rendered as null.
		{nil, "null"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			assert.Equal(t, tt.want, yamlScalar(tt.input))
		})
	}
}

func TestTranslate_ConstBooleanString(t *testing.T) {
	constVal := any("off")
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"flag": {
				Type:  "string",
				Const: &constVal,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_equal_to", "flag")
	assert.Equal(t, "off", checks[0].Check.Arguments["value"])
}

func TestTranslate_PatternWithHash(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"color": {
				Type:    "string",
				Pattern: `^#[0-9a-fA-F]{6}$`,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "regex_match", "color")
	assert.Equal(t, `^#[0-9a-fA-F]{6}$`, checks[0].Check.Arguments["regex"])
}

func TestTranslate_EnumWithBooleanStrings(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"toggle": {
				Type: "string",
				Enum: []any{"on", "off", "auto"},
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_in_list", "toggle")
	allowed := checks[0].Check.Arguments["allowed"].([]any)
	assert.Equal(t, "on", allowed[0])
	assert.Equal(t, "off", allowed[1])
	assert.Equal(t, "auto", allowed[2])
}

func TestTranslate_PatternWithBracket(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"code": {
				Type:    "string",
				Pattern: `[a-z]+`,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "regex_match", "code")
	assert.Equal(t, "[a-z]+", checks[0].Check.Arguments["regex"])
}

func TestTranslate_ConstPreservesStringType(t *testing.T) {
	constVal := any("123")
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"code": {
				Type:  "string",
				Const: &constVal,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_equal_to", "code")
	assert.Equal(t, "123", checks[0].Check.Arguments["value"])
}

func TestTranslate_ConstNumeric(t *testing.T) {
	constVal := any(float64(42))
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"version": {
				Const: &constVal,
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_equal_to", "version")
	assert.EqualValues(t, 42, checks[0].Check.Arguments["value"])
}

func TestTranslate_EnumWithNumericStrings(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"code": {
				Type: "string",
				Enum: []any{"123", "456"},
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_in_list", "code")
	allowed := checks[0].Check.Arguments["allowed"].([]any)
	assert.Equal(t, "123", allowed[0])
	assert.Equal(t, "456", allowed[1])
}

func TestTranslate_EnumWithSpecialChars(t *testing.T) {
	schema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"status": {
				Enum: []any{"on", "key: value", "[bracket]"},
			},
		},
	}

	checks := translateSchema(t, schema)

	require.Len(t, checks, 1)
	assertCheck(t, checks[0], "is_in_list", "status")
	allowed := checks[0].Check.Arguments["allowed"].([]any)
	assert.Equal(t, "on", allowed[0])
	assert.Equal(t, "key: value", allowed[1])
	assert.Equal(t, "[bracket]", allowed[2])
}

func TestFileExtension(t *testing.T) {
	translator := &Translator{}
	assert.Equal(t, ".yaml", translator.FileExtension())
}

// Test helpers

func translateSchema(t *testing.T, schema *jsonschema.Schema) []testCheck {
	t.Helper()
	translator := &Translator{}
	output, err := translator.Translate("test", schema, "")
	require.NoError(t, err)

	var checks []testCheck
	require.NoError(t, yaml.Unmarshal(output, &checks))
	return checks
}

func findCheckByFunction(checks []testCheck, function string) *testCheck {
	for i := range checks {
		if checks[i].Check.Function == function {
			return &checks[i]
		}
	}
	return nil
}

func assertCheck(t *testing.T, c testCheck, function, column string) {
	t.Helper()
	assert.Equal(t, "error", c.Criticality)
	assert.Equal(t, function, c.Check.Function)
	if column != "" {
		assert.Equal(t, column, c.Check.Arguments["column"])
	}
}

func assertCheckExists(t *testing.T, checks []testCheck, function, column string) {
	t.Helper()
	for _, c := range checks {
		if c.Check.Function == function && c.Check.Arguments["column"] == column {
			assert.Equal(t, "error", c.Criticality)
			return
		}
	}
	t.Errorf("check with function=%q column=%q not found", function, column)
}
