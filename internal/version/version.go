// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package version provides build-time version information.
package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags.
var (
	// Version is the semantic version (e.g., "0.1.0", "0.1.0-alpha.1").
	Version = "dev"
	// Commit is the git commit SHA.
	Commit = "none"
	// Date is the build date in RFC3339 format.
	Date = "unknown"
)

// Info returns formatted version information.
func Info() string {
	return fmt.Sprintf("daco version %s (commit: %s, built: %s, go: %s)",
		Version, Commit, Date, runtime.Version())
}

// Short returns just the version string.
func Short() string {
	return Version
}
