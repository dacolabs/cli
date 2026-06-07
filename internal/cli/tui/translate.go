// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/rivo/tview"

	"github.com/dacolabs/daco/internal/cli/engine"
	"github.com/dacolabs/daco/internal/opendpi"
	"github.com/dacolabs/daco/internal/translate/registry"
)

const (
	pageTranslate          = "translate"
	defaultTranslateFormat = "pydantic"
)

func openSchemaTranslate(s *state, schemaName string) {
	path, ok := s.project.Schemas[schemaName]
	if !ok {
		s.showError(fmt.Errorf("schema %q not found", schemaName))
		return
	}
	abs := resolveRel(s.project.Path, path)
	openTranslateView(s, translateTarget{
		name:       schemaName,
		title:      "translate · " + schemaName,
		sourcePath: abs,
		sourceLang: "yaml",
		loadSchema: func() (*jsonschema.Schema, error) { return loadJSONSchema(abs) },
		save: func(format, outputDir string) error {
			_, err := engine.SchemasTranslate(context.Background(), engine.SchemasTranslateInput{
				Prj: s.project, Schema: schemaName, Format: format, OutputDir: outputDir,
			})
			return err
		},
	})
}

func openPortTranslate(s *state, product, port string) {
	doc, err := opendpi.Load(resolveRel(s.project.Path, s.project.Products[product]))
	if err != nil {
		s.showError(err)
		return
	}
	p, ok := doc.Ports[port]
	if !ok {
		s.showError(fmt.Errorf("port %q not found in %q", port, product))
		return
	}
	if p.Schema == nil {
		s.showError(fmt.Errorf("port %q has no schema", port))
		return
	}

	var sourcePath string
	var loader func() (*jsonschema.Schema, error)
	productAbs := resolveRel(s.project.Path, s.project.Products[product])

	if p.Schema.Ref != "" {
		ref := p.Schema.Ref
		if !filepath.IsAbs(ref) {
			ref = filepath.Join(filepath.Dir(productAbs), ref)
		}
		sourcePath = ref
		loader = func() (*jsonschema.Schema, error) { return loadJSONSchema(ref) }
	} else {
		sourcePath = productAbs
		schemaCopy := p.Schema.Schema
		loader = func() (*jsonschema.Schema, error) { return &schemaCopy, nil }
	}

	openTranslateView(s, translateTarget{
		name:       port,
		title:      "translate · " + product + " / " + port,
		sourcePath: sourcePath,
		sourceLang: "yaml",
		loadSchema: loader,
		save: func(format, outputDir string) error {
			_, err := engine.PortsTranslate(context.Background(), engine.PortsTranslateInput{
				Prj: s.project, Product: product, Port: port, Format: format, OutputDir: outputDir,
			})
			return err
		},
	})
}

type translateTarget struct {
	name       string
	title      string
	sourcePath string
	sourceLang string
	loadSchema func() (*jsonschema.Schema, error)
	save       func(format, outputDir string) error
}

