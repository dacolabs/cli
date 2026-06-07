// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runs `daco` with no args and no TTY. tcell can't initialize a screen against
// /dev/null so it fails fast — proving the TUI entry path was taken (cobra
// found no subcommand, RunE fired, tui.Run was called). The alternative would
// be that cobra fell back to printing help (no `tui:` in the error).
func TestE2E_TUI_LaunchAttempted(t *testing.T) {
	home := t.TempDir()
	cwd := t.TempDir()
	cmd := exec.Command(binaryPath)
	cmd.Dir = cwd
	cmd.Env = append([]string{"HOME=" + home, "TERM=dumb"}, "PATH=/usr/bin:/bin")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.Error(t, err, "daco with no args + no TTY should error from tui init, not exit cleanly")
	output := stderr.String()
	assert.True(t,
		strings.Contains(output, "tui:") ||
			strings.Contains(output, "tcell") ||
			strings.Contains(output, "open /dev/tty"),
		"expected TUI initialization error in stderr, got: %s", output)
}