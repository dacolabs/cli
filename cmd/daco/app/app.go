// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package app

import (
	"context"

	"github.com/dacolabs/daco/internal/cli"
	"github.com/dacolabs/daco/internal/cli/commands"
	"github.com/dacolabs/daco/internal/cli/tui"
)

func Run(ctx context.Context, _ func(string) string) error {
	cli.TUIRunner = tui.Run
	root := cli.New()
	root.AddCommand(commands.Init())
	root.AddCommand(commands.Products())
	root.AddCommand(commands.Connections())
	root.AddCommand(commands.Schemas())
	root.AddCommand(commands.Ports())
	root.AddCommand(commands.Format())
	root.AddCommand(commands.Lint())
	root.AddCommand(commands.Version())
	return root.ExecuteContext(ctx)
}