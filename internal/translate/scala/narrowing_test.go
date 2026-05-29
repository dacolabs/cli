// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package scala

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowing_ByteIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerByteRange(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), ": Byte")
}

func TestNarrowing_ShortIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerShortRange(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), ": Short")
}

func TestNarrowing_Int32IntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerInt32Range(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), ": Int")
}

func TestNarrowing_DecimalFromMultipleOf(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.DecimalFromMultipleOf(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), ": BigDecimal")
}
