// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli/engine"
)

func Ports() *cobra.Command {
	group := &cobra.Command{
		Use:   "ports",
		Short: "Manage product ports",
	}
	group.AddCommand(portsCreate())
	group.AddCommand(portsList())
	group.AddCommand(portsDescribe())
	group.AddCommand(portsDelete())
	group.AddCommand(portsLink())
	group.AddCommand(portsUnlink())
	group.AddCommand(portsBind())
	group.AddCommand(portsUnbind())
	group.AddCommand(portsTranslate())
	return group
}

func portsUnlink() *cobra.Command {
	var in engine.PortsUnlinkSchemaInput
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Unlink the schema from a port",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj
			out, err := engine.PortsUnlinkSchema(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Unlinked schema from port %q in product %q\n", out.Port, out.Product)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVar(&in.Port, "port", "", "Port name")
	return cmd
}

func portsBind() *cobra.Command {
	var in engine.PortsBindConnectionInput
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind a port to a connection at a specific location",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj
			out, err := engine.PortsBindConnection(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Bound port %q in product %q to connection %q at location %q\n",
				out.Port, out.Product, out.Connection, out.Location)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVar(&in.Port, "port", "", "Port name")
	cmd.Flags().StringVarP(&in.Connection, "connection", "c", "", "Connection name (must already be linked to the product)")
	cmd.Flags().StringVarP(&in.Location, "location", "l", "", "Location string (e.g. schema.table, topic name, bucket path)")
	return cmd
}

func portsUnbind() *cobra.Command {
	var in engine.PortsUnbindConnectionInput
	cmd := &cobra.Command{
		Use:   "unbind",
		Short: "Remove a port-to-connection binding",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj
			out, err := engine.PortsUnbindConnection(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Removed %d binding(s) for connection %q from port %q in product %q\n",
				out.Removed, out.Connection, out.Port, out.Product)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVar(&in.Port, "port", "", "Port name")
	cmd.Flags().StringVarP(&in.Connection, "connection", "c", "", "Connection name")
	cmd.Flags().StringVarP(&in.Location, "location", "l", "", "Location to remove (optional; empty removes all bindings of this connection)")
	return cmd
}

func portsTranslate() *cobra.Command {
	var in engine.PortsTranslateInput
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate a port's schema to a target format",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.PortsTranslate(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", out.OutputFile)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVar(&in.Port, "port", "", "Port name")
	cmd.Flags().StringVarP(&in.Format, "format", "f", "", "Target format (e.g. pydantic, gotypes, protobuf)")
	cmd.Flags().StringVarP(&in.OutputDir, "output", "o", "", "Output directory")
	return cmd
}

func portsCreate() *cobra.Command {
	var in engine.PortsCreateInput
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a port in a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.PortsCreate(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created port %q in product %q\n", out.Name, out.Product)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Port name")
	cmd.Flags().StringVarP(&in.Description, "description", "d", "", "Port description")
	return cmd
}

func portsList() *cobra.Command {
	var in engine.PortsListInput
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ports in a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.PortsList(cmd.Context(), in)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			for _, p := range out.Ports {
				if p.Description != "" {
					fmt.Fprintf(w, "%s\t%s\n", p.Name, p.Description)
				} else {
					fmt.Fprintln(w, p.Name)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	return cmd
}

func portsDescribe() *cobra.Command {
	var in engine.PortsDescribeInput
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a port",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.PortsDescribe(cmd.Context(), in)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Product:     %s\n", out.Product)
			fmt.Fprintf(w, "Name:        %s\n", out.Name)
			if out.Description != "" {
				fmt.Fprintf(w, "Description: %s\n", out.Description)
			}
			fmt.Fprintf(w, "Connections: %d\n", out.ConnectionsCount)
			for _, b := range out.Connections {
				fmt.Fprintf(w, "  - %s @ %s\n", b.Connection, b.Location)
			}
			fmt.Fprintf(w, "Has schema:  %t\n", out.HasSchema)
			if out.HasSchema {
				if out.SchemaRef != "" {
					fmt.Fprintf(w, "Schema $ref: %s\n", out.SchemaRef)
				}
				if out.SchemaType != "" {
					fmt.Fprintf(w, "Schema type: %s\n", out.SchemaType)
				}
				if out.SchemaTitle != "" {
					fmt.Fprintf(w, "Schema title: %s\n", out.SchemaTitle)
				}
				fmt.Fprintf(w, "Properties:  %d\n", out.SchemaProperties)
				fmt.Fprintf(w, "Required:    %d\n", out.SchemaRequired)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Port name")
	return cmd
}

func portsDelete() *cobra.Command {
	var in engine.PortsDeleteInput
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a port from a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.PortsDelete(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed port %q from product %q\n", out.Name, out.Product)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Port name")
	return cmd
}

func portsLink() *cobra.Command {
	var in engine.PortsLinkSchemaInput
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link a project-level schema to a port (writes a $ref into the product yaml)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.PortsLinkSchema(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Linked schema %q (%s) to port %q in product %q via $ref %s\n",
				out.Schema, out.SchemaType, out.Port, out.Product, out.Ref)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Product, "product", "p", "", "Product name")
	cmd.Flags().StringVar(&in.Port, "port", "", "Port name")
	cmd.Flags().StringVarP(&in.Schema, "schema", "s", "", "Schema name (project-level)")
	return cmd
}