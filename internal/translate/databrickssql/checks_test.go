// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package databrickssql

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecks_EnumProducesCheckIn(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringEnum(), "")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "CONSTRAINT status_enum CHECK (`status` IN ('ACTIVE', 'INACTIVE', 'PENDING'))")
}

func TestChecks_PatternProducesRLike(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringPattern(), "")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "CONSTRAINT email_pattern CHECK (`email` RLIKE '")
}

func TestChecks_ConstProducesEquality(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringConst(), "")
	require.NoError(t, err)
	result := string(out)
	assert.Contains(t, result, "CONSTRAINT version_const CHECK (`version` = 'v1')")
}
