// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"errors"

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
				Validate(func(s string) error {
					if s == "" {
						return errors.New("data product name is required")
					}
					return nil
				}).
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
