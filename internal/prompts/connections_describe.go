// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
)

// RunConnectionsDescribeForm prompts the user to select a connection to describe.
func RunConnectionsDescribeForm(value *string, connections map[string]opendpi.Connection) error {
	names := make([]string, 0, len(connections))
	for name := range connections {
		names = append(names, name)
	}
	sort.Strings(names)

	options := make([]huh.Option[string], 0, len(connections))
	for _, name := range names {
		conn := connections[name]
		label := fmt.Sprintf("%s (%s://%s)", name, conn.Type, conn.Host)
		options = append(options, huh.NewOption(label, name))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select connection to describe").
				Options(options...).
				Filtering(true).
				Value(value).
				Height(10),
		),
	).WithTheme(Theme()).Run()
}
