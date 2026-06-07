// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// searchTable is a vertical Flex containing a Table on top and a single-line
// bordered search field at the bottom (matching the original daco TUI). The
// table is filtered case-insensitively across every cell as the user types.
// Callers always operate on "original row" indices that are stable across
// filtering.
type searchTable struct {
	flex    *tview.Flex
	search  *tview.InputField
	table   *tview.Table
	headers []string

	allRows       [][]string
	visibleToOrig []int
	filter        string

	onChanged func(origRow int)
}

func newSearchTable(s *state, headers ...string) *searchTable {
	st := &searchTable{headers: headers}

	st.table = tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)
	st.table.SetBorder(true)

	st.search = tview.NewInputField().
		SetLabel(" Search: ").
		SetFieldWidth(0)
	st.search.SetBorder(true)

	st.flex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(st.table, 0, 1, true).
		AddItem(st.search, 3, 0, false)

	st.search.SetChangedFunc(func(text string) {
		st.filter = strings.ToLower(text)
		st.render()
	})
	st.search.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			st.filter = ""
			st.search.SetText("")
			st.render()
		}
		s.app.SetFocus(st.table)
	})
	st.search.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyUp, tcell.KeyDown:
			s.app.SetFocus(st.table)
			return ev
		}
		return ev
	})

	st.table.SetSelectionChangedFunc(func(_, _ int) {
		if st.onChanged != nil {
			st.onChanged(st.CurrentOrigRow())
		}
	})

	return st
}

func (st *searchTable) Primitive() tview.Primitive { return st.flex }
func (st *searchTable) Table() *tview.Table        { return st.table }
func (st *searchTable) FocusSearch(s *state)       { s.app.SetFocus(st.search) }

// SetTableTitle sets the bordered table's title.
func (st *searchTable) SetTableTitle(t string) {
	st.table.SetTitle(t).SetTitleAlign(tview.AlignLeft)
}

func (st *searchTable) SetRows(rows [][]string) {
	st.allRows = rows
	st.render()
}

func (st *searchTable) CurrentOrigRow() int {
	row, _ := st.table.GetSelection()
	if row < 1 || row > len(st.visibleToOrig) {
		return -1
	}
	return st.visibleToOrig[row-1]
}

func (st *searchTable) SelectOrigRow(origIdx int) {
	for visIdx, origi := range st.visibleToOrig {
		if origi == origIdx {
			st.table.Select(visIdx+1, 0)
			return
		}
	}
	if len(st.visibleToOrig) > 0 {
		st.table.Select(1, 0)
	}
}

func (st *searchTable) OnChanged(fn func(origRow int)) { st.onChanged = fn }

func (st *searchTable) render() {
	st.table.Clear()
	for i, h := range st.headers {
		st.table.SetCell(0, i,
			tview.NewTableCell(h).
				SetSelectable(false).
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetExpansion(1))
	}
	st.visibleToOrig = st.visibleToOrig[:0]
	for i, row := range st.allRows {
		if !st.matches(row) {
			continue
		}
		st.visibleToOrig = append(st.visibleToOrig, i)
		r := len(st.visibleToOrig)
		for j, cell := range row {
			st.table.SetCell(r, j, tview.NewTableCell(cell).SetExpansion(1))
		}
	}
	if len(st.visibleToOrig) > 0 {
		st.table.Select(1, 0)
		return
	}
	msg := "(empty)"
	if st.filter != "" {
		msg = "(no matches)"
	}
	st.table.SetCell(1, 0, tview.NewTableCell(msg).
		SetSelectable(false).
		SetTextColor(tcell.ColorGray))
	if st.onChanged != nil {
		st.onChanged(-1)
	}
}

func (st *searchTable) matches(row []string) bool {
	if st.filter == "" {
		return true
	}
	for _, cell := range row {
		if strings.Contains(strings.ToLower(cell), st.filter) {
			return true
		}
	}
	return false
}