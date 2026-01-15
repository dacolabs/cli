// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package internal contains the main application logic for the CLI.
package internal

import (
	"context"

	"github.com/dacolabs/cli/internal/commands"
)

// Run is the main application logic, extracted for testability.
// It accepts OS dependencies as parameters (context, env lookup).
func Run(ctx context.Context, getenv func(string) string) error {
	rootCmd := commands.NewRootCmd()
	return rootCmd.ExecuteContext(ctx)
}
