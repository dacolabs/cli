// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func registerInitCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:  "init",
		RunE: runInit,
	}

	parent.AddCommand(cmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Hello World")
	return nil
}
