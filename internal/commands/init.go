// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dacolabs/cli/internal/config"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/spf13/cobra"
)

type initOptions struct {
	extends        string
	createSpec     bool
	name           string
	path           string
	version        string
	schemaOrg      string
	format         string
	nonInteractive bool
}

func newInitCmd() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new daco project",
		Long: `Initialize a new daco project with a daco.yaml configuration file.
Can create a new spec, use an existing one, or extend a parent config.`,
		Example: `  # Interactive mode
  daco init

  # Non-interactive
  daco init --name "my-product" --non-interactive
  daco init --extends ../daco.yaml --non-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&opts.path, "path", "p", "./spec", "Path to spec folder")
	cmd.Flags().StringVarP(&opts.extends, "extends", "e", "", "Path to parent daco.yaml")
	cmd.Flags().StringVarP(&opts.version, "version", "v", "1.0.0", "Initial spec version")
	cmd.Flags().StringVarP(&opts.schemaOrg, "schema-organization", "s", "modular", "Schema organization (modular, components, or inline)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "yaml", "Spec format (yaml or json)")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run without prompts (requires --name or --extends)")

	return cmd
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
		if opts.name == "" && opts.extends == "" {
			return errors.New("non-interactive mode requires either --name or --extends")
		}
	} else {
		opts.createSpec = true
		if err := prompts.RunInitForm(
			&opts.extends,
			&opts.name,
			&opts.path,
			&opts.version,
			&opts.schemaOrg,
			&opts.format,
			&opts.createSpec,
		); err != nil {
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
	specFolder := opts.path
	if !filepath.IsAbs(specFolder) {
		specFolder = filepath.Join(cwd, specFolder)
	}

	if !opts.createSpec {
		_, err1 := os.Stat(filepath.Join(specFolder, "opendpi.yaml"))
		_, err2 := os.Stat(filepath.Join(specFolder, "opendpi.json"))
		if err1 != nil && err2 != nil {
			return fmt.Errorf("spec file not found in: %s", opts.path)
		}
	} else {
		specFileName := "opendpi.yaml"
		if opts.format == "json" {
			specFileName = "opendpi.json"
		}
		specPath := filepath.Join(specFolder, specFileName)

		if _, err := os.Stat(specPath); err == nil {
			return fmt.Errorf("spec file already exists: %s", opts.path)
		}

		if err := os.MkdirAll(specFolder, 0o750); err != nil {
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

		var writer opendpi.Writer
		if opts.format == "json" {
			writer = opendpi.JSONWriter
		} else {
			writer = opendpi.YAMLWriter
		}

		if err := writer.Write(spec, &cfg); err != nil {
			return fmt.Errorf("failed to write spec file: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if err := cfg.Save(filepath.Join(cwd, "daco.yaml")); err != nil {
		return fmt.Errorf("config file couldn't be saved: %w", err)
	}
	fmt.Printf("Initialization completed\n")

	return nil
}
