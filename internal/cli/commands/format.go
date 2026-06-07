// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli/engine"
)

func Format() *cobra.Command {
	return &cobra.Command{
		Use:   "format",
		Short: "Format every product, connection, and schema YAML in the project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}

			out, err := engine.ProjectFormat(cmd.Context(), engine.ProjectFormatInput{Prj: prj})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			for _, f := range out.Files {
				fmt.Fprintln(w, f)
			}
			fmt.Fprintf(w, "Formatted %d file(s)\n", len(out.Files))
			return nil
		},
	}
}
