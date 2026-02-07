// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/prompts"
	"github.com/dacolabs/cli/internal/session"
	"github.com/spf13/cobra"
)

type productUpgradeOptions struct {
	bump string
}

func newProductUpgradeCmd() *cobra.Command {
	opts := &productUpgradeOptions{}

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the data product version",
		Long: `Upgrade the data product version by bumping major, minor, or patch.
In interactive mode, a form is shown to select the bump type.`,
		Example: `  # Interactive mode
  daco product upgrade

  # Non-interactive
  daco product upgrade --bump minor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return runProductUpgrade(cmd, ctx, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.bump, "bump", "b", "", "Bump type: major, minor, or patch")

	return cmd
}

func runProductUpgrade(cmd *cobra.Command, ctx *session.Context, opts *productUpgradeOptions) error {
	currentVersion := ctx.Spec.Info.Version

	var bumpType string
	if cmd.Flags().Changed("bump") {
		bumpType = opts.bump
		if bumpType != "major" && bumpType != "minor" && bumpType != "patch" {
			return fmt.Errorf("--bump must be one of: major, minor, patch")
		}
	} else {
		if err := prompts.RunProductUpgradeForm(&bumpType, currentVersion); err != nil {
			return err
		}
	}

	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format %q: expected X.Y.Z", currentVersion)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	ctx.Spec.Info.Version = newVersion

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specDir := ctx.Config.Path
	if !filepath.IsAbs(specDir) {
		specDir = filepath.Join(cwd, specDir)
	}

	var writer opendpi.Writer
	if _, err := os.Stat(filepath.Join(specDir, "opendpi.json")); err == nil {
		writer = opendpi.JSONWriter
	} else {
		writer = opendpi.YAMLWriter
	}

	if err := writer.Write(ctx.Spec, specDir); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	prompts.PrintResult([]prompts.ResultField{
		{Label: "Previous version", Value: currentVersion},
		{Label: "New version", Value: newVersion},
	}, "✓ Version upgraded")

	return nil
}
