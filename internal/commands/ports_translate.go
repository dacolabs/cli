// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/spf13/cobra"

	// Import translator to auto-register
	_ "github.com/dacolabs/cli/internal/translate/pyspark"
)

type portsTranslateOptions struct {
	name   string
	format string
	output string
	all    bool
}

func newPortsTranslateCmd(translators translate.Register) *cobra.Command {
	opts := &portsTranslateOptions{}

	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate a port schema to a target format",
		Long: fmt.Sprintf(`Translate a port schema to a target format.

Available formats: %s`, strings.Join(translators.Available(), ", ")),
		Example: `  # Interactive mode
  daco ports translate

  # Translate specific port
  daco ports translate --name my-port --format pyspark

  # Translate multiple ports
  daco ports translate --name port-a,port-b --format pyspark

  # Translate all ports
  daco ports translate --all --format pyspark

  # Translate to a custom output directory (also sets package name for Go/Protobuf/Scala)
  daco ports translate --all --format gotypes --output models`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPortsTranslate(cmd, translators, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Port name(s), comma-separated")
	cmd.Flags().StringVar(&opts.format, "format", "", fmt.Sprintf("Output format (%s)", strings.Join(translators.Available(), ", ")))
	cmd.Flags().StringVarP(&opts.output, "output", "o", "schemas", "Output directory (also used as package name for Go/Protobuf/Scala)")
	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Translate all ports")

	return cmd
}

func runPortsTranslate(cmd *cobra.Command, translators translate.Register, opts *portsTranslateOptions) error {
	ctx, err := session.RequireFromCommand(cmd)
	if err != nil {
		return err
	}

	if len(ctx.Spec.Ports) == 0 {
		return fmt.Errorf("no ports defined")
	}

	// Validate mutually exclusive flags
	if opts.all && opts.name != "" {
		return fmt.Errorf("--all and --name are mutually exclusive")
	}

	// Resolve selected ports from flags
	var selectedPorts []string

	if opts.all {
		for name := range ctx.Spec.Ports {
			selectedPorts = append(selectedPorts, name)
		}
	} else if opts.name != "" {
		for _, n := range strings.Split(opts.name, ",") {
			n = strings.TrimSpace(n)
			if n == "" {
				continue
			}
			if _, ok := ctx.Spec.Ports[n]; !ok {
				return fmt.Errorf("port %q not found in spec", n)
			}
			selectedPorts = append(selectedPorts, n)
		}
	}

	format := opts.format
	output := opts.output

	// Prompt for any missing values
	err = prompts.RunTranslateForm(
		&selectedPorts, &format, &output,
		!cmd.Flags().Changed("output"),
		ctx.Spec.Ports, translators.Available(),
	)
	if err != nil {
		return err
	}

	if len(selectedPorts) == 0 {
		return fmt.Errorf("no ports selected")
	}

	// Execute translations
	fmt.Printf("Translating %d port(s) to %s...\n", len(selectedPorts), format)

	var errors []string
	successCount := 0

	translator, err := translators.Get(format)
	if err != nil {
		return fmt.Errorf("unsupported format %q. Available formats: %s",
			format, strings.Join(translators.Available(), ", "))
	}

	if err := os.MkdirAll(output, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, portName := range selectedPorts {
		port := ctx.Spec.Ports[portName]
		if port.Schema == nil {
			fmt.Printf("  Skipping %s (no schema defined)\n", portName)
			continue
		}

		data, err := translator.Translate(portName, port.Schema, output)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", portName, err))
			continue
		}

		outFile := filepath.Join(output, portName+translator.FileExtension())

		if err := os.WriteFile(outFile, data, 0o600); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", portName, err))
			continue
		}
		fmt.Printf("  %s\n", outFile)
		successCount++
	}

	fmt.Printf("\nSuccessfully translated %d port(s)\n", successCount)

	if len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("failed to translate %d port(s)", len(errors))
	}

	return nil
}
