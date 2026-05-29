// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package avro

import (
	"encoding/json"
	"testing"

	"github.com/dacolabs/daco/internal/translate/testcases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnum_NativeAvroEnum(t *testing.T) {
	out, err := (&Translator{}).Translate("t", testcases.StringEnum(), "schemas")
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(out, &doc))

	fieldTypes := extractFieldTypes(doc["fields"].([]any))
	status, ok := fieldTypes["status"].(map[string]any)
	require.True(t, ok, "status field type should be a map")
	assert.Equal(t, "enum", status["type"])
	assert.Equal(t, "Status", status["name"])

	symbols := status["symbols"].([]any)
	assert.Equal(t, []any{"ACTIVE", "INACTIVE", "PENDING"}, symbols)
}
