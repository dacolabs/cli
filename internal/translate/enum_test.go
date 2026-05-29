// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumStringSymbols(t *testing.T) {
	assert.Nil(t, EnumStringSymbols(Constraints{}))
	assert.Equal(t, []string{"a", "b"}, EnumStringSymbols(Constraints{Enum: []any{"a", "b"}}))
	// Mixed types — no native string enum possible.
	assert.Nil(t, EnumStringSymbols(Constraints{Enum: []any{"a", 1}}))
}

func TestFormatLiteral(t *testing.T) {
	assert.Equal(t, `"hello"`, FormatLiteral("hello"))
	assert.Equal(t, "True", FormatLiteral(true))
	assert.Equal(t, "False", FormatLiteral(false))
	assert.Equal(t, "42", FormatLiteral(42.0))
	assert.Equal(t, "3.14", FormatLiteral(3.14))
	assert.Equal(t, "None", FormatLiteral(nil))
}
