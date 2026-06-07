// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/engine"
	"github.com/dacolabs/daco/internal/opendpi"
)

func newSchemasList(s *state) tview.Primitive {
	st := newSearchTable(s, "NAME", "PATH")
	st.SetTableTitle(" Schemas ")

	var names []string
	refresh := func(selectName string) {
		out, err := engine.SchemasList(context.Background(), engine.SchemasListInput{Prj: s.project})
		if err != nil {
			s.showError(err)
			return
		}
		names = nil
		rows := [][]string{}
		selectIdx := -1
		for i, sc := range out.Schemas {
			names = append(names, sc.Name)
			rows = append(rows, []string{sc.Name, sc.Path})
			if sc.Name == selectName {
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
		s.pushPage("detail-schema-"+names[idx], newSchemaDetail(s, names[idx]))
	})

	shell := shellShortcut(s)
	st.Table().SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Rune() {
		case '/':
			st.FocusSearch(s)
			return nil
		case 'n', 'N':
			schemasCreateForm(s, refresh)
			return nil
		case 'd', 'D':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			name := names[idx]
			s.showConfirm("Delete schema", fmt.Sprintf("Delete schema %q?", name), func() {
				if _, err := engine.SchemasDelete(context.Background(), engine.SchemasDeleteInput{Prj: s.project, Name: name}); err != nil {
					s.showError(err)
					return
				}
				refresh("")
			})
			return nil
		case 't', 'T':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			openSchemaTranslate(s, names[idx])
			return nil
		}
		return shell(ev)
	})

	refresh("")
	return st.Primitive()
}

func newSchemaDetail(s *state, name string) tview.Primitive {
	out, err := engine.SchemasDescribe(context.Background(), engine.SchemasDescribeInput{Prj: s.project, Name: name})
	if err != nil {
		s.showError(err)
		return tview.NewBox()
	}

	title, titleH := titleBanner(name)
	hint := hintBar("esc back  ·  1-2 panes / tab  ·  t translate  ·  q quit")

	infoLines := []string{
		fmt.Sprintf("[gray]Path:[-]        %s", tview.Escape(out.Path)),
		fmt.Sprintf("[gray]Exists:[-]      %t", out.Exists),
	}
	if out.Schema != nil {
		if t := opendpi.SchemaTypeString(out.Schema); t != "" {
			infoLines = append(infoLines, fmt.Sprintf("[gray]Type:[-]        %s", tview.Escape(t)))
		}
		if out.Schema.Title != "" {
			infoLines = append(infoLines, fmt.Sprintf("[gray]Title:[-]       %s", tview.Escape(out.Schema.Title)))
		}
		if out.Schema.Description != "" {
			infoLines = append(infoLines, fmt.Sprintf("[gray]Description:[-] %s", tview.Escape(out.Schema.Description)))
		}
		infoLines = append(infoLines,
			fmt.Sprintf("[gray]Properties:[-]  %d", len(out.Schema.Properties)),
			fmt.Sprintf("[gray]Required:[-]    %d", len(out.Schema.Required)),
		)
	}
	info := tview.NewTextView().SetDynamicColors(true)
	info.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignLeft)
	info.SetText(joinLines(infoLines))

	body := tview.NewTextView().SetDynamicColors(true).SetWrap(false).SetScrollable(true)
	body.SetBorder(true).SetTitle(" Schema ").SetTitleAlign(tview.AlignLeft)
	if raw, err := readFile(resolveRel(s.project.Path, out.Path)); err == nil {
		body.SetText(highlight(raw, "yaml"))
	} else {
		body.SetText("[red]" + err.Error() + "[-]")
	}

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, titleH, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(info, 0, 1, false).
		AddItem(body, 0, 3, true)

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
		case 't', 'T':
			openSchemaTranslate(s, name)
			return nil
		}
		return ev
	}
	installShortcuts(shortcuts, info, body)

	return page
}

func schemasCreateForm(s *state, after func(selectName string)) {
	var in engine.SchemasCreateInput
	in.Prj = s.project
	in.Type = "object"
	form := tview.NewForm().
		AddInputField("Name", "", 30, nil, func(t string) { in.Name = t }).
		AddInputField("Path", "", 50, nil, func(t string) { in.Path = t }).
		AddInputField("Type", "object", 30, nil, func(t string) { in.Type = t }).
		AddInputField("Title", "", 50, nil, func(t string) { in.Title = t }).
		AddInputField("Description", "", 50, nil, func(t string) { in.Description = t })
	form.AddButton("Create", func() {
		if _, err := engine.SchemasCreate(context.Background(), in); err != nil {
			s.showError(err)
			return
		}
		s.pages.RemovePage("form-schema-create")
		after(in.Name)
	})
	form.AddButton("Cancel", func() { s.pages.RemovePage("form-schema-create") })
	form.SetCancelFunc(func() { s.pages.RemovePage("form-schema-create") })
	form.SetBorder(true).SetTitle(" Create Schema ").SetTitleAlign(tview.AlignLeft)
	s.pages.AddPage("form-schema-create", centered(form, 70, 14), true, true)
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}