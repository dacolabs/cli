// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

// Package tui provides an interactive terminal UI for browsing daco specs.
package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dacolabs/daco/internal/cli/session"
	"github.com/dacolabs/daco/internal/opendpi"
	"github.com/dacolabs/daco/internal/translate"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	pageList      = "list"
	pageDetail    = "detail"
	pageTranslate = "translate"
)

// Run starts the daco TUI against the loaded session context. It blocks until
// the user quits the application.
func Run(sctx *session.Context, translators translate.Register) error {
	app := tview.NewApplication()
	pages := tview.NewPages()

	allNames := sortedPortNames(sctx.Spec.Ports)

	title, titleHeight := buildTitleBanner(sctx.Spec.Info.Title)

	table := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)
	table.SetBorder(true).SetTitle(" Ports ")
	table.SetSelectedFunc(func(row, _ int) {
		if row < 1 {
			return
		}
		cell := table.GetCell(row, 0)
		if cell == nil {
			return
		}
		if _, ok := sctx.Spec.Ports[cell.Text]; ok {
			showDetail(app, pages, sctx, translators, cell.Text)
		}
	})

	populate := func(query string) {
		table.Clear()
		table.SetCell(0, 0, tview.NewTableCell("NAME").
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false))
		table.SetCell(0, 1, tview.NewTableCell("DESCRIPTION").
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))

		q := strings.ToLower(strings.TrimSpace(query))
		row := 1
		for _, name := range allNames {
			port := sctx.Spec.Ports[name]
			if q != "" &&
				!strings.Contains(strings.ToLower(name), q) &&
				!strings.Contains(strings.ToLower(port.Description), q) {
				continue
			}
			desc := port.Description
			if desc == "" {
				desc = "-"
			}
			table.SetCell(row, 0, tview.NewTableCell(name))
			table.SetCell(row, 1, tview.NewTableCell(desc).SetExpansion(1))
			row++
		}
		if row == 1 {
			msg := "(no ports defined)"
			if q != "" {
				msg = "(no matches)"
			}
			table.SetCell(1, 0, tview.NewTableCell(msg).
				SetTextColor(tcell.ColorGray).
				SetSelectable(false))
		} else {
			table.Select(1, 0)
		}
	}
	populate("")

	search := tview.NewInputField().
		SetLabel(" Search: ").
		SetFieldWidth(0)
	search.SetBorder(true)
	search.SetChangedFunc(populate)
	search.SetDoneFunc(func(_ tcell.Key) {
		app.SetFocus(table)
	})

	listPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(hintBar("/ search · ↑/↓ navigate · enter details · q quit"), 1, 0, false).
		AddItem(title, titleHeight, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(search, 3, 0, false)

	pages.AddPage(pageList, listPage, true, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		name, _ := pages.GetFrontPage()
		_, isInput := app.GetFocus().(*tview.InputField)

		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyEsc, tcell.KeyLeft:
			if name == pageTranslate {
				if isInput {
					return event
				}
				pages.SwitchToPage(pageDetail)
				return nil
			}
			if name == pageDetail {
				pages.SwitchToPage(pageList)
				return nil
			}
			if isInput && name == pageList {
				app.SetFocus(table)
				return nil
			}
		case tcell.KeyUp, tcell.KeyDown:
			if isInput && name == pageList {
				app.SetFocus(table)
				return event
			}
		}
		if !isInput {
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case '/':
				if name == pageList {
					app.SetFocus(search)
					return nil
				}
			}
		}
		return event
	})

	return app.SetRoot(pages, true).EnableMouse(true).Run()
}

// buildTitleBanner renders the product title as a single bold yellow line,
// centered with blank-line padding above and below.
func buildTitleBanner(title string) (*tview.TextView, int) {
	if strings.TrimSpace(title) == "" {
		title = "daco"
	}
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	tv.SetText("\n[yellow::b]" + tview.Escape(title) + "[-:-:-]\n")
	return tv, 3
}

