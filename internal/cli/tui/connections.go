// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/engine"
)

func newConnectionsList(s *state) tview.Primitive {
	st := newSearchTable(s, "NAME", "PATH")
	st.SetTableTitle(" Connections ")

	var names []string
	refresh := func(selectName string) {
		out, err := engine.ConnectionsList(context.Background(), engine.ConnectionsListInput{Prj: s.project})
		if err != nil {
			s.showError(err)
			return
		}
		names = nil
		rows := [][]string{}
		selectIdx := -1
		for i, c := range out.Connections {
			names = append(names, c.Name)
			rows = append(rows, []string{c.Name, c.Path})
			if c.Name == selectName {
				selectIdx = i
			}
		}
		st.SetRows(rows)
		if selectIdx >= 0 {
			st.SelectOrigRow(selectIdx)
		}
	}

	st.Table().SetSelectedFunc(func(_, _ int) {
		idx := st.CurrentOrigRow()
		if idx < 0 {
			return
		}
		s.pushPage("detail-connection-"+names[idx], newConnectionDetail(s, names[idx]))
	})

	shell := shellShortcut(s)
	st.Table().SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Rune() {
		case '/':
			st.FocusSearch(s)
			return nil
		case 'n', 'N':
			connectionsCreateForm(s, refresh)
			return nil
		case 'd', 'D':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			name := names[idx]
			s.showConfirm("Delete connection", fmt.Sprintf("Delete connection %q?", name), func() {
				if _, err := engine.ConnectionsDelete(context.Background(), engine.ConnectionsDeleteInput{Prj: s.project, Name: name}); err != nil {
					s.showError(err)
					return
				}
				refresh("")
			})
			return nil
		}
		return shell(ev)
	})

	refresh("")
	return st.Primitive()
}

func newConnectionDetail(s *state, name string) tview.Primitive {
	out, err := engine.ConnectionsDescribe(context.Background(), engine.ConnectionsDescribeInput{Prj: s.project, Name: name})
	if err != nil {
		s.showError(err)
		return tview.NewBox()
	}

	title, titleH := titleBanner(name)
	hint := hintBar("esc back  ·  1-2 panes / tab  ·  q quit")

	infoLines := []string{
		fmt.Sprintf("[gray]Path:[-]        %s", tview.Escape(out.Path)),
		fmt.Sprintf("[gray]Exists:[-]      %t", out.Exists),
	}
	if out.Conn != nil {
		infoLines = append(infoLines,
			fmt.Sprintf("[gray]Type:[-]        %s", tview.Escape(out.Conn.Type)),
			fmt.Sprintf("[gray]Host:[-]        %s", tview.Escape(out.Conn.Host)),
		)
		if out.Conn.Description != "" {
			infoLines = append(infoLines, fmt.Sprintf("[gray]Description:[-] %s", tview.Escape(out.Conn.Description)))
		}
		infoLines = append(infoLines, fmt.Sprintf("[gray]Variables:[-]   %d", len(out.Conn.Variables)))
	}
	info := tview.NewTextView().SetDynamicColors(true)
	info.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignLeft)
	info.SetText(joinLines(infoLines))

	body := tview.NewTextView().SetDynamicColors(true).SetWrap(false).SetScrollable(true)
	body.SetBorder(true).SetTitle(" Definition ").SetTitleAlign(tview.AlignLeft)
	if raw, err := readFile(resolveRel(s.project.Path, out.Path)); err == nil {
		body.SetText(highlight(raw, "yaml"))
	} else {
		body.SetText("[red]" + err.Error() + "[-]")
	}

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, titleH, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(info, 0, 1, false).
		AddItem(body, 0, 2, true)

	focusables := []tview.Primitive{info, body}
	shortcuts := func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc, tcell.KeyLeft:
			s.popPage()
			return nil
		case tcell.KeyTab:
			cycleFocus(s, focusables)
			return nil
		}
		switch ev.Rune() {
		case '1':
			s.app.SetFocus(info)
			return nil
		case '2':
			s.app.SetFocus(body)
			return nil
		case 'q', 'Q':
			s.app.Stop()
			return nil
		}
		return ev
	}
	installShortcuts(shortcuts, info, body)

	return page
}

func connectionsCreateForm(s *state, after func(selectName string)) {
	var in engine.ConnectionsCreateInput
	in.Prj = s.project
	var varsBlob string
	form := tview.NewForm().
		AddInputField("Name", "", 30, nil, func(t string) { in.Name = t }).
		AddInputField("Path", "", 50, nil, func(t string) { in.Path = t }).
		AddInputField("Type", "", 30, nil, func(t string) { in.Type = t }).
		AddInputField("Host", "", 50, nil, func(t string) { in.Host = t }).
		AddInputField("Description", "", 50, nil, func(t string) { in.Description = t }).
		AddTextArea("Variables", "", 50, 4, 0, func(t string) { varsBlob = t })
	form.AddButton("Create", func() {
		vars, err := parseVarBlob(varsBlob)
		if err != nil {
			s.showError(err)
			return
		}
		in.Variables = vars
		if _, err := engine.ConnectionsCreate(context.Background(), in); err != nil {
			s.showError(err)
			return
		}
		s.pages.RemovePage("form-connection-create")
		after(in.Name)
	})
	form.AddButton("Cancel", func() { s.pages.RemovePage("form-connection-create") })
	form.SetCancelFunc(func() { s.pages.RemovePage("form-connection-create") })
	form.SetBorder(true).SetTitle(" Create Connection (Variables: one k=v per line) ").SetTitleAlign(tview.AlignLeft)
	s.pages.AddPage("form-connection-create", centered(form, 70, 18), true, true)
}

func parseVarBlob(blob string) (map[string]any, error) {
	blob = strings.TrimSpace(blob)
	if blob == "" {
		return nil, nil
	}
	out := map[string]any{}
	for _, line := range strings.Split(blob, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("invalid variable line %q (expected k=v)", line)
		}
		out[k] = v
	}
	return out, nil
}