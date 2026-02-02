// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

func newPortsDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [PORT_NAME]",
		Short: "Show detailed information about a port",
		Long:  `Display complete port definition including schema and connections. If no port name is provided, an interactive selection prompt is shown.`,
		Example: `  # Interactive selection
  daco ports describe

  # Show port details
  daco ports describe user_events`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runPortsDescribe(ctx, args)
		},
	}

	return cmd
}

func runPortsDescribe(ctx *session.Context, args []string) error {
	var portName string
	if len(args) > 0 {
		portName = args[0]
	} else {
		if len(ctx.Spec.Ports) == 0 {
			return fmt.Errorf("no ports defined")
		}
		if err := prompts.RunPortsDescribeForm(&portName, ctx.Spec.Ports); err != nil {
			return err
		}
	}

	port, exists := ctx.Spec.Ports[portName]
	if !exists {
		return fmt.Errorf("port %q not found", portName)
	}

	fields := []prompts.ResultField{
		{Label: "Name", Value: portName},
	}
	if port.Description != "" {
		fields = append(fields, prompts.ResultField{Label: "Description", Value: port.Description})
	}
	if len(port.Connections) > 0 {
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "\nNAME\tHOST\tLOCATION")
		for _, pc := range port.Connections {
			connName := findConnectionName(pc.Connection, ctx.Spec.Connections)
			host := "-"
			if pc.Connection != nil {
				host = fmt.Sprintf("%s://%s", pc.Connection.Type, pc.Connection.Host)
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", connName, host, pc.Location)
		}
		_ = w.Flush()
		fields = append(fields, prompts.ResultField{Label: "Connections", Value: strings.TrimRight(buf.String(), "\n")})
	}

	prompts.PrintResult(fields, "")
	return nil
}
