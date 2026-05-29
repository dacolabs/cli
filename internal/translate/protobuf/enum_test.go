// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package protobuf

import (
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnum_TopLevelDeclaration(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringEnum(), "schemas")
	require.NoError(t, err)
	result := string(out)

	assert.Contains(t, result, "enum Status {")
	assert.Contains(t, result, "STATUS_UNSPECIFIED = 0;")
	assert.Contains(t, result, "ACTIVE = 1;")
	assert.Contains(t, result, "INACTIVE = 2;")
	assert.Contains(t, result, "PENDING = 3;")
	assert.Contains(t, result, "Status status = 1;")
}
