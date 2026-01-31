// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
)

// RunAddNewConnectionForm runs the interactive form for creating a new connection.
func RunAddNewConnectionForm(
	name, protocol, host, description *string,
	existingConns map[string]opendpi.Connection,
) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Connection name").
				Prompt(": ").
				Inline(true).
				Value(name).
				Validate(identifierValidator(existingConns)),
			huh.NewInput().
				Title("Type").
				Prompt(": ").
				Inline(true).
				Placeholder("kafka, postgresql, s3, ...").
				Value(protocol).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("type is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Host").
				Prompt(": ").
				Inline(true).
				Value(host).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("host is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Description").
				Prompt(": ").
				Inline(true).
				Placeholder("optional").
				Value(description),
		),
	).WithTheme(Theme()).Run()
}
