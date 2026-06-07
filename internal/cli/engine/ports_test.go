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

	"github.com/dacolabs/daco/internal/opendpi"
)

func TestPortsCreate_Happy(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	out, err := PortsCreate(context.Background(), PortsCreateInput{
		Prj: prj, Product: "orders", Name: "daily", Description: "Daily metrics",
	})
	require.NoError(t, err)
	assert.Equal(t, "daily", out.Name)
	assert.Equal(t, "orders", out.Product)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	require.Contains(t, doc.Ports, "daily")
	assert.Equal(t, "Daily metrics", doc.Ports["daily"].Description)
}

func TestPortsCreate_Duplicate(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	_, err := PortsCreate(context.Background(), PortsCreateInput{
		Prj: prj, Product: "orders", Name: "daily",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPortsCreate_MissingProduct(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := PortsCreate(context.Background(), PortsCreateInput{
		Prj: prj, Product: "nope", Name: "daily",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product \"nope\" not found")
}

func TestPortsCreate_Validation(t *testing.T) {
	_, err := PortsCreate(context.Background(), PortsCreateInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"name: required", "prj: project not loaded", "product: required"},
		errs,
	)
}

func TestPortsList_SortedAndEmpty(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	out, err := PortsList(context.Background(), PortsListInput{Prj: prj, Product: "orders"})
	require.NoError(t, err)
	assert.Empty(t, out.Ports)

	mustCreatePort(t, prj, "orders", "zebra")
	mustCreatePort(t, prj, "orders", "apple")
	out, err = PortsList(context.Background(), PortsListInput{Prj: prj, Product: "orders"})
	require.NoError(t, err)
	require.Len(t, out.Ports, 2)
	assert.Equal(t, "apple", out.Ports[0].Name)
	assert.Equal(t, "zebra", out.Ports[1].Name)
}

func TestPortsList_Validation(t *testing.T) {
	_, err := PortsList(context.Background(), PortsListInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Contains(t, errs.Error(), "product: required")
	assert.Contains(t, errs.Error(), "prj: project not loaded")
}

func TestPortsDescribe_NoSchema(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	out, err := PortsDescribe(context.Background(), PortsDescribeInput{
		Prj: prj, Product: "orders", Name: "daily",
	})
	require.NoError(t, err)
	assert.False(t, out.HasSchema)
	assert.Empty(t, out.SchemaRef)
	assert.Empty(t, out.SchemaType)
	assert.Equal(t, 0, out.SchemaProperties)
}

func TestPortsDescribe_RefSchema(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	writeFile(t, filepath.Join(prjDir, "schemas/user.yaml"), `type: object
title: User
required: [id]
properties:
  id: {type: integer}
  name: {type: string}
`)
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.NoError(t, err)

	out, err := PortsDescribe(context.Background(), PortsDescribeInput{
		Prj: prj, Product: "orders", Name: "daily",
	})
	require.NoError(t, err)
	assert.True(t, out.HasSchema)
	assert.Equal(t, "../schemas/user.yaml", out.SchemaRef)
	// Describe follows the ref and surfaces the resolved schema's info.
	assert.Equal(t, "object", out.SchemaType)
	assert.Equal(t, "User", out.SchemaTitle)
	assert.Equal(t, 2, out.SchemaProperties)
	assert.Equal(t, 1, out.SchemaRequired)
}

func TestPortsDescribe_InlineSchema(t *testing.T) {
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
      title: Inline
      required: [a]
      properties:
        a: {type: string}
        b: {type: number}
`)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	out, err := PortsDescribe(context.Background(), PortsDescribeInput{
		Prj: prj, Product: "orders", Name: "daily",
	})
	require.NoError(t, err)
	assert.True(t, out.HasSchema)
	assert.Empty(t, out.SchemaRef)
	assert.Equal(t, "object", out.SchemaType)
	assert.Equal(t, "Inline", out.SchemaTitle)
	assert.Equal(t, 2, out.SchemaProperties)
	assert.Equal(t, 1, out.SchemaRequired)
}

func TestPortsDescribe_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := PortsDescribe(context.Background(), PortsDescribeInput{
		Prj: prj, Product: "orders", Name: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port \"nope\" not found")
}

func TestPortsDescribe_Validation(t *testing.T) {
	_, err := PortsDescribe(context.Background(), PortsDescribeInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Contains(t, errs.Error(), "product: required")
	assert.Contains(t, errs.Error(), "name: required")
}

func TestPortsDelete_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreatePort(t, prj, "orders", "events")
	out, err := PortsDelete(context.Background(), PortsDeleteInput{
		Prj: prj, Product: "orders", Name: "events",
	})
	require.NoError(t, err)
	assert.Equal(t, "events", out.Name)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, doc.Ports, "events")
	assert.Contains(t, doc.Ports, "daily")
}

func TestPortsDelete_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := PortsDelete(context.Background(), PortsDeleteInput{
		Prj: prj, Product: "orders", Name: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port \"nope\" not found")
}

func TestPortsLinkSchema_Happy(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	out, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.NoError(t, err)
	assert.Equal(t, "../schemas/user.yaml", out.Ref)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	port := doc.Ports["daily"]
	require.NotNil(t, port.Schema)
	assert.Equal(t, "../schemas/user.yaml", port.Schema.Ref)
	assert.Empty(t, port.Schema.Type)
	assert.Empty(t, port.Schema.Properties)
}

func TestPortsLinkSchema_RejectsExistingRef(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.NoError(t, err)

	_, err = PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has a schema")
}

func TestPortsLinkSchema_RejectsInlineSchema(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	docPath := filepath.Join(prjDir, "products/orders.yaml")
	writeFile(t, docPath, `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections: {}
ports:
  daily:
    schema:
      type: object
      properties:
        a: {type: string}
`)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has a schema")
}

func TestPortsLinkSchema_MissingSchema(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema \"nope\" not found")
}

func TestPortsLinkSchema_MissingPort(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "nope", Schema: "user",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port \"nope\" not found")
}

func TestPortsUnlinkSchema_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateSchema(t, prj, "user", "schemas/user.yaml")
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily", Schema: "user",
	})
	require.NoError(t, err)

	out, err := PortsUnlinkSchema(context.Background(), PortsUnlinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily",
	})
	require.NoError(t, err)
	assert.Equal(t, "daily", out.Port)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	port := doc.Ports["daily"]
	assert.Nil(t, port.Schema)
}

func TestPortsUnlinkSchema_NoSchemaErrors(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	_, err := PortsUnlinkSchema(context.Background(), PortsUnlinkSchemaInput{
		Prj: prj, Product: "orders", Port: "daily",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no schema to unlink")
}

func TestPortsBindConnection_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "warehouse",
	})
	require.NoError(t, err)

	out, err := PortsBindConnection(context.Background(), PortsBindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily",
		Connection: "warehouse", Location: "analytics.daily_orders",
	})
	require.NoError(t, err)
	assert.Equal(t, "warehouse", out.Connection)
	assert.Equal(t, "analytics.daily_orders", out.Location)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	port := doc.Ports["daily"]
	require.Len(t, port.Connections, 1)
	assert.Equal(t, "warehouse", port.Connections[0].Connection)
	assert.Equal(t, "analytics.daily_orders", port.Connections[0].Location)
}

func TestPortsBindConnection_ConnectionNotLinkedToProduct(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	// We did NOT call products link — connection isn't in the product's connections map.

	_, err := PortsBindConnection(context.Background(), PortsBindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily",
		Connection: "warehouse", Location: "loc",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not linked to product")
}

func TestPortsBindConnection_Duplicate(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "warehouse",
	})
	require.NoError(t, err)
	_, err = PortsBindConnection(context.Background(), PortsBindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily",
		Connection: "warehouse", Location: "loc",
	})
	require.NoError(t, err)
	_, err = PortsBindConnection(context.Background(), PortsBindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily",
		Connection: "warehouse", Location: "loc",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already bound")
}

func TestPortsBindConnection_Validation(t *testing.T) {
	_, err := PortsBindConnection(context.Background(), PortsBindConnectionInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"connection: required", "location: required", "port: required", "prj: project not loaded", "product: required"},
		errs)
}

func TestPortsUnbindConnection_RemovesBindings(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "warehouse",
	})
	require.NoError(t, err)

	for _, loc := range []string{"loc_a", "loc_b"} {
		_, err = PortsBindConnection(context.Background(), PortsBindConnectionInput{
			Prj: prj, Product: "orders", Port: "daily",
			Connection: "warehouse", Location: loc,
		})
		require.NoError(t, err)
	}

	// Remove a specific location.
	out, err := PortsUnbindConnection(context.Background(), PortsUnbindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily",
		Connection: "warehouse", Location: "loc_a",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, out.Removed)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	require.Len(t, doc.Ports["daily"].Connections, 1)
	assert.Equal(t, "loc_b", doc.Ports["daily"].Connections[0].Location)

	// Remove without location → matches remaining one.
	out, err = PortsUnbindConnection(context.Background(), PortsUnbindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily", Connection: "warehouse",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, out.Removed)
}

func TestPortsUnbindConnection_NoMatch(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	_, err := PortsUnbindConnection(context.Background(), PortsUnbindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily", Connection: "warehouse",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no binding")
}

func TestPortsDescribe_SurfacesBindings(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreatePort(t, prj, "orders", "daily")
	mustCreateConnection(t, prj, "warehouse", "connections/warehouse.yaml", "postgresql", "h:5432")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "warehouse",
	})
	require.NoError(t, err)
	_, err = PortsBindConnection(context.Background(), PortsBindConnectionInput{
		Prj: prj, Product: "orders", Port: "daily",
		Connection: "warehouse", Location: "daily_orders",
	})
	require.NoError(t, err)

	out, err := PortsDescribe(context.Background(), PortsDescribeInput{
		Prj: prj, Product: "orders", Name: "daily",
	})
	require.NoError(t, err)
	require.Len(t, out.Connections, 1)
	assert.Equal(t, "warehouse", out.Connections[0].Connection)
	assert.Equal(t, "daily_orders", out.Connections[0].Location)
}

func TestPortsLinkSchema_Validation(t *testing.T) {
	_, err := PortsLinkSchema(context.Background(), PortsLinkSchemaInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"port: required", "prj: project not loaded", "product: required", "schema: required"},
		errs,
	)
}