// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package protobuf

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNarrowing_ByteIntegerRangeUsesUInt32(t *testing.T) {
	// Protobuf has no 8/16-bit scalars; non-negative byte range widens to uint32.
	out, err := (&Translator{}).Translate("t", testcases.IntegerByteRange(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), "uint32 count")
}

func TestNarrowing_ShortIntegerSignedUsesInt32(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerShortRange(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), "int32 count")
}

func TestNarrowing_Int32NonNegativeUsesUInt32(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.IntegerInt32Range(), "schemas")
	require.NoError(t, err)
	assert.Contains(t, string(out), "uint32 count")
}
