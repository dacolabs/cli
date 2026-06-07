// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// titleBanner renders text as a centered yellow bold line padded by a blank
// line on each side (3 lines total). Returns the widget and its row height.
func titleBanner(text string) (*tview.TextView, int) {
	if strings.TrimSpace(text) == "" {
		text = "daco"
	}
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	tv.SetText("\n[yellow::b]" + tview.Escape(text) + "[-:-:-]\n")
	return tv, 3
}

// dacoLogo is the ASCII-art wordmark with the user's exact spacing. `█` are
// foreground blocks (yellow); `░` are shadow blocks (rendered dimmer for a
// 3D look). Trailing whitespace on each line is intentional — the lines are
// padded to a uniform width before rendering so AlignCenter keeps the columns
// in step.
const dacoLogo = `██████████        █████████        █████████        ███████
░░███░░░░███      ███░░░░░███      ███░░░░░███     ███░░░░░███
 ░███   ░░███    ░███    ░███     ███     ░░░     ███     ░░███
 ░███    ░███    ░███████████    ░███            ░███      ░███
 ░███    ░███    ░███░░░░░███    ░███            ░███      ░███
 ░███    ███     ░███    ░███    ░░███     ███   ░░███     ███
 ██████████      █████   █████    ░░█████████     ░░░███████░
░░░░░░░░░░      ░░░░░   ░░░░░      ░░░░░░░░░        ░░░░░░░    `

// dacoTagline is rendered under the wordmark as a subtle subtitle. The full-
// width Unicode glyphs read at roughly the same column-width as the logo.
const dacoTagline = "Ｌｏｃａｌ－ｆｉｒｓｔ，  ｂｙ  ｄｅｓｉｇｎ"

// dacoLogoBanner returns a centered colorized version of the wordmark with the
// tagline below it. Each line of the logo is right-padded to the max line
// width so AlignCenter keeps the columns aligned.
func dacoLogoBanner() (*tview.TextView, int) {
	colored := colorizeLogo(padLines(dacoLogo))
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	// One blank top + 8 logo rows + one blank gap + tagline + one blank bottom.
	tv.SetText("\n" + colored + "\n\n[#888888]" + dacoTagline + "[-]\n")
	return tv, 12
}

// padLines right-pads every line of in to the width of the widest line, so
// centered rendering keeps the columns in lockstep.
func padLines(in string) string {
	lines := strings.Split(in, "\n")
	max := 0
	for _, l := range lines {
		if w := runeCount(l); w > max {
			max = w
		}
	}
	for i, l := range lines {
		if pad := max - runeCount(l); pad > 0 {
			lines[i] = l + strings.Repeat(" ", pad)
		}
	}
	return strings.Join(lines, "\n")
}

func runeCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// colorizeLogo wraps consecutive runs of `█` in yellow and runs of `░` in
// dimmer gray. Spaces stay untagged. Runs are coalesced for compact output.
func colorizeLogo(in string) string {
	var b strings.Builder
	current := byte('n') // 'y' yellow, 'g' gray, 'n' none
	for _, ch := range in {
		var next byte
		switch ch {
		case '█':
			next = 'y'
		case '░':
			next = 'g'
		default:
			next = 'n'
		}
		if next != current {
			if current == 'y' || current == 'g' {
				b.WriteString("[-]")
			}
			switch next {
			case 'y':
				b.WriteString("[yellow]")
			case 'g':
				b.WriteString("[#666666]")
			}
			current = next
		}
		b.WriteRune(ch)
	}
	if current == 'y' || current == 'g' {
		b.WriteString("[-]")
	}
	return b.String()
}

// hintBar renders a centered gray hint line.
func hintBar(text string) *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]" + text + "[-]")
}

// sectionLabels are the in-shell section names, in their key-shortcut order.
var sectionLabels = []string{"Products", "Connections", "Schemas", "Ports"}

// sectionHint returns the footer hint string for a given section. Only the
// shortcuts that actually do something in that section are listed.
func sectionHint(section string) string {
	common := "1-4 sections  ·  enter detail  ·  / search  ·  p switch project  ·  q quit"
	switch section {
	case "Products":
		return "n new  ·  d delete  ·  l link connection  ·  " + common
	case "Connections":
		return "n new  ·  d delete  ·  " + common
	case "Schemas":
		return "n new  ·  d delete  ·  t translate  ·  " + common
	case "Ports":
		return "n new  ·  d delete  ·  l link schema  ·  t translate  ·  " + common
	}
	return common
}

