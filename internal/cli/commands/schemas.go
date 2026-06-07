// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli/engine"
	"github.com/dacolabs/daco/internal/opendpi"
)

func Schemas() *cobra.Command {
	group := &cobra.Command{
		Use:   "schemas",
		Short: "Manage schemas",
	}
	group.AddCommand(schemasCreate())
	group.AddCommand(schemasList())
	group.AddCommand(schemasDescribe())
	group.AddCommand(schemasDelete())
	group.AddCommand(schemasTranslate())
	return group
}

func schemasTranslate() *cobra.Command {
	var in engine.SchemasTranslateInput
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate a schema to a target format",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.SchemasTranslate(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", out.OutputFile)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Schema, "name", "n", "", "Schema name")
	cmd.Flags().StringVarP(&in.Format, "format", "f", "", "Target format (e.g. pydantic, gotypes, protobuf)")
	cmd.Flags().StringVarP(&in.OutputDir, "output", "o", "", "Output directory")
	return cmd
}

func schemasCreate() *cobra.Command {
	var in engine.SchemasCreateInput
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.SchemasCreate(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if out.Scaffolded {
				fmt.Fprintf(w, "Scaffolded %s and registered schema %q\n", out.Path, out.Name)
			} else {
				fmt.Fprintf(w, "Registered schema %q -> %s\n", out.Name, out.Path)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Schema name")
	cmd.Flags().StringVarP(&in.Path, "path", "p", "", "Path to the schema yaml (relative to the project root)")
	cmd.Flags().StringVarP(&in.Type, "type", "t", "object", "JSON Schema type")
	cmd.Flags().StringVar(&in.Title, "title", "", "Schema title")
	cmd.Flags().StringVarP(&in.Description, "description", "d", "", "Schema description")
	return cmd
}

func schemasList() *cobra.Command {
	var in engine.SchemasListInput
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schemas",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.SchemasList(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			for _, s := range out.Schemas {
				fmt.Fprintf(w, "%s\t%s\n", s.Name, s.Path)
			}
			return nil
		},
	}
	return cmd
}

func schemasDescribe() *cobra.Command {
	var in engine.SchemasDescribeInput
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.SchemasDescribe(cmd.Context(), in)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Name:        %s\n", out.Name)
			fmt.Fprintf(w, "Path:        %s\n", out.Path)
			fmt.Fprintf(w, "Exists:      %t\n", out.Exists)
			if out.Schema != nil {
				if t := opendpi.SchemaTypeString(out.Schema); t != "" {
					fmt.Fprintf(w, "Type:        %s\n", t)
				}
				if out.Schema.Title != "" {
					fmt.Fprintf(w, "Title:       %s\n", out.Schema.Title)
				}
				if out.Schema.Description != "" {
					fmt.Fprintf(w, "Description: %s\n", out.Schema.Description)
				}
				fmt.Fprintf(w, "Properties:  %d\n", len(out.Schema.Properties))
				fmt.Fprintf(w, "Required:    %d\n", len(out.Schema.Required))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Schema name")
	return cmd
}

func schemasDelete() *cobra.Command {
	var in engine.SchemasDeleteInput
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, err := loadProject(cmd)
			if err != nil {
				return err
			}
			in.Prj = prj

			out, err := engine.SchemasDelete(cmd.Context(), in)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed schema %q\n", out.Name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Schema name")
	return cmd
}