// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package sparksql

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowing_ByteIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerByteRange(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "count TINYINT")
}

func TestNarrowing_ShortIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerShortRange(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "count SMALLINT")
}

func TestNarrowing_Int32IntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerInt32Range(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "count INT")
}

func TestNarrowing_DecimalFromMultipleOf(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.DecimalFromMultipleOf(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "price DECIMAL(38, 2)")
}

func TestNarrowing_DecimalBounded(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.DecimalBounded(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "price DECIMAL(7, 2)")
}

func TestNarrowing_VarcharFromMaxLength(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringMaxLength(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "code VARCHAR(32)")
}

func TestNarrowing_NumberRangeStaysDouble(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.NumberRange(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "temperature DOUBLE")
}
