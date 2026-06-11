// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// KV is a single active constraint: a JSON Schema keyword name and its value.
// Value is the raw Go value (float64 for numerics, int for counts, string for
// pattern/format, []any for enum, the dereferenced value for const).
type KV struct {
	Key   string
	Value any
}

// ActiveConstraints returns the constraints that are actually present on a field,
// in a fixed, deterministic order keyed by JSON Schema keyword name. This is the
// single source every translator's "preserve/validate" pass iterates so that output
// ordering is identical across targets and stable across runs.
func ActiveConstraints(c Constraints) []KV { //nolint:gocyclo // a flat keyword switch, intentionally linear
	var out []KV
	if c.Minimum != nil {
		out = append(out, KV{"minimum", *c.Minimum})
	}
	if c.Maximum != nil {
		out = append(out, KV{"maximum", *c.Maximum})
	}
	if c.ExclusiveMinimum != nil {
		out = append(out, KV{"exclusiveMinimum", *c.ExclusiveMinimum})
	}
	if c.ExclusiveMaximum != nil {
		out = append(out, KV{"exclusiveMaximum", *c.ExclusiveMaximum})
	}
	if c.MultipleOf != nil {
		out = append(out, KV{"multipleOf", *c.MultipleOf})
	}
	if c.MinLength != nil {
		out = append(out, KV{"minLength", *c.MinLength})
	}
	if c.MaxLength != nil {
		out = append(out, KV{"maxLength", *c.MaxLength})
	}
	if c.Pattern != "" {
		out = append(out, KV{"pattern", c.Pattern})
	}
	if len(c.Enum) > 0 {
		out = append(out, KV{"enum", c.Enum})
	}
	if c.Const != nil {
		out = append(out, KV{"const", *c.Const})
	}
	if c.MinItems != nil {
		out = append(out, KV{"minItems", *c.MinItems})
	}
	if c.MaxItems != nil {
		out = append(out, KV{"maxItems", *c.MaxItems})
	}
	if c.Format != "" {
		out = append(out, KV{"format", c.Format})
	}
	return out
}

// ConstraintEntriesJSON returns active constraints as rendered JSON `"key": value`
// entry strings, in canonical order, for composing into a JSON object (optionally with
// extra leading entries such as a comment). Used by Avro and Spark Scala.
func ConstraintEntriesJSON(c Constraints) []string {
	kvs := ActiveConstraints(c)
	out := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		v, err := json.Marshal(kv.Value)
		if err != nil {
			continue
		}
		key, _ := json.Marshal(kv.Key)
		out = append(out, fmt.Sprintf("%s: %s", key, v))
	}
	return out
}

// ConstraintEntriesPy returns active constraints as rendered Python `"key": value`
// entry strings, in canonical order, for composing into a dict literal (optionally with
// extra leading entries such as a comment). Used by PySpark.
func ConstraintEntriesPy(c Constraints) []string {
	kvs := ActiveConstraints(c)
	out := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		out = append(out, fmt.Sprintf("%q: %s", kv.Key, pyValue(kv.Value)))
	}
	return out
}

// ConstraintsJSON renders the active constraints as a compact JSON object, preserving
// ActiveConstraints order (standard library map marshalling would sort keys). Returns
// "" when there are no constraints. Used for Avro field custom properties and Spark
// Scala Metadata.fromJson(...).
func ConstraintsJSON(c Constraints) string {
	entries := ConstraintEntriesJSON(c)
	if len(entries) == 0 {
		return ""
	}
	return "{" + strings.Join(entries, ", ") + "}"
}

// ConstraintsPyDict renders the active constraints as a Python dict literal, preserving
// ActiveConstraints order. Returns "" when there are no constraints. Used for PySpark
// StructField metadata={...}.
func ConstraintsPyDict(c Constraints) string {
	entries := ConstraintEntriesPy(c)
	if len(entries) == 0 {
		return ""
	}
	return "{" + strings.Join(entries, ", ") + "}"
}

// pyValue renders a constraint value as a Python literal, recursing into enum lists.
func pyValue(v any) string {
	if list, ok := v.([]any); ok {
		items := make([]string, len(list))
		for i, e := range list {
			items[i] = FormatLiteral(e)
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	return FormatLiteral(v)
}

// ConstraintsText renders the active constraints as a compact, human-readable list,
// e.g. `minimum: 0, maximum: 150, pattern: ^x`. Returns "" when there are none. Used
// for comment channels (Spark SQL inline comments, protobuf leading comments).
func ConstraintsText(c Constraints) string {
	kvs := ActiveConstraints(c)
	if len(kvs) == 0 {
		return ""
	}
	parts := make([]string, len(kvs))
	for i, kv := range kvs {
		parts[i] = kv.Key + ": " + textValue(kv.Value)
	}
	return strings.Join(parts, ", ")
}

// textValue renders a constraint value as plain text, recursing into enum lists.
func textValue(v any) string {
	switch x := v.(type) {
	case []any:
		items := make([]string, len(x))
		for i, e := range x {
			items[i] = textValue(e)
		}
		return "[" + strings.Join(items, ", ") + "]"
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x))
		}
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
