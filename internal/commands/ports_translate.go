// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"

	// Import translator to auto-register
	_ "github.com/dacolabs/cli/internal/translate/pyspark"
)

type portsTranslateOptions struct {
	name       string
	format     string
	outputFile string
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
  daco ports translate --name my-port --format pyspark --output schema.py`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPortsTranslate(cmd, translators, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Port name (translates all if not specified)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "", fmt.Sprintf("Output format (%s)", strings.Join(translators.Available(), ", ")))
	cmd.Flags().StringVarP(&opts.outputFile, "output", "o", "", "Output file path (only valid when translating a single port)")

	return cmd
}

func runPortsTranslate(cmd *cobra.Command, translators translate.Register, opts *portsTranslateOptions) error {
	ctx, err := session.RequireFromCommand(cmd)
	if err != nil {
		return err
	}

	format := opts.format
	if format == "" {
		formats := translators.Available()
		if len(formats) == 0 {
			return fmt.Errorf("no translation formats available")
		}
		err := huh.NewForm(
			huh.NewGroup(prompts.RunTranslateFormatSelect(&format, formats)),
		).Run()
		if err != nil {
			return err
		}
	}

	if opts.name == "" {
		spec := ctx.Spec
		if opts.outputFile != "" {
			return fmt.Errorf("--output flag is only valid when translating a single port")
		}
		if len(ctx.Spec.Ports) == 0 {
			return fmt.Errorf("no ports found in spec")
		}

		fmt.Printf("Translating %d port(s) to %s...\n", len(spec.Ports), format)

		var errors []string
		successCount := 0

		for portName, port := range spec.Ports {
			if port.Schema == nil {
				fmt.Printf("  Skipping %s (no schema defined)\n", portName)
				continue
			}

			err := translatePort(portName, translators, port.Schema, format, "")
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", portName, err))
			} else {
				successCount++
			}
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

	port, ok := ctx.Spec.Ports[opts.name]
	if !ok {
		return fmt.Errorf("port %q not found in spec", opts.name)
	}
	if port.Schema == nil {
		return fmt.Errorf("port %q has no schema defined", opts.name)
	}

	return translatePort(opts.name, translators, port.Schema, format, opts.outputFile)
}

func translatePort(portName string, translators translate.Register, schema *jsonschema.Schema, format, outputFile string) error {
	translator, err := translators.Get(format)
	if err != nil {
		return fmt.Errorf("unsupported format %q. Available formats: %s",
			format, strings.Join(translators.Available(), ", "))
	}

	output, err := translator.Translate(portName, schema)
	if err != nil {
		return fmt.Errorf("failed to translate schema: %w", err)
	}

	if outputFile == "" {
		schemasDir := "schemas"
		if err := os.MkdirAll(schemasDir, 0o750); err != nil {
			return fmt.Errorf("failed to create schemas directory: %w", err)
		}
		outputFile = filepath.Join(schemasDir, portName+translator.FileExtension())
	}

	if err := os.WriteFile(outputFile, output, 0o600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	fmt.Printf("  %s\n", outputFile)

	return nil
}
