// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/dacolabs/daco/internal/cli/settings"
	"github.com/dacolabs/daco/internal/opendpi"
)

type PortEntry struct {
	Name        string
	Description string
}

type PortsCreateInput struct {
	Prj         *settings.Project
	Product     string
	Name        string
	Description string
}

type PortsCreateOutput struct {
	Product string
	Name    string
}

func (in PortsCreateInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func PortsCreate(ctx context.Context, in PortsCreateInput) (*PortsCreateOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	if doc.Ports == nil {
		doc.Ports = map[string]opendpi.Port{}
	}
	if _, exists := doc.Ports[in.Name]; exists {
		return nil, fmt.Errorf("port %q already exists in product %q", in.Name, in.Product)
	}
	doc.Ports[in.Name] = opendpi.Port{Description: in.Description}
	if err := opendpi.Save(productPath, doc); err != nil {
		return nil, err
	}
	return &PortsCreateOutput{Product: in.Product, Name: in.Name}, nil
}

type PortsListInput struct {
	Prj     *settings.Project
	Product string
}

type PortsListOutput struct {
	Ports []PortEntry
}

func (in PortsListInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	return problems
}

func PortsList(ctx context.Context, in PortsListInput) (*PortsListOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, _, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(doc.Ports))
	for name := range doc.Ports {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]PortEntry, 0, len(names))
	for _, name := range names {
		out = append(out, PortEntry{Name: name, Description: doc.Ports[name].Description})
	}
	return &PortsListOutput{Ports: out}, nil
}

type PortsDescribeInput struct {
	Prj     *settings.Project
	Product string
	Name    string
}

type PortConnectionBinding struct {
	Connection string
	Location   string
}

type PortsDescribeOutput struct {
	Product          string
	Name             string
	Description      string
	ConnectionsCount int
	Connections      []PortConnectionBinding
	HasSchema        bool
	SchemaRef        string
	SchemaType       string
	SchemaTitle      string
	SchemaProperties int
	SchemaRequired   int
}

func (in PortsDescribeInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func PortsDescribe(ctx context.Context, in PortsDescribeInput) (*PortsDescribeOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	port, ok := doc.Ports[in.Name]
	if !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Name, in.Product)
	}
	out := &PortsDescribeOutput{
		Product:          in.Product,
		Name:             in.Name,
		Description:      port.Description,
		ConnectionsCount: len(port.Connections),
	}
	for _, pc := range port.Connections {
		out.Connections = append(out.Connections, PortConnectionBinding{
			Connection: pc.Connection,
			Location:   pc.Location,
		})
	}
	if port.Schema != nil {
		out.HasSchema = true
		if port.Schema.Ref != "" {
			out.SchemaRef = port.Schema.Ref
			// Try to resolve and surface the referenced schema's type/title.
			schemaPath := port.Schema.Ref
			if !filepath.IsAbs(schemaPath) {
				schemaPath = filepath.Join(filepath.Dir(productPath), schemaPath)
			}
			if resolved, err := opendpi.LoadSchema(schemaPath); err == nil {
				out.SchemaType = opendpi.SchemaTypeString(resolved)
				out.SchemaTitle = resolved.Title
				out.SchemaProperties = len(resolved.Properties)
				out.SchemaRequired = len(resolved.Required)
			}
		} else {
			out.SchemaType = opendpi.SchemaTypeString(port.Schema)
			out.SchemaTitle = port.Schema.Title
			out.SchemaProperties = len(port.Schema.Properties)
			out.SchemaRequired = len(port.Schema.Required)
		}
	}
	return out, nil
}

type PortsDeleteInput struct {
	Prj     *settings.Project
	Product string
	Name    string
}

type PortsDeleteOutput struct {
	Product string
	Name    string
}

func (in PortsDeleteInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func PortsDelete(ctx context.Context, in PortsDeleteInput) (*PortsDeleteOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	if _, ok := doc.Ports[in.Name]; !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Name, in.Product)
	}
	delete(doc.Ports, in.Name)
	if err := opendpi.Save(productPath, doc); err != nil {
		return nil, err
	}
	return &PortsDeleteOutput{Product: in.Product, Name: in.Name}, nil
}

type PortsLinkSchemaInput struct {
	Prj     *settings.Project
	Product string
	Port    string
	Schema  string
}

type PortsLinkSchemaOutput struct {
	Product    string
	Port       string
	Schema     string
	Ref        string
	SchemaType string
}

func (in PortsLinkSchemaInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Port == "" {
		problems["port"] = "required"
	}
	if in.Schema == "" {
		problems["schema"] = "required"
	}
	return problems
}