func openTranslateView(s *state, target translateTarget) {
	reg := registry.Default()
	formats := reg.Available()
	sort.Strings(formats)

	title, titleH := titleBanner(target.title)
	hint := hintBar("↑/↓ pick format  ·  tab focus save-as  ·  enter save  ·  esc back  ·  q quit")

	// Format list.
	formatList := tview.NewList().ShowSecondaryText(false)
	formatList.SetBorder(true).SetTitle(" Format ").SetTitleAlign(tview.AlignLeft)
	for _, f := range formats {
		formatList.AddItem(f, "", 0, nil)
	}

	schemaView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(false)
	schemaView.SetBorder(true).SetTitle(" Source ").SetTitleAlign(tview.AlignLeft)
	if raw, err := os.ReadFile(target.sourcePath); err == nil {
		schemaView.SetText(highlight(string(raw), target.sourceLang))
	} else {
		schemaView.SetText("[red]" + err.Error() + "[-]")
	}

	translationView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(false)
	translationView.SetBorder(true).SetTitle(" Translation ").SetTitleAlign(tview.AlignLeft)

	saveAs := tview.NewInputField().SetLabel(" Save as: ").SetFieldWidth(0)
	saveAs.SetBorder(true)
	status := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	pathEdited := false

	defaultPathFor := func(format string) string {
		t, ok := reg[format]
		if !ok {
			return ""
		}
		return filepath.Join("generated", target.name+t.FileExtension())
	}

	renderTranslation := func(format string) {
		t, err := reg.Get(format)
		if err != nil {
			translationView.SetText("[red]" + err.Error() + "[-]")
			return
		}
		schema, err := target.loadSchema()
		if err != nil {
			translationView.SetText("[red]load schema: " + err.Error() + "[-]")
			return
		}
		translationView.SetTitle(fmt.Sprintf(" Translation (%s) ", format))
		outputDir := filepath.Dir(saveAs.GetText())
		if outputDir == "" || outputDir == "." {
			outputDir = "generated"
		}
		data, err := t.Translate(target.name, schema, outputDir)
		if err != nil {
			translationView.SetText(fmt.Sprintf("[red](translation failed: %s)[-]", tview.Escape(err.Error())))
			return
		}
		translationView.SetText(highlight(string(data), langForFormat(format)))
		translationView.ScrollToBeginning()
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

	saveAs.SetChangedFunc(func(_ string) { pathEdited = true })
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
		t := reg[f]
		outputDir := filepath.Dir(path)
		schema, err := target.loadSchema()
		if err != nil {
			status.SetText(fmt.Sprintf("[red]%s[-]", tview.Escape(err.Error())))
			return
		}
		data, err := t.Translate(target.name, schema, outputDir)
		if err != nil {
			status.SetText(fmt.Sprintf("[red]translate failed: %s[-]", tview.Escape(err.Error())))
			return
		}
		if outputDir != "" && outputDir != "." {
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				status.SetText(fmt.Sprintf("[red]mkdir failed: %s[-]", tview.Escape(err.Error())))
				return
			}
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			status.SetText(fmt.Sprintf("[red]save failed: %s[-]", tview.Escape(err.Error())))
			return
		}
		status.SetText(fmt.Sprintf("[green]wrote %s[-]", tview.Escape(path)))
	})

	// Default selection.
	defaultIdx := 0
	for i, f := range formats {
		if f == defaultTranslateFormat {
			defaultIdx = i
			break
		}
	}
	formatList.SetCurrentItem(defaultIdx)
	if len(formats) > 0 {
		saveAs.SetText(defaultPathFor(formats[defaultIdx]))
		pathEdited = false
		renderTranslation(formats[defaultIdx])
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
		AddItem(title, titleH, 0, false).
		AddItem(hint, 1, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(footer, 4, 0, false)

	page.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if _, isInput := s.app.GetFocus().(*tview.InputField); isInput {
			// Let inputs handle their own keys; only Esc closes the page.
			if ev.Key() == tcell.KeyEsc {
				s.popPage()
				return nil
			}
			return ev
		}
		switch ev.Key() {
		case tcell.KeyTab:
			if s.app.GetFocus() == saveAs {
				s.app.SetFocus(formatList)
			} else {
				s.app.SetFocus(saveAs)
			}
			return nil
		case tcell.KeyEsc, tcell.KeyLeft:
			s.popPage()
			return nil
		}
		switch ev.Rune() {
		case 'q', 'Q':
			s.app.Stop()
			return nil
		}
		return ev
	})

	s.pushPage(pageTranslate, page)
	s.app.SetFocus(formatList)
}

func loadJSONSchema(path string) (*jsonschema.Schema, error) {
	s, err := opendpi.LoadSchema(path)
	if err != nil {
		return nil, err
	}
	return &s.Schema, nil
}