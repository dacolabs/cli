package cli

import (
	"github.com/dacolabs/daco/internal/cli/commands"
	"github.com/dacolabs/daco/internal/cli/session"
	"github.com/dacolabs/daco/internal/cli/tui"
	"github.com/dacolabs/daco/internal/translate"
	"github.com/dacolabs/daco/internal/version"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root cobra command with all subcommands attached.
func NewRootCmd(translators translate.Register) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "daco",
		Short:         "daco is the command-line interface for the daco platform",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := session.PreRunLoad(cmd, nil); err != nil {
				return err
			}
			sctx, err := session.RequireFromCommand(cmd)
			if err != nil {
				return err
			}
			return tui.Run(sctx, translators)
		},
	}
	rootCmd.SetVersionTemplate(version.Info() + "\n")

	addCommands(rootCmd, translators)

	return rootCmd
}

func addCommands(rootCmd *cobra.Command, translators translate.Register) {
	portsCmd := &cobra.Command{
		Use:   "ports",
		Short: "Manage data product ports",
		Long: `Manage data product ports defined in the OpenDPI spec. Ports represent
data interfaces (input or output) with associated JSON schemas and connections.
Use subcommands to add, list, describe, remove, or translate port schemas.`,
		PersistentPreRunE: session.PreRunLoad,
	}
	portsCmd.AddCommand(
		commands.NewPortsAddCmd(),
		commands.NewPortsDescribeCmd(),
		commands.NewPortsListCmd(),
		commands.NewPortsRemoveCmd(),
		commands.NewPortsTranslateCmd(translators))

	connsCmd := &cobra.Command{
		Use:   "connections",
		Short: "Manage data product connections",
		Long: `Manage infrastructure connections defined in the OpenDPI spec. Connections
	describe how to reach external systems (Kafka, PostgreSQL, S3, HTTP, etc.)
	and are referenced by ports. Use subcommands to add, list, describe, or remove connections.`,
		PersistentPreRunE: session.PreRunLoad,
	}
	connsCmd.AddCommand(
		commands.NewConnectionsAddCmd(),
		commands.NewConnectionsDescribeCmd(),
		commands.NewConnectionsListCmd(),
		commands.NewConnectionsRemoveCmd())

	productCmd := &cobra.Command{
		Use:   "product",
		Short: "Manage data product metadata",
		Long: `Manage data product metadata defined in the OpenDPI spec.
	Use subcommands to view or upgrade the product version.`,
		PersistentPreRunE: session.PreRunLoad,
	}
	productCmd.AddCommand(
		commands.NewProductVersionCmd(),
		commands.NewProductUpgradeCmd())

	rootCmd.AddCommand(
		commands.NewInitCmd(),
		commands.NewDescribeCmd(),
		portsCmd,
		connsCmd,
		productCmd)
}
