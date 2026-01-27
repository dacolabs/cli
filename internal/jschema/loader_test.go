// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package jschema

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile_YAML(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("simple.yaml")
	require.NoError(t, err)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "age")
}

func TestLoadFile_JSON(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("simple.json")
	require.NoError(t, err)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "age")
}

func TestLoadFile_NotFound(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	_, err := loader.LoadFile("nonexistent.yaml")
	require.Error(t, err)
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	fsys := fstest.MapFS{
		"invalid.yaml": &fstest.MapFile{Data: []byte("{{invalid yaml")},
	}
	loader := NewLoader(fsys)
	_, err := loader.LoadFile("invalid.yaml")
	require.Error(t, err)
}

func TestLoadFile_InvalidJSON(t *testing.T) {
	fsys := fstest.MapFS{
		"invalid.json": &fstest.MapFile{Data: []byte("{invalid json}")},
	}
	loader := NewLoader(fsys)
	_, err := loader.LoadFile("invalid.json")
	require.Error(t, err)
}

func TestResolveRefs_SimpleFileRef(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("with-file-ref.yaml")
	require.NoError(t, err)

	// Before resolution, the data property has a $ref
	assert.Equal(t, "./external.yaml", schema.Properties["data"].Ref)

	// Resolve refs
	err = loader.ResolveRefs(schema, ".")
	require.NoError(t, err)

	// After resolution, the ref is replaced with the loaded schema
	dataProp := schema.Properties["data"]
	assert.Empty(t, dataProp.Ref)
	assert.Equal(t, "object", dataProp.Type)
	assert.Contains(t, dataProp.Properties, "id")
	assert.Contains(t, dataProp.Properties, "value")
}

func TestResolveRefs_NestedFileRefs(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("nested/main.yaml")
	require.NoError(t, err)

	err = loader.ResolveRefs(schema, "nested")
	require.NoError(t, err)

	parentProp := schema.Properties["parent"]
	assert.Empty(t, parentProp.Ref)
	assert.Equal(t, "object", parentProp.Type)
	assert.Contains(t, parentProp.Properties, "id")
}

func TestResolveRefs_DeepNestedPaths(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("nested/deep/deep.yaml")
	require.NoError(t, err)

	err = loader.ResolveRefs(schema, "nested/deep")
	require.NoError(t, err)

	rootProp := schema.Properties["root"]
	assert.Empty(t, rootProp.Ref)
	assert.Equal(t, "object", rootProp.Type)
	assert.Contains(t, rootProp.Properties, "id")
}

func TestResolveRefs_SkipsInternalRefs(t *testing.T) {
	loader := NewLoader(os.DirFS("testdata"))
	schema, err := loader.LoadFile("with-defs.yaml")
	require.NoError(t, err)

	err = loader.ResolveRefs(schema, ".")
	require.NoError(t, err)

	// Internal refs should be preserved
	assert.Equal(t, "#/$defs/address", schema.Properties["address"].Ref)
}

func TestResolveRefs_MissingFile(t *testing.T) {
	fsys := fstest.MapFS{
		"schema.yaml": &fstest.MapFile{Data: []byte(`
type: object
properties:
  missing:
    $ref: "./does-not-exist.yaml"
`)},
	}
	loader := NewLoader(fsys)
	schema, err := loader.LoadFile("schema.yaml")
	require.NoError(t, err)

	err = loader.ResolveRefs(schema, ".")
	require.Error(t, err)
}
