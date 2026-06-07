// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/settings"
)

func newProjectsView(s *state) tview.Primitive {
	title, titleH := dacoLogoBanner()
	hint := hintBar("↑/↓ navigate  ·  enter open  ·  / search  ·  q quit")

	st := newSearchTable(s, "NAME", "PATH")
	st.SetTableTitle(" Projects ")

	names := make([]string, 0, len(s.usr.Projects))
	for n := range s.usr.Projects {
		names = append(names, n)
	}
	sort.Strings(names)

	rows := make([][]string, 0, len(names))
	for _, n := range names {
		rows = append(rows, []string{n, s.usr.Projects[n]})
	}
	st.SetRows(rows)

	st.Table().SetSelectedFunc(func(_, _ int) {
		idx := st.CurrentOrigRow()
		if idx < 0 {
			return
		}
		name := names[idx]
		dir := s.usr.Projects[name]
		prj, err := settings.LoadProjectAt(dir)
		if err != nil {
			s.showError(fmt.Errorf("open %q: %w", name, err))
			return
		}
		s.switchToShell(prj)
	})

	st.Table().SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyCtrlC:
			s.app.Stop()
			return nil
		}
		switch ev.Rune() {
		case 'q', 'Q':
			s.app.Stop()
			return nil
		case '/':
			st.FocusSearch(s)
			return nil
		}
		return ev
	})

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, titleH, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(st.Primitive(), 0, 1, true)
}