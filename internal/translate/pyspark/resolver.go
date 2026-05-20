// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dacolabs/daco/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date":
			return "T.DateType()"
		case "date-time":
			return "T.TimestampType()"
		case "uuid":
			return "T.StringType()"
		}
	}

	switch schemaType {
	case "string":
		return "T.StringType()"
	case "integer":
		return "T.LongType()"
	case "number":
		return "T.DoubleType()"
	case "boolean":
		return "T.BooleanType()"
	default:
		return "T.StringType()"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return fmt.Sprintf("T.ArrayType(%s)", elemType)
}

func (r *resolver) MapType(keyType, valueType string) string {
	return fmt.Sprintf("T.MapType(%s, %s)", keyType, valueType)
}

func (r *resolver) RefType(defName string) string {
	return "_" + translate.ToSnakeCase(defName)
}

func (r *resolver) FormatDefName(defName string) string {
	return "_" + translate.ToSnakeCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return portName + "_schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	c := f.Constraints

	switch f.Type {
	case "T.DoubleType()":
		if c.MultipleOf != nil {
			if scale := computeDecimalScale(*c.MultipleOf); scale > 0 || (scale == 0 && (c.Minimum != nil || c.Maximum != nil)) {
				precision := computeDecimalPrecision(c.Minimum, c.Maximum, scale)
				f.Type = fmt.Sprintf("T.DecimalType(%d, %d)", precision, scale)
				return
			}
		}
		if c.Minimum != nil && c.Maximum != nil {
			f.Type = inferNumberType(*c.Minimum, *c.Maximum)
		}

	case "T.LongType()":
		if c.Minimum != nil && c.Maximum != nil {
			f.Type = inferIntegerType(*c.Minimum, *c.Maximum)
		}
	}
}

// computeDecimalPrecision derives precision from minimum and maximum bounds.
// Returns 38 (Spark default) if both bounds are nil.
func computeDecimalPrecision(minimum, maximum *float64, scale int) int {
	var absMax float64
	switch {
	case minimum != nil && maximum != nil:
		absMax = math.Max(math.Abs(*minimum), math.Abs(*maximum))
	case maximum != nil:
		absMax = math.Abs(*maximum)
	case minimum != nil:
		absMax = math.Abs(*minimum)
	default:
		return 38
	}
	if absMax < 1 {
		return scale
	}
	intDigits := len(strconv.FormatFloat(math.Floor(absMax), 'f', 0, 64))
	return intDigits + scale
}

// computeDecimalScale returns the number of decimal places in multipleOf.
// Returns -1 if multipleOf is >= 1 (not a decimal fraction).
func computeDecimalScale(multipleOf float64) int {
	if multipleOf <= 0 {
		return -1
	}
	if multipleOf >= 1 {
		if multipleOf == math.Floor(multipleOf) {
			return 0
		}
		return -1
	}
	s := strconv.FormatFloat(multipleOf, 'f', -1, 64)
	if i := strings.Index(s, "."); i >= 0 {
		return len(s) - i - 1
	}
	return -1
}

// inferIntegerType returns a narrower integer type if min/max allow it.
func inferIntegerType(lo, hi float64) string {
	switch {
	case lo >= -128 && hi <= 127:
		return "T.ByteType()"
	case lo >= -32768 && hi <= 32767:
		return "T.ShortType()"
	case lo >= -2147483648 && hi <= 2147483647:
		return "T.IntegerType()"
	default:
		return "T.LongType()"
	}
}

// inferNumberType returns the Spark type for a JSON number field.
func inferNumberType(_, _ float64) string {
	return "T.DoubleType()"
}
