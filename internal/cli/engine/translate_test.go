// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemasTranslate_Happy_Pydantic(t *testing.T) {
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
	out, err := SchemasTranslate(context.Background(), SchemasTranslateInput{
		Prj: prj, Schema: "user", Format: "pydantic", OutputDir: filepath.Join(prjDir, "models"),
	})
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(prjDir, "models", "user.py"), out.OutputFile)
	body, err := readFileString(out.OutputFile)
	require.NoError(t, err)
	assert.Contains(t, body, "class User")
	assert.Contains(t, body, "id:")
	assert.Contains(t, body, "email:")
}

func TestSchemasTranslate_Happy_Gotypes(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "schemas/user.yaml"), `type: object
title: User
properties:
  id: {type: integer}
  name: {type: string}
`)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	out, err := SchemasTranslate(context.Background(), SchemasTranslateInput{
		Prj: prj, Schema: "user", Format: "gotypes", OutputDir: filepath.Join(prjDir, "gen"),
	})
	require.NoError(t, err)
	body, err := readFileString(out.OutputFile)
	require.NoError(t, err)
	assert.Contains(t, body, "type User")
	assert.Contains(t, body, "ID")
}

func TestSchemasTranslate_UnknownFormat(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := SchemasTranslate(context.Background(), SchemasTranslateInput{
		Prj: prj, Schema: "user", Format: "nope", OutputDir: filepath.Join(prjDir, "out"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope")
	assert.Contains(t, err.Error(), "pydantic")
}

func TestSchemasTranslate_MissingSchema(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	_, err := SchemasTranslate(context.Background(), SchemasTranslateInput{
		Prj: prj, Schema: "nope", Format: "pydantic", OutputDir: filepath.Join(prjDir, "out"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema \"nope\" not found")
}

func TestSchemasTranslate_Validation(t *testing.T) {
	_, err := SchemasTranslate(context.Background(), SchemasTranslateInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"format: required", "output-dir: required", "prj: project not loaded", "schema: required"},
		errs,
	)
}

func TestPortsTranslate_InlineSchema(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	docPath := filepath.Join(prjDir, "products/orders.yaml")
	writeFile(t, docPath, `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections: {}
ports:
  daily:
    description: Daily
    schema:
      type: object
      title: Daily
      properties:
        amount: {type: number}
        currency: {type: string}
`)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	out, err := PortsTranslate(context.Background(), PortsTranslateInput{
		Prj: prj, Product: "orders", Port: "daily", Format: "pydantic", OutputDir: filepath.Join(prjDir, "out"),
	})
	require.NoError(t, err)
	body, err := readFileString(out.OutputFile)
	require.NoError(t, err)
	assert.Contains(t, body, "amount")
	assert.Contains(t, body, "currency")
}

func TestPortsTranslate_RefSchema(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	writeFile(t, filepath.Join(prjDir, "schemas/user.yaml"), `type: object
title: User
properties:
  id: {type: integer}
  email: {type: string}
`)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.NoError(t, err)

	out, err := PortsTranslate(context.Background(), PortsTranslateInput{
		Prj: prj, Product: "orders", Port: "daily", Format: "pydantic", OutputDir: filepath.Join(prjDir, "out"),
	})
	require.NoError(t, err)
	body, err := readFileString(out.OutputFile)
	require.NoError(t, err)
	assert.Contains(t, body, "id")
	assert.Contains(t, body, "email")
}

func TestPortsTranslate_NoSchemaErrors(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	_, err := PortsTranslate(context.Background(), PortsTranslateInput{
		Prj: prj, Product: "orders", Port: "daily", Format: "pydantic", OutputDir: filepath.Join(prjDir, "out"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no schema")
}

func TestPortsTranslate_MissingPort(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := PortsTranslate(context.Background(), PortsTranslateInput{
		Prj: prj, Product: "orders", Port: "nope", Format: "pydantic", OutputDir: filepath.Join(prjDir, "out"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port \"nope\" not found")
}

func TestPortsTranslate_Validation(t *testing.T) {
	_, err := PortsTranslate(context.Background(), PortsTranslateInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	missing := errs.Error()
	for _, m := range []string{"product: required", "port: required", "format: required", "output-dir: required", "prj: project not loaded"} {
		assert.True(t, strings.Contains(missing, m), "missing %q in %q", m, missing)
	}
}