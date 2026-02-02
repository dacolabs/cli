// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"sort"
	"unicode/utf8"
	"strings"
	"text/tabwriter"

	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

func newPortsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all ports in the OpenDPI spec",
		Long: `List all ports defined in the OpenDPI spec.
Displays port names, schema types, descriptions, and connection information.`,
		Example: `  # List ports
  daco ports list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runPortsList(ctx)
		},
	}

	return cmd
}

func runPortsList(ctx *session.Context) error {
	if len(ctx.Spec.Ports) == 0 {
		fmt.Println("No ports defined.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tCONNECTIONS")

	names := make([]string, 0, len(ctx.Spec.Ports))
	for name := range ctx.Spec.Ports {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		port := ctx.Spec.Ports[name]

		desc := port.Description
		if utf8.RuneCountInString(desc) > 40 {
			desc = string([]rune(desc)[:37]) + "..."
		}
		if desc == "" {
			desc = "-"
		}

		connStrs := make([]string, 0, len(port.Connections))
		for _, pc := range port.Connections {
			connName := findConnectionName(pc.Connection, ctx.Spec.Connections)
			connStrs = append(connStrs, fmt.Sprintf("%s:%s", connName, pc.Location))
		}
		connsDisplay := "-"
		if len(connStrs) > 0 {
			connsDisplay = strings.Join(connStrs, ", ")
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", name, desc, connsDisplay)
	}

	return w.Flush()
}
