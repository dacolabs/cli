// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault_ContainsExpectedTranslators(t *testing.T) {
	reg := Default()
	expected := []string{
		"avro",
		"databricks-pyspark",
		"databricks-scala",
		"databricks-sql",
		"dqx-yaml",
		"gotypes",
		"markdown",
		"protobuf",
		"pydantic",
		"pyspark",
		"python",
		"scala",
		"spark-scala",
		"spark-sql",
	}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			tr, ok := reg[name]
			require.True(t, ok, "translator %q must be registered", name)
			require.NotNil(t, tr)
			assert.NotEmpty(t, tr.FileExtension(), "translator %q must report a file extension", name)
		})
	}
}

func TestDefault_TranslatorsHaveDistinctExtensionsOrAreReused(t *testing.T) {
	// Sanity check — every translator returns a non-empty extension. Multiple
	// translators may share an extension (e.g. .py).
	for name, tr := range Default() {
		assert.NotEmpty(t, tr.FileExtension(), name)
	}
}
