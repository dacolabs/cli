// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/engine"
	"github.com/dacolabs/daco/internal/opendpi"
)

func newProductsList(s *state) tview.Primitive {
	st := newSearchTable(s, "NAME", "PATH")
	st.SetTableTitle(" Products ")

	var names []string

	refresh := func(selectName string) {
		out, err := engine.ProductsList(context.Background(), engine.ProductsListInput{Prj: s.project})
		if err != nil {
			s.showError(err)
			return
		}
		names = nil
		rows := [][]string{}
		selectIdx := -1
		for i, p := range out.Products {
			names = append(names, p.Name)
			rows = append(rows, []string{p.Name, p.Path})
			if p.Name == selectName {
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
		s.pushPage("detail-product-"+names[idx], newProductDetail(s, names[idx]))
	})

	shell := shellShortcut(s)
	st.Table().SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Rune() {
		case '/':
			st.FocusSearch(s)
			return nil
		case 'n', 'N':
			productsCreateForm(s, refresh)
			return nil
		case 'd', 'D':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			name := names[idx]
			s.showConfirm("Delete product", fmt.Sprintf("Delete product %q?", name), func() {
				if _, err := engine.ProductsDelete(context.Background(), engine.ProductsDeleteInput{Prj: s.project, Name: name}); err != nil {
					s.showError(err)
					return
				}
				refresh("")
			})
			return nil
		case 'l', 'L':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			productsLinkForm(s, names[idx], func() { refresh(names[idx]) })
			return nil
		}
		return shell(ev)
	})

	refresh("")
	return st.Primitive()
}

func newProductDetail(s *state, name string) tview.Primitive {
	out, err := engine.ProductsDescribe(context.Background(), engine.ProductsDescribeInput{Prj: s.project, Name: name})
	if err != nil {
		s.showError(err)
		return tview.NewBox()
	}

	title, titleH := titleBanner(name)
	hint := hintBar("esc back  ·  1-3 panes / tab  ·  / search ports  ·  n new port  ·  l link conn  ·  u unlink conn  ·  q quit")

	info := tview.NewTextView().SetDynamicColors(true)
	info.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignLeft)
	infoLines := []string{
		fmt.Sprintf("[gray]Path:[-]        %s", tview.Escape(out.Path)),
		fmt.Sprintf("[gray]Exists:[-]      %t", out.Exists),
	}
	if out.Info != nil {
		infoLines = append(infoLines,
			fmt.Sprintf("[gray]Title:[-]       %s", tview.Escape(out.Info.Title)),
			fmt.Sprintf("[gray]Version:[-]     %s", tview.Escape(out.Info.Version)),
		)
	}
	infoLines = append(infoLines,
		fmt.Sprintf("[gray]Connections:[-] %d", out.ConnectionsCount),
		fmt.Sprintf("[gray]Ports:[-]       %d", out.PortsCount),
	)
	info.SetText(joinLines(infoLines))

	connTable := tview.NewTable().SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
	connTable.SetBorder(true).SetTitle(" Connections (u to unlink) ").SetTitleAlign(tview.AlignLeft)
	connTable.SetCell(0, 0, tableHeader("NAME"))
	connTable.SetCell(0, 1, tableHeader("FORM"))

	portsST := newSearchTable(s, "NAME", "DESCRIPTION", "SCHEMA")
	portsST.SetTableTitle(" Ports ")

	productAbs := resolveRel(s.project.Path, out.Path)
	doc, derr := opendpi.Load(productAbs)
	var portNames []string
	var connKeys []string
	if derr == nil {
		for k := range doc.Connections {
			connKeys = append(connKeys, k)
		}
		sort.Strings(connKeys)
		for i, k := range connKeys {
			c := doc.Connections[k]
			form := "inline"
			if c.Ref != "" {
				form = "$ref " + c.Ref
			}
			connTable.SetCell(i+1, 0, tview.NewTableCell(k).SetExpansion(1))
			connTable.SetCell(i+1, 1, tview.NewTableCell(form).SetExpansion(2))
		}
		if len(connKeys) == 0 {
			connTable.SetCell(1, 0, tview.NewTableCell("(none)").SetTextColor(tcell.ColorGray))
		}

		for k := range doc.Ports {
			portNames = append(portNames, k)
		}
		sort.Strings(portNames)
		portRows := make([][]string, 0, len(portNames))
		for _, k := range portNames {
			p := doc.Ports[k]
			form := "—"
			if p.Schema != nil {
				if p.Schema.Ref != "" {
					form = "$ref " + p.Schema.Ref
				} else {
					form = "inline"
				}
			}
			portRows = append(portRows, []string{k, p.Description, form})
		}
		portsST.SetRows(portRows)
	} else {
		s.showError(derr)
	}

	portsST.Table().SetSelectedFunc(func(_, _ int) {
		idx := portsST.CurrentOrigRow()
		if idx < 0 {
			return
		}
		s.pushPage("detail-port-"+name+"-"+portNames[idx], newPortDetail(s, name, portNames[idx]))
	})

	refreshSelf := func() {
		s.popPage()
		s.pushPage("detail-product-"+name, newProductDetail(s, name))
	}

	focusables := []tview.Primitive{info, connTable, portsST.Table()}
	// Shared shortcuts that should work no matter which focusable child
	// currently has focus.
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
			s.app.SetFocus(connTable)
			return nil
		case '3':
			s.app.SetFocus(portsST.Table())
			return nil
		case 'q', 'Q':
			s.app.Stop()
			return nil
		case 'l', 'L':
			productsLinkForm(s, name, refreshSelf)
			return nil
		case 'n', 'N':
			portsCreateForm(s, name, func(string) { refreshSelf() })
			return nil
		case '/':
			s.app.SetFocus(portsST.Table())
			portsST.FocusSearch(s)
			return nil
		}
		return ev
	}

	// On the connections table, `u` unlinks the highlighted connection.
	installShortcutsWithFallback(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Rune() != 'u' && ev.Rune() != 'U' {
			return ev
		}
		row, _ := connTable.GetSelection()
		if row == 0 || row > len(connKeys) {
			return nil
		}
		conn := connKeys[row-1]
		s.showConfirm("Unlink connection", fmt.Sprintf("Unlink connection %q from product %q?", conn, name), func() {
			if _, err := engine.ProductsUnlink(context.Background(), engine.ProductsUnlinkInput{
				Prj: s.project, ProductName: name, ConnectionName: conn,
			}); err != nil {
				s.showError(err)
				return
			}
			refreshSelf()
		})
		return nil
	}, connTable, shortcuts)
	installShortcuts(shortcuts, info, portsST.Table())

	body := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(info, 0, 1, false).
		AddItem(connTable, 0, 1, false).
		AddItem(portsST.Primitive(), 0, 2, true)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, titleH, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(body, 0, 1, true)

	return page
}

