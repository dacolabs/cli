// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show data product overview with ports and connections",
		Long: `Show a summary of the data product including its metadata,
all defined ports, and all defined connections.`,
		Example: `  # Describe the data product
  daco describe`,
		PreRunE: session.PreRunLoad,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runDescribe(ctx)
		},
	}
	return cmd
}

func runDescribe(ctx *session.Context) error {
	fields := []prompts.ResultField{
		{Label: "Title", Value: ctx.Spec.Info.Title},
		{Label: "Version", Value: ctx.Spec.Info.Version},
		{Label: "Description", Value: ctx.Spec.Info.Description},
	}
	prompts.PrintResult(fields, "")

	prompts.PrintResult([]prompts.ResultField{{Label: "Ports", Value: ""}}, "")
	if err := runPortsList(ctx); err != nil {
		return err
	}

	prompts.PrintResult([]prompts.ResultField{{Label: "Connections", Value: ""}}, "")
	return runConnectionsList(ctx)
}
