// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package version provides build-time version information.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
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

func init() {
	// If version wasn't set via ldflags, try to get it from build info.
	// This works when installed via "go install module@version".
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}

	// Try to get commit and date from build settings if not set via ldflags.
	if Commit == "none" || Date == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					if Commit == "none" && len(setting.Value) >= 7 {
						Commit = setting.Value[:7]
					}
				case "vcs.time":
					if Date == "unknown" {
						Date = setting.Value
					}
				}
			}
		}
	}
}

// Info returns formatted version information.
func Info() string {
	return fmt.Sprintf("daco version %s (commit: %s, built: %s, go: %s)",
		Version, Commit, Date, runtime.Version())
}

// Short returns just the version string.
func Short() string {
	return Version
}
