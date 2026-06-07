// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli"
	"github.com/dacolabs/daco/internal/cli/engine"
	"github.com/dacolabs/daco/internal/cli/settings"
)

func Products() *cobra.Command {
	group := &cobra.Command{
		Use:   "products",
		Short: "Manage products",
	}
	group.AddCommand(productsCreate())
	group.AddCommand(productsList())
	group.AddCommand(productsDescribe())
	group.AddCommand(productsDelete())
	group.AddCommand(productsLink())
	group.AddCommand(productsUnlink())
	return group
}

func productsUnlink() *cobra.Command {
	var in engine.ProductsUnlinkInput
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Unlink a connection from a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ProductsUnlink(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Unlinked connection %q from product %q\n", out.ConnectionName, out.ProductName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.ProductName, "product", "p", "", "Product name")
	cmd.Flags().StringVarP(&in.ConnectionName, "connection", "c", "", "Connection name")
	return cmd
}

func loadProject(cmd *cobra.Command) (*settings.Project, error) {
	usr := cli.User(cmd.Context())
	if usr == nil {
		return nil, fmt.Errorf("user settings not loaded")
	}
	return settings.LoadProject(usr)
}

func productsCreate() *cobra.Command {
	var in engine.ProductsCreateInput
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ProductsCreate(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if out.Scaffolded {
				fmt.Fprintf(w, "Scaffolded %s and registered product %q\n", out.Path, out.Name)
			} else {
				fmt.Fprintf(w, "Registered product %q -> %s\n", out.Name, out.Path)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Product name")
	cmd.Flags().StringVarP(&in.Path, "path", "p", "", "Path to the OpenDPI yaml (relative to the project root)")
	return cmd
}

func productsList() *cobra.Command {
	var in engine.ProductsListInput
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List products",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ProductsList(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			for _, p := range out.Products {
				fmt.Fprintf(w, "%s\t%s\n", p.Name, p.Path)
			}
			return nil
		},
	}
	return cmd
}

func productsDescribe() *cobra.Command {
	var in engine.ProductsDescribeInput
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ProductsDescribe(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Name:        %s\n", out.Name)
			fmt.Fprintf(w, "Path:        %s\n", out.Path)
			fmt.Fprintf(w, "Exists:      %t\n", out.Exists)
			if out.Info != nil {
				fmt.Fprintf(w, "Title:       %s\n", out.Info.Title)
				fmt.Fprintf(w, "Version:     %s\n", out.Info.Version)
			}
			fmt.Fprintf(w, "Ports:       %d\n", out.PortsCount)
			fmt.Fprintf(w, "Connections: %d\n", out.ConnectionsCount)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Product name")
	return cmd
}

func productsDelete() *cobra.Command {
	var in engine.ProductsDeleteInput
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a product",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ProductsDelete(cmd.Context(), in)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed product %q\n", out.Name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Product name")
	return cmd
}

func productsLink() *cobra.Command {
	var in engine.ProductsLinkInput
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link a product to an existing connection (writes a $ref into the product yaml)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.ProductsLink(cmd.Context(), in)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(),
				"Linked connection %q (%s @ %s) to product %q via $ref %s\n",
				out.ConnectionName, out.ConnectionType, out.ConnectionHost, out.ProductName, out.Ref)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.ProductName, "product", "p", "", "Product name")
	cmd.Flags().StringVarP(&in.ConnectionName, "connection", "c", "", "Connection name")
	return cmd
}