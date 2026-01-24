// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package prompts provides interactive terminal prompts for CLI commands.
package prompts

import (
	"errors"

	"github.com/charmbracelet/huh"
)

// RunInitForm runs the interactive form for the init command.
// It fills the provided pointers with user input.
func RunInitForm(extends, name, path, version, schemaOrg, format *string, createSpec *bool) error {
	isExtends := false
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Configuration type").
				Options(
					huh.NewOption("Root configuration", false),
					huh.NewOption("Extend existing configuration", true),
				).
				Value(&isExtends),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Path to parent config").
				Placeholder("../daco.yaml").
				Validate(func(s string) error {
					if s == "" {
						return errors.New("parent config path is required")
					}
					return nil
				}).
				Value(extends),
		).WithHideFunc(func() bool { return !isExtends }),
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Spec source").
				Options(
					huh.NewOption("Create new spec", true),
					huh.NewOption("Use existing spec", false),
				).
				Height(3).
				Value(createSpec),
		).WithHideFunc(func() bool { return isExtends }),
		huh.NewGroup(
			huh.NewInput().
				TitleFunc(func() string {
					if *createSpec {
						return "Path for new spec"
					}
					return "Path to existing spec"
				}, createSpec).
				PlaceholderFunc(func() string {
					if *createSpec {
						return "./spec"
					}
					return ""
				}, createSpec).
				Validate(func(s string) error {
					if s == "" && !*createSpec {
						return errors.New("spec path is required")
					}
					return nil
				}).
				Value(path),
		).WithHideFunc(func() bool { return isExtends }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Spec format").
				Options(
					huh.NewOption("YAML (recommended)", "yaml"),
					huh.NewOption("JSON", "json"),
				).
				Value(format),
		).WithHideFunc(func() bool { return isExtends || !*createSpec }),
		huh.NewGroup(
			huh.NewInput().
				Title("Product name").
				Validate(func(s string) error {
					if s == "" {
						return errors.New("product name is required")
					}
					return nil
				}).
				Value(name),
			huh.NewInput().
				Title("Version").
				Placeholder("1.0.0").
				Value(version),
		).WithHideFunc(func() bool { return isExtends || !*createSpec }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Schema organization").
				Options(
					huh.NewOption("Modular (separate files)", "modular"),
					huh.NewOption("Components (in spec file)", "components"),
					huh.NewOption("Inline (in port definition)", "inline"),
				).
				Value(schemaOrg),
		).WithHideFunc(func() bool { return isExtends }),
	).Run()
}