func showDetail(app *tview.Application, pages *tview.Pages, sctx *session.Context, translators translate.Register, name string) {
	port, ok := sctx.Spec.Ports[name]
	if !ok {
		return
	}

	header := tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	header.SetBorder(true).SetTitle(fmt.Sprintf(" %s ", name))
	header.SetText(renderHeader(port))

	connections := buildConnectionsTable(port, sctx.Spec)

	schema := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	schema.SetBorder(true).SetTitle(" Schema ")
	schema.SetText(renderSchema(port.Schema))
	schema.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 't' {
			showTranslate(app, pages, sctx, translators, name)
			return nil
		}
		return event
	})

	body := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 0, 1, false).
		AddItem(connections, 0, 1, false).
		AddItem(schema, 0, 3, true)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(hintBar("t translate · esc/← back · q quit"), 1, 0, false).
		AddItem(body, 0, 1, true)

	pages.AddAndSwitchToPage(pageDetail, page, true)
	app.SetFocus(schema)
}

func showTranslate(app *tview.Application, pages *tview.Pages, sctx *session.Context, translators translate.Register, portName string) {
	port, ok := sctx.Spec.Ports[portName]
	if !ok {
		return
	}

	formats := make([]string, 0, len(translators))
	for k := range translators {
		formats = append(formats, k)
	}
	sort.Strings(formats)

	formatList := tview.NewList().ShowSecondaryText(false)
	formatList.SetBorder(true).SetTitle(" Format ")

	schemaView := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	schemaView.SetBorder(true).SetTitle(" Schema ")
	schemaView.SetText(renderSchema(port.Schema))

	translationView := tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	translationView.SetBorder(true).SetTitle(" Translation ")

	saveAs := tview.NewInputField().SetLabel(" Save as: ").SetFieldWidth(0)
	saveAs.SetBorder(true)

	status := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)

	pathEdited := false

	defaultPathFor := func(format string) string {
		t, ok := translators[format]
		if !ok {
			return ""
		}
		return filepath.Join("schemas", portName+t.FileExtension())
	}

	renderTranslation := func(format string) {
		t, ok := translators[format]
		if !ok {
			translationView.SetText("(unknown format)")
			return
		}
		translationView.SetTitle(fmt.Sprintf(" Translation (%s) ", format))
		outputDir := filepath.Dir(saveAs.GetText())
		if outputDir == "" || outputDir == "." {
			outputDir = "schemas"
		}
		data, err := t.Translate(portName, port.Schema, outputDir)
		if err != nil {
			translationView.SetText(fmt.Sprintf("[red](translation failed: %s)[-]", tview.Escape(err.Error())))
			return
		}
		translationView.SetText(highlightCode(string(data), langForFormat(format)))
	}

	currentFormat := func() string {
		idx := formatList.GetCurrentItem()
		if idx < 0 || idx >= len(formats) {
			return ""
		}
		return formats[idx]
	}

	formatList.SetChangedFunc(func(_ int, _, _ string, _ rune) {
		f := currentFormat()
		if f == "" {
			return
		}
		if !pathEdited {
			saveAs.SetText(defaultPathFor(f))
		}
		renderTranslation(f)
	})

	for _, f := range formats {
		formatList.AddItem(f, "", 0, nil)
	}

	saveAs.SetChangedFunc(func(_ string) {
		pathEdited = true
	})

	saveAs.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		path := strings.TrimSpace(saveAs.GetText())
		if path == "" {
			status.SetText("[red]save-as path is empty[-]")
			return
		}
		f := currentFormat()
		if f == "" {
			status.SetText("[red]no format selected[-]")
			return
		}
		t := translators[f]
		outputDir := filepath.Dir(path)
		data, err := t.Translate(portName, port.Schema, outputDir)
		if err != nil {
			status.SetText(fmt.Sprintf("[red]translate failed: %s[-]", tview.Escape(err.Error())))
			return
		}
		if outputDir != "" && outputDir != "." {
			if err := os.MkdirAll(outputDir, 0o750); err != nil {
				status.SetText(fmt.Sprintf("[red]mkdir failed: %s[-]", tview.Escape(err.Error())))
				return
			}
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			status.SetText(fmt.Sprintf("[red]save failed: %s[-]", tview.Escape(err.Error())))
			return
		}
		pages.SwitchToPage(pageDetail)
		pages.RemovePage(pageTranslate)
	})

	// Initialise to first format.
	if len(formats) > 0 {
		saveAs.SetText(defaultPathFor(formats[0]))
		pathEdited = false
		renderTranslation(formats[0])
	} else {
		translationView.SetText("(no translators registered)")
	}

	previews := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(schemaView, 0, 1, false).
		AddItem(translationView, 0, 1, false)

	body := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(formatList, 22, 0, true).
		AddItem(previews, 0, 1, false)

	footer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(saveAs, 3, 0, false).
		AddItem(status, 1, 0, false)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(hintBar(fmt.Sprintf(" %s · ↑/↓ pick format · tab save-as · enter save · esc back ", portName)), 1, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(footer, 4, 0, false)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() != tcell.KeyTab {
			return event
		}
		if app.GetFocus() == saveAs {
			app.SetFocus(formatList)
		} else {
			app.SetFocus(saveAs)
		}
		return nil
	})

	pages.AddAndSwitchToPage(pageTranslate, page, true)
	app.SetFocus(formatList)
}

