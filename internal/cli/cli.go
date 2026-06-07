// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dacolabs/daco/internal/cli/settings"
)

// TUIRunner is set by the cmd/daco bootstrap (which imports both cli and tui)
// to break the otherwise-cyclic dependency. When nil, `daco` with no args
// falls back to printing help.
var TUIRunner func(ctx context.Context) error

func tuiRunner(ctx context.Context) error {
	if TUIRunner == nil {
		return errors.New("TUI not registered")
	}
	return TUIRunner(ctx)
}

func New() *cobra.Command {
	return &cobra.Command{
		Use:           "daco",
		Short:         "Daco CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			usr, err := loadOrCreateUser()
			if err != nil {
				return err
			}
			cmd.SetContext(WithUser(cmd.Context(), usr))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return cmd.Help()
			}
			return tuiRunner(cmd.Context())
		},
	}
}

func loadOrCreateUser() (*settings.User, error) {
	usr, err := settings.LoadUser()
	if err == nil {
		return usr, nil
	}
	if !errors.Is(err, settings.ErrUserNotFound) {
		return nil, err
	}
	usr = &settings.User{Projects: map[string]string{}}
	if err := usr.Save(); err != nil {
		return nil, fmt.Errorf("create user settings: %w", err)
	}
	return usr, nil
}