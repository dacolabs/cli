// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"context"
	"errors"
	"fmt"

	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli"
	"github.com/dacolabs/daco/internal/cli/settings"
)

const (
	pageProjects = "home-projects"
	pageShell    = "home-shell"
)

func Run(ctx context.Context) error {
	usr := cli.User(ctx)
	if usr == nil {
		return errors.New("user settings not loaded")
	}

	app := tview.NewApplication()
	pages := tview.NewPages()
	s := newState(app, pages, usr)

	prj, err := settings.LoadProject(usr)
	switch {
	case err == nil:
		s.setProject(prj)
		pages.AddPage(pageShell, newShell(s), true, true)
	case errors.Is(err, settings.ErrProjectNotFound):
		pages.AddPage(pageProjects, newProjectsView(s), true, true)
	default:
		return fmt.Errorf("load project: %w", err)
	}

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}

// switchToProjects swaps the root to the projects registry view.
func (s *state) switchToProjects() {
	s.setProject(nil)
	s.stack = nil
	s.pages.RemovePage(pageShell)
	if !s.pages.HasPage(pageProjects) {
		s.pages.AddPage(pageProjects, newProjectsView(s), true, true)
	} else {
		s.pages.SwitchToPage(pageProjects)
	}
}

// switchToShell rebuilds the per-project shell for the given project.
func (s *state) switchToShell(prj *settings.Project) {
	s.setProject(prj)
	s.stack = nil
	s.pages.RemovePage(pageShell)
	s.pages.AddPage(pageShell, newShell(s), true, true)
	s.pages.SwitchToPage(pageShell)
}