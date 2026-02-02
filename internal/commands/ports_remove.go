// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

type portsRemoveOptions struct {
	force bool
}

func newPortsRemoveCmd() *cobra.Command {
	opts := &portsRemoveOptions{}

	cmd := &cobra.Command{
		Use:   "remove [PORT_NAME]",
		Short: "Remove a port from the OpenDPI spec",
		Long: `Remove a port from the OpenDPI spec.
If no port name is provided, an interactive selection prompt is shown.
Requires confirmation unless --force is specified.`,
		Example: `  # Interactive selection
  daco ports remove

  # Remove with confirmation prompt
  daco ports remove user_events

  # Remove without confirmation
  daco ports remove user_events --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runPortsRemove(ctx, args, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runPortsRemove(ctx *session.Context, args []string, opts *portsRemoveOptions) error {
	var portName string
	if len(args) > 0 {
		portName = args[0]
	} else {
		if len(ctx.Spec.Ports) == 0 {
			return fmt.Errorf("no ports defined")
		}

		names := make([]string, 0, len(ctx.Spec.Ports))
		for name := range ctx.Spec.Ports {
			names = append(names, name)
		}
		sort.Strings(names)

		options := make([]huh.Option[string], 0, len(ctx.Spec.Ports))
		for _, name := range names {
			port := ctx.Spec.Ports[name]
			label := name
			if port.Description != "" {
				desc := port.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				label = fmt.Sprintf("%s - %s", name, desc)
			}
			options = append(options, huh.NewOption(label, name))
		}

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select port to remove").
					Options(options...).
					Filtering(true).
					Value(&portName).
					Height(10),
			),
		).WithTheme(prompts.Theme()).Run(); err != nil {
			return err
		}
	}

	port, exists := ctx.Spec.Ports[portName]
	if !exists {
		return fmt.Errorf("port %q not found", portName)
	}

	// Show port summary
	fmt.Printf("Port: %s\n", portName)
	if port.Description != "" {
		fmt.Printf("Description: %s\n", port.Description)
	}
	fmt.Printf("Connections: %d\n", len(port.Connections))
	fmt.Println()

	// Confirmation
	if !opts.force {
		var confirmed bool
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Are you sure you want to remove this port?").
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

	delete(ctx.Spec.Ports, portName)

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
		{Label: "Port", Value: portName},
	}, "✓ Port removed")

	return nil
}
