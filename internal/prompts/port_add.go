// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
)

// RunPortsAddForm runs the interactive form for adding a port.
// It fills the provided pointers with user input.
func RunPortsAddForm(
	name, description, connection, location *string,
	existingPorts map[string]opendpi.Port,
	existingConns map[string]opendpi.Connection,
) error {
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Port name").
				Prompt(": ").
				Inline(true).
				Value(name).
				Validate(identifierValidator(existingPorts)),
			huh.NewInput().
				Title("Description").
				Prompt(": ").
				Inline(true).
				Value(description),
		),
	).WithTheme(Theme()).Run(); err != nil {
		return err
	}

	if len(existingConns) > 0 {
		options := []huh.Option[string]{
			huh.NewOption("None", ""),
		}
		for n, c := range existingConns {
			label := fmt.Sprintf("%s (%s://%s)", n, c.Type, c.Host)
			options = append(options, huh.NewOption(label, n))
		}

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Connection").
					Options(options...).
					Value(connection),
				huh.NewInput().
					Title("Location").
					Prompt(": ").
					Inline(true).
					Placeholder("e.g., events_topic, users_table").
					Value(location).
					Validate(requiredValidator("location")),
			),
		).WithTheme(Theme()).Run(); err != nil {
			return err
		}
	}

	return nil
}
