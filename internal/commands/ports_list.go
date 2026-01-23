// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/dacolabs/cli/internal/context"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type portsListOptions struct {
	output string // output format: table, json, yaml
}

func registerPortsListCmd(parent *cobra.Command) {
	opts := &portsListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all ports in the OpenDPI spec",
		Long:  `List all ports defined in the OpenDPI spec with their names, types, and descriptions.`,
		Example: `  # List ports in table format
  daco ports list

  # List ports as JSON
  daco ports list -o json

  # List ports as YAML
  daco ports list -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := context.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runPortsList(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "table", "Output format (table, json, yaml)")

	parent.AddCommand(cmd)
}

func runPortsList(ctx *context.Context, opts *portsListOptions) error {
	if len(ctx.Spec.Ports) == 0 {
		fmt.Println("No ports defined.")
		return nil
	}

	switch opts.output {
	case "json":
		return printPortsJSON(ctx.Spec.Ports)
	case "yaml":
		return printPortsYAML(ctx.Spec.Ports)
	default:
		return printPortsTable(ctx.Spec.Ports, ctx.Spec.Connections)
	}
}

func printPortsTable(ports map[string]opendpi.Port, connections map[string]opendpi.Connection) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION\tCONNECTIONS")

	// Sort port names for consistent output
	names := make([]string, 0, len(ports))
	for name := range ports {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		port := ports[name]

		// Determine schema type
		schemaType := "-"
		if port.Schema != nil && port.Schema.Type != "" {
			schemaType = port.Schema.Type
		}

		// Truncate description
		desc := port.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		if desc == "" {
			desc = "-"
		}

		// Format connections
		connStrs := make([]string, 0, len(port.Connections))
		for _, pc := range port.Connections {
			connName := findConnectionName(pc.Connection, connections)
			connStrs = append(connStrs, fmt.Sprintf("%s (%s)", connName, pc.Location))
		}
		connsDisplay := "-"
		if len(connStrs) > 0 {
			connsDisplay = strings.Join(connStrs, ", ")
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, schemaType, desc, connsDisplay)
	}

	return w.Flush()
}

func printPortsJSON(ports map[string]opendpi.Port) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(ports)
}

func printPortsYAML(ports map[string]opendpi.Port) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(ports)
}

func findConnectionName(conn *opendpi.Connection, connections map[string]opendpi.Connection) string {
	if conn == nil {
		return "unknown"
	}
	for name, c := range connections {
		if c.Protocol == conn.Protocol && c.Host == conn.Host {
			return name
		}
	}
	return "unknown"
}
