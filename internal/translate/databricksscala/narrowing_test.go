// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databricksscala

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowing_ByteIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerByteRange(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "ByteType")
}

func TestNarrowing_ShortIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerShortRange(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "ShortType")
}

func TestNarrowing_DecimalBounded(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.DecimalBounded(), "")
	require.NoError(t, err)
	assert.Contains(t, string(out), "DecimalType(7, 2)")
}
