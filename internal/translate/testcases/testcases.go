// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package testcases holds canonical JSON Schema inputs used by each translator's
// tests to verify constraint-driven type narrowing and runtime validation.
//
// Each fixture is a *jsonschema.Schema with a single property covering one
// narrowing or validation axis (byte integer, decimal-from-multipleOf, varchar,
// pattern, etc.). Translators import this package from _test.go files so the
// matrix of "which constraint, which target" stays in lockstep.
package testcases

import "github.com/dacolabs/jsonschema-go/jsonschema"

func f64(v float64) *float64 { return &v }
func iptr(v int) *int        { return &v }

// IntegerByteRange — `count` is an integer in [0, 127], should narrow to Int8.
func IntegerByteRange() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"count"},
		Properties: map[string]*jsonschema.Schema{
			"count": {Type: "integer", Minimum: f64(0), Maximum: f64(127)},
		},
	}
}

// IntegerShortRange — `count` is an integer in [-32768, 32767], narrows to Int16.
func IntegerShortRange() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"count"},
		Properties: map[string]*jsonschema.Schema{
			"count": {Type: "integer", Minimum: f64(-32768), Maximum: f64(32767)},
		},
	}
}

// IntegerInt32Range — `count` is a non-negative integer up to int32 max, narrows to Int32 / uint32.
func IntegerInt32Range() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"count"},
		Properties: map[string]*jsonschema.Schema{
			"count": {Type: "integer", Minimum: f64(0), Maximum: f64(2147483647)},
		},
	}
}

// DecimalFromMultipleOf — `price` is a number with 2-decimal step, narrows to Decimal(38, 2).
func DecimalFromMultipleOf() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"price"},
		Properties: map[string]*jsonschema.Schema{
			"price": {Type: "number", MultipleOf: f64(0.01)},
		},
	}
}

// DecimalBounded — `price` has 2-decimal step AND maximum, narrows to Decimal(7, 2).
func DecimalBounded() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"price"},
		Properties: map[string]*jsonschema.Schema{
			"price": {Type: "number", MultipleOf: f64(0.01), Maximum: f64(99999.99)},
		},
	}
}

// StringMaxLength — `code` has maxLength 32, targets with VARCHAR / max_length validation use it.
func StringMaxLength() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"code"},
		Properties: map[string]*jsonschema.Schema{
			"code": {Type: "string", MaxLength: iptr(32)},
		},
	}
}

// StringPattern — `email` has a regex pattern, used by validators like Pydantic.
func StringPattern() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"email"},
		Properties: map[string]*jsonschema.Schema{
			"email": {Type: "string", Pattern: `^[^@]+@[^@]+\.[^@]+$`},
		},
	}
}

// NumberRange — `temperature` has min/max but no multipleOf; stays double.
func NumberRange() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"temperature"},
		Properties: map[string]*jsonschema.Schema{
			"temperature": {Type: "number", Minimum: f64(-273.15), Maximum: f64(1000)},
		},
	}
}

// StringEnum — `status` is constrained to a closed set of string values.
func StringEnum() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"status"},
		Properties: map[string]*jsonschema.Schema{
			"status": {Type: "string", Enum: []any{"ACTIVE", "INACTIVE", "PENDING"}},
		},
	}
}

// StringConst — `version` is pinned to a single string value.
func StringConst() *jsonschema.Schema {
	v := any("v1")
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"version"},
		Properties: map[string]*jsonschema.Schema{
			"version": {Type: "string", Const: &v},
		},
	}
}
