// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
)

// RunPortsDescribeForm prompts the user to select a port to describe.
func RunPortsDescribeForm(value *string, ports map[string]opendpi.Port) error {
	names := make([]string, 0, len(ports))
	for name := range ports {
		names = append(names, name)
	}
	sort.Strings(names)

	options := make([]huh.Option[string], 0, len(ports))
	for _, name := range names {
		port := ports[name]
		label := name
		if port.Description != "" {
			desc := port.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			label = fmt.Sprintf("%s - %s", name, desc)
		}
		options = append(options, huh.NewOption(label, name))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select port to describe").
				Options(options...).
				Filtering(true).
				Value(value).
				Height(10),
		),
	).WithTheme(Theme()).Run()
}
