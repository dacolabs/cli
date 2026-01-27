// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/session"
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type portsDescribeOptions struct {
	output string // output format: text, json, yaml
}

func newPortsDescribeCmd() *cobra.Command {
	opts := &portsDescribeOptions{}

	cmd := &cobra.Command{
		Use:   "describe [PORT_NAME]",
		Short: "Show detailed information about a port",
		Long:  `Display complete port definition including schema and connections. If no port name is provided, an interactive selection prompt is shown.`,
		Example: `  # Interactive selection
  daco ports describe

  # Show port details in human-readable format
  daco ports describe user_events

  # Show port details as JSON
  daco ports describe user_events -o json

  # Show port details as YAML
  daco ports describe user_events -o yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}

			var portName string
			if len(args) > 0 {
				portName = args[0]
			} else {
				portName, err = selectPortToDescribe(ctx.Spec.Ports)
				if err != nil {
					return err
				}
			}
			return runPortsDescribe(ctx, portName, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "text", "Output format (text, json, yaml)")

	return cmd
}

func selectPortToDescribe(ports map[string]opendpi.Port) (string, error) {
	if len(ports) == 0 {
		return "", fmt.Errorf("no ports defined")
	}

	names := make([]string, 0, len(ports))
	for name := range ports {
		names = append(names, name)
	}
	sort.Strings(names)

	options := make([]huh.Option[string], 0, len(ports))
	for _, name := range names {
		port := ports[name]
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

	var selected string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select port to describe").
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

func runPortsDescribe(ctx *session.Context, portName string, opts *portsDescribeOptions) error {
	port, exists := ctx.Spec.Ports[portName]
	if !exists {
		return fmt.Errorf("port %q not found", portName)
	}

	switch opts.output {
	case "json":
		output := map[string]any{"name": portName, "port": port}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)

	case "yaml":
		output := map[string]any{"name": portName, "port": port}
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		defer func() { _ = enc.Close() }()
		return enc.Encode(output)

	default:
		// Text format
		fmt.Printf("Name:        %s\n", portName)
		if port.Description != "" {
			fmt.Printf("Description: %s\n", port.Description)
		}
		fmt.Println()

		// Schema section
		if port.Schema != nil {
			fmt.Println("Schema:")
			printSchemaText(port.Schema, "  ")
		} else {
			fmt.Println("Schema:      (none)")
		}
		fmt.Println()

		// Connections section
		if len(port.Connections) > 0 {
			fmt.Println("Connections:")
			for _, pc := range port.Connections {
				connName := findConnectionName(pc.Connection, ctx.Spec.Connections)
				if pc.Connection != nil {
					fmt.Printf("  - %s (%s://%s) -> %s\n", connName, pc.Connection.Protocol, pc.Connection.Host, pc.Location)
				} else {
					fmt.Printf("  - %s -> %s\n", connName, pc.Location)
				}
			}
		} else {
			fmt.Println("Connections: (none)")
		}
		return nil
	}
}

func printSchemaText(schema *jsonschema.Schema, indent string) {
	if schema.Ref != "" {
		fmt.Printf("%sRef: %s\n", indent, schema.Ref)
		return
	}

	if schema.Type != "" {
		fmt.Printf("%sType: %s\n", indent, schema.Type)
	}

	if schema.Description != "" {
		fmt.Printf("%sDescription: %s\n", indent, schema.Description)
	}

	if schema.Format != "" {
		fmt.Printf("%sFormat: %s\n", indent, schema.Format)
	}

	if len(schema.Enum) > 0 {
		enumStrs := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			enumStrs[i] = fmt.Sprintf("%v", v)
		}
		fmt.Printf("%sEnum: [%s]\n", indent, strings.Join(enumStrs, ", "))
	}

	if schema.Minimum != nil {
		fmt.Printf("%sMinimum: %v\n", indent, *schema.Minimum)
	}

	if schema.Maximum != nil {
		fmt.Printf("%sMaximum: %v\n", indent, *schema.Maximum)
	}

	// Object properties
	if len(schema.Properties) > 0 {
		fmt.Printf("%sProperties:\n", indent)

		propNames := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)

		requiredSet := make(map[string]bool)
		for _, r := range schema.Required {
			requiredSet[r] = true
		}

		for _, propName := range propNames {
			propSchema := schema.Properties[propName]
			required := ""
			if requiredSet[propName] {
				required = " (required)"
			}

			propType := propSchema.Type
			if propSchema.Ref != "" {
				propType = propSchema.Ref
			}
			if propType == "" {
				propType = "any"
			}

			extra := ""
			if propSchema.Format != "" {
				extra = fmt.Sprintf(", format: %s", propSchema.Format)
			}
			if len(propSchema.Enum) > 0 {
				enumStrs := make([]string, len(propSchema.Enum))
				for i, v := range propSchema.Enum {
					enumStrs[i] = fmt.Sprintf("%v", v)
				}
				extra = fmt.Sprintf(", enum: [%s]", strings.Join(enumStrs, ", "))
			}

			fmt.Printf("%s  - %s (%s%s)%s\n", indent, propName, propType, extra, required)
		}
	}

	// Array items
	if schema.Items != nil {
		fmt.Printf("%sItems:\n", indent)
		printSchemaText(schema.Items, indent+"  ")
	}

	// $defs
	if len(schema.Defs) > 0 {
		fmt.Printf("%s$defs:\n", indent)
		defNames := make([]string, 0, len(schema.Defs))
		for name := range schema.Defs {
			defNames = append(defNames, name)
		}
		sort.Strings(defNames)

		for _, defName := range defNames {
			fmt.Printf("%s  %s:\n", indent, defName)
			printSchemaText(schema.Defs[defName], indent+"    ")
		}
	}
}
