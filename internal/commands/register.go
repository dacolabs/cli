// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package commands contains all CLI command definitions.
package commands

import (
	"github.com/dacolabs/cli/internal/context"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for the CLI.
func NewRootCmd(translators translate.Register) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "daco",
	}

	registerInitCmd(rootCmd)
	cmd := &cobra.Command{
		Use:               "ports",
		Short:             "Manage data product ports",
		PersistentPreRunE: context.PreRunLoad,
	}

	registerPortTranslateCmd(cmd, translators)

	rootCmd.AddCommand(cmd)
	
	return rootCmd
}
