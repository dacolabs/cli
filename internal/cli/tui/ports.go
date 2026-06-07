// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/engine"
	"github.com/dacolabs/daco/internal/opendpi"
)

// portRef is the flat record used by the ports list and detail navigation.
type portRef struct {
	product string
	name    string
}

func newPortsList(s *state) tview.Primitive {
	st := newSearchTable(s, "PRODUCT", "PORT", "DESCRIPTION")
	st.SetTableTitle(" Ports ")

	var refs []portRef

	refresh := func(selectKey string) {
		refs = nil
		rows := [][]string{}
		products := sortedKeys(s.project.Products)
		selectIdx := -1
		for _, prod := range products {
			out, err := engine.PortsList(context.Background(), engine.PortsListInput{Prj: s.project, Product: prod})
			if err != nil {
				s.showError(err)
				return
			}
			for _, p := range out.Ports {
				ref := portRef{product: prod, name: p.Name}
				if portKey(ref) == selectKey {
					selectIdx = len(refs)
				}
				refs = append(refs, ref)
				rows = append(rows, []string{prod, p.Name, p.Description})
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
		r := refs[idx]
		s.pushPage("detail-port-"+r.product+"-"+r.name, newPortDetail(s, r.product, r.name))
	})

	shell := shellShortcut(s)
	st.Table().SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Rune() {
		case '/':
			st.FocusSearch(s)
			return nil
		case 'n', 'N':
			if len(s.project.Products) == 0 {
				s.showError(fmt.Errorf("no products yet — create one first"))
				return nil
			}
			s.showPicker("Create port in product", sortedKeys(s.project.Products), func(_ int, product string) {
				portsCreateForm(s, product, func(string) { refresh("") })
			})
			return nil
		case 'd', 'D':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			r := refs[idx]
			s.showConfirm("Delete port", fmt.Sprintf("Delete port %q from %q?", r.name, r.product), func() {
				if _, err := engine.PortsDelete(context.Background(), engine.PortsDeleteInput{Prj: s.project, Product: r.product, Name: r.name}); err != nil {
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
			r := refs[idx]
			portsLinkSchemaForm(s, r.product, r.name, func() { refresh(portKey(r)) })
			return nil
		case 't', 'T':
			idx := st.CurrentOrigRow()
			if idx < 0 {
				return nil
			}
			r := refs[idx]
			openPortTranslate(s, r.product, r.name)
			return nil
		}
		return shell(ev)
	})

	refresh("")
	return st.Primitive()
}

func portKey(r portRef) string { return r.product + "/" + r.name }

func newPortDetail(s *state, product, name string) tview.Primitive {
	out, err := engine.PortsDescribe(context.Background(), engine.PortsDescribeInput{Prj: s.project, Product: product, Name: name})
	if err != nil {
		s.showError(err)
		return tview.NewBox()
	}

	title, titleH := titleBanner(product + " / " + name)
	hint := hintBar("esc back  ·  1-3 panes / tab  ·  t translate  ·  l link schema  ·  u unlink schema  ·  c bind connection  ·  b unbind connection  ·  q quit")

	infoLines := []string{
		fmt.Sprintf("[gray]Product:[-]      %s", tview.Escape(out.Product)),
		fmt.Sprintf("[gray]Name:[-]         %s", tview.Escape(out.Name)),
	}
	if out.Description != "" {
		infoLines = append(infoLines, fmt.Sprintf("[gray]Description:[-]  %s", tview.Escape(out.Description)))
	}
	infoLines = append(infoLines,
		fmt.Sprintf("[gray]Connections:[-]  %d", out.ConnectionsCount),
		fmt.Sprintf("[gray]Has schema:[-]   %t", out.HasSchema),
	)
	if out.SchemaRef != "" {
		infoLines = append(infoLines, fmt.Sprintf("[gray]Schema $ref:[-]  %s", tview.Escape(out.SchemaRef)))
	}
	if out.SchemaType != "" {
		infoLines = append(infoLines, fmt.Sprintf("[gray]Schema type:[-]  %s", tview.Escape(out.SchemaType)))
	}
	if out.SchemaTitle != "" {
		infoLines = append(infoLines, fmt.Sprintf("[gray]Schema title:[-] %s", tview.Escape(out.SchemaTitle)))
	}
	if out.HasSchema {
		infoLines = append(infoLines,
			fmt.Sprintf("[gray]Properties:[-]   %d", out.SchemaProperties),
			fmt.Sprintf("[gray]Required:[-]     %d", out.SchemaRequired),
		)
	}

	info := tview.NewTextView().SetDynamicColors(true)
	info.SetBorder(true).SetTitle(" Info ").SetTitleAlign(tview.AlignLeft)
	info.SetText(joinLines(infoLines))

	bindingsTable := tview.NewTable().SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
	bindingsTable.SetBorder(true).SetTitle(" Connection bindings (U to unlink) ").SetTitleAlign(tview.AlignLeft)
	bindingsTable.SetCell(0, 0, tableHeader("CONNECTION"))
	bindingsTable.SetCell(0, 1, tableHeader("LOCATION"))
	for i, b := range out.Connections {
		bindingsTable.SetCell(i+1, 0, tview.NewTableCell(b.Connection).SetExpansion(1))
		bindingsTable.SetCell(i+1, 1, tview.NewTableCell(b.Location).SetExpansion(2))
	}
	if len(out.Connections) == 0 {
		bindingsTable.SetCell(1, 0, tview.NewTableCell("(none)").SetTextColor(tcell.ColorGray))
	} else {
		bindingsTable.Select(1, 0)
	}

	body := tview.NewTextView().SetDynamicColors(true).SetWrap(false).SetScrollable(true)
	body.SetBorder(true).SetTitle(" Schema ").SetTitleAlign(tview.AlignLeft)
	body.SetText(loadPortSchemaText(s, product, name))

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, titleH, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(info, 0, 1, false).
		AddItem(bindingsTable, 0, 1, false).
		AddItem(body, 0, 3, true)

	refreshSelf := func() {
		s.popPage()
		s.pushPage("detail-port-"+product+"-"+name, newPortDetail(s, product, name))
	}

	focusables := []tview.Primitive{info, bindingsTable, body}
	// Page-wide shortcuts that work no matter which child has focus.
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
			s.app.SetFocus(bindingsTable)
			return nil
		case '3':
			s.app.SetFocus(body)
			return nil
		case 'q', 'Q':
			s.app.Stop()
			return nil
		case 't', 'T':
			openPortTranslate(s, product, name)
			return nil
		case 'l', 'L':
			portsLinkSchemaForm(s, product, name, refreshSelf)
			return nil
		case 'u':
			s.showConfirm("Unlink schema", fmt.Sprintf("Unlink schema from port %q?", name), func() {
				if _, err := engine.PortsUnlinkSchema(context.Background(), engine.PortsUnlinkSchemaInput{
					Prj: s.project, Product: product, Port: name,
				}); err != nil {
					s.showError(err)
					return
				}
				refreshSelf()
			})
			return nil
		case 'c':
			portsBindConnectionForm(s, product, name, refreshSelf)
			return nil
		case 'b':
			// Picker: choose which binding to unbind.
			if len(out.Connections) == 0 {
				s.showError(fmt.Errorf("port %q has no connection bindings", name))
				return nil
			}
			items := make([]string, len(out.Connections))
			for i, b := range out.Connections {
				items[i] = b.Connection + " @ " + b.Location
			}
			s.showPicker("Unbind connection from "+name, items, func(idx int, _ string) {
				b := out.Connections[idx]
				if _, err := engine.PortsUnbindConnection(context.Background(), engine.PortsUnbindConnectionInput{
					Prj: s.project, Product: product, Port: name,
					Connection: b.Connection, Location: b.Location,
				}); err != nil {
					s.showError(err)
					return
				}
				refreshSelf()
			})
			return nil
		}
		return ev
	}

	// On the bindings table, `U` unlinks the highlighted port-connection
	// binding. Everything else falls through to the shared shortcuts.
	installShortcutsWithFallback(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Rune() != 'U' {
			return ev
		}
		row, _ := bindingsTable.GetSelection()
		if row == 0 || row > len(out.Connections) {
			return nil
		}
		b := out.Connections[row-1]
		s.showConfirm("Unlink binding", fmt.Sprintf("Remove binding %s @ %s from port %q?", b.Connection, b.Location, name), func() {
			if _, err := engine.PortsUnbindConnection(context.Background(), engine.PortsUnbindConnectionInput{
				Prj: s.project, Product: product, Port: name,
				Connection: b.Connection, Location: b.Location,
			}); err != nil {
				s.showError(err)
				return
			}
			refreshSelf()
		})
		return nil
	}, bindingsTable, shortcuts)
	installShortcuts(shortcuts, info, body)

	return page
}

