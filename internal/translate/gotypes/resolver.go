// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package gotypes

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dacolabs/daco/internal/translate"
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

func (r *resolver) MapType(keyType, valueType string) string {
	return "map[" + keyType + "]" + valueType
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
	if f.Type == "int64" {
		f.Type = goIntType(translate.NarrowInteger(f.Constraints), translate.IsNonNegative(f.Constraints))
	}

	jsonTag := f.Name
	if f.Nullable {
		jsonTag += ",omitempty"
		f.Type = "*" + f.Type
	}
	tag := "`json:\"" + jsonTag + "\""
	if rules := goValidateRules(f); rules != "" {
		tag += " validate:\"" + rules + "\""
	}
	tag += "`"
	f.Tag = tag
	f.Name = toPascalCase(f.Name)
}

// goValidateRules maps constraints to a go-playground/validator tag body (e.g.
// `min=0,max=150,oneof=a b`). pattern and multipleOf have no validator builtin and are
// omitted. Optional (pointer) fields are prefixed with omitempty so nil passes.
func goValidateRules(f *translate.Field) string {
	c := f.Constraints
	var rules []string
	if c.Minimum != nil {
		rules = append(rules, "min="+goNum(*c.Minimum))
	}
	if c.Maximum != nil {
		rules = append(rules, "max="+goNum(*c.Maximum))
	}
	if c.ExclusiveMinimum != nil {
		rules = append(rules, "gt="+goNum(*c.ExclusiveMinimum))
	}
	if c.ExclusiveMaximum != nil {
		rules = append(rules, "lt="+goNum(*c.ExclusiveMaximum))
	}
	if c.MinLength != nil {
		rules = append(rules, fmt.Sprintf("min=%d", *c.MinLength))
	}
	if c.MaxLength != nil {
		rules = append(rules, fmt.Sprintf("max=%d", *c.MaxLength))
	}
	if c.MinItems != nil {
		rules = append(rules, fmt.Sprintf("min=%d", *c.MinItems))
	}
	if c.MaxItems != nil {
		rules = append(rules, fmt.Sprintf("max=%d", *c.MaxItems))
	}
	if len(c.Enum) > 0 {
		parts := make([]string, len(c.Enum))
		for i, v := range c.Enum {
			parts[i] = goOneofValue(v)
		}
		rules = append(rules, "oneof="+strings.Join(parts, " "))
	}
	if c.Const != nil {
		rules = append(rules, "eq="+goOneofValue(*c.Const))
	}
	if len(rules) == 0 {
		return ""
	}
	if f.Nullable {
		return "omitempty," + strings.Join(rules, ",")
	}
	return strings.Join(rules, ",")
}

// goNum renders a float bound as a Go-style numeric literal, dropping the trailing .0.
func goNum(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// goOneofValue renders an enum/const value for a validator oneof/eq rule.
func goOneofValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return goNum(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// goIntType picks the narrowest Go integer type for the constraint kind.
// nonNegative promotes signed kinds to their unsigned counterparts.
func goIntType(k translate.IntKind, nonNegative bool) string {
	if nonNegative {
		switch k {
		case translate.Int8:
			return "uint8"
		case translate.Int16:
			return "uint16"
		case translate.Int32:
			return "uint32"
		default:
			return "uint64"
		}
	}
	switch k {
	case translate.Int8:
		return "int8"
	case translate.Int16:
		return "int16"
	case translate.Int32:
		return "int32"
	default:
		return "int64"
	}
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
