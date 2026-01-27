// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package commands contains all CLI command definitions.
package commands

import (
	"github.com/dacolabs/cli/internal/session"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for the CLI.
func NewRootCmd(translators translate.Register) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "daco",
	}

	portsCmd := &cobra.Command{
		Use:               "ports",
		Short:             "Manage data product ports",
		PersistentPreRunE: session.PreRunLoad,
	}
	portsCmd.AddCommand(
		newPortsAddCmd(),
		newPortsListCmd(),
		newPortsRemoveCmd(),
		newPortsTranslateCmd(translators))

	connsCmd := &cobra.Command{
		Use:               "connections",
		Short:             "Manage data product connections",
		PersistentPreRunE: session.PreRunLoad,
	}
	connsCmd.AddCommand(
		newConnectionsAddCmd(),
		newConnectionsDescribeCmd(),
		newConnectionsListCmd(),
		newConnectionsRemoveCmd())

	rootCmd.AddCommand(
		newInitCmd(),
		portsCmd,
		connsCmd)
	return rootCmd
}
