// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import (
	"math"
	"strconv"
	"strings"
)

// IntKind names a narrowed signed-integer kind derived from Minimum/Maximum bounds.
type IntKind int

const (
	Int8 IntKind = iota
	Int16
	Int32
	Int64
)

// NumberKind names a narrowed number kind derived from MultipleOf/Minimum/Maximum.
type NumberKind int

const (
	NumberDouble NumberKind = iota
	NumberDecimal
)

// DecimalShape carries the precision (total digits) and scale (digits after the decimal point)
// of a narrowed fixed-point decimal type.
type DecimalShape struct {
	Precision int
	Scale     int
}

// NarrowInteger picks the tightest signed integer kind for the given bounds.
// When either bound is missing, returns Int64 (the safe default).
func NarrowInteger(c Constraints) IntKind {
	if c.Minimum == nil || c.Maximum == nil {
		return Int64
	}
	lo, hi := *c.Minimum, *c.Maximum
	switch {
	case lo >= -128 && hi <= 127:
		return Int8
	case lo >= -32768 && hi <= 32767:
		return Int16
	case lo >= -2147483648 && hi <= 2147483647:
		return Int32
	}
	return Int64
}

// NarrowNumber decides whether a JSON Schema "number" should map to a fixed-point
// decimal (when MultipleOf indicates a fractional step) or to floating-point double.
// For NumberDecimal, the returned DecimalShape carries Precision and Scale.
func NarrowNumber(c Constraints) (NumberKind, DecimalShape) {
	if c.MultipleOf == nil {
		return NumberDouble, DecimalShape{}
	}
	scale := DecimalScale(*c.MultipleOf)
	if scale < 0 {
		return NumberDouble, DecimalShape{}
	}
	// An integer-valued multipleOf alone (e.g. 5) doesn't imply a decimal — only
	// promote to decimal when paired with bounds that nail down precision.
	if scale == 0 && c.Minimum == nil && c.Maximum == nil {
		return NumberDouble, DecimalShape{}
	}
	return NumberDecimal, DecimalShape{
		Precision: DecimalPrecision(c.Minimum, c.Maximum, scale),
		Scale:     scale,
	}
}

// DecimalScale returns the number of decimal places in multipleOf, or -1 if multipleOf
// is not a recognizable decimal fraction (e.g. 1.5, irrational steps).
func DecimalScale(multipleOf float64) int {
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

// DecimalPrecision derives total digit count from minimum and maximum bounds plus scale.
// Returns 38 (a common decimal-default) when both bounds are nil.
func DecimalPrecision(minimum, maximum *float64, scale int) int {
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

// IsNonNegative reports whether the constraints restrict the value to >= 0.
// Useful for translators with native unsigned integer types.
func IsNonNegative(c Constraints) bool {
	if c.Minimum != nil && *c.Minimum >= 0 {
		return true
	}
	if c.ExclusiveMinimum != nil && *c.ExclusiveMinimum >= 0 {
		return true
	}
	return false
}

// MaxStringLength returns the maximum string length when set.
func MaxStringLength(c Constraints) (int, bool) {
	if c.MaxLength != nil {
		return *c.MaxLength, true
	}
	return 0, false
}
