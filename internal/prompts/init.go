// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package prompts provides interactive terminal prompts for CLI commands.
package prompts

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// RunInitForm runs the interactive form for the init command.
// It fills the provided pointers with user input.
func RunInitForm(name, version, description *string) error {
	theme := huh.ThemeBase16()
	theme.FieldSeparator = lipgloss.NewStyle().SetString("\n").MarginBottom(1)
	theme.Form.Base = theme.Form.Base.MarginTop(1)
	theme.Group.Base = theme.Group.Base.MarginTop(1)
	theme.Focused.Title = theme.Focused.Title.Foreground(lipgloss.Color("#f9ca24"))
	theme.Blurred.Title = theme.Blurred.Title.Foreground(lipgloss.Color("#bababa"))

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
	).WithTheme(theme).Run()
}
