// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package python

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiteral_StringEnum(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringEnum(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "from typing import Optional, Literal")
	assert.Contains(t, result, `status: Literal["ACTIVE", "INACTIVE", "PENDING"]`)
}

func TestLiteral_StringConst(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringConst(), "schemas")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, `version: Literal["v1"]`)
}