func buildConnectionsTable(port opendpi.Port, spec *opendpi.Spec) *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(false, false)
	table.SetBorder(true).SetTitle(" Connections ")

	headers := []string{"NAME", "HOST", "LOCATION"}
	for col, h := range headers {
		table.SetCell(0, col, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetExpansion(1))
	}

	if len(port.Connections) == 0 {
		table.SetCell(1, 0, tview.NewTableCell("(none)").SetExpansion(1))
		return table
	}

	for i, pc := range port.Connections {
		row := i + 1
		host := "-"
		if pc.Connection != nil {
			host = fmt.Sprintf("%s://%s", pc.Connection.Type, pc.Connection.Host)
		}
		table.SetCell(row, 0, tview.NewTableCell(spec.ConnectionName(pc.Connection)).SetExpansion(1))
		table.SetCell(row, 1, tview.NewTableCell(host).SetExpansion(1))
		table.SetCell(row, 2, tview.NewTableCell(pc.Location).SetExpansion(1))
	}
	return table
}

func renderHeader(port opendpi.Port) string {
	var b strings.Builder
	desc := port.Description
	if desc == "" {
		desc = "(no description)"
	}
	fmt.Fprintf(&b, "[gray]Description:[-] %s\n", tview.Escape(desc))
	ref := port.SchemaRef
	if ref == "" {
		ref = "(inline)"
	}
	fmt.Fprintf(&b, "[gray]Schema ref:[-]  %s", tview.Escape(ref))
	return b.String()
}

func renderSchema(schema any) string {
	if schema == nil {
		return "(no schema)"
	}
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Sprintf("(failed to render schema: %v)", err)
	}
	return highlightCode(string(data), "json")
}

// chromaStyle is the colour theme used for all syntax-highlighted panes.
// Picked once at init so we don't re-resolve it on every keystroke.
var chromaStyle = func() *chroma.Style {
	if s := styles.Get("monokai"); s != nil {
		return s
	}
	return styles.Fallback
}()

// chromaFormatter outputs ANSI 256-colour escapes; tview.TranslateANSI then
// converts them into tview color tags so SetDynamicColors(true) renders them.
var chromaFormatter = func() chroma.Formatter {
	if f := formatters.Get("terminal256"); f != nil {
		return f
	}
	return formatters.Fallback
}()

// highlightCode returns syntax-highlighted source ready to feed into a tview
// TextView with SetDynamicColors(true). Falls back to escaped plain text on any
// error (unknown language, lex failure, format failure).
func highlightCode(source, lang string) string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return tview.Escape(source)
	}
	var buf bytes.Buffer
	if err := chromaFormatter.Format(&buf, chromaStyle, iterator); err != nil {
		return tview.Escape(source)
	}
	return tview.TranslateANSI(buf.String())
}

// formatToLang maps daco translator keys to chroma language identifiers.
// Anything not in the map falls through to chroma's auto-detection.
var formatToLang = map[string]string{
	"pyspark":            "python",
	"databricks-pyspark": "python",
	"pydantic":           "python",
	"python":             "python",
	"gotypes":            "go",
	"avro":               "json",
	"protobuf":           "protobuf",
	"scala":              "scala",
	"spark-scala":        "scala",
	"databricks-scala":   "scala",
	"databricks-sql":     "sql",
	"spark-sql":          "sql",
	"dqx-yaml":           "yaml",
	"markdown":           "markdown",
}

func langForFormat(format string) string {
	if lang, ok := formatToLang[format]; ok {
		return lang
	}
	return ""
}

func hintBar(text string) *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]" + text + "[-]")
}

func sortedPortNames(ports map[string]opendpi.Port) []string {
	names := make([]string, 0, len(ports))
	for name := range ports {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
