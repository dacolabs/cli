// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"github.com/charmbracelet/huh"
)

// RunInitForm runs the interactive form for the init command.
func RunInitForm(name, version, description *string) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Data product name").
				Prompt(": ").
				Inline(true).
				Validate(requiredValidator("data product name")).
				Value(name),
			huh.NewInput().
				Title("Version").
				Prompt(": ").
				Inline(true).
				Placeholder("1.0.0").
				Value(version),
			huh.NewInput().
				Title("Description").
				Prompt(": ").
				Inline(true).
				Value(description),
		),
	).WithTheme(Theme()).Run()
}
