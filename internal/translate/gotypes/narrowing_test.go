// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package gotypes

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowing_UnsignedByteIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerByteRange(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), "uint8")
}

func TestNarrowing_SignedShortIntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerShortRange(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), "int16")
}

func TestNarrowing_UnsignedInt32IntegerRange(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerInt32Range(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), "uint32")
}
