// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli/engine"
)

func Connections() *cobra.Command {
	group := &cobra.Command{
		Use:   "connections",
		Short: "Manage connections",
	}
	group.AddCommand(connectionsCreate())
	group.AddCommand(connectionsList())
	group.AddCommand(connectionsDescribe())
	group.AddCommand(connectionsDelete())
	return group
}

func connectionsCreate() *cobra.Command {
	var in engine.ConnectionsCreateInput
	var vars []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a connection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			parsed, err := parseVars(vars)
			if err != nil {
				return err
			}
			in.Variables = parsed

			out, err := engine.ConnectionsCreate(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if out.Scaffolded {
				fmt.Fprintf(w, "Scaffolded %s and registered connection %q\n", out.Path, out.Name)
			} else {
				fmt.Fprintf(w, "Registered connection %q -> %s\n", out.Name, out.Path)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Connection name")
	cmd.Flags().StringVarP(&in.Path, "path", "p", "", "Path to the connection yaml (relative to the project root)")
	cmd.Flags().StringVarP(&in.Type, "type", "t", "", "Connection type (e.g. postgresql, kafka, s3)")
	cmd.Flags().StringVar(&in.Host, "host", "", "Host or endpoint")
	cmd.Flags().StringVarP(&in.Description, "description", "d", "", "Description")
	cmd.Flags().StringArrayVar(&vars, "var", nil, "Variable as key=value (repeatable)")
	return cmd
}

func connectionsList() *cobra.Command {
	var in engine.ConnectionsListInput
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ConnectionsList(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			for _, c := range out.Connections {
				fmt.Fprintf(w, "%s\t%s\n", c.Name, c.Path)
			}
			return nil
		},
	}
	return cmd
}

func connectionsDescribe() *cobra.Command {
	var in engine.ConnectionsDescribeInput
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a connection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ConnectionsDescribe(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Name:        %s\n", out.Name)
			fmt.Fprintf(w, "Path:        %s\n", out.Path)
			fmt.Fprintf(w, "Exists:      %t\n", out.Exists)
			if out.Conn != nil {
				fmt.Fprintf(w, "Type:        %s\n", out.Conn.Type)
				fmt.Fprintf(w, "Host:        %s\n", out.Conn.Host)
				if out.Conn.Description != "" {
					fmt.Fprintf(w, "Description: %s\n", out.Conn.Description)
				}
				fmt.Fprintf(w, "Variables:   %d\n", len(out.Conn.Variables))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Connection name")
	return cmd
}

func connectionsDelete() *cobra.Command {
	var in engine.ConnectionsDeleteInput
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a connection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ConnectionsDelete(cmd.Context(), in)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed connection %q\n", out.Name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Connection name")
	return cmd
}

func parseVars(vars []string) (map[string]any, error) {
	if len(vars) == 0 {
		return nil, nil
	}
	out := make(map[string]any, len(vars))
	for _, v := range vars {
		k, val, ok := strings.Cut(v, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("invalid --var %q (expected key=value)", v)
		}
		out[k] = val
	}
	return out, nil
}