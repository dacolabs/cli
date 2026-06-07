// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package opendpi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Covers map[any]any branch in toJSONCompatible. yaml.v3 normally produces
// map[string]any but nested mappings can carry non-string keys depending on
// the source.
func TestToJSONCompatible_NonStringKey(t *testing.T) {
	// Build the structure directly because yaml.v3 doesn't easily produce
	// non-string keys.
	in := map[any]any{
		"valid": "x",
		123:     "skipped",
		"nested": map[any]any{
			"k": "v",
		},
	}
	out := toJSONCompatible(in)
	cast, ok := out.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "x", cast["valid"])
	_, present := cast["123"]
	assert.False(t, present)
	nested, ok := cast["nested"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "v", nested["k"])
}

func TestToJSONCompatible_Slice(t *testing.T) {
	out := toJSONCompatible([]any{"a", map[any]any{"k": "v"}})
	slice, ok := out.([]any)
	require.True(t, ok)
	require.Len(t, slice, 2)
	assert.Equal(t, "a", slice[0])
	nested, ok := slice[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "v", nested["k"])
}

func TestToJSONCompatible_Passthrough(t *testing.T) {
	assert.Equal(t, "literal", toJSONCompatible("literal"))
	assert.Equal(t, 42, toJSONCompatible(42))
}

// Hit the remaining branches of collectDefRefs (Items, AdditionalProperties,
// AllOf/AnyOf/OneOf, $defs walking).
func TestTraverseDefs_RefsAcrossAllContainers(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
$defs:
  Item: { type: string }
  Wrap:
    type: object
    properties:
      list:
        type: array
        items:
          $ref: '#/$defs/Item'
  WithAdditional:
    type: object
    additionalProperties:
      $ref: '#/$defs/Item'
  WithUnion:
    allOf:
      - $ref: '#/$defs/Item'
    anyOf:
      - $ref: '#/$defs/Item'
    oneOf:
      - $ref: '#/$defs/Item'
  WithDefRefViaDefs:
    $defs:
      nested: { $ref: '#/$defs/Item' }
`)
	s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
	order := TraverseDefs(&s.Schema)
	first := order[0].Name
	assert.Equal(t, "Item", first, "Item must come before everything referencing it")
}

// findChild on a non-mapping node returns nil.
func TestPopulatePropertyOrder_NonMappingNode(t *testing.T) {
	// Schema is `[]`, a sequence node — populatePropertyOrder should no-op.
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "seq.yaml"), "- a\n- b\n")
	_, err := LoadSchema(filepath.Join(dir, "seq.yaml"))
	// Parsing a list into a Schema struct yields an error (unmarshal).
	assert.Error(t, err)
}

// writeYAML failure: target path is inside a directory that exists as a file.
func TestWriteYAML_BlockedByExistingFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "block")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	err := writeYAML(filepath.Join(blocker, "nested.yaml"), &Connection{Type: "x", Host: "y"})
	assert.Error(t, err)
}

// LoadSchema with valid YAML that fails json.Unmarshal because the doc isn't a
// schema (e.g. a top-level list rather than mapping).
func TestLoadSchema_NonObjectRoot(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "list.yaml"), "- a\n- b\n")
	_, err := LoadSchema(filepath.Join(dir, "list.yaml"))
	assert.Error(t, err)
}

// SaveSchema target inside a blocker path.
func TestSaveSchema_BlockedByExistingFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "block")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o644))
	s := &Schema{}
	s.Type = "object"
	err := SaveSchema(filepath.Join(blocker, "out.yaml"), s)
	assert.Error(t, err)
}

func TestSchema_UnmarshalYAML_NestedItemsPopulatesOrder(t *testing.T) {
	// Exercises Schema.UnmarshalYAML via a document where the schema field is
	// embedded in a Port.
	dir := t.TempDir()
	docPath := filepath.Join(dir, "daco.yaml")
	writeTestYAML(t, docPath, `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections: {}
ports:
  daily:
    schema:
      type: object
      properties:
        zeta: { type: integer }
        alpha: { type: string }
`)
	doc, err := Load(docPath)
	require.NoError(t, err)
	require.Contains(t, doc.Ports, "daily")
	require.NotNil(t, doc.Ports["daily"].Schema)
	assert.Equal(t, []string{"zeta", "alpha"}, doc.Ports["daily"].Schema.PropertyOrder)
}