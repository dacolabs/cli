// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemasCreate_Happy(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	out, err := SchemasCreate(context.Background(), SchemasCreateInput{
		Prj: prj, Name: "user", Path: "schemas/user.yaml",
		Type: "object", Title: "User", Description: "a user",
	})
	require.NoError(t, err)
	assert.True(t, out.Scaffolded)
	assert.FileExists(t, filepath.Join(prjDir, "schemas/user.yaml"))
	body, err := readFileString(filepath.Join(prjDir, "schemas/user.yaml"))
	require.NoError(t, err)
	assert.Contains(t, body, "type: object")
	assert.Contains(t, body, "title: User")
}

func TestSchemasCreate_DefaultsToObject(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	_, err := SchemasCreate(context.Background(), SchemasCreateInput{
		Prj: prj, Name: "x", Path: "schemas/x.yaml",
	})
	require.NoError(t, err)
	body, err := readFileString(filepath.Join(prjDir, "schemas/x.yaml"))
	require.NoError(t, err)
	assert.Contains(t, body, "type: object")
}

func TestSchemasCreate_PreservesExistingFile(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "schemas/user.yaml"), "hand-written\n")
	out, err := SchemasCreate(context.Background(), SchemasCreateInput{
		Prj: prj, Name: "user", Path: "schemas/user.yaml",
	})
	require.NoError(t, err)
	assert.False(t, out.Scaffolded)
	body, err := readFileString(filepath.Join(prjDir, "schemas/user.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "hand-written\n", body)
}

func TestSchemasCreate_Duplicate(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := SchemasCreate(context.Background(), SchemasCreateInput{
		Prj: prj, Name: "user", Path: "schemas/dup.yaml",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSchemasCreate_Validation(t *testing.T) {
	_, err := SchemasCreate(context.Background(), SchemasCreateInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"name: required", "path: required", "prj: project not loaded"},
		errs,
	)
}

func TestSchemasList_SortedAndEmpty(t *testing.T) {
	_, prj, _ := newTestProject(t)
	out, err := SchemasList(context.Background(), SchemasListInput{Prj: prj})
	require.NoError(t, err)
	assert.Empty(t, out.Schemas)

	mustCreateSchema(t, prj, "zebra", "schemas/zebra.yaml")
	mustCreateSchema(t, prj, "apple", "schemas/apple.yaml")
	out, err = SchemasList(context.Background(), SchemasListInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Schemas, 2)
	assert.Equal(t, "apple", out.Schemas[0].Name)
	assert.Equal(t, "zebra", out.Schemas[1].Name)
}

func TestSchemasList_Validation(t *testing.T) {
	_, err := SchemasList(context.Background(), SchemasListInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prj: project not loaded")
}

func TestSchemasDescribe_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "schemas/user.yaml"), `type: object
title: User
required: [id, email]
properties:
  id: {type: integer}
  email: {type: string}
  name: {type: string}
`)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	out, err := SchemasDescribe(context.Background(), SchemasDescribeInput{Prj: prj, Name: "user"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	require.NotNil(t, out.Schema)
	assert.Equal(t, "User", out.Schema.Title)
	assert.Equal(t, 3, len(out.Schema.Properties))
	assert.Equal(t, 2, len(out.Schema.Required))
}

func TestSchemasDescribe_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := SchemasDescribe(context.Background(), SchemasDescribeInput{Prj: prj, Name: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSchemasDescribe_RefOnlyFile(t *testing.T) {
	// Schema file is just `$ref: ...`. Current behavior: load succeeds,
	// Schema.Ref is populated, Type/Title empty, counts zero.
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "schemas/user.yaml"), "$ref: ../user-real.yaml\n")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	out, err := SchemasDescribe(context.Background(), SchemasDescribeInput{Prj: prj, Name: "user"})
	require.NoError(t, err)
	require.NotNil(t, out.Schema)
	assert.Equal(t, "../user-real.yaml", out.Schema.Ref)
	assert.Empty(t, out.Schema.Type)
	assert.Empty(t, out.Schema.Title)
	assert.Empty(t, out.Schema.Properties)
	assert.Empty(t, out.Schema.Required)
}

func TestSchemasDescribe_MalformedYAMLTolerant(t *testing.T) {
	// describe tolerates parse failures — returns Exists=true with no Schema.
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "schemas/bad.yaml"), ":\n  -\n -")
	mustCreateSchema(t, prj, "bad", "schemas/bad.yaml")
	out, err := SchemasDescribe(context.Background(), SchemasDescribeInput{Prj: prj, Name: "bad"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	assert.Nil(t, out.Schema)
}

func TestSchemasDelete_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	out, err := SchemasDelete(context.Background(), SchemasDeleteInput{Prj: prj, Name: "user"})
	require.NoError(t, err)
	assert.Equal(t, "user", out.Name)
	assert.FileExists(t, filepath.Join(prjDir, "schemas/user.yaml"))
	assert.NotContains(t, prj.Schemas, "user")
}

func TestSchemasDelete_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := SchemasDelete(context.Background(), SchemasDeleteInput{Prj: prj, Name: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}