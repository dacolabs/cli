// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package commands contains all CLI command definitions.
package commands

import (
	"github.com/dacolabs/cli/internal/opendpi"
	"github.com/dacolabs/cli/internal/session"
	"github.com/dacolabs/cli/internal/translate"
	"github.com/dacolabs/cli/internal/version"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for the CLI.
func NewRootCmd(translators translate.Register) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "daco",
		Short: "Data product CLI tool",
		Long: `daco is a CLI for defining, managing, and translating data product specifications
using the OpenDPI standard. It supports managing ports (with JSON schemas),
infrastructure connections, and translating schemas to formats such as PySpark,
Avro, Protobuf, Go types, and more.`,
		Version: version.Short(),
	}
	rootCmd.SetVersionTemplate(version.Info() + "\n")

	portsCmd := &cobra.Command{
		Use:   "ports",
		Short: "Manage data product ports",
		Long: `Manage data product ports defined in the OpenDPI spec. Ports represent
data interfaces (input or output) with associated JSON schemas and connections.
Use subcommands to add, list, describe, remove, or translate port schemas.`,
		PersistentPreRunE: session.PreRunLoad,
	}
	portsCmd.AddCommand(
		newPortsAddCmd(),
		newPortsDescribeCmd(),
		newPortsListCmd(),
		newPortsRemoveCmd(),
		newPortsTranslateCmd(translators))

	connsCmd := &cobra.Command{
		Use:   "connections",
		Short: "Manage data product connections",
		Long: `Manage infrastructure connections defined in the OpenDPI spec. Connections
describe how to reach external systems (Kafka, PostgreSQL, S3, HTTP, etc.)
and are referenced by ports. Use subcommands to add, list, describe, or remove connections.`,
		PersistentPreRunE: session.PreRunLoad,
	}
	connsCmd.AddCommand(
		newConnectionsAddCmd(),
		newConnectionsDescribeCmd(),
		newConnectionsListCmd(),
		newConnectionsRemoveCmd())

	productCmd := &cobra.Command{
		Use:   "product",
		Short: "Manage data product metadata",
		Long: `Manage data product metadata defined in the OpenDPI spec.
Use subcommands to view or upgrade the product version.`,
		PersistentPreRunE: session.PreRunLoad,
	}
	productCmd.AddCommand(
		newProductVersionCmd(),
		newProductUpgradeCmd())

	rootCmd.AddCommand(
		newInitCmd(),
		newDescribeCmd(),
		portsCmd,
		connsCmd,
		productCmd)
	return rootCmd
}

func findConnectionName(conn *opendpi.Connection, connections map[string]opendpi.Connection) string {
	if conn == nil {
		return "unknown"
	}
	for name, c := range connections {
		if c.Type == conn.Type && c.Host == conn.Host {
			return name
		}
	}
	return "unknown"
}
