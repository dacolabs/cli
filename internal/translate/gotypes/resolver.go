// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package gotypes

import (
	"strings"
	"unicode"

	"github.com/dacolabs/cli/internal/translate"
)

type resolver struct{}

func (r *resolver) PrimitiveType(schemaType, format string) string {
	if format != "" {
		switch format {
		case "date", "date-time":
			return "time.Time"
		case "uuid":
			return "string"
		}
	}

	switch schemaType {
	case "string":
		return "string"
	case "integer":
		return "int64"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}

func (r *resolver) ArrayType(elemType string) string {
	return "[]" + elemType
}

func (r *resolver) RefType(defName string) string {
	return toPascalCase(defName)
}

func (r *resolver) FormatDefName(defName string) string {
	return toPascalCase(defName)
}

func (r *resolver) FormatRootName(portName string) string {
	return toPascalCase(portName) + "Schema"
}

func (r *resolver) EnrichField(f *translate.Field) {
	tag := f.Name
	if f.Nullable {
		tag += ",omitempty"
		f.Type = "*" + f.Type
	}
	f.Tag = "`json:\"" + tag + "\"`"
	f.Name = toPascalCase(f.Name)
}

// toPascalCase converts a snake_case or camelCase string to PascalCase.
// It handles common Go acronyms (ID, URL, HTTP, API, JSON, XML, SQL, HTML).
func toPascalCase(s string) string {
	// Common Go acronyms that should be fully uppercased.
	acronyms := map[string]string{
		"id":   "ID",
		"url":  "URL",
		"http": "HTTP",
		"api":  "API",
		"json": "JSON",
		"xml":  "XML",
		"sql":  "SQL",
		"html": "HTML",
		"ip":   "IP",
		"tcp":  "TCP",
		"udp":  "UDP",
		"tls":  "TLS",
		"ssl":  "SSL",
		"ssh":  "SSH",
		"cpu":  "CPU",
		"uri":  "URI",
	}

	parts := splitWords(s)

	var sb strings.Builder
	for _, part := range parts {
		lower := strings.ToLower(part)
		if acronym, ok := acronyms[lower]; ok {
			sb.WriteString(acronym)
		} else if part != "" {
			sb.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}

	return sb.String()
}

// splitWords splits a string into words by underscores, hyphens, and camelCase boundaries.
func splitWords(s string) []string {
	var parts []string
	var current strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '_' || r == '-' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}
		if unicode.IsUpper(r) && current.Len() > 0 {
			// lowercase→uppercase: start new word
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				parts = append(parts, current.String())
				current.Reset()
			} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				// uppercase run followed by lowercase: split before current (e.g. HTTPServer → HTTP, Server)
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
