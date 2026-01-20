// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"github.com/spf13/cobra"
)

func registerPortCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "port",
		Short: "Manage data product ports",
	}

	registerPortTranslateCmd(cmd)

	parent.AddCommand(cmd)
}
