// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dacolabs/daco/internal/cli/settings"
)

// fakePrj is a non-nil placeholder so each Input only flags the field we're
// probing in isolation, not the prj sentinel.
var fakePrj = &settings.Project{Products: map[string]string{}, Connections: map[string]string{}, Schemas: map[string]string{}}

func TestValid_Init_AllVariants(t *testing.T) {
	cases := []struct {
		name     string
		in       InitInput
		expected []string
	}{
		{"all empty", InitInput{}, []string{"cwd", "usr"}},
		{"only usr", InitInput{Usr: &settings.User{}}, []string{"cwd"}},
		{"only cwd", InitInput{Cwd: "/x"}, []string{"usr"}},
		{"all set", InitInput{Usr: &settings.User{}, Cwd: "/x"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problems := tc.in.Valid(context.Background())
			assertKeys(t, problems, tc.expected)
		})
	}
}

func TestValid_Products_AllVariants(t *testing.T) {
	cases := []struct {
		name     string
		in       Validator
		expected []string
	}{
		{"create empty", ProductsCreateInput{}, []string{"name", "path", "prj"}},
		{"create only prj", ProductsCreateInput{Prj: fakePrj}, []string{"name", "path"}},
		{"create only name", ProductsCreateInput{Prj: fakePrj, Name: "x"}, []string{"path"}},
		{"create only path", ProductsCreateInput{Prj: fakePrj, Path: "x.yaml"}, []string{"name"}},
		{"create all", ProductsCreateInput{Prj: fakePrj, Name: "x", Path: "x.yaml"}, nil},
		{"list empty", ProductsListInput{}, []string{"prj"}},
		{"list ok", ProductsListInput{Prj: fakePrj}, nil},
		{"describe empty", ProductsDescribeInput{}, []string{"name", "prj"}},
		{"describe only prj", ProductsDescribeInput{Prj: fakePrj}, []string{"name"}},
		{"describe ok", ProductsDescribeInput{Prj: fakePrj, Name: "x"}, nil},
		{"delete empty", ProductsDeleteInput{}, []string{"name", "prj"}},
		{"delete only prj", ProductsDeleteInput{Prj: fakePrj}, []string{"name"}},
		{"delete ok", ProductsDeleteInput{Prj: fakePrj, Name: "x"}, nil},
		{"link empty", ProductsLinkInput{}, []string{"connection", "prj", "product"}},
		{"link only prj", ProductsLinkInput{Prj: fakePrj}, []string{"connection", "product"}},
		{"link only product", ProductsLinkInput{Prj: fakePrj, ProductName: "x"}, []string{"connection"}},
		{"link all", ProductsLinkInput{Prj: fakePrj, ProductName: "x", ConnectionName: "y"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertKeys(t, tc.in.Valid(context.Background()), tc.expected)
		})
	}
}

func TestValid_Connections_AllVariants(t *testing.T) {
	cases := []struct {
		name     string
		in       Validator
		expected []string
	}{
		{"create empty", ConnectionsCreateInput{}, []string{"host", "name", "path", "prj", "type"}},
		{"create all", ConnectionsCreateInput{Prj: fakePrj, Name: "x", Path: "x.yaml", Type: "t", Host: "h"}, nil},
		{"list", ConnectionsListInput{}, []string{"prj"}},
		{"describe empty", ConnectionsDescribeInput{}, []string{"name", "prj"}},
		{"describe ok", ConnectionsDescribeInput{Prj: fakePrj, Name: "x"}, nil},
		{"delete empty", ConnectionsDeleteInput{}, []string{"name", "prj"}},
		{"delete ok", ConnectionsDeleteInput{Prj: fakePrj, Name: "x"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertKeys(t, tc.in.Valid(context.Background()), tc.expected)
		})
	}
}

func TestValid_Schemas_AllVariants(t *testing.T) {
	cases := []struct {
		name     string
		in       Validator
		expected []string
	}{
		{"create empty", SchemasCreateInput{}, []string{"name", "path", "prj"}},
		{"create ok", SchemasCreateInput{Prj: fakePrj, Name: "x", Path: "x.yaml"}, nil},
		{"list", SchemasListInput{}, []string{"prj"}},
		{"describe empty", SchemasDescribeInput{}, []string{"name", "prj"}},
		{"describe ok", SchemasDescribeInput{Prj: fakePrj, Name: "x"}, nil},
		{"delete empty", SchemasDeleteInput{}, []string{"name", "prj"}},
		{"delete ok", SchemasDeleteInput{Prj: fakePrj, Name: "x"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertKeys(t, tc.in.Valid(context.Background()), tc.expected)
		})
	}
}

func TestValid_Ports_AllVariants(t *testing.T) {
	cases := []struct {
		name     string
		in       Validator
		expected []string
	}{
		{"create empty", PortsCreateInput{}, []string{"name", "prj", "product"}},
		{"create ok", PortsCreateInput{Prj: fakePrj, Product: "x", Name: "y"}, nil},
		{"list empty", PortsListInput{}, []string{"prj", "product"}},
		{"list ok", PortsListInput{Prj: fakePrj, Product: "x"}, nil},
		{"describe empty", PortsDescribeInput{}, []string{"name", "prj", "product"}},
		{"describe ok", PortsDescribeInput{Prj: fakePrj, Product: "x", Name: "y"}, nil},
		{"delete empty", PortsDeleteInput{}, []string{"name", "prj", "product"}},
		{"delete ok", PortsDeleteInput{Prj: fakePrj, Product: "x", Name: "y"}, nil},
		{"link empty", PortsLinkSchemaInput{}, []string{"port", "prj", "product", "schema"}},
		{"link ok", PortsLinkSchemaInput{Prj: fakePrj, Product: "x", Port: "y", Schema: "z"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertKeys(t, tc.in.Valid(context.Background()), tc.expected)
		})
	}
}

func TestValid_Translate_AllVariants(t *testing.T) {
	cases := []struct {
		name     string
		in       Validator
		expected []string
	}{
		{"schemas empty", SchemasTranslateInput{}, []string{"format", "output-dir", "prj", "schema"}},
		{"schemas ok", SchemasTranslateInput{Prj: fakePrj, Schema: "x", Format: "p", OutputDir: "o"}, nil},
		{"ports empty", PortsTranslateInput{}, []string{"format", "output-dir", "port", "prj", "product"}},
		{"ports ok", PortsTranslateInput{Prj: fakePrj, Product: "x", Port: "y", Format: "p", OutputDir: "o"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertKeys(t, tc.in.Valid(context.Background()), tc.expected)
		})
	}
}

func assertKeys(t *testing.T, problems map[string]string, expected []string) {
	t.Helper()
	gotKeys := make([]string, 0, len(problems))
	for k := range problems {
		gotKeys = append(gotKeys, k)
	}
	assert.ElementsMatch(t, expected, gotKeys, "problems=%v", problems)
}