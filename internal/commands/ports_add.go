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
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
)

type portsAddOptions struct {
	name        string
	description string
	connection  string
	location    string
}

func newPortsAddCmd() *cobra.Command {
	opts := &portsAddOptions{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new port to the OpenDPI spec",
		Long: `Add a new port to the OpenDPI spec.
An empty schema file is created in the schemas/ directory for you to fill in.`,
		Example: `  # Interactive mode
  daco ports add

  # Non-interactive
  daco ports add --name events --connection kafka --location events_topic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runPortsAdd(cmd, ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Port name")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Port description")
	cmd.Flags().StringVarP(&opts.connection, "connection", "c", "", "Connection name")
	cmd.Flags().StringVarP(&opts.location, "location", "l", "", "Location (table, topic, path, etc.)")

	return cmd
}

func runPortsAdd(cmd *cobra.Command, ctx *session.Context, opts *portsAddOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var name, description, connName, location string

	if cmd.Flags().Changed("name") {
		// Non-interactive: use flags
		name = opts.name
		description = opts.description
		connName = opts.connection
		location = opts.location

		if name == "" {
			return fmt.Errorf("--name cannot be empty")
		}
		if _, exists := ctx.Spec.Ports[name]; exists {
			return fmt.Errorf("port %q already exists", name)
		}
		if connName != "" {
			if _, exists := ctx.Spec.Connections[connName]; !exists {
				return fmt.Errorf("connection %q not found", connName)
			}
			if location == "" {
				return fmt.Errorf("--location is required when --connection is specified")
			}
		}
	} else {
		// Interactive mode
		if err := prompts.RunPortsAddForm(
			&name, &description, &connName, &location,
			ctx.Spec.Ports, ctx.Spec.Connections,
		); err != nil {
			return err
		}
	}

	// Build port
	schemaRef := "schemas/" + name + ".schema.yaml"
	port := opendpi.Port{
		Description: description,
		SchemaRef:   schemaRef,
		Schema: &jsonschema.Schema{
			Ref: schemaRef,
		},
	}

	if connName != "" {
		conn := ctx.Spec.Connections[connName]
		port.Connections = []opendpi.PortConnection{
			{Connection: &conn, Location: location},
		}
	}

	ctx.Spec.Ports[name] = port

	// Determine spec directory
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

	if err := writer.Write(ctx.Spec, specDir); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	// Create empty schema file
	schemasDir := filepath.Join(specDir, "schemas")
	if err := opendpi.WriteEmptySchemaYAML(schemasDir, name); err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}

	fields := []prompts.ResultField{
		{Label: "Port", Value: name},
	}
	if description != "" {
		fields = append(fields, prompts.ResultField{Label: "Description", Value: description})
	}
	if connName != "" {
		fields = append(fields, prompts.ResultField{Label: "Connection", Value: connName + ":" + location})
	}
	prompts.PrintResult(fields, "✓ Port added, schema created in schemas/"+name+".schema.yaml")

	return nil
}
