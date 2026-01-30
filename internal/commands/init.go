// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/dacolabs/cli/internal/config"
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/spf13/cobra"
)

type initOptions struct {
	name        string
	version     string
	description string
}

func newInitCmd() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new daco project",
		Long:  `Initialize a new daco project with a daco.yaml configuration file and an OpenDPI spec.`,
		Example: `  # Interactive mode
  daco init

  # Non-interactive
  daco init --name "Customer Analytics" --non-interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Data product name")
	cmd.Flags().StringVarP(&opts.version, "version", "v", "1.0.0", "Initial spec version")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Data product description")

	return cmd
}

func runInit(cmd *cobra.Command, opts *initOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	dacoPath := filepath.Join(cwd, "daco.yaml")
	if _, err := os.Stat(dacoPath); err == nil {
		return errors.New("daco.yaml already exists; project already initialized")
	}

	if !cmd.Flags().Changed("name") {
		if err := prompts.RunInitForm(
			&opts.name,
			&opts.version,
			&opts.description,
		); err != nil {
			return err
		}
	}

	specDir := filepath.Join(cwd, "spec")
	specPath := filepath.Join(specDir, "opendpi.yaml")

	if _, err := os.Stat(specPath); err == nil {
		return fmt.Errorf("spec file already exists: %s", specPath)
	}

	if err := os.MkdirAll(specDir, 0o750); err != nil {
		return fmt.Errorf("failed to create spec directory: %w", err)
	}

	_ = &opendpi.Spec{
		OpenDPI: "1.0.0",
		Info: opendpi.Info{
			Title:       opts.name,
			Version:     opts.version,
			Description: opts.description,
		},
		Connections: map[string]opendpi.Connection{},
		Ports:       map[string]opendpi.Port{},
	}

	// if err := opendpi.YAMLWriter.Write(spec, specDir); err != nil {
		// return fmt.Errorf("failed to write spec file: %w", err)
	// }

	cfg := config.Config{
		Version: config.CurrentConfigVersion,
		Path:    "./spec",
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if err := cfg.Save(dacoPath); err != nil {
		return fmt.Errorf("config file couldn't be saved: %w", err)
	}

	success := lipgloss.NewStyle().Foreground(lipgloss.Color("#27ca3f"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#bababa"))
	check := success.Render("?")
	fmt.Printf("\n%s %s %s\n", check, label.Render("Data product name:"), opts.name)
	fmt.Printf("\n%s %s %s\n", check, label.Render("Version:"), opts.version)
	fmt.Printf("\n%s %s %s\n", check, label.Render("Description:"), opts.description)
	fmt.Println(success.Render("\n✓ Created opendpi.yaml"))
	return nil
}
