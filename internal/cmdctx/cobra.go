// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package cmdctx

import (
	"errors"

	"github.com/spf13/cobra"
)

// FromCommand extracts the daco Context from a cobra.Command's context.
// Returns nil if no Context is stored.
func FromCommand(cmd *cobra.Command) *Context {
	return From(cmd.Context())
}

// RequireFromCommand extracts the daco Context from a cobra.Command's context,
// returning an error if not found.
func RequireFromCommand(cmd *cobra.Command) (*Context, error) {
	ctx := FromCommand(cmd)
	if ctx == nil {
		return nil, errors.New("project context not loaded")
	}
	return ctx, nil
}

// PreRunLoad returns a PersistentPreRunE function that loads the project
// context and stores it in the command's context.
func PreRunLoad(cmd *cobra.Command, _ []string) error {
	ctx, err := Load(cmd.Context())
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	return nil
}
