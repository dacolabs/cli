// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/version"
)

func Version() *cobra.Command {
	var short bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		// Skip the root PersistentPreRunE: version should work without touching user settings.
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			if short {
				fmt.Fprintln(w, version.Short())
				return nil
			}
			fmt.Fprintln(w, version.Info())
			return nil
		},
	}
	cmd.Flags().BoolVar(&short, "short", false, "Print only the version string")
	return cmd
}
