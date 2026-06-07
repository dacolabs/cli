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

func TestSchema_PropertyOrder_Root(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
type: object
properties:
  zeta:
    type: string
  alpha:
    type: integer
  beta:
    type: boolean
`)
	s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
	assert.Equal(t, []string{"zeta", "alpha", "beta"}, s.PropertyOrder)
}

func TestSchema_PropertyOrder_Nested(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
type: object
properties:
  user:
    type: object
    properties:
      id:
        type: integer
      email:
        type: string
      name:
        type: string
  meta:
    type: object
    properties:
      created_at:
        type: string
      updated_at:
        type: string
`)
	s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
	assert.Equal(t, []string{"user", "meta"}, s.PropertyOrder)
	require.NotNil(t, s.Properties["user"])
	assert.Equal(t, []string{"id", "email", "name"}, s.Properties["user"].PropertyOrder)
	require.NotNil(t, s.Properties["meta"])
	assert.Equal(t, []string{"created_at", "updated_at"}, s.Properties["meta"].PropertyOrder)
}

func TestSchema_PropertyOrder_DefsAndItems(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
type: object
properties:
  list:
    type: array
    items:
      type: object
      properties:
        b: { type: string }
        a: { type: string }
$defs:
  Order:
    type: object
    properties:
      total: { type: number }
      currency: { type: string }
`)
	s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
	require.NotNil(t, s.Properties["list"])
	require.NotNil(t, s.Properties["list"].Items)
	assert.Equal(t, []string{"b", "a"}, s.Properties["list"].Items.PropertyOrder)
	require.NotNil(t, s.Defs["Order"])
	assert.Equal(t, []string{"total", "currency"}, s.Defs["Order"].PropertyOrder)
}

func TestSchema_PropertyOrder_AllOfAnyOfOneOf(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
allOf:
  - type: object
    properties:
      a: { type: string }
      b: { type: string }
anyOf:
  - type: object
    properties:
      c: { type: string }
      d: { type: string }
oneOf:
  - type: object
    properties:
      e: { type: string }
      f: { type: string }
`)
	s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
	require.Len(t, s.AllOf, 1)
	require.Len(t, s.AnyOf, 1)
	require.Len(t, s.OneOf, 1)
	assert.Equal(t, []string{"a", "b"}, s.AllOf[0].PropertyOrder)
	assert.Equal(t, []string{"c", "d"}, s.AnyOf[0].PropertyOrder)
	assert.Equal(t, []string{"e", "f"}, s.OneOf[0].PropertyOrder)
}

func TestSchema_PropertyOrder_AdditionalProperties(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
type: object
additionalProperties:
  type: object
  properties:
    z: { type: string }
    a: { type: string }
`)
	s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
	require.NotNil(t, s.AdditionalProperties)
	assert.Equal(t, []string{"z", "a"}, s.AdditionalProperties.PropertyOrder)
}

func TestLoadSchema_RefOnly(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "ref.yaml"), `$ref: ../other.yaml
`)
	s := mustLoadSchema(t, filepath.Join(dir, "ref.yaml"))
	assert.Equal(t, "../other.yaml", s.Ref)
	assert.Empty(t, s.PropertyOrder)
	assert.Empty(t, s.Type)
}

func TestLoadSchema_FileMissing(t *testing.T) {
	_, err := LoadSchema(filepath.Join(t.TempDir(), "nope.yaml"))
	assert.Error(t, err)
}

func TestLoadSchema_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "bad.yaml"), ":\n  -\n -")
	_, err := LoadSchema(filepath.Join(dir, "bad.yaml"))
	assert.Error(t, err)
}

func TestSaveSchema_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.yaml")
	writeTestYAML(t, src, `
type: object
title: User
required: [id, email]
properties:
  id: { type: integer }
  email: { type: string }
  name: { type: string }
`)
	loaded := mustLoadSchema(t, src)
	out := filepath.Join(dir, "out.yaml")
	require.NoError(t, SaveSchema(out, loaded))
	reloaded := mustLoadSchema(t, out)
	assert.Equal(t, loaded.Title, reloaded.Title)
	assert.Equal(t, loaded.Type, reloaded.Type)
	assert.ElementsMatch(t, loaded.Required, reloaded.Required)
	assert.Equal(t, len(loaded.Properties), len(reloaded.Properties))
}

func TestSchemaTypeString(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, "", SchemaTypeString(nil))
	})
	t.Run("single type", func(t *testing.T) {
		s := &Schema{}
		s.Type = "object"
		assert.Equal(t, "object", SchemaTypeString(s))
	})
	t.Run("types array", func(t *testing.T) {
		s := &Schema{}
		s.Types = []string{"string", "null"}
		assert.Equal(t, "string|null", SchemaTypeString(s))
	})
	t.Run("empty", func(t *testing.T) {
		s := &Schema{}
		assert.Equal(t, "", SchemaTypeString(s))
	})
}

func TestTraverseDefs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := &Schema{}
		assert.Empty(t, TraverseDefs(&s.Schema))
		assert.Empty(t, TraverseDefs(nil))
	})

	t.Run("linear A->B->C", func(t *testing.T) {
		dir := t.TempDir()
		writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
$defs:
  C: { type: string }
  B:
    type: object
    properties:
      c: { $ref: '#/$defs/C' }
  A:
    type: object
    properties:
      b: { $ref: '#/$defs/B' }
`)
		s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
		order := TraverseDefs(&s.Schema)
		names := []string{}
		for _, e := range order {
			names = append(names, e.Name)
		}
		assert.Equal(t, []string{"C", "B", "A"}, names)
	})

	t.Run("independent alphabetical", func(t *testing.T) {
		dir := t.TempDir()
		writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
$defs:
  Zebra: { type: string }
  Apple: { type: string }
  Mango: { type: string }
`)
		s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
		order := TraverseDefs(&s.Schema)
		names := []string{}
		for _, e := range order {
			names = append(names, e.Name)
		}
		assert.Equal(t, []string{"Apple", "Mango", "Zebra"}, names)
	})

	t.Run("branching A->B,A->C", func(t *testing.T) {
		dir := t.TempDir()
		writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
$defs:
  B: { type: string }
  C: { type: string }
  A:
    type: object
    properties:
      b: { $ref: '#/$defs/B' }
      c: { $ref: '#/$defs/C' }
`)
		s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
		order := TraverseDefs(&s.Schema)
		names := []string{}
		for _, e := range order {
			names = append(names, e.Name)
		}
		// A must come last; B and C come first in alphabetical order.
		assert.Equal(t, []string{"B", "C", "A"}, names)
	})

	t.Run("cycle A<->B alphabetical break", func(t *testing.T) {
		dir := t.TempDir()
		writeTestYAML(t, filepath.Join(dir, "s.yaml"), `
$defs:
  A:
    type: object
    properties:
      b: { $ref: '#/$defs/B' }
  B:
    type: object
    properties:
      a: { $ref: '#/$defs/A' }
`)
		s := mustLoadSchema(t, filepath.Join(dir, "s.yaml"))
		order := TraverseDefs(&s.Schema)
		names := []string{}
		for _, e := range order {
			names = append(names, e.Name)
		}
		assert.Equal(t, []string{"B", "A"}, names)
	})
}

