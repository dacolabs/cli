// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/engine"
)

const (
	pageError   = "modal-error"
	pageConfirm = "modal-confirm"
	pagePicker  = "modal-picker"
)

// showError pops a modal that lists the validation/domain error and dismisses
// itself on OK. engine.Errors is rendered as a bulleted list.
func (s *state) showError(err error) {
	if err == nil {
		return
	}
	modal := tview.NewModal().
		SetText(formatError(err)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(int, string) {
			s.pages.RemovePage(pageError)
		})
	s.pages.AddPage(pageError, modal, true, true)
}

func formatError(err error) string {
	var list engine.Errors
	if errors.As(err, &list) {
		var b strings.Builder
		b.WriteString("Validation failed:\n")
		for _, e := range list {
			fmt.Fprintf(&b, "  • %s\n", e)
		}
		return strings.TrimRight(b.String(), "\n")
	}
	return err.Error()
}

// showConfirm prompts yes/no and invokes onYes only if the user chooses Yes.
func (s *state) showConfirm(title, message string, onYes func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(_ int, label string) {
			s.pages.RemovePage(pageConfirm)
			if label == "Yes" {
				onYes()
			}
		})
	s.pages.AddPage(pageConfirm, modal, true, true)
}

// showPicker presents a single-column searchable table. onPick(idx, label)
// fires when the user selects a row (Enter); Esc dismisses without calling.
func (s *state) showPicker(title string, items []string, onPick func(idx int, label string)) {
	st := newSearchTable(s, "NAME")
	rows := make([][]string, len(items))
	for i, it := range items {
		rows[i] = []string{it}
	}
	st.SetRows(rows)

	st.Table().SetSelectedFunc(func(_, _ int) {
		idx := st.CurrentOrigRow()
		if idx < 0 {
			return
		}
		s.pages.RemovePage(pagePicker)
		onPick(idx, items[idx])
	})
	st.Table().SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			s.pages.RemovePage(pagePicker)
			return nil
		}
		switch ev.Rune() {
		case '/':
			st.FocusSearch(s)
			return nil
		}
		return ev
	})

	box := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText(" "+title+" ").SetTextColor(tcell.ColorYellow), 1, 0, false).
		AddItem(st.Primitive(), 0, 1, true).
		AddItem(tview.NewTextView().SetText(" [Enter] pick  [/] search  [Esc] cancel ").SetTextColor(tcell.ColorGray), 1, 0, false)
	box.SetBorder(true).SetTitleAlign(tview.AlignLeft)

	frame := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(box, 0, 2, true).
			AddItem(nil, 0, 1, false),
			0, 2, true).
		AddItem(nil, 0, 1, false)

	s.pages.AddPage(pagePicker, frame, true, true)
}