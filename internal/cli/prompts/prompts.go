// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package prompts provides interactive terminal prompts for CLI commands.
package prompts

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Theme returns the shared huh theme used across all CLI forms.
func Theme() *huh.Theme {
	theme := huh.ThemeBase16()
	theme.FieldSeparator = lipgloss.NewStyle().SetString("\n").MarginBottom(1)
	theme.Form.Base = theme.Form.Base.MarginTop(1)
	theme.Group.Base = theme.Group.Base.MarginTop(1)
	theme.Focused.Title = theme.Focused.Title.Foreground(lipgloss.Color("#f9ca24"))
	theme.Blurred.Title = theme.Blurred.Title.Foreground(lipgloss.Color("#bababa"))
	return theme
}

// ResultField is a label-value pair for PrintResult.
type ResultField struct {
	Label string
	Value string
}

// PrintResult prints a styled summary with green checkmarks and gray labels.
func PrintResult(fields []ResultField, successMsg string) {
	success := lipgloss.NewStyle().Foreground(lipgloss.Color("#27ca3f"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#bababa"))
	check := success.Render("✓")

	fmt.Println()
	for _, f := range fields {
		fmt.Printf("%s %s %s\n", check, label.Render(f.Label+":"), f.Value)
	}

	if successMsg != "" {
		fmt.Println(success.Render("\n" + successMsg))
	}
}

func identifierValidator[T any](existing map[string]T) func(string) error {
	return func(s string) error {
		if s == "" {
			return errors.New("name is required")
		}
		for i, r := range s {
			if i == 0 && !unicode.IsLetter(r) && r != '_' {
				return errors.New("must start with letter or underscore")
			}
			if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return errors.New("must contain only letters, numbers, underscores")
			}
		}
		if _, exists := existing[s]; exists {
			return fmt.Errorf("%q already exists", s)
		}
		return nil
	}
}

func requiredValidator(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}
