// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/config"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/spf13/cobra"
)

type initOptions struct {
	extends        string // path to parent config
	createSpec     bool   // create spec
	name           string // product name
	path           string // path to spec file
	version        string // spec version
	schemaOrg      string // schema organization
	nonInteractive bool
}

func registerInitCmd(parent *cobra.Command) {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new daco project",
		Long: `Initialize a new daco project with a daco.yaml configuration file.
Can create a new spec, use an existing one, or extend a parent config.`,
		Example: `  daco init
  daco init --name "my-product" --non-interactive
  daco init --extends ../daco.yaml --non-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&opts.path, "path", "p", "spec", "Path to spec folder")
	cmd.Flags().StringVarP(&opts.extends, "extends", "e", "", "Path to parent daco.yaml")
	cmd.Flags().StringVarP(&opts.version, "version", "v", "1.0.0", "Initial spec version")
	cmd.Flags().StringVarP(&opts.schemaOrg, "schema-organization", "s", "modular", "Schema organization (modular, components, or inline)")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run without prompts (requires --name or--extends)")

	parent.AddCommand(cmd)
}

func runInit(opts *initOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check that the current directory isn't already initialized
	dacoPath := filepath.Join(cwd, "daco.yaml")
	if _, err := os.Stat(dacoPath); err == nil {
		return errors.New("daco.yaml already exists; project already initialized")
	}

	if opts.nonInteractive {
		if opts.name == "" && opts.extends != "" {
			return errors.New("non-interactive mode requires either --name or --extends")
		}
	} else {
		extends := false
		if err := huh.NewForm(
			huh.NewGroup(
				prompts.ConfigTypeSelect(&extends),
			).WithHide(opts.path != ""),
			huh.NewGroup(
				prompts.ParentConfigPathInput(&opts.path),
			).WithHideFunc(func() bool { return !extends }),
			huh.NewGroup(
				prompts.NewSpecSourceSelect(&opts.createSpec),
			).WithHideFunc(func() bool { return extends }),
			huh.NewGroup(
				prompts.SpecPathInput(&opts.path, &opts.createSpec),
			).WithHideFunc(func() bool { return extends }),
			huh.NewGroup(
				prompts.ProductNameInput(&opts.name),
			).WithHideFunc(func() bool { return extends || opts.name != "" }),
			huh.NewGroup(
				prompts.SpecVersionInput(&opts.version),
			).WithHideFunc(func() bool { return extends }),
			huh.NewGroup(
				prompts.SchemaOrganizationSelect(&opts.schemaOrg),
			).WithHideFunc(func() bool { return extends }),
		).Run(); err != nil {
			return err
		}
	}

	// handle child config
	if parentPath := opts.extends; parentPath != "" {
		if !filepath.IsAbs(parentPath) {
			parentPath = filepath.Join(cwd, parentPath)
		}
		if _, err := os.Stat(parentPath); os.IsNotExist(err) {
			return fmt.Errorf("parent config not found: %s", opts.extends)
		}
		cfg := config.Config{
			Version: config.CurrentConfigVersion,
			Extends: opts.extends,
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}
		if err := cfg.Save(filepath.Join(cwd, "daco.yaml")); err != nil {
			return fmt.Errorf("failed to write daco.yaml: %w", err)
		}
		fmt.Print("Initialization completed")
		return nil
	}
	cfg := config.Config{
		Version: config.CurrentConfigVersion,
		Path:    opts.path,
		Schema: config.SchemaConfig{
			Organization: config.SchemaOrganization(opts.schemaOrg),
		},
	}
	specPath := filepath.Join(opts.path, "opendpi.yaml")
	if !filepath.IsAbs(specPath) {
		specPath = filepath.Join(cwd, specPath)
	}
	if !opts.createSpec {
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			return fmt.Errorf("spec file not found in: %s", opts.path)
		}
	} else {
		if _, err := os.Stat(specPath); err == nil {
			return fmt.Errorf("spec file already exists: %s", opts.path)
		}

		if err := os.MkdirAll(filepath.Dir(opts.path), 0o750); err != nil {
			return fmt.Errorf("failed to create spec directory: %w", err)
		}

		spec := &opendpi.Spec{
			OpenDPI: "1.0.0",
			Info: opendpi.Info{
				Title:   opts.name,
				Version: opts.version,
			},
			Connections: map[string]opendpi.Connection{},
			Ports:       map[string]opendpi.Port{},
		}

		if err := opendpi.YAMLWriter.Write(spec, &cfg); err != nil {
			return fmt.Errorf("failed to write spec file: %w", err)
		}
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	fmt.Print("Initialization completed")

	return nil
}
