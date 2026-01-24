// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/context"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/spf13/cobra"

	// Import translator to auto-register
	_ "github.com/dacolabs/cli/internal/translate/pyspark"
)

func registerPortTranslateCmd(parent *cobra.Command) {
	var portName string
	var format string
	var outputFile string

	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate a port schema to a target format",
		Long: fmt.Sprintf(`Translate a port schema to a target format.

Available formats: %s`, strings.Join(translate.Available(), ", ")),
		Example: `  daco ports translate
		daco ports translate --port-name "my-port" --format pyspark --output schema.py`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPortTranslate(cmd, portName, format, outputFile)
		},
	}

	cmd.Flags().StringVar(&portName, "port-name", "", "Name of the port to translate (translates all ports if not specified)")
	cmd.Flags().StringVar(&format, "format", "", fmt.Sprintf("Output format (options: %s)", strings.Join(translate.Available(), ", ")))
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (only valid when translating a single port)")

	parent.AddCommand(cmd)
}

func runPortTranslate(cmd *cobra.Command, portName, format, outputFile string) error {
	ctx, err := context.RequireFromCommand(cmd)
	if err != nil {
		return err
	}

	if format == "" {
		formats := translate.Available()
		if len(formats) == 0 {
			return fmt.Errorf("no translation formats available")
		}
		err := huh.NewForm(
			huh.NewGroup(prompts.TranslateFormatSelect(&format, formats)),
		).Run()
		if err != nil {
			return err
		}
	}

	if portName == "" {
		spec := ctx.Spec
		if outputFile != "" {
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

			err := translatePort(portName, port.Schema, format, "")
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

	port, ok := ctx.Spec.Ports[portName]
	if !ok {
		return fmt.Errorf("port %q not found in spec", portName)
	}
	if port.Schema == nil {
		return fmt.Errorf("port %q has no schema defined", portName)
	}

	return translatePort(portName, port.Schema, format, outputFile)
}

func translatePort(portName string, schema *jschema.Schema, format, outputFile string) error {
	translator, err := translate.Get(format)
	if err != nil {
		return fmt.Errorf("unsupported format %q. Available formats: %s",
			format, strings.Join(translate.Available(), ", "))
	}

	rawJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to serialize schema: %w", err)
	}

	output, err := translator.Translate(portName, schema, rawJSON)
	if err != nil {
		return fmt.Errorf("failed to translate schema: %w", err)
	}

	if outputFile == "" {
		schemasDir := "schemas"
		if err := os.MkdirAll(schemasDir, 0755); err != nil {
			return fmt.Errorf("failed to create schemas directory: %w", err)
		}
		outputFile = filepath.Join(schemasDir, portName+translator.FileExtension())
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	fmt.Printf("  ✓ %s\n", outputFile)

	return nil
}
