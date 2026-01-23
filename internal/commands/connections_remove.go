// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/context"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/spf13/cobra"
)

type connectionsRemoveOptions struct {
	force bool
}

func registerConnectionsRemoveCmd(parent *cobra.Command) {
	opts := &connectionsRemoveOptions{}

	cmd := &cobra.Command{
		Use:   "remove [CONNECTION_NAME]",
		Short: "Remove a connection from the OpenDPI spec",
		Long:  `Remove a connection from the OpenDPI spec. Cannot remove connections that are in use by ports. If no connection name is provided, an interactive selection prompt is shown.`,
		Example: `  # Interactive selection
  daco connections remove

  # Remove with confirmation
  daco connections remove unused_conn

  # Remove without confirmation
  daco connections remove unused_conn --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := context.RequireFromCommand(cmd)
			if err != nil {
				return err
			}

			var connName string
			if len(args) > 0 {
				connName = args[0]
			} else {
				var err error
				connName, err = selectConnectionToRemove(ctx.Spec.Connections)
				if err != nil {
					return err
				}
			}
			return runConnectionsRemove(ctx, connName, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	parent.AddCommand(cmd)
}

func runConnectionsRemove(ctx *context.Context, connName string, opts *connectionsRemoveOptions) error {
	conn, exists := ctx.Spec.Connections[connName]
	if !exists {
		return fmt.Errorf("connection %q not found", connName)
	}

	// Check if connection is in use
	usingPorts := findPortsUsingConnection(&conn, ctx.Spec.Ports)
	if len(usingPorts) > 0 {
		return fmt.Errorf("cannot remove connection %q: used by ports: %s", connName, strings.Join(usingPorts, ", "))
	}

	// Show connection summary
	fmt.Printf("Connection: %s\n", connName)
	fmt.Printf("Protocol:   %s\n", conn.Protocol)
	fmt.Printf("Host:       %s\n", conn.Host)
	if conn.Description != "" {
		fmt.Printf("Description: %s\n", conn.Description)
	}
	fmt.Println()

	// Confirmation
	if !opts.force {
		var confirmed bool
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Are you sure you want to remove this connection?").
					Affirmative("Yes, remove").
					Negative("No, cancel").
					Value(&confirmed),
			),
		).Run(); err != nil {
			return err
		}

		if !confirmed {
			fmt.Println("Removal canceled.")
			return nil
		}
	}

	// Remove the connection
	delete(ctx.Spec.Connections, connName)

	// Get working directory and spec directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specDir := ctx.Config.Path
	if !filepath.IsAbs(specDir) {
		specDir = filepath.Join(cwd, specDir)
	}

	// Determine writer format based on existing spec file
	var writer opendpi.Writer
	if _, err := os.Stat(filepath.Join(specDir, "opendpi.json")); err == nil {
		writer = opendpi.JSONWriter
	} else {
		writer = opendpi.YAMLWriter
	}

	if err := writer.Write(ctx.Spec, ctx.Config); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	fmt.Printf("Connection %q removed.\n", connName)
	return nil
}

func selectConnectionToRemove(conns map[string]opendpi.Connection) (string, error) {
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
		label := fmt.Sprintf("%s (%s://%s)", name, conn.Protocol, conn.Host)
		options = append(options, huh.NewOption(label, name))
	}

	var selected string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select connection to remove").
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

func findPortsUsingConnection(conn *opendpi.Connection, ports map[string]opendpi.Port) []string {
	var usingPorts []string
	for portName, port := range ports {
		for _, pc := range port.Connections {
			if pc.Connection != nil &&
				pc.Connection.Protocol == conn.Protocol &&
				pc.Connection.Host == conn.Host {
				usingPorts = append(usingPorts, portName)
				break
			}
		}
	}
	sort.Strings(usingPorts)
	return usingPorts
}
