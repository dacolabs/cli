// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type connectionsListOptions struct {
	output string
}

func newConnectionsListCmd() *cobra.Command {
	opts := &connectionsListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all connections in the OpenDPI spec",
		Long:  `List all connections defined in the OpenDPI spec with their protocols and hosts.`,
		Example: `  # List connections in table format
  daco connections list

  # List connections as JSON
  daco connections list -o json

  # List connections as YAML
  daco connections list -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runConnectionsList(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func runConnectionsList(ctx *session.Context, opts *connectionsListOptions) error {
	if len(ctx.Spec.Connections) == 0 {
		fmt.Println("No connections defined.")
		return nil
	}

	switch opts.output {
	case "json":
		return printConnectionsJSON(ctx.Spec.Connections)
	case "yaml":
		return printConnectionsYAML(ctx.Spec.Connections)
	default:
		return printConnectionsTable(ctx.Spec.Connections, ctx.Spec.Ports)
	}
}

func printConnectionsTable(connections map[string]opendpi.Connection, ports map[string]opendpi.Port) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tPROTOCOL\tHOST\tDESCRIPTION\tUSED BY")

	// Sort connection names for consistent output
	names := make([]string, 0, len(connections))
	for name := range connections {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		conn := connections[name]

		// Truncate description
		desc := conn.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		if desc == "" {
			desc = "-"
		}

		// Count ports using this connection
		usageCount := countConnectionUsage(&conn, ports)
		usageStr := fmt.Sprintf("%d ports", usageCount)
		if usageCount == 1 {
			usageStr = "1 port"
		}
		if usageCount == 0 {
			usageStr = "-"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, conn.Protocol, conn.Host, desc, usageStr)
	}

	return w.Flush()
}

func printConnectionsJSON(connections map[string]opendpi.Connection) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(connections)
}

func printConnectionsYAML(connections map[string]opendpi.Connection) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(connections)
}

func countConnectionUsage(conn *opendpi.Connection, ports map[string]opendpi.Port) int {
	count := 0
	for _, port := range ports {
		for _, pc := range port.Connections {
			if pc.Connection != nil &&
				pc.Connection.Protocol == conn.Protocol &&
				pc.Connection.Host == conn.Host {
				count++
				break
			}
		}
	}
	return count
}
