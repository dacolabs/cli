// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dacolabs/daco/internal/cli/settings"
	"github.com/dacolabs/daco/internal/opendpi"
)

type ProductEntry struct {
	Name string
	Path string
}

type ProductsCreateInput struct {
	Prj  *settings.Project
	Name string
	Path string
}

type ProductsCreateOutput struct {
	Name       string
	Path       string
	Scaffolded bool
}

func (in ProductsCreateInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	if in.Path == "" {
		problems["path"] = "required"
	}
	return problems
}

func ProductsCreate(ctx context.Context, in ProductsCreateInput) (*ProductsCreateOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	if _, exists := in.Prj.Products[in.Name]; exists {
		return nil, fmt.Errorf("product %q already exists", in.Name)
	}

	abs := resolveProjectPath(in.Prj, in.Path)
	scaffolded := false
	if _, err := os.Stat(abs); errors.Is(err, os.ErrNotExist) {
		if err := opendpi.Scaffold(abs, in.Name); err != nil {
			return nil, fmt.Errorf("scaffold %s: %w", abs, err)
		}
		scaffolded = true
	} else if err != nil {
		return nil, err
	}

	in.Prj.Products[in.Name] = in.Path
	if err := in.Prj.Save(); err != nil {
		return nil, err
	}
	return &ProductsCreateOutput{Name: in.Name, Path: in.Path, Scaffolded: scaffolded}, nil
}

type ProductsListInput struct {
	Prj *settings.Project
}

type ProductsListOutput struct {
	Products []ProductEntry
}

func (in ProductsListInput) Valid(ctx context.Context) map[string]string {
	if in.Prj == nil {
		return map[string]string{"prj": "project not loaded"}
	}
	return nil
}

func ProductsList(ctx context.Context, in ProductsListInput) (*ProductsListOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(in.Prj.Products))
	for name := range in.Prj.Products {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]ProductEntry, 0, len(names))
	for _, name := range names {
		out = append(out, ProductEntry{Name: name, Path: in.Prj.Products[name]})
	}
	return &ProductsListOutput{Products: out}, nil
}

type ProductsDescribeInput struct {
	Prj  *settings.Project
	Name string
}

type ProductsDescribeOutput struct {
	Name             string
	Path             string
	Exists           bool
	Info             *opendpi.Info
	PortsCount       int
	ConnectionsCount int
}

func (in ProductsDescribeInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func ProductsDescribe(ctx context.Context, in ProductsDescribeInput) (*ProductsDescribeOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	path, ok := in.Prj.Products[in.Name]
	if !ok {
		return nil, fmt.Errorf("product %q not found", in.Name)
	}

	abs := resolveProjectPath(in.Prj, path)
	out := &ProductsDescribeOutput{Name: in.Name, Path: path}
	if _, err := os.Stat(abs); err == nil {
		out.Exists = true
		if doc, err := opendpi.Load(abs); err == nil {
			info := doc.Info
			out.Info = &info
			out.PortsCount = len(doc.Ports)
			out.ConnectionsCount = len(doc.Connections)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return out, nil
}

type ProductsDeleteInput struct {
	Prj  *settings.Project
	Name string
}

type ProductsDeleteOutput struct {
	Name string
}

func (in ProductsDeleteInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func ProductsDelete(ctx context.Context, in ProductsDeleteInput) (*ProductsDeleteOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	if _, ok := in.Prj.Products[in.Name]; !ok {
		return nil, fmt.Errorf("product %q not found", in.Name)
	}
	delete(in.Prj.Products, in.Name)
	if err := in.Prj.Save(); err != nil {
		return nil, err
	}
	return &ProductsDeleteOutput{Name: in.Name}, nil
}

type ProductsLinkInput struct {
	Prj            *settings.Project
	ProductName    string
	ConnectionName string
}

type ProductsLinkOutput struct {
	ProductName    string
	ConnectionName string
	Ref            string
	ConnectionType string
	ConnectionHost string
}

func (in ProductsLinkInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.ProductName == "" {
		problems["product"] = "required"
	}
	if in.ConnectionName == "" {
		problems["connection"] = "required"
	}
	return problems
}

func ProductsLink(ctx context.Context, in ProductsLinkInput) (*ProductsLinkOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}

	productPath, ok := in.Prj.Products[in.ProductName]
	if !ok {
		return nil, fmt.Errorf("product %q not found", in.ProductName)
	}
	connectionPath, ok := in.Prj.Connections[in.ConnectionName]
	if !ok {
		return nil, fmt.Errorf("connection %q not found", in.ConnectionName)
	}

	absProduct := resolveProjectPath(in.Prj, productPath)
	absConnection := resolveProjectPath(in.Prj, connectionPath)

	conn, err := opendpi.LoadConnection(absConnection)
	if err != nil {
		return nil, fmt.Errorf("load connection %s: %w", absConnection, err)
	}
	doc, err := opendpi.Load(absProduct)
	if err != nil {
		return nil, fmt.Errorf("load product %s: %w", absProduct, err)
	}
	if doc.Connections == nil {
		doc.Connections = map[string]opendpi.Connection{}
	}
	if _, exists := doc.Connections[in.ConnectionName]; exists {
		return nil, fmt.Errorf("product %q already linked to connection %q", in.ProductName, in.ConnectionName)
	}

	rel, err := filepath.Rel(filepath.Dir(absProduct), absConnection)
	if err != nil {
		return nil, fmt.Errorf("compute relative ref: %w", err)
	}
	doc.Connections[in.ConnectionName] = opendpi.Connection{Ref: rel}

	if err := opendpi.Save(absProduct, doc); err != nil {
		return nil, fmt.Errorf("save product %s: %w", absProduct, err)
	}
	return &ProductsLinkOutput{
		ProductName:    in.ProductName,
		ConnectionName: in.ConnectionName,
		Ref:            rel,
		ConnectionType: conn.Type,
		ConnectionHost: conn.Host,
	}, nil
}

type ProductsUnlinkInput struct {
	Prj            *settings.Project
	ProductName    string
	ConnectionName string
}

type ProductsUnlinkOutput struct {
	ProductName    string
	ConnectionName string
}

func (in ProductsUnlinkInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.ProductName == "" {
		problems["product"] = "required"
	}
	if in.ConnectionName == "" {
		problems["connection"] = "required"
	}
	return problems
}

func ProductsUnlink(ctx context.Context, in ProductsUnlinkInput) (*ProductsUnlinkOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	productPath, ok := in.Prj.Products[in.ProductName]
	if !ok {
		return nil, fmt.Errorf("product %q not found", in.ProductName)
	}
	abs := resolveProjectPath(in.Prj, productPath)
	doc, err := opendpi.Load(abs)
	if err != nil {
		return nil, fmt.Errorf("load product %s: %w", abs, err)
	}
	if _, exists := doc.Connections[in.ConnectionName]; !exists {
		return nil, fmt.Errorf("product %q is not linked to connection %q", in.ProductName, in.ConnectionName)
	}
	delete(doc.Connections, in.ConnectionName)
	if err := opendpi.Save(abs, doc); err != nil {
		return nil, fmt.Errorf("save product %s: %w", abs, err)
	}
	return &ProductsUnlinkOutput{ProductName: in.ProductName, ConnectionName: in.ConnectionName}, nil
}

func resolveProjectPath(prj *settings.Project, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(filepath.Dir(prj.Path), p)
}