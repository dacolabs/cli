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
	name        string
	connType    string
	host        string
	description string
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
  daco connections add -n kafka_prod -t kafka --host broker:9092`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runConnectionsAdd(cmd, ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Connection name")
	cmd.Flags().StringVarP(&opts.connType, "type", "t", "", "Connection type (kafka, postgresql, mysql, s3, http, etc.)")
	cmd.Flags().StringVar(&opts.host, "host", "", "Host/endpoint")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Description")

	return cmd
}

func runConnectionsAdd(cmd *cobra.Command, ctx *session.Context, opts *connectionsAddOptions) error {
	var name, connType, host, description string

	if cmd.Flags().Changed("name") {
		name = opts.name
		connType = opts.connType
		host = opts.host
		description = opts.description

		if name == "" {
			return fmt.Errorf("--name cannot be empty")
		}
		if connType == "" {
			return fmt.Errorf("--type is required when --name is specified")
		}
		if host == "" {
			return fmt.Errorf("--host is required when --name is specified")
		}
		if _, exists := ctx.Spec.Connections[name]; exists {
			return fmt.Errorf("connection %q already exists", name)
		}
	} else {
		if err := prompts.RunConnectionsAddForm(
			&name, &connType, &host, &description,
			ctx.Spec.Connections,
		); err != nil {
			return err
		}
	}

	if ctx.Spec.Connections == nil {
		ctx.Spec.Connections = make(map[string]opendpi.Connection)
	}
	ctx.Spec.Connections[name] = opendpi.Connection{
		Type:        connType,
		Host:        host,
		Description: description,
	}

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
		{Label: "Connection", Value: name},
		{Label: "Type", Value: connType},
		{Label: "Host", Value: host},
	}, "✓ Connection added")

	return nil
}
