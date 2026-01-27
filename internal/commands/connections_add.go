// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

type connectionsAddOptions struct {
	name           string
	protocol       string
	host           string
	description    string
	nonInteractive bool
}

func newConnectionsAddCmd() *cobra.Command {
	opts := &connectionsAddOptions{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new connection to the OpenDPI spec",
		Long:  `Add a new infrastructure connection to the OpenDPI spec.`,
		Example: `  # Interactive mode
  daco connections add

  # Non-interactive
  daco connections add -n kafka_prod -p kafka --host broker:9092 --non-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runConnectionsAdd(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Connection name")
	cmd.Flags().StringVarP(&opts.protocol, "protocol", "p", "", "Protocol (kafka, postgresql, mysql, s3, http, etc.)")
	cmd.Flags().StringVar(&opts.host, "host", "", "Host/endpoint")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Description")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run without prompts")

	return cmd
}

func runConnectionsAdd(ctx *session.Context, opts *connectionsAddOptions) error {
	// Non-interactive validation
	if opts.nonInteractive {
		if opts.name == "" {
			return fmt.Errorf("--name is required in non-interactive mode")
		}
		if opts.protocol == "" {
			return fmt.Errorf("--protocol is required in non-interactive mode")
		}
		if opts.host == "" {
			return fmt.Errorf("--host is required in non-interactive mode")
		}
		if _, exists := ctx.Spec.Connections[opts.name]; exists {
			return fmt.Errorf("connection %q already exists", opts.name)
		}
	}

	var result prompts.ConnectionAddResult
	if opts.nonInteractive {
		result = prompts.ConnectionAddResult{
			Name: opts.name,
			Connection: opendpi.Connection{
				Protocol:    opts.protocol,
				Host:        opts.host,
				Description: opts.description,
			},
		}
	} else {
		var err error
		result, err = prompts.RunAddNewConnectionForm(ctx.Spec.Connections)
		if err != nil {
			return err
		}
	}

	// Initialize connections map if nil
	if ctx.Spec.Connections == nil {
		ctx.Spec.Connections = make(map[string]opendpi.Connection)
	}

	// Add the connection
	ctx.Spec.Connections[result.Name] = result.Connection

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

	fmt.Printf("Connection %q added.\n", result.Name)
	return nil
}
