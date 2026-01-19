// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package prompts provides interactive terminal prompts for CLI commands.
package prompts

import (
	"errors"

	"github.com/charmbracelet/huh"
)

// Field factories for init command

// ProductNameInput returns an input field for the product name.
func ProductNameInput(value *string) *huh.Input {
	return huh.NewInput().
		Title("Product name").
		Value(value).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("product name is required")
			}
			return nil
		})
}

// ConfigTypeSelect returns a select field for choosing configuration type.
func ConfigTypeSelect(value *bool) *huh.Select[bool] {
	return huh.NewSelect[bool]().
		Title("Configuration type").
		Options(
			huh.NewOption("Root configuration", false),
			huh.NewOption("Extend existing configuration", true),
		).
		Value(value)
}

// ParentConfigPathInput returns an input field for the parent config path.
func ParentConfigPathInput(value *string) *huh.Input {
	return huh.NewInput().
		Title("Path to parent config").
		Placeholder("../daco.yaml").
		Value(value).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("parent config path is required")
			}
			return nil
		})
}

// SpecPathOverrideInput returns an input field for overriding the spec path.
func SpecPathOverrideInput(value *string) *huh.Input {
	return huh.NewInput().
		Title("Override spec path? (leave empty to inherit)").
		Value(value)
}

// NewSpecSourceSelect returns a select field for choosing spec source.
func NewSpecSourceSelect(value *bool) *huh.Select[bool] {
	return huh.NewSelect[bool]().
		Title("Spec source").
		Options(
			huh.NewOption("Create new spec", true),
			huh.NewOption("Use existing spec", false),
		).
		Value(value)
}

// SpecPathInput returns an input field for the spec path.
// The title and placeholder change based on whether creating a new spec or using an existing one.
func SpecPathInput(value *string, createSpec *bool) *huh.Input {
	return huh.NewInput().
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
		Value(value).
		Validate(func(s string) error {
			if s == "" && !*createSpec {
				return errors.New("spec path is required")
			}
			return nil
		})
}

// SpecVersionInput returns an input field for the spec version.
func SpecVersionInput(value *string) *huh.Input {
	return huh.NewInput().
		Title("Version").
		Placeholder("1.0.0").
		Value(value)
}

// SchemaOrganizationSelect returns a select field for choosing schema organization.
func SchemaOrganizationSelect(value *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		Title("Schema organization").
		Options(
			huh.NewOption("Modular (separate files)", "modular"),
			huh.NewOption("Components (in spec file)", "components"),
			huh.NewOption("Inline (in port definition)", "inline"),
		).
		Value(value)
}

// SpecFormatSelect returns a select field for choosing spec format.
func SpecFormatSelect(value *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		Title("Spec format").
		Options(
			huh.NewOption("YAML (recommended)", "yaml"),
			huh.NewOption("JSON", "json"),
		).
		Value(value)
}