func portsBindConnectionForm(s *state, product, port string, after func()) {
	// Only show connections that are already linked to this product.
	doc, err := opendpi.Load(resolveRel(s.project.Path, s.project.Products[product]))
	if err != nil {
		s.showError(err)
		return
	}
	var connNames []string
	for k := range doc.Connections {
		connNames = append(connNames, k)
	}
	if len(connNames) == 0 {
		s.showError(fmt.Errorf("no connections linked to %q yet — use products link first", product))
		return
	}
	sort.Strings(connNames)

	s.showPicker("Bind port "+port+" to connection", connNames, func(_ int, conn string) {
		var location string
		form := tview.NewForm().
			AddInputField("Location", "", 50, nil, func(t string) { location = t })
		form.AddButton("Bind", func() {
			if _, err := engine.PortsBindConnection(context.Background(), engine.PortsBindConnectionInput{
				Prj: s.project, Product: product, Port: port,
				Connection: conn, Location: location,
			}); err != nil {
				s.showError(err)
				return
			}
			s.pages.RemovePage("form-port-bind")
			after()
		})
		form.AddButton("Cancel", func() { s.pages.RemovePage("form-port-bind") })
		form.SetCancelFunc(func() { s.pages.RemovePage("form-port-bind") })
		form.SetBorder(true).
			SetTitle(" Bind " + port + " → " + conn + " ").
			SetTitleAlign(tview.AlignLeft)
		s.pages.AddPage("form-port-bind", centered(form, 70, 8), true, true)
	})
}

