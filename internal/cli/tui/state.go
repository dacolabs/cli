// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/settings"
)

// state is the runtime shared by every page. It exposes the App + Pages and
// provides a tiny page-stack convenience so views can push/pop detail screens
// without juggling page-name strings everywhere.
type state struct {
	app   *tview.Application
	pages *tview.Pages
	usr   *settings.User
	stack []string // page names pushed by openPage; popPage walks back through them

	project *settings.Project

	// onShellShown is invoked by popPage after the last detail page is popped,
	// so the per-project shell can rebuild its current section to reflect any
	// mutations made from a detail view.
	onShellShown func()

	// shellSwitchSection swaps the active section. Set by the shell so that
	// section list views can wire it up to 1-4 key shortcuts.
	shellSwitchSection func(string)
}

func newState(app *tview.Application, pages *tview.Pages, usr *settings.User) *state {
	return &state{app: app, pages: pages, usr: usr}
}

func (s *state) setProject(p *settings.Project) { s.project = p }

// pushPage stacks a new page on top of whatever is currently shown and
// explicitly moves focus to it — otherwise the previously-focused widget
// (e.g. the shell's table) keeps receiving key events and shortcuts on the
// new page never fire.
func (s *state) pushPage(name string, p tview.Primitive) {
	s.pages.AddPage(name, p, true, true)
	s.stack = append(s.stack, name)
	s.app.SetFocus(p)
}

// popPage removes the top page and falls back to the page beneath it. When the
// stack is empty (i.e. we're back at the shell), the onShellShown hook fires
// so the shell can rebuild the active section.
func (s *state) popPage() {
	if len(s.stack) == 0 {
		return
	}
	top := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	s.pages.RemovePage(top)
	if len(s.stack) == 0 && s.onShellShown != nil {
		s.onShellShown()
	}
}