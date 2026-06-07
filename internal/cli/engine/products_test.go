// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dacolabs/daco/internal/opendpi"
)

func TestProductsCreate_ScaffoldsWhenAbsent(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	out, err := ProductsCreate(context.Background(), ProductsCreateInput{
		Prj: prj, Name: "orders", Path: "products/orders.yaml",
	})
	require.NoError(t, err)
	assert.True(t, out.Scaffolded)
	assert.FileExists(t, filepath.Join(prjDir, "products/orders.yaml"))
	assert.Equal(t, "products/orders.yaml", prj.Products["orders"])
}

func TestProductsCreate_PreservesExistingFile(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "products/orders.yaml"), "hand-written\n")
	out, err := ProductsCreate(context.Background(), ProductsCreateInput{
		Prj: prj, Name: "orders", Path: "products/orders.yaml",
	})
	require.NoError(t, err)
	assert.False(t, out.Scaffolded)
	body, err := readFileString(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "hand-written\n", body)
}

func TestProductsCreate_Duplicate(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := ProductsCreate(context.Background(), ProductsCreateInput{
		Prj: prj, Name: "orders", Path: "products/dup.yaml",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestProductsCreate_Validation(t *testing.T) {
	_, err := ProductsCreate(context.Background(), ProductsCreateInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.Equal(t, Errors{"name: required", "path: required", "prj: project not loaded"}, errs)
}

func TestProductsList_SortedAndEmpty(t *testing.T) {
	_, prj, _ := newTestProject(t)
	out, err := ProductsList(context.Background(), ProductsListInput{Prj: prj})
	require.NoError(t, err)
	assert.Empty(t, out.Products)

	mustCreateProduct(t, prj, "zebra", "products/zebra.yaml")
	mustCreateProduct(t, prj, "apple", "products/apple.yaml")
	mustCreateProduct(t, prj, "mango", "products/mango.yaml")

	out, err = ProductsList(context.Background(), ProductsListInput{Prj: prj})
	require.NoError(t, err)
	require.Len(t, out.Products, 3)
	assert.Equal(t, "apple", out.Products[0].Name)
	assert.Equal(t, "mango", out.Products[1].Name)
	assert.Equal(t, "zebra", out.Products[2].Name)
}

func TestProductsDescribe_HappyPath(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	out, err := ProductsDescribe(context.Background(), ProductsDescribeInput{Prj: prj, Name: "orders"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	require.NotNil(t, out.Info)
	assert.Equal(t, "orders", out.Info.Title)
	assert.Equal(t, "1.0.0", out.Info.Version)
}

func TestProductsDescribe_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := ProductsDescribe(context.Background(), ProductsDescribeInput{Prj: prj, Name: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProductsDescribe_NonOpenDPIFile(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	writeFile(t, filepath.Join(prjDir, "products/garbage.yaml"), "hand-written\n")
	mustCreateProduct(t, prj, "garbage", "products/garbage.yaml")
	out, err := ProductsDescribe(context.Background(), ProductsDescribeInput{Prj: prj, Name: "garbage"})
	require.NoError(t, err)
	assert.True(t, out.Exists)
	assert.Nil(t, out.Info)
}

func TestProductsDescribe_CountsBothConnectionForms(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	// Write a product doc manually with one $ref and one inline connection.
	docPath := filepath.Join(prjDir, "products/mix.yaml")
	writeFile(t, docPath, `opendpi: "1.0.0"
info:
  title: mix
  version: "1.0.0"
connections:
  refconn:
    $ref: ../connections/other.yaml
  inline:
    type: postgresql
    host: localhost:5432
ports: {}
`)
	mustCreateProduct(t, prj, "mix", "products/mix.yaml")
	out, err := ProductsDescribe(context.Background(), ProductsDescribeInput{Prj: prj, Name: "mix"})
	require.NoError(t, err)
	assert.Equal(t, 2, out.ConnectionsCount)
}

func TestProductsDelete_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	out, err := ProductsDelete(context.Background(), ProductsDeleteInput{Prj: prj, Name: "orders"})
	require.NoError(t, err)
	assert.Equal(t, "orders", out.Name)
	// File is left on disk.
	assert.FileExists(t, filepath.Join(prjDir, "products/orders.yaml"))
	// Registry no longer contains the entry.
	assert.NotContains(t, prj.Products, "orders")
}

func TestProductsDelete_Unknown(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := ProductsDelete(context.Background(), ProductsDeleteInput{Prj: prj, Name: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestProductsLink_Happy(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	out, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "db",
	})
	require.NoError(t, err)
	assert.Equal(t, "../connections/db.yaml", out.Ref)
	assert.Equal(t, "postgresql", out.ConnectionType)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	require.Contains(t, doc.Connections, "db")
	conn := doc.Connections["db"]
	assert.Equal(t, "../connections/db.yaml", conn.Ref)
	assert.Empty(t, conn.Type)
	assert.Empty(t, conn.Host)
}

func TestProductsLink_RejectsInlineCollision(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	// Hand-write a product doc that already has an inline connection named "db".
	docPath := filepath.Join(prjDir, "products/orders.yaml")
	writeFile(t, docPath, `opendpi: "1.0.0"
info:
  title: orders
  version: "1.0.0"
connections:
  db:
    type: mysql
    host: hardcoded:3306
ports: {}
`)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "db",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already linked")
}

func TestProductsLink_MissingProduct(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "nope", ConnectionName: "db",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product \"nope\" not found")
}

func TestProductsLink_MissingConnection(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection \"nope\" not found")
}

func TestProductsUnlink_HappyPath(t *testing.T) {
	_, prj, prjDir := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	mustCreateConnection(t, prj, "db", "connections/db.yaml", "postgresql", "localhost:5432")
	_, err := ProductsLink(context.Background(), ProductsLinkInput{Prj: prj, ProductName: "orders", ConnectionName: "db"})
	require.NoError(t, err)

	out, err := ProductsUnlink(context.Background(), ProductsUnlinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "db",
	})
	require.NoError(t, err)
	assert.Equal(t, "db", out.ConnectionName)

	doc, err := opendpi.Load(filepath.Join(prjDir, "products/orders.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, doc.Connections, "db")
}

func TestProductsUnlink_NotLinked(t *testing.T) {
	_, prj, _ := newTestProject(t)
	mustCreateProduct(t, prj, "orders", "products/orders.yaml")
	_, err := ProductsUnlink(context.Background(), ProductsUnlinkInput{
		Prj: prj, ProductName: "orders", ConnectionName: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not linked to connection")
}

func TestProductsUnlink_MissingProduct(t *testing.T) {
	_, prj, _ := newTestProject(t)
	_, err := ProductsUnlink(context.Background(), ProductsUnlinkInput{
		Prj: prj, ProductName: "nope", ConnectionName: "db",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `product "nope" not found`)
}

func TestProductsUnlink_Validation(t *testing.T) {
	_, err := ProductsUnlink(context.Background(), ProductsUnlinkInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t, Errors{"connection: required", "prj: project not loaded", "product: required"}, errs)
}

func TestProductsLink_Validation(t *testing.T) {
	_, err := ProductsLink(context.Background(), ProductsLinkInput{})
	var errs Errors
	require.True(t, errors.As(err, &errs))
	assert.ElementsMatch(t,
		Errors{"connection: required", "prj: project not loaded", "product: required"},
		errs,
	)
}

func readFileString(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}