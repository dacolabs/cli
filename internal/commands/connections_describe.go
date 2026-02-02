// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

func newConnectionsDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [CONNECTION_NAME]",
		Short: "Show detailed information about a connection",
		Long:  `Display complete connection details including which ports use it. If no connection name is provided, an interactive selection prompt is shown.`,
		Example: `  # Interactive selection
  daco connections describe

  # Show connection details
  daco connections describe kafka_prod`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runConnectionsDescribe(ctx, args)
		},
	}

	return cmd
}

func runConnectionsDescribe(ctx *session.Context, args []string) error {
	var connName string
	if len(args) > 0 {
		connName = args[0]
	} else {
		if len(ctx.Spec.Connections) == 0 {
			return fmt.Errorf("no connections defined")
		}
		if err := prompts.RunConnectionsDescribeForm(&connName, ctx.Spec.Connections); err != nil {
			return err
		}
	}

	conn, exists := ctx.Spec.Connections[connName]
	if !exists {
		return fmt.Errorf("connection %q not found", connName)
	}

	fields := []prompts.ResultField{
		{Label: "Name", Value: connName},
		{Label: "Type", Value: conn.Type},
		{Label: "Host", Value: conn.Host},
	}
	if conn.Description != "" {
		fields = append(fields, prompts.ResultField{Label: "Description", Value: conn.Description})
	}

	// Variables section
	if len(conn.Variables) > 0 {
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		varNames := make([]string, 0, len(conn.Variables))
		for k := range conn.Variables {
			varNames = append(varNames, k)
		}
		sort.Strings(varNames)
		for _, k := range varNames {
			_, _ = fmt.Fprintf(w, "%s\t%v\n", k, conn.Variables[k])
		}
		_ = w.Flush()
		fields = append(fields, prompts.ResultField{Label: "Variables", Value: strings.TrimRight(buf.String(), "\n")})
	}

	// Ports that use this connection
	var portRows []string
	portNames := make([]string, 0, len(ctx.Spec.Ports))
	for name := range ctx.Spec.Ports {
		portNames = append(portNames, name)
	}
	sort.Strings(portNames)
	for _, portName := range portNames {
		port := ctx.Spec.Ports[portName]
		for _, pc := range port.Connections {
			if pc.Connection != nil && pc.Connection.Type == conn.Type && pc.Connection.Host == conn.Host {
				desc := port.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				if desc == "" {
					desc = "-"
				}
				portRows = append(portRows, fmt.Sprintf("%s\t%s\t%s", portName, desc, pc.Location))
			}
		}
	}
	if len(portRows) > 0 {
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "\nNAME\tDESCRIPTION\tLOCATION")
		for _, row := range portRows {
			_, _ = fmt.Fprintln(w, row)
		}
		_ = w.Flush()
		fields = append(fields, prompts.ResultField{Label: "Ports", Value: strings.TrimRight(buf.String(), "\n")})
	}

	prompts.PrintResult(fields, "")
	return nil
}
