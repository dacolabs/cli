// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package pyspark

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dacolabs/cli/internal/translate"
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

func (r *resolver) RefType(defName string) string {
	return "_" + defName
}

func (r *resolver) FormatDefName(defName string) string {
	return "_" + defName
}

func (r *resolver) FormatRootName(portName string) string {
	return portName + "_schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	c := f.Constraints

	switch f.Type {
	case "T.DoubleType()":
		if c.MultipleOf != nil {
			if scale := computeDecimalScale(*c.MultipleOf); scale > 0 {
				precision := computeDecimalPrecision(c.Maximum, scale)
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

// computeDecimalPrecision derives precision from maximum if available.
// Returns 38 (Spark default) if maximum is nil.
func computeDecimalPrecision(maximum *float64, scale int) int {
	if maximum == nil {
		return 38
	}
	max := math.Abs(*maximum)
	if max < 1 {
		return scale
	}
	intDigits := len(strconv.FormatFloat(math.Floor(max), 'f', 0, 64))
	return intDigits + scale
}

// computeDecimalScale returns the number of decimal places in multipleOf.
// Returns -1 if multipleOf is >= 1 (not a decimal fraction).
func computeDecimalScale(multipleOf float64) int {
	if multipleOf >= 1 || multipleOf <= 0 {
		return -1
	}
	s := strconv.FormatFloat(multipleOf, 'f', -1, 64)
	if i := strings.Index(s, "."); i >= 0 {
		return len(s) - i - 1
	}
	return -1
}

// inferIntegerType returns a narrower integer type if min/max allow it.
func inferIntegerType(min, max float64) string {
	switch {
	case min >= -128 && max <= 127:
		return "T.ByteType()"
	case min >= -32768 && max <= 32767:
		return "T.ShortType()"
	case min >= -2147483648 && max <= 2147483647:
		return "T.IntegerType()"
	default:
		return "T.LongType()"
	}
}

// inferNumberType returns FloatType if min/max fit in float32 bounds.
func inferNumberType(min, max float64) string {
	if min >= -math.MaxFloat32 && max <= math.MaxFloat32 {
		return "T.FloatType()"
	}
	return "T.DoubleType()"
}