// sectionBar renders a row like:
//
//	[1 Products]  [2 Connections]  [3 Schemas]  [4 Ports]
//
// with the current section highlighted.
func sectionBar(current string) *tview.TextView {
	parts := make([]string, 0, len(sectionLabels))
	for i, name := range sectionLabels {
		raw := fmt.Sprintf("%d %s", i+1, name)
		if name == current {
			parts = append(parts, "[yellow::b]["+tview.Escape(raw)+"][-:-:-]")
		} else {
			parts = append(parts, "[gray] "+tview.Escape(raw)+" [-]")
		}
	}
	return tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(strings.Join(parts, "   "))
}

// chromaStyle / chromaFormatter are resolved once. terminal256 emits ANSI;
// tview.TranslateANSI converts those into color tags so DynamicColors render.
var chromaStyle = func() *chroma.Style {
	if s := styles.Get("monokai"); s != nil {
		return s
	}
	return styles.Fallback
}()

var chromaFormatter = func() chroma.Formatter {
	if f := formatters.Get("terminal256"); f != nil {
		return f
	}
	return formatters.Fallback
}()

// highlight returns syntax-highlighted source ready for a tview TextView with
// SetDynamicColors(true). Falls back to escaped plain text on any error.
func highlight(source, lang string) string {
	if lang == "" {
		return tview.Escape(source)
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)
	it, err := lexer.Tokenise(nil, source)
	if err != nil {
		return tview.Escape(source)
	}
	var buf bytes.Buffer
	if err := chromaFormatter.Format(&buf, chromaStyle, it); err != nil {
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

// captureSetter is any tview primitive that exposes SetInputCapture. Every
// Box-derived widget satisfies it.
type captureSetter interface {
	SetInputCapture(func(*tcell.EventKey) *tcell.EventKey) *tview.Box
}

// installShortcuts puts the same shortcuts handler on every passed primitive.
// It's the workaround for tview not firing a parent Flex's input-capture when
// focus is on a leaf child: each focusable child gets the same handler so the
// page-level shortcuts (Esc back, q quit, u/l/c/t actions) work regardless of
// which sub-widget currently has focus.
func installShortcuts(handler func(*tcell.EventKey) *tcell.EventKey, primitives ...captureSetter) {
	for _, p := range primitives {
		p.SetInputCapture(handler)
	}
}

// installShortcutsWithFallback wraps `primary` so its existing behavior runs
// AFTER the shortcuts have a chance to handle the key.
func installShortcutsWithFallback(handler func(*tcell.EventKey) *tcell.EventKey, p captureSetter, primaryCapture func(*tcell.EventKey) *tcell.EventKey) {
	p.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if r := handler(ev); r == nil {
			return nil
		}
		if primaryCapture != nil {
			return primaryCapture(ev)
		}
		return ev
	})
}

// cycleFocus moves focus to the next primitive in `focusables`, wrapping at
// the end. Used by detail pages to let Tab cycle through their children.
func cycleFocus(s *state, focusables []tview.Primitive) {
	cur := s.app.GetFocus()
	for i, p := range focusables {
		if p == cur {
			s.app.SetFocus(focusables[(i+1)%len(focusables)])
			return
		}
	}
	if len(focusables) > 0 {
		s.app.SetFocus(focusables[0])
	}
}

// shellShortcut is the shared handler applied to every section list view's
// table so the shell-level keys (1-4 sections, p switch project, q quit)
// keep working even when the focus is on a deeply-nested table.
func shellShortcut(s *state) func(ev *tcell.EventKey) *tcell.EventKey {
	return func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyCtrlC:
			s.app.Stop()
			return nil
		}
		switch ev.Rune() {
		case 'q', 'Q':
			s.app.Stop()
			return nil
		case 'p', 'P':
			s.switchToProjects()
			return nil
		case '1':
			if s.shellSwitchSection != nil {
				s.shellSwitchSection("Products")
			}
			return nil
		case '2':
			if s.shellSwitchSection != nil {
				s.shellSwitchSection("Connections")
			}
			return nil
		case '3':
			if s.shellSwitchSection != nil {
				s.shellSwitchSection("Schemas")
			}
			return nil
		case '4':
			if s.shellSwitchSection != nil {
				s.shellSwitchSection("Ports")
			}
			return nil
		}
		return ev
	}
}