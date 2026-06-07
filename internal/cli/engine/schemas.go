// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Daco Labs

package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/dacolabs/daco/internal/cli/settings"
	"github.com/dacolabs/daco/internal/opendpi"
)

type SchemaEntry struct {
	Name string
	Path string
}

type SchemasCreateInput struct {
	Prj         *settings.Project
	Name        string
	Path        string
	Type        string
	Title       string
	Description string
}

type SchemasCreateOutput struct {
	Name       string
	Path       string
	Scaffolded bool
}

func (in SchemasCreateInput) Valid(ctx context.Context) map[string]string {
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

func SchemasCreate(ctx context.Context, in SchemasCreateInput) (*SchemasCreateOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	if _, exists := in.Prj.Schemas[in.Name]; exists {
		return nil, fmt.Errorf("schema %q already exists", in.Name)
	}

	abs := resolveProjectPath(in.Prj, in.Path)
	scaffolded := false
	if _, err := os.Stat(abs); errors.Is(err, os.ErrNotExist) {
		t := in.Type
		if t == "" {
			t = "object"
		}
		schema := &opendpi.Schema{}
		schema.Type = t
		schema.Title = in.Title
		schema.Description = in.Description
		if err := opendpi.SaveSchema(abs, schema); err != nil {
			return nil, fmt.Errorf("scaffold %s: %w", abs, err)
		}
		scaffolded = true
	} else if err != nil {
		return nil, err
	}

	in.Prj.Schemas[in.Name] = in.Path
	if err := in.Prj.Save(); err != nil {
		return nil, err
	}
	return &SchemasCreateOutput{Name: in.Name, Path: in.Path, Scaffolded: scaffolded}, nil
}

type SchemasListInput struct {
	Prj *settings.Project
}

type SchemasListOutput struct {
	Schemas []SchemaEntry
}

func (in SchemasListInput) Valid(ctx context.Context) map[string]string {
	if in.Prj == nil {
		return map[string]string{"prj": "project not loaded"}
	}
	return nil
}

func SchemasList(ctx context.Context, in SchemasListInput) (*SchemasListOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(in.Prj.Schemas))
	for name := range in.Prj.Schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]SchemaEntry, 0, len(names))
	for _, name := range names {
		out = append(out, SchemaEntry{Name: name, Path: in.Prj.Schemas[name]})
	}
	return &SchemasListOutput{Schemas: out}, nil
}

type SchemasDescribeInput struct {
	Prj  *settings.Project
	Name string
}

type SchemasDescribeOutput struct {
	Name   string
	Path   string
	Exists bool
	Schema *opendpi.Schema
}

func (in SchemasDescribeInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func SchemasDescribe(ctx context.Context, in SchemasDescribeInput) (*SchemasDescribeOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	path, ok := in.Prj.Schemas[in.Name]
	if !ok {
		return nil, fmt.Errorf("schema %q not found", in.Name)
	}

	abs := resolveProjectPath(in.Prj, path)
	out := &SchemasDescribeOutput{Name: in.Name, Path: path}
	if _, err := os.Stat(abs); err == nil {
		out.Exists = true
		if schema, err := opendpi.LoadSchema(abs); err == nil {
			out.Schema = schema
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return out, nil
}

type SchemasDeleteInput struct {
	Prj  *settings.Project
	Name string
}

type SchemasDeleteOutput struct {
	Name string
}

func (in SchemasDeleteInput) Valid(ctx context.Context) map[string]string {
	problems := map[string]string{}
	if in.Prj == nil {
		problems["prj"] = "project not loaded"
	}
	if in.Name == "" {
		problems["name"] = "required"
	}
	return problems
}

func SchemasDelete(ctx context.Context, in SchemasDeleteInput) (*SchemasDeleteOutput, error) {
	if err := validate(ctx, in); err != nil {
		return nil, err
	}
	if _, ok := in.Prj.Schemas[in.Name]; !ok {
		return nil, fmt.Errorf("schema %q not found", in.Name)
	}
	delete(in.Prj.Schemas, in.Name)
	if err := in.Prj.Save(); err != nil {
		return nil, err
	}
	return &SchemasDeleteOutput{Name: in.Name}, nil
}