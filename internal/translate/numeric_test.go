// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func f64(v float64) *float64 { return &v }
func iptr(v int) *int        { return &v }

func TestNarrowInteger(t *testing.T) {
	cases := []struct {
		name string
		c    Constraints
		want IntKind
	}{
		{"no bounds", Constraints{}, Int64},
		{"only min", Constraints{Minimum: f64(0)}, Int64},
		{"only max", Constraints{Maximum: f64(127)}, Int64},
		{"byte signed", Constraints{Minimum: f64(-128), Maximum: f64(127)}, Int8},
		{"byte non-neg", Constraints{Minimum: f64(0), Maximum: f64(127)}, Int8},
		{"short", Constraints{Minimum: f64(-32768), Maximum: f64(32767)}, Int16},
		{"int", Constraints{Minimum: f64(0), Maximum: f64(2147483647)}, Int32},
		{"long", Constraints{Minimum: f64(0), Maximum: f64(9999999999)}, Int64},
		{"byte+1 overflows", Constraints{Minimum: f64(0), Maximum: f64(128)}, Int16},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, NarrowInteger(c.c))
		})
	}
}

func TestDecimalScale(t *testing.T) {
	cases := []struct {
		multipleOf float64
		want       int
	}{
		{0.01, 2},
		{0.001, 3},
		{0.1, 1},
		{1, 0},
		{5, 0},
		{1.5, -1}, // non-decimal step
		{0, -1},
		{-1, -1},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, DecimalScale(c.multipleOf), "multipleOf=%v", c.multipleOf)
	}
}

func TestDecimalPrecision(t *testing.T) {
	assert.Equal(t, 38, DecimalPrecision(nil, nil, 2))
	assert.Equal(t, 7, DecimalPrecision(nil, f64(99999.99), 2))
	assert.Equal(t, 7, DecimalPrecision(f64(-99999.99), f64(99999.99), 2))
	assert.Equal(t, 2, DecimalPrecision(nil, f64(0.99), 2))
	assert.Equal(t, 4, DecimalPrecision(f64(0), f64(1000), 0))
}

func TestNarrowNumber(t *testing.T) {
	t.Run("no multipleOf is double", func(t *testing.T) {
		kind, _ := NarrowNumber(Constraints{Minimum: f64(0), Maximum: f64(100)})
		assert.Equal(t, NumberDouble, kind)
	})

	t.Run("multipleOf 0.01 is decimal", func(t *testing.T) {
		kind, shape := NarrowNumber(Constraints{MultipleOf: f64(0.01)})
		assert.Equal(t, NumberDecimal, kind)
		assert.Equal(t, DecimalShape{Precision: 38, Scale: 2}, shape)
	})

	t.Run("multipleOf 0.01 with max bounds precision", func(t *testing.T) {
		kind, shape := NarrowNumber(Constraints{MultipleOf: f64(0.01), Maximum: f64(99999.99)})
		assert.Equal(t, NumberDecimal, kind)
		assert.Equal(t, DecimalShape{Precision: 7, Scale: 2}, shape)
	})

	t.Run("integer multipleOf alone stays double", func(t *testing.T) {
		kind, _ := NarrowNumber(Constraints{MultipleOf: f64(5)})
		assert.Equal(t, NumberDouble, kind)
	})

	t.Run("integer multipleOf with bounds becomes decimal", func(t *testing.T) {
		kind, shape := NarrowNumber(Constraints{MultipleOf: f64(5), Maximum: f64(1000)})
		assert.Equal(t, NumberDecimal, kind)
		assert.Equal(t, DecimalShape{Precision: 4, Scale: 0}, shape)
	})

	t.Run("non-decimal multipleOf stays double", func(t *testing.T) {
		kind, _ := NarrowNumber(Constraints{MultipleOf: f64(1.5)})
		assert.Equal(t, NumberDouble, kind)
	})
}

func TestIsNonNegative(t *testing.T) {
	assert.False(t, IsNonNegative(Constraints{}))
	assert.True(t, IsNonNegative(Constraints{Minimum: f64(0)}))
	assert.True(t, IsNonNegative(Constraints{Minimum: f64(1)}))
	assert.False(t, IsNonNegative(Constraints{Minimum: f64(-1)}))
	assert.True(t, IsNonNegative(Constraints{ExclusiveMinimum: f64(0)}))
}

func TestMaxStringLength(t *testing.T) {
	n, ok := MaxStringLength(Constraints{})
	assert.False(t, ok)
	assert.Equal(t, 0, n)

	n, ok = MaxStringLength(Constraints{MaxLength: iptr(64)})
	assert.True(t, ok)
	assert.Equal(t, 64, n)
}
