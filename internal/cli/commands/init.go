// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli"
	"github.com/dacolabs/daco/internal/cli/engine"
)

func Init() *cobra.Command {
	var in engine.InitInput

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a daco project in the current directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			in.Usr = cli.User(cmd.Context())
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			in.Cwd = cwd

			out, err := engine.Init(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if out.Created {
				fmt.Fprintf(w, "Created %s and registered %q\n", out.ProjectPath, out.Name)
			} else {
				fmt.Fprintf(w, "Registered %q at %s\n", out.Name, out.ProjectPath)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Project name")
	return cmd
}