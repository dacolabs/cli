// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package commands contains all CLI command definitions.
package commands

import (
	"github.com/dacolabs/cli/internal/context"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for the CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "daco",
	}

	registerInitCmd(rootCmd)
	registerConnectionsCmd(rootCmd)

	return rootCmd
}

func registerConnectionsCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:               "connections",
		Short:             "Manage infrastructure connections",
		PersistentPreRunE: context.PreRunLoad,
	}

	registerConnectionsListCmd(cmd)

	parent.AddCommand(cmd)
}
