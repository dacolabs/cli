// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package translate

import "testing"

func ptr[T any](v T) *T { return &v }

func TestActiveConstraintsOrder(t *testing.T) {
	c := Constraints{
		Maximum:    ptr(150.0),
		Minimum:    ptr(0.0),
		MultipleOf: ptr(0.01),
		Pattern:    "^x",
		MaxLength:  ptr(10),
		Format:     "uuid",
	}
	got := ActiveConstraints(c)
	want := []string{"minimum", "maximum", "multipleOf", "maxLength", "pattern", "format"}
	if len(got) != len(want) {
		t.Fatalf("got %d constraints, want %d: %+v", len(got), len(want), got)
	}
	for i, kv := range got {
		if kv.Key != want[i] {
			t.Errorf("position %d: got key %q, want %q", i, kv.Key, want[i])
		}
	}
}

func TestActiveConstraintsEmpty(t *testing.T) {
	if got := ActiveConstraints(Constraints{}); got != nil {
		t.Errorf("expected nil for empty constraints, got %+v", got)
	}
}

func TestConstraintsJSON(t *testing.T) {
	c := Constraints{
		Minimum: ptr(0.0),
		Maximum: ptr(150.0),
		Pattern: `^a"b`,
		Enum:    []any{"a", "b"},
	}
	got := ConstraintsJSON(c)
	want := `{"minimum": 0, "maximum": 150, "pattern": "^a\"b", "enum": ["a","b"]}`
	if got != want {
		t.Errorf("ConstraintsJSON:\n got: %s\nwant: %s", got, want)
	}
	if ConstraintsJSON(Constraints{}) != "" {
		t.Error("expected empty string for no constraints")
	}
}

func TestConstraintsPyDict(t *testing.T) {
	c := Constraints{
		Minimum:    ptr(0.0),
		MultipleOf: ptr(0.01),
		Enum:       []any{"a", true, 2.0},
		MaxLength:  ptr(10),
	}
	got := ConstraintsPyDict(c)
	want := `{"minimum": 0, "multipleOf": 0.01, "maxLength": 10, "enum": ["a", True, 2]}`
	if got != want {
		t.Errorf("ConstraintsPyDict:\n got: %s\nwant: %s", got, want)
	}
	if ConstraintsPyDict(Constraints{}) != "" {
		t.Error("expected empty string for no constraints")
	}
}
