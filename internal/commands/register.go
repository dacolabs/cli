// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package commands contains all CLI command definitions.
package commands

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for the CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "daco",
	}

	registerCommands(rootCmd)

	return rootCmd
}

func registerCommands(rootCmd *cobra.Command) {
	registerInitCmd(rootCmd)
	registerPortCmd(rootCmd)
}
