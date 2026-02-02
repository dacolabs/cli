// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

func newConnectionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all connections in the OpenDPI spec",
		Long:  `List all connections defined in the OpenDPI spec with their protocols and hosts.`,
		Example: `  # List connections
  daco connections list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runConnectionsList(ctx)
		},
	}

	return cmd
}

func runConnectionsList(ctx *session.Context) error {
	if len(ctx.Spec.Connections) == 0 {
		fmt.Println("No connections defined.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tTYPE\tHOST\tDESCRIPTION\tUSED BY")

	names := make([]string, 0, len(ctx.Spec.Connections))
	for name := range ctx.Spec.Connections {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		conn := ctx.Spec.Connections[name]

		desc := conn.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		if desc == "" {
			desc = "-"
		}

		usageCount := 0
		for _, port := range ctx.Spec.Ports {
			for _, pc := range port.Connections {
				if pc.Connection != nil &&
					pc.Connection.Type == conn.Type &&
					pc.Connection.Host == conn.Host {
					usageCount++
					break
				}
			}
		}

		usageStr := fmt.Sprintf("%d ports", usageCount)
		if usageCount == 1 {
			usageStr = "1 port"
		}
		if usageCount == 0 {
			usageStr = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, conn.Type, conn.Host, desc, usageStr)
	}

	return w.Flush()
}