func PortsLinkSchema(ctx context.Context, in PortsLinkSchemaInput) (*PortsLinkSchemaOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}

	schemaPath, ok := in.Prj.Schemas[in.Schema]
	if !ok {
		return nil, fmt.Errorf("schema %q not found", in.Schema)
	}
	absSchema := resolveProjectPath(in.Prj, schemaPath)
	resolvedSchema, err := opendpi.LoadSchema(absSchema)
	if err != nil {
		return nil, fmt.Errorf("load schema %s: %w", absSchema, err)
	}

	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	port, ok := doc.Ports[in.Port]
	if !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Port, in.Product)
	}
	if port.Schema != nil && (port.Schema.Ref != "" || port.Schema.Type != "" || len(port.Schema.Types) > 0) {
		return nil, fmt.Errorf("port %q already has a schema", in.Port)
	}

	rel, err := filepath.Rel(filepath.Dir(productPath), absSchema)
	if err != nil {
		return nil, fmt.Errorf("compute relative ref: %w", err)
	}
	schemaRef := &opendpi.Schema{}
	schemaRef.Ref = rel
	port.Schema = schemaRef
	doc.Ports[in.Port] = port

	if err := opendpi.Save(productPath, doc); err != nil {
		return nil, err
	}

	return &PortsLinkSchemaOutput{
		Product:    in.Product,
		Port:       in.Port,
		Schema:     in.Schema,
		Ref:        rel,
		SchemaType: opendpi.SchemaTypeString(resolvedSchema),
	}, nil
}

type PortsUnlinkSchemaInput struct {
	Prj     *settings.Project
	Product string
	Port    string
}

type PortsUnlinkSchemaOutput struct {
	Product string
	Port    string
}

func (in PortsUnlinkSchemaInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Port == "" {
		problems["port"] = "required"
	}
	return problems
}

func PortsUnlinkSchema(ctx context.Context, in PortsUnlinkSchemaInput) (*PortsUnlinkSchemaOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	port, ok := doc.Ports[in.Port]
	if !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Port, in.Product)
	}
	if port.Schema == nil {
		return nil, fmt.Errorf("port %q has no schema to unlink", in.Port)
	}
	port.Schema = nil
	doc.Ports[in.Port] = port
	if err := opendpi.Save(productPath, doc); err != nil {
		return nil, err
	}
	return &PortsUnlinkSchemaOutput{Product: in.Product, Port: in.Port}, nil
}

type PortsBindConnectionInput struct {
	Prj        *settings.Project
	Product    string
	Port       string
	Connection string
	Location   string
}

type PortsBindConnectionOutput struct {
	Product    string
	Port       string
	Connection string
	Location   string
}

func (in PortsBindConnectionInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Port == "" {
		problems["port"] = "required"
	}
	if in.Connection == "" {
		problems["connection"] = "required"
	}
	if in.Location == "" {
		problems["location"] = "required"
	}
	return problems
}

func PortsBindConnection(ctx context.Context, in PortsBindConnectionInput) (*PortsBindConnectionOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	if _, ok := doc.Connections[in.Connection]; !ok {
		return nil, fmt.Errorf("connection %q is not linked to product %q (use products link first)", in.Connection, in.Product)
	}
	port, ok := doc.Ports[in.Port]
	if !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Port, in.Product)
	}
	for _, pc := range port.Connections {
		if pc.Connection == in.Connection && pc.Location == in.Location {
			return nil, fmt.Errorf("port %q already bound to connection %q at location %q", in.Port, in.Connection, in.Location)
		}
	}
	port.Connections = append(port.Connections, opendpi.PortConnection{
		Connection: in.Connection,
		Location:   in.Location,
	})
	doc.Ports[in.Port] = port
	if err := opendpi.Save(productPath, doc); err != nil {
		return nil, err
	}
	return &PortsBindConnectionOutput{
		Product:    in.Product,
		Port:       in.Port,
		Connection: in.Connection,
		Location:   in.Location,
	}, nil
}

type PortsUnbindConnectionInput struct {
	Prj        *settings.Project
	Product    string
	Port       string
	Connection string
	Location   string // optional; empty matches every binding for this connection
}

type PortsUnbindConnectionOutput struct {
	Product    string
	Port       string
	Connection string
	Removed    int
}

func (in PortsUnbindConnectionInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Product == "" {
		problems["product"] = "required"
	}
	if in.Port == "" {
		problems["port"] = "required"
	}
	if in.Connection == "" {
		problems["connection"] = "required"
	}
	return problems
}

func PortsUnbindConnection(ctx context.Context, in PortsUnbindConnectionInput) (*PortsUnbindConnectionOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	doc, productPath, err := loadProductDoc(in.Prj, in.Product)
	if err != nil {
		return nil, err
	}
	port, ok := doc.Ports[in.Port]
	if !ok {
		return nil, fmt.Errorf("port %q not found in product %q", in.Port, in.Product)
	}
	kept := port.Connections[:0]
	removed := 0
	for _, pc := range port.Connections {
		if pc.Connection == in.Connection && (in.Location == "" || pc.Location == in.Location) {
			removed++
			continue
		}
		kept = append(kept, pc)
	}
	if removed == 0 {
		return nil, fmt.Errorf("no binding to connection %q on port %q", in.Connection, in.Port)
	}
	port.Connections = kept
	doc.Ports[in.Port] = port
	if err := opendpi.Save(productPath, doc); err != nil {
		return nil, err
	}
	return &PortsUnbindConnectionOutput{
		Product:    in.Product,
		Port:       in.Port,
		Connection: in.Connection,
		Removed:    removed,
	}, nil
}

func loadProductDoc(prj *settings.Project, productName string) (*opendpi.Document, string, error) {
	path, ok := prj.Products[productName]
	if !ok {
		return nil, "", fmt.Errorf("product %q not found", productName)
	}
	abs := resolveProjectPath(prj, path)
	doc, err := opendpi.Load(abs)
	if err != nil {
		return nil, "", fmt.Errorf("load product %s: %w", abs, err)
	}
	return doc, abs, nil
}