// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package avro

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dacolabs/jsonschema-go/jsonschema"
	"github.com/dacolabs/daco/internal/translate"
)

// Translator translates JSON schemas to Apache Avro schema definitions.
type Translator struct{}

// FileExtension returns the file extension for Avro schema files.
func (t *Translator) FileExtension() string {
	return ".avsc"
}

// avroRecord represents an Avro record schema.
type avroRecord struct {
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Fields    []avroField `json:"fields"`
}

// avroField represents a field within an Avro record.
type avroField struct {
	Name string `json:"name"`
	Type any    `json:"type"`
}

// avroArray represents an Avro array type.
type avroArray struct {
	Type  string `json:"type"`
	Items any    `json:"items"`
}

// avroLogicalType represents an Avro logical type.
type avroLogicalType struct {
	Type        string `json:"type"`
	LogicalType string `json:"logicalType"`
}

// avroDecimal represents an Avro decimal logical type with precision and scale.
type avroDecimal struct {
	Type        string `json:"type"`
	LogicalType string `json:"logicalType"`
	Precision   int    `json:"precision"`
	Scale       int    `json:"scale"`
}

// avroEnum represents an Avro enum type with a named symbol set.
type avroEnum struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Symbols []string `json:"symbols"`
}

// Translate converts a JSON schema to an Avro schema JSON document.
func (t *Translator) Translate(portName string, schema *jsonschema.Schema, _ string) ([]byte, error) {
	data, err := translate.Prepare(portName, schema, &resolver{})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema data: %w", err)
	}

	// Build a map of def name → TypeDef for inline-at-first-use
	defsByName := make(map[string]*translate.TypeDef, len(data.Defs))
	for i := range data.Defs {
		defsByName[data.Defs[i].Name] = &data.Defs[i]
	}

	inlined := make(map[string]bool)

	// Build the root record
	root := avroRecord{
		Type:      "record",
		Name:      data.Root.Name,
		Namespace: "schemas",
		Fields:    buildFields(data.Root.Fields, defsByName, inlined),
	}

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Avro schema: %w", err)
	}

	return append(out, '\n'), nil
}

// buildFields converts translate.Fields to avroFields, inlining defs at first use.
func buildFields(fields []translate.Field, defs map[string]*translate.TypeDef, inlined map[string]bool) []avroField {
	result := make([]avroField, 0, len(fields))
	for i := range fields {
		avroType := buildAvroType(fields[i].Type, defs, inlined)
		if fields[i].Nullable {
			avroType = []any{"null", avroType}
		}
		result = append(result, avroField{
			Name: fields[i].Name,
			Type: avroType,
		})
	}
	return result
}

// buildAvroType converts a resolver type string to an Avro type value.
func buildAvroType(typeStr string, defs map[string]*translate.TypeDef, inlined map[string]bool) any {
	// Handle ref markers
	if name, ok := strings.CutPrefix(typeStr, "ref:"); ok {
		if def, exists := defs[name]; exists && !inlined[name] {
			inlined[name] = true
			return avroRecord{
				Type:   "record",
				Name:   name,
				Fields: buildFields(def.Fields, defs, inlined),
			}
		}
		return name
	}

	// Handle array markers
	if elemStr, ok := strings.CutPrefix(typeStr, "array:"); ok {
		return avroArray{
			Type:  "array",
			Items: buildAvroType(elemStr, defs, inlined),
		}
	}

	// Handle enum markers (name and comma-separated symbols carried in the marker string)
	if rest, ok := strings.CutPrefix(typeStr, "enum:"); ok {
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) == 2 {
			return avroEnum{Type: "enum", Name: parts[0], Symbols: strings.Split(parts[1], ",")}
		}
	}

	// Handle decimal markers (precision/scale carried in the marker string)
	if rest, ok := strings.CutPrefix(typeStr, "decimal:"); ok {
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) == 2 {
			p, errP := strconv.Atoi(parts[0])
			s, errS := strconv.Atoi(parts[1])
			if errP == nil && errS == nil {
				return avroDecimal{Type: "bytes", LogicalType: "decimal", Precision: p, Scale: s}
			}
		}
	}

	// Handle logical types
	switch typeStr {
	case "date":
		return avroLogicalType{Type: "int", LogicalType: "date"}
	case "timestamp-millis":
		return avroLogicalType{Type: "long", LogicalType: "timestamp-millis"}
	case "uuid":
		return avroLogicalType{Type: "string", LogicalType: "uuid"}
	}

	// Primitive types pass through as strings
	return typeStr
}
