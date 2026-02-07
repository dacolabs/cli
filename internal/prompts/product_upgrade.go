// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

// RunProductUpgradeForm runs the interactive form for selecting a version bump type.
func RunProductUpgradeForm(bumpType *string, currentVersion string) error {
	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("patch  (%s)", previewBump(currentVersion, "patch")), "patch"),
		huh.NewOption(fmt.Sprintf("minor  (%s)", previewBump(currentVersion, "minor")), "minor"),
		huh.NewOption(fmt.Sprintf("major  (%s)", previewBump(currentVersion, "major")), "major"),
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Current version: %s — select bump type", currentVersion)).
				Options(options...).
				Value(bumpType),
		),
	).WithTheme(Theme()).Run()
}

// previewBump returns the bumped version string for display purposes.
func previewBump(version, bumpType string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "?"
	}

	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return "?"
	}

	switch bumpType {
	case "major":
		return fmt.Sprintf("%d.0.0", major+1)
	case "minor":
		return fmt.Sprintf("%d.%d.0", major, minor+1)
	case "patch":
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
	}
	return "?"
}