func tableHeader(s string) *tview.TableCell {
	return tview.NewTableCell(s).
		SetSelectable(false).
		SetTextColor(tcell.ColorYellow).
		SetAttributes(tcell.AttrBold).
		SetExpansion(1)
}

func joinLines(lines []string) string {
	out := "\n"
	for _, l := range lines {
		out += "  " + l + "\n"
	}
	return out
}

func productsCreateForm(s *state, after func(selectName string)) {
	var name, path string
	form := tview.NewForm().
		AddInputField("Name", "", 30, nil, func(t string) { name = t }).
		AddInputField("Path", "", 50, nil, func(t string) { path = t })
	form.AddButton("Create", func() {
		_, err := engine.ProductsCreate(context.Background(), engine.ProductsCreateInput{Prj: s.project, Name: name, Path: path})
		if err != nil {
			s.showError(err)
			return
		}
		s.pages.RemovePage("form-product-create")
		after(name)
	})
	form.AddButton("Cancel", func() { s.pages.RemovePage("form-product-create") })
	form.SetCancelFunc(func() { s.pages.RemovePage("form-product-create") })
	form.SetBorder(true).SetTitle(" Create Product ").SetTitleAlign(tview.AlignLeft)
	s.pages.AddPage("form-product-create", centered(form, 60, 9), true, true)
}

func productsLinkForm(s *state, productName string, after func()) {
	conns := sortedKeys(s.project.Connections)
	if len(conns) == 0 {
		s.showError(fmt.Errorf("no connections to link — create one first"))
		return
	}
	s.showPicker("Link "+productName+" to connection", conns, func(_ int, label string) {
		_, err := engine.ProductsLink(context.Background(), engine.ProductsLinkInput{
			Prj: s.project, ProductName: productName, ConnectionName: label,
		})
		if err != nil {
			s.showError(err)
			return
		}
		after()
	})
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func centered(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 0, true).
			AddItem(nil, 0, 1, false),
			width, 0, true).
		AddItem(nil, 0, 1, false)
}

// resolveRel resolves p against the dir of projectFile when p is relative.
func resolveRel(projectFile, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(filepath.Dir(projectFile), p)
}