// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package commands contains all CLI command definitions.
package commands

import (
	"github.com/dacolabs/cli/internal/cmdctx"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for the CLI.
func NewRootCmd(translators translate.Register) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "daco",
	}

	registerInitCmd(rootCmd)

	portsCmd := &cobra.Command{
		Use:               "ports",
		Short:             "Manage data product ports",
		PersistentPreRunE: cmdctx.PreRunLoad,
	}
	registerPortsAddCmd(portsCmd)
	registerPortsListCmd(portsCmd)
	registerPortsRemoveCmd(portsCmd)
	registerPortTranslateCmd(portsCmd, translators)
	rootCmd.AddCommand(portsCmd)

	connsCmd := &cobra.Command{
		Use:               "conections",
		Short:             "Manage data product connections",
		PersistentPreRunE: cmdctx.PreRunLoad,
	}
	registerConnectionsDescribeCmd(connsCmd)
	registerConnectionsListCmd(portsCmd)
	registerConnectionsRemoveCmd(connsCmd)
	rootCmd.AddCommand(connsCmd)

	return rootCmd
}
