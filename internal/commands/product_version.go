// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

func newProductVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the data product spec version",
		Long:  `Show the current version of the data product as defined in the OpenDPI spec.`,
		Example: `  # Show the product version
  daco product version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runProductVersion(ctx)
		},
	}
	return cmd
}

func runProductVersion(ctx *session.Context) error {
	prompts.PrintResult([]prompts.ResultField{
		{Label: "Version", Value: ctx.Spec.Info.Version},
	}, "")
	return nil
}