func TestConnection_RoundTrip_Inline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db.yaml")
	c := Connection{
		Type:        "postgresql",
		Host:        "localhost:5432",
		Description: "primary",
		Variables:   map[string]any{"db": "x"},
	}
	require.NoError(t, ScaffoldConnection(path, c))
	got, err := LoadConnection(path)
	require.NoError(t, err)
	assert.Equal(t, c.Type, got.Type)
	assert.Equal(t, c.Host, got.Host)
	assert.Equal(t, c.Description, got.Description)
	assert.Equal(t, c.Variables, got.Variables)
	assert.Empty(t, got.Ref)
}

func TestConnection_RoundTrip_Ref(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ref.yaml")
	c := Connection{Ref: "../other.yaml"}
	require.NoError(t, ScaffoldConnection(path, c))
	got, err := LoadConnection(path)
	require.NoError(t, err)
	assert.Equal(t, "../other.yaml", got.Ref)
	assert.Empty(t, got.Type)
	assert.Empty(t, got.Host)
}

func TestConnection_Scaffold_EmptyVariablesOmitted(t *testing.T) {
	// ScaffoldConnection defaults nil Variables to an empty map, but yaml's
	// omitempty on Variables strips it back out on serialization. Behavior:
	// empty Variables doesn't survive a round trip — this test pins that.
	dir := t.TempDir()
	path := filepath.Join(dir, "db.yaml")
	require.NoError(t, ScaffoldConnection(path, Connection{Type: "x", Host: "y"}))
	got, err := LoadConnection(path)
	require.NoError(t, err)
	assert.Nil(t, got.Variables)
}

func TestConnection_IsRef(t *testing.T) {
	assert.True(t, Connection{Ref: "x"}.IsRef())
	assert.False(t, Connection{Type: "x"}.IsRef())
}

func TestConnection_Resolve_RefForm(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yaml")
	require.NoError(t, ScaffoldConnection(target, Connection{Type: "kafka", Host: "k:9092"}))
	c := Connection{Ref: "target.yaml"}
	got, err := c.Resolve(dir)
	require.NoError(t, err)
	assert.Equal(t, "kafka", got.Type)
	assert.Equal(t, "k:9092", got.Host)
}

func TestConnection_Resolve_RefAbsolute(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yaml")
	require.NoError(t, ScaffoldConnection(target, Connection{Type: "kafka", Host: "k:9092"}))
	c := Connection{Ref: target}
	got, err := c.Resolve("")
	require.NoError(t, err)
	assert.Equal(t, "kafka", got.Type)
}

func TestConnection_Resolve_Inline(t *testing.T) {
	c := Connection{Type: "postgresql", Host: "h"}
	got, err := c.Resolve("")
	require.NoError(t, err)
	assert.Equal(t, "postgresql", got.Type)
}

