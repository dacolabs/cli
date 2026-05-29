// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package pydantic provides Pydantic BaseModel schema translation utilities.
package pydantic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dacolabs/daco/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if schemaType == "string" && format != "" {
		switch format {
		case "date":
			return "datetime.date"
		case "date-time":
			return "datetime.datetime"
		}
	}

	switch schemaType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	default:
		return "str"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return "list[" + elemType + "]"
}

func (r *resolver) MapType(keyType, valueType string) string {
	return "dict[" + keyType + ", " + valueType + "]"
}

func (r *resolver) RefType(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatDefName(defName string) string {
	return translate.ToPascalCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return translate.ToPascalCase(portName) + "Schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	if lit := buildLiteralType(f.Constraints); lit != "" {
		f.Type = lit
	}

	params := buildFieldParams(f)

	if f.Nullable {
		f.Type = "Optional[" + f.Type + "]"
		if len(params) == 0 {
			f.Tag = " = None"
		} else {
			f.Tag = " = Field(default=None, " + strings.Join(params, ", ") + ")"
		}
		return
	}

	if len(params) > 0 {
		f.Tag = " = Field(" + strings.Join(params, ", ") + ")"
	}
}

// buildLiteralType returns a typing.Literal[...] type when the field is constrained
// to a fixed value (const) or a closed set (enum). Returns "" when neither applies.
func buildLiteralType(c translate.Constraints) string {
	if c.Const != nil {
		return "Literal[" + translate.FormatLiteral(*c.Const) + "]"
	}
	if len(c.Enum) > 0 {
		parts := make([]string, len(c.Enum))
		for i, v := range c.Enum {
			parts[i] = translate.FormatLiteral(v)
		}
		return "Literal[" + strings.Join(parts, ", ") + "]"
	}
	return ""
}

// buildFieldParams maps JSON Schema constraints to Pydantic Field keyword arguments.
// Output is in a stable, human-readable order.
func buildFieldParams(f *translate.Field) []string {
	var params []string
	c := f.Constraints

	// Constraints that conflict with Literal types (enum/const) — Pydantic raises
	// at model build time if both are present, so skip the numeric/length params
	// when the type is already restricted to specific values.
	literal := c.Const != nil || len(c.Enum) > 0

	if !literal && c.Minimum != nil {
		params = append(params, "ge="+formatNumber(*c.Minimum))
	}
	if !literal && c.Maximum != nil {
		params = append(params, "le="+formatNumber(*c.Maximum))
	}
	if !literal && c.ExclusiveMinimum != nil {
		params = append(params, "gt="+formatNumber(*c.ExclusiveMinimum))
	}
	if !literal && c.ExclusiveMaximum != nil {
		params = append(params, "lt="+formatNumber(*c.ExclusiveMaximum))
	}
	if !literal && c.MultipleOf != nil {
		params = append(params, "multiple_of="+formatNumber(*c.MultipleOf))
	}
	if !literal && c.MinLength != nil {
		params = append(params, fmt.Sprintf("min_length=%d", *c.MinLength))
	}
	if !literal && c.MaxLength != nil {
		params = append(params, fmt.Sprintf("max_length=%d", *c.MaxLength))
	}
	// MinItems/MaxItems use the same Pydantic v2 keyword on list fields.
	if !literal && c.MinItems != nil {
		params = append(params, fmt.Sprintf("min_length=%d", *c.MinItems))
	}
	if !literal && c.MaxItems != nil {
		params = append(params, fmt.Sprintf("max_length=%d", *c.MaxItems))
	}
	if !literal && c.Pattern != "" {
		params = append(params, "pattern="+strconv.Quote(c.Pattern))
	}
	if f.Description != "" {
		params = append(params, "description="+strconv.Quote(f.Description))
	}
	return params
}

// formatNumber renders a float64 with the same minimum-digits representation Python uses.
func formatNumber(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}
