// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package gotypes

import (
	"strings"

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

	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})

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