func TestConnection_Resolve_Empty(t *testing.T) {
	_, err := Connection{}.Resolve("")
	assert.Error(t, err)
}

func TestLoadConnection_FileMissing(t *testing.T) {
	_, err := LoadConnection(filepath.Join(t.TempDir(), "nope.yaml"))
	assert.Error(t, err)
}

func TestLoadConnection_Malformed(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "bad.yaml"), ":\n  -\n -")
	_, err := LoadConnection(filepath.Join(dir, "bad.yaml"))
	assert.Error(t, err)
}

func TestDocument_Scaffold(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daco.yaml")
	require.NoError(t, Scaffold(path, "acme"))
	doc, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "acme", doc.Info.Title)
	assert.Equal(t, "1.0.0", doc.Info.Version)
	assert.NotNil(t, doc.Connections)
	assert.NotNil(t, doc.Ports)
	assert.Empty(t, doc.Connections)
	assert.Empty(t, doc.Ports)

	data, err := readFile(t, path)
	require.NoError(t, err)
	// Connections + Ports must serialize as {} not null.
	assert.Contains(t, data, "connections: {}")
	assert.Contains(t, data, "ports: {}")
}

func TestDocument_LoadSave_MixedForms(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "daco.yaml")
	writeTestYAML(t, docPath, `
opendpi: "1.0.0"
info:
    title: acme
    version: "1.0.0"
connections:
    refconn:
        $ref: ../connections/other.yaml
    inline:
        type: postgresql
        host: localhost:5432
ports:
    p1:
        description: First port
        schema:
            $ref: ../schemas/user.yaml
    p2:
        description: Inline port
        schema:
            type: object
            properties:
                value:
                    type: number
                label:
                    type: string
components:
    schemas:
        User:
            type: object
`)
	doc, err := Load(docPath)
	require.NoError(t, err)
	require.Len(t, doc.Connections, 2)
	require.Len(t, doc.Ports, 2)
	require.NotNil(t, doc.Components)

	// Round-trip via Save preserves the structure.
	out := filepath.Join(dir, "out.yaml")
	require.NoError(t, Save(out, doc))
	reloaded, err := Load(out)
	require.NoError(t, err)
	assert.Len(t, reloaded.Connections, 2)
	assert.Len(t, reloaded.Ports, 2)
	assert.NotNil(t, reloaded.Components)

	// $ref form preserved.
	assert.Equal(t, "../connections/other.yaml", reloaded.Connections["refconn"].Ref)
	// Inline form preserved.
	assert.Equal(t, "postgresql", reloaded.Connections["inline"].Type)
	// Port $ref form preserved.
	require.NotNil(t, reloaded.Ports["p1"].Schema)
	assert.Equal(t, "../schemas/user.yaml", reloaded.Ports["p1"].Schema.Ref)
	// Port inline form preserved.
	require.NotNil(t, reloaded.Ports["p2"].Schema)
	assert.Equal(t, "object", reloaded.Ports["p2"].Schema.Type)
	assert.Contains(t, reloaded.Ports["p2"].Schema.Properties, "value")
	assert.Contains(t, reloaded.Ports["p2"].Schema.Properties, "label")
}

func TestLoad_FileMissing(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	assert.Error(t, err)
}

// TestSpecExamples_Parse walks every example under opendpi/v1/examples/
// (the vendored spec examples) and confirms they all parse cleanly. The
// $ref-connection example additionally verifies Connection.Resolve.
func TestSpecExamples_Parse(t *testing.T) {
	root := filepath.Join("..", "..", "opendpi", "v1", "examples")
	matches, err := filepath.Glob(filepath.Join(root, "*.yaml"))
	require.NoError(t, err)
	jsonMatches, err := filepath.Glob(filepath.Join(root, "*.json"))
	require.NoError(t, err)
	matches = append(matches, jsonMatches...)
	require.NotEmpty(t, matches, "no spec examples found")

	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			doc, err := Load(path)
			require.NoError(t, err)
			require.NotEmpty(t, doc.OpenDPI)

			for name, conn := range doc.Connections {
				if !conn.IsRef() {
					continue
				}
				resolved, err := conn.Resolve(filepath.Dir(path))
				require.NoErrorf(t, err, "resolve %s in %s", name, path)
				assert.NotEmpty(t, resolved.Type)
				assert.NotEmpty(t, resolved.Host)
			}
		})
	}
}

func TestLoad_Malformed(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, filepath.Join(dir, "bad.yaml"), ":\n  -\n -")
	_, err := Load(filepath.Join(dir, "bad.yaml"))
	assert.Error(t, err)
}

func TestSave_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deeper", "daco.yaml")
	require.NoError(t, Scaffold(path, "x"))
	_, err := Load(path)
	assert.NoError(t, err)
}

func readFile(t *testing.T, path string) (string, error) {
	t.Helper()
	b, err := os.ReadFile(path)
	return string(b), err
}