// loadPortSchemaText returns highlighted source for the port's schema. If $ref
// → loads the referenced file (yaml). If inline → marshals to JSON for a
// uniform read.
func loadPortSchemaText(s *state, product, name string) string {
	productPath, ok := s.project.Products[product]
	if !ok {
		return "[red](product not found)[-]"
	}
	productAbs := resolveRel(s.project.Path, productPath)
	doc, err := opendpi.Load(productAbs)
	if err != nil {
		return "[red]" + err.Error() + "[-]"
	}
	p, ok := doc.Ports[name]
	if !ok || p.Schema == nil {
		return "[gray](no schema)[-]"
	}
	if p.Schema.Ref != "" {
		path := p.Schema.Ref
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(productAbs), path)
		}
		raw, err := readFile(path)
		if err != nil {
			return "[red]" + err.Error() + "[-]"
		}
		return highlight(raw, "yaml")
	}
	b, err := json.MarshalIndent(&p.Schema.Schema, "", "  ")
	if err != nil {
		return "[red]" + err.Error() + "[-]"
	}
	return highlight(string(b), "json")
}

func portsCreateForm(s *state, product string, after func(selectName string)) {
	in := engine.PortsCreateInput{Prj: s.project, Product: product}
	form := tview.NewForm().
		AddInputField("Product", product, 30, nil, func(t string) { in.Product = t }).
		AddInputField("Name", "", 30, nil, func(t string) { in.Name = t }).
		AddInputField("Description", "", 50, nil, func(t string) { in.Description = t })
	form.AddButton("Create", func() {
		if _, err := engine.PortsCreate(context.Background(), in); err != nil {
			s.showError(err)
			return
		}
		s.pages.RemovePage("form-port-create")
		after(in.Name)
	})
	form.AddButton("Cancel", func() { s.pages.RemovePage("form-port-create") })
	form.SetCancelFunc(func() { s.pages.RemovePage("form-port-create") })
	form.SetBorder(true).SetTitle(" Create Port ").SetTitleAlign(tview.AlignLeft)
	s.pages.AddPage("form-port-create", centered(form, 70, 10), true, true)
}

func portsLinkSchemaForm(s *state, product, port string, after func()) {
	schemas := sortedKeys(s.project.Schemas)
	if len(schemas) == 0 {
		s.showError(fmt.Errorf("no schemas to link — create one first"))
		return
	}
	s.showPicker("Link "+port+" → schema", schemas, func(_ int, label string) {
		_, err := engine.PortsLinkSchema(context.Background(), engine.PortsLinkSchemaInput{
			Prj: s.project, Product: product, Port: port, Schema: label,
		})
		if err != nil {
			s.showError(err)
			return
		}
		after()
	})
}

// sortedKeysSorted returns the keys of m in sorted order (kept for symmetry).
func sortedKeysAny(m map[string]opendpi.Port) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}