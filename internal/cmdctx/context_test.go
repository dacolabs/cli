// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package cmdctx

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		dir       string // relative to testdata, empty means use t.TempDir()
		wantErr   error
		wantTitle string // only checked if wantErr is nil
		wantPath  string // only checked if wantErr is nil
	}{
		{
			name:    "not initialized",
			dir:     "", // empty dir with no daco.yaml
			wantErr: ErrNotInitialized,
		},
		{
			name:    "invalid config",
			dir:     "testdata/invalid-config",
			wantErr: ErrInvalidConfig,
		},
		{
			name:    "spec not found",
			dir:     "testdata/missing-spec",
			wantErr: ErrSpecNotFound,
		},
		{
			name:      "valid",
			dir:       "testdata/valid",
			wantErr:   nil,
			wantTitle: "My Data Product",
			wantPath:  "spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testDir string
			if tt.dir == "" {
				testDir = t.TempDir()
			} else {
				var err error
				testDir, err = filepath.Abs(tt.dir)
				require.NoError(t, err)
			}

			origDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(origDir) }()
			require.NoError(t, os.Chdir(testDir))

			ctx, err := Load(context.Background())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			dacoCtx := From(ctx)
			require.NotNil(t, dacoCtx)
			assert.Equal(t, tt.wantPath, dacoCtx.Config.Path)
			assert.Equal(t, tt.wantTitle, dacoCtx.Spec.Info.Title)
		})
	}
}

func TestFrom_NoContextStored(t *testing.T) {
	assert.Nil(t, From(context.Background()))
}
