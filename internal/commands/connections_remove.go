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
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

type connectionsRemoveOptions struct {
	force bool
}

func newConnectionsRemoveCmd() *cobra.Command {
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
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runConnectionsRemove(ctx, args, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runConnectionsRemove(ctx *session.Context, args []string, opts *connectionsRemoveOptions) error {
	var connName string
	if len(args) > 0 {
		connName = args[0]
	} else {
		if len(ctx.Spec.Connections) == 0 {
			return fmt.Errorf("no connections defined")
		}

		names := make([]string, 0, len(ctx.Spec.Connections))
		for name := range ctx.Spec.Connections {
			names = append(names, name)
		}
		sort.Strings(names)

		options := make([]huh.Option[string], 0, len(ctx.Spec.Connections))
		for _, name := range names {
			conn := ctx.Spec.Connections[name]
			label := fmt.Sprintf("%s (%s://%s)", name, conn.Type, conn.Host)
			options = append(options, huh.NewOption(label, name))
		}

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select connection to remove").
					Options(options...).
					Filtering(true).
					Value(&connName).
					Height(10),
			),
		).WithTheme(prompts.Theme()).Run(); err != nil {
			return err
		}
	}

	conn, exists := ctx.Spec.Connections[connName]
	if !exists {
		return fmt.Errorf("connection %q not found", connName)
	}

	// Check if connection is in use
	var usingPorts []string
	for portName, port := range ctx.Spec.Ports {
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

	if len(usingPorts) > 0 {
		return fmt.Errorf("cannot remove connection %q: used by ports: %s", connName, strings.Join(usingPorts, ", "))
	}

	// Show connection summary
	fmt.Printf("Connection: %s\n", connName)
	fmt.Printf("Type:       %s\n", conn.Type)
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
		).WithTheme(prompts.Theme()).Run(); err != nil {
			return err
		}

		if !confirmed {
			fmt.Println("Removal canceled.")
			return nil
		}
	}

	delete(ctx.Spec.Connections, connName)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specDir := ctx.Config.Path
	if !filepath.IsAbs(specDir) {
		specDir = filepath.Join(cwd, specDir)
	}

	var writer opendpi.Writer
	if _, err := os.Stat(filepath.Join(specDir, "opendpi.json")); err == nil {
		writer = opendpi.JSONWriter
	} else {
		writer = opendpi.YAMLWriter
	}

	if err := writer.Write(ctx.Spec, specDir); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	prompts.PrintResult([]prompts.ResultField{
		{Label: "Connection", Value: connName},
	}, "✓ Connection removed")

	return nil
}
