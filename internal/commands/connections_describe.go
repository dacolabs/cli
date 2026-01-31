// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type connectionsDescribeOptions struct {
	output string
}

func newConnectionsDescribeCmd() *cobra.Command {
	opts := &connectionsDescribeOptions{}

	cmd := &cobra.Command{
		Use:   "describe [CONNECTION_NAME]",
		Short: "Show detailed information about a connection",
		Long:  `Display complete connection details including which ports use it. If no connection name is provided, an interactive selection prompt is shown.`,
		Example: `  # Interactive selection
  daco connections describe

  # Show connection details
  daco connections describe kafka_prod

  # Show as JSON
  daco connections describe kafka_prod -o json

  # Show as YAML
  daco connections describe kafka_prod -o yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}

			var connName string
			if len(args) > 0 {
				connName = args[0]
			} else {
				connName, err = selectConnectionToDescribe(ctx.Spec.Connections)
				if err != nil {
					return err
				}
			}
			return runConnectionsDescribe(ctx, connName, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "text", "Output format (text, json, yaml)")

	return cmd
}

func selectConnectionToDescribe(conns map[string]opendpi.Connection) (string, error) {
	if len(conns) == 0 {
		return "", fmt.Errorf("no connections defined")
	}

	names := make([]string, 0, len(conns))
	for name := range conns {
		names = append(names, name)
	}
	sort.Strings(names)

	options := make([]huh.Option[string], 0, len(conns))
	for _, name := range names {
		conn := conns[name]
		label := fmt.Sprintf("%s (%s://%s)", name, conn.Type, conn.Host)
		options = append(options, huh.NewOption(label, name))
	}

	var selected string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select connection to describe").
				Options(options...).
				Filtering(true).
				Value(&selected).
				Height(10),
		),
	).Run(); err != nil {
		return "", err
	}

	return selected, nil
}

func runConnectionsDescribe(ctx *session.Context, connName string, opts *connectionsDescribeOptions) error {
	conn, exists := ctx.Spec.Connections[connName]
	if !exists {
		return fmt.Errorf("connection %q not found", connName)
	}

	switch opts.output {
	case "json":
		usedBy := findPortsUsingConnectionNames(&conn, ctx.Spec.Ports)
		output := map[string]any{
			"name":       connName,
			"connection": conn,
			"usedBy":     usedBy,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)

	case "yaml":
		usedBy := findPortsUsingConnectionNames(&conn, ctx.Spec.Ports)
		output := map[string]any{
			"name":       connName,
			"connection": conn,
			"usedBy":     usedBy,
		}
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		defer func() { _ = enc.Close() }()
		return enc.Encode(output)

	default:
		// Text format
		fmt.Printf("Name:        %s\n", connName)
		fmt.Printf("Protocol:    %s\n", conn.Type)
		fmt.Printf("Host:        %s\n", conn.Host)

		if conn.Description != "" {
			fmt.Printf("Description: %s\n", conn.Description)
		}

		// Variables
		if len(conn.Variables) > 0 {
			fmt.Println("\nVariables:")
			varNames := make([]string, 0, len(conn.Variables))
			for k := range conn.Variables {
				varNames = append(varNames, k)
			}
			sort.Strings(varNames)
			for _, k := range varNames {
				fmt.Printf("  %s: %v\n", k, conn.Variables[k])
			}
		}

		// Used by ports
		usedBy := findPortsUsingConnectionWithLocation(&conn, ctx.Spec.Ports)
		if len(usedBy) > 0 {
			fmt.Println("\nUsed by ports:")
			for _, usage := range usedBy {
				fmt.Printf("  - %s (location: %s)\n", usage.portName, usage.location)
			}
		} else {
			fmt.Println("\nUsed by ports: (none)")
		}
		return nil
	}
}

type portUsage struct {
	portName string
	location string
}

func findPortsUsingConnectionWithLocation(conn *opendpi.Connection, ports map[string]opendpi.Port) []portUsage {
	var usages []portUsage
	for portName, port := range ports {
		for _, pc := range port.Connections {
			if pc.Connection != nil &&
				pc.Connection.Type == conn.Type &&
				pc.Connection.Host == conn.Host {
				usages = append(usages, portUsage{
					portName: portName,
					location: pc.Location,
				})
			}
		}
	}
	sort.Slice(usages, func(i, j int) bool {
		return usages[i].portName < usages[j].portName
	})
	return usages
}

func findPortsUsingConnectionNames(conn *opendpi.Connection, ports map[string]opendpi.Port) []string {
	var usingPorts []string
	for portName, port := range ports {
		for _, pc := range port.Connections {
			if pc.Connection != nil &&
				pc.Connection.Type == conn.Type &&
				pc.Connection.Host == conn.Host {
				usingPorts = append(usingPorts, portName)
				break
			}
		}
	}
	sort.Strings(usingPorts)
	return usingPorts
}
