// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// newShell builds the per-project home: logo + project subtitle + section bar
// + hint + the active section's content. Sections are rebuilt every time the
// user switches to them (or after popping back from a detail page) so any
// mutations show up immediately.
func newShell(s *state) tview.Primitive {
	current := sectionLabels[0]
	logo, logoH := dacoLogoBanner()
	subtitle := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow::b]— " + tview.Escape(s.project.Name) + " —[-:-:-]")
	section := sectionBar(current)
	hint := hintBar(sectionHint(current))

	content := tview.NewPages()
	build := map[string]func(s *state) tview.Primitive{
		"Products":    newProductsList,
		"Connections": newConnectionsList,
		"Schemas":     newSchemasList,
		"Ports":       newPortsList,
	}

	mountCurrent := func() {
		p := build[current](s)
		content.RemovePage(current)
		content.AddPage(current, p, true, true)
		content.SwitchToPage(current)
		s.app.SetFocus(p)
	}
	mountCurrent()

	switchTo := func(name string) {
		current = name
		section.Clear()
		section.SetText(sectionBar(current).GetText(true))
		hint.SetText("[gray]" + sectionHint(current) + "[-]")
		mountCurrent()
	}
	s.shellSwitchSection = switchTo

	// Hook the back-to-shell refresh so popping from a detail rebuilds the
	// section content.
	s.onShellShown = mountCurrent

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(logo, logoH, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(section, 1, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(content, 0, 1, true)

	root.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		// Skip these hotkeys when the user is typing into a search input or
		// otherwise editing text — let the input have the key.
		if _, isInput := s.app.GetFocus().(*tview.InputField); isInput {
			return ev
		}
		if _, isArea := s.app.GetFocus().(*tview.TextArea); isArea {
			return ev
		}
		switch ev.Key() {
		case tcell.KeyCtrlC:
			s.app.Stop()
			return nil
		}
		switch ev.Rune() {
		case '1':
			switchTo("Products")
			return nil
		case '2':
			switchTo("Connections")
			return nil
		case '3':
			switchTo("Schemas")
			return nil
		case '4':
			switchTo("Ports")
			return nil
		case 'q', 'Q':
			s.app.Stop()
			return nil
		case 'p', 'P':
			s.switchToProjects()
			return nil
		}
		return ev
	})

	return root
}