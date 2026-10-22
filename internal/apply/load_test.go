package apply_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dacolabs/cli/internal/apply"
)

func TestLoadMergesMultiDocAndDedupesIdentical(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ds.yaml")
	content := `
kind: Dataset
urn: urn:daco:dataset:orders
version: "1.0.0"
title: Orders
description: Orders data
metadata: {}
contract:
  schema: {"type":"object"}
  metadata: {}
---
kind: Dataset
urn: urn:daco:dataset:orders
version: "1.0.0"
title: Orders
description: Orders data
metadata: {}
contract:
  schema: {"type":"object"}
  metadata: {}
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	units, err := apply.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 {
		t.Fatalf("got %d units, want 1", len(units))
	}
	if units[0].Input.Urn != "urn:daco:dataset:orders" || units[0].Input.Version != "1.0.0" {
		t.Fatalf("unexpected unit: %+v", units[0].Input)
	}
}

func TestLoadFailsOnConflictingDuplicates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.yaml")
	b := filepath.Join(dir, "b.yaml")
	doc := func(title string) string {
		return `
kind: Dataset
urn: urn:daco:dataset:orders
version: "1.0.0"
title: ` + title + `
description: Orders data
metadata: {}
contract:
  schema: {"type":"object"}
  metadata: {}
`
	}
	if err := os.WriteFile(a, []byte(doc("Orders")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(doc("Orders v2")), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := apply.Load([]string{a, b})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("error = %v, want conflict", err)
	}
}

func TestLoadRejectsUnknownKind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "x.yaml")
	if err := os.WriteFile(path, []byte("kind: Widget\nname: x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := apply.Load([]string{path})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadFromDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sub, "ds.yml")
	content := `
kind: Dataset
urn: urn:daco:dataset:inv
version: "0.1.0"
title: Inventory
description: Inventory
metadata: {}
contract:
  schema: {"type":"object"}
  metadata: {}
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	units, err := apply.Load([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Input.Urn != "urn:daco:dataset:inv" {
		t.Fatalf("got %+v", units)
	}
}
