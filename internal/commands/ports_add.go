// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dacolabs/cli/internal/jschema"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
)

type portsAddOptions struct {
	name           string
	description    string
	schemaFile     string
	connectionName string
	location       string
	nonInteractive bool
}

func newPortsAddCmd() *cobra.Command {
	opts := &portsAddOptions{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new port to the OpenDPI spec",
		Long: `Add a new port to the OpenDPI spec with schema and connection configuration.
Ports can have optional schemas (from file or created interactively) and
connections that define where data flows.`,
		Example: `  # Interactive mode
  daco ports add

  # Non-interactive with existing connection
  daco ports add --name events --schema-file ./schemas/event.json \
    --connection kafka --location events_topic --non-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runPortsAdd(ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Port name")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Port description")
	cmd.Flags().StringVarP(&opts.schemaFile, "schema-file", "s", "", "Path to JSON schema file")
	cmd.Flags().StringVarP(&opts.connectionName, "connection", "c", "", "Connection name")
	cmd.Flags().StringVarP(&opts.location, "location", "l", "", "Location (table, topic, path, etc.)")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run without prompts")

	return cmd
}

func runPortsAdd(ctx *session.Context, opts *portsAddOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Non-interactive validation
	if opts.nonInteractive {
		if opts.name == "" {
			return fmt.Errorf("--name is required in non-interactive mode")
		}
		if opts.connectionName != "" && opts.location == "" {
			return fmt.Errorf("--location is required when --connection is specified")
		}
		if opts.connectionName != "" {
			if _, exists := ctx.Spec.Connections[opts.connectionName]; !exists {
				return fmt.Errorf("connection %q not found", opts.connectionName)
			}
		}
		if _, exists := ctx.Spec.Ports[opts.name]; exists {
			return fmt.Errorf("port %q already exists", opts.name)
		}
	}

	var name, description string
	var schema *jsonschema.Schema
	var conns []opendpi.PortConnection

	// Load schema from file if specified via flag
	if opts.schemaFile != "" {
		loader := jschema.NewLoader(os.DirFS(cwd))
		schema, err = loader.LoadFile(opts.schemaFile)
		if err != nil {
			return fmt.Errorf("failed to load schema file: %w", err)
		}
	}

	if opts.nonInteractive {
		name = opts.name
		description = opts.description
		if opts.connectionName != "" {
			conn := ctx.Spec.Connections[opts.connectionName]
			conns = append(conns, opendpi.PortConnection{
				Connection: &conn,
				Location:   opts.location,
			})
		}
	} else {
		// Interactive mode
		result, err := prompts.RunAddPortForm(
			ctx.Spec.Ports,
			ctx.Spec.Connections,
			ctx.Spec.Schemas,
		)
		if err != nil {
			return err
		}

		name = result.Name
		description = result.Description
		schema = result.Schema
		conns = result.Connections

		// Add any newly created connections to the spec
		for connName, conn := range result.NewConns {
			ctx.Spec.Connections[connName] = conn
		}

		// Load schema from file if user chose file option
		if result.SchemaPath != "" && schema == nil {
			loader := jschema.NewLoader(os.DirFS(cwd))
			schema, err = loader.LoadFile(result.SchemaPath)
			if err != nil {
				return fmt.Errorf("failed to load schema file: %w", err)
			}
		}
	}

	// Create port
	ctx.Spec.Ports[name] = opendpi.Port{
		Description: description,
		Connections: conns,
		Schema:      schema,
	}

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

	return writer.Write(ctx.Spec, ctx.Config)
}
