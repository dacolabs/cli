// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package prompts

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/dacolabs/cli/internal/opendpi"
)

// RunTranslateForm runs the interactive form for translating ports.
// It only prompts for fields that are not already set:
//   - selectedPorts: if empty, prompts with a multi-select of available ports
//   - format: if empty, prompts with a select of available formats
//   - outputDir: if promptOutputDir is true, prompts with a text input
func RunTranslateForm(
	selectedPorts *[]string,
	format *string,
	outputDir *string,
	promptOutputDir bool,
	ports map[string]opendpi.Port,
	formats []string,
) error {
	var groups []*huh.Group

	if len(*selectedPorts) == 0 {
		names := make([]string, 0, len(ports))
		for name := range ports {
			names = append(names, name)
		}
		sort.Strings(names)

		options := make([]huh.Option[string], 0, len(ports))
		for _, name := range names {
			port := ports[name]
			label := name
			if port.Description != "" {
				desc := port.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				label = fmt.Sprintf("%s - %s", name, desc)
			}
			options = append(options, huh.NewOption(label, name))
		}

		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select ports to translate").
				Options(options...).
				Value(selectedPorts).
				Height(10),
		))
	}

	if *format == "" {
		if len(formats) == 0 {
			return fmt.Errorf("no translation formats available")
		}
		fmtOptions := make([]huh.Option[string], len(formats))
		for i, f := range formats {
			fmtOptions[i] = huh.NewOption(f, f)
		}
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title("Output format").
				Options(fmtOptions...).
				Filtering(true).
				Value(format).
				Height(10),
		))
	}

	if promptOutputDir {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Output directory").
				Prompt(": ").
				Inline(true).
				Value(outputDir).
				Validate(requiredValidator("output directory")),
		))
	}

	if len(groups) == 0 {
		return nil
	}

	return huh.NewForm(groups...).WithTheme(Theme()).Run()
}
