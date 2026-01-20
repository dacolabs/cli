// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromCommand(t *testing.T) {
	testDir, err := filepath.Abs("testdata/valid")
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	require.NoError(t, os.Chdir(testDir))

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// Before PreRunLoad
	assert.Nil(t, FromCommand(cmd))

	// After PreRunLoad
	require.NoError(t, PreRunLoad(cmd, nil))
	dacoCtx := FromCommand(cmd)
	require.NotNil(t, dacoCtx)
	assert.Equal(t, "My Data Product", dacoCtx.Spec.Info.Title)
}

func TestRequireFromCommand(t *testing.T) {
	tests := []struct {
		name      string
		setupDir  string // testdata path, empty means no setup needed
		loadFirst bool   // whether to call PreRunLoad before RequireFromCommand
		wantErr   bool
		wantTitle string
	}{
		{
			name:      "not loaded",
			setupDir:  "",
			loadFirst: false,
			wantErr:   true,
		},
		{
			name:      "loaded",
			setupDir:  "testdata/valid",
			loadFirst: true,
			wantErr:   false,
			wantTitle: "My Data Product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(origDir) }()

			if tt.setupDir != "" {
				testDir, err := filepath.Abs(tt.setupDir)
				require.NoError(t, err)
				require.NoError(t, os.Chdir(testDir))
			}

			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())

			if tt.loadFirst {
				require.NoError(t, PreRunLoad(cmd, nil))
			}

			dacoCtx, err := RequireFromCommand(cmd)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, dacoCtx.Spec.Info.Title)
		})
	}
}

func TestPreRunLoad_WithCommandExecution(t *testing.T) {
	tests := []struct {
		name      string
		dir       string // testdata path, empty means use t.TempDir()
		wantErr   error
		wantTitle string
		wantPath  string
	}{
		{
			name:      "valid project",
			dir:       "testdata/valid",
			wantErr:   nil,
			wantTitle: "My Data Product",
			wantPath:  "spec",
		},
		{
			name:    "not initialized",
			dir:     "",
			wantErr: ErrNotInitialized,
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

			var capturedCtx *Context

			rootCmd := &cobra.Command{
				Use:               "test",
				PersistentPreRunE: PreRunLoad,
			}

			subCmd := &cobra.Command{
				Use: "sub",
				RunE: func(cmd *cobra.Command, args []string) error {
					ctx, requireErr := RequireFromCommand(cmd)
					capturedCtx = ctx
					return requireErr
				},
			}
			rootCmd.AddCommand(subCmd)

			rootCmd.SetArgs([]string{"sub"})
			err := rootCmd.ExecuteContext(context.Background())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, capturedCtx)
			assert.Equal(t, tt.wantTitle, capturedCtx.Spec.Info.Title)
			assert.Equal(t, tt.wantPath, capturedCtx.Config.Path)
		})
	}
}
