// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli/engine"
)

func Lint() *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Validate every product, connection, and schema YAML in the project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}

			out, err := engine.ProjectLint(cmd.Context(), engine.ProjectLintInput{Prj: prj})
			if err != nil {
				return err
			}

			if len(out.Issues) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK")
				return nil
			}

			msgs := make(engine.Errors, len(out.Issues))
			for i, iss := range out.Issues {
				msgs[i] = fmt.Sprintf("%s: %s", iss.Path, iss.Message)
			}
			return msgs
		},
	}
}